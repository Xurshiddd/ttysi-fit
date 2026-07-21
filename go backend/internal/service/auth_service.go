package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/middleware"
	"github.com/ttysi-fit/backend/pkg/hemis"
	"github.com/ttysi-fit/backend/pkg/security"
)

// AuthService — autentifikatsiya use-case qatlami.
type AuthService struct {
	users    domain.UserRepository
	jwt      *security.JWTManager
	redis    *redis.Client
	oauth    *hemis.OAuthClient // HEMIS OAuth (nil bo'lsa — o'chiq)
	stateTTL time.Duration
	codeTTL  time.Duration // bir martalik exchange code muddati
	// sessions — qurilma sessiyalari. nil bo'lsa qurilma cheklovi
	// ishlamaydi va login avvalgidek o'tadi.
	sessions domain.SessionRepository
}

// SetSessions — qurilma sessiyalari repozitoriysini ulaydi.
//
// Alohida setter: NewAuthService imzosi allaqachon uzun va u bir necha
// joydan chaqiriladi; ixtiyoriy bog'liqlikni shu tarzda qo'shish
// chaqiruvchilarni buzmaydi.
func (s *AuthService) SetSessions(r domain.SessionRepository) { s.sessions = r }

func NewAuthService(
	users domain.UserRepository,
	jwt *security.JWTManager,
	rdb *redis.Client,
	oauth *hemis.OAuthClient,
	stateTTL time.Duration,
	codeTTL time.Duration,
) *AuthService {
	if stateTTL <= 0 {
		stateTTL = 10 * time.Minute
	}
	if codeTTL <= 0 {
		codeTTL = 60 * time.Second
	}
	return &AuthService{users: users, jwt: jwt, redis: rdb, oauth: oauth, stateTTL: stateTTL, codeTTL: codeTTL}
}

func refreshKey(userID uuid.UUID) string {
	return fmt.Sprintf("refresh:%s", userID)
}

func hemisStateKey(state string) string {
	return fmt.Sprintf("hemis_oauth_state:%s", state)
}

func hemisCodeKey(code string) string {
	return fmt.Sprintf("hemis_oauth_code:%s", code)
}

// Register — yangi foydalanuvchi yaratadi va token jufti qaytaradi.
// requestLocale — so'rovdan aniqlangan til (req.Language bo'sh bo'lsa ishlatiladi).
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, requestLocale string) (*dto.TokenResponse, error) {
	// Email band emasligini tekshirish
	if _, err := s.users.GetByEmail(ctx, req.Email); err == nil {
		return nil, domain.ErrAlreadyExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("AuthService.Register: %w", err)
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("AuthService.Register: hash: %w", err)
	}

	role := domain.RoleStudent
	if req.Role == string(domain.RoleEmployee) {
		role = domain.RoleEmployee
	}

	// Til: aniq ko'rsatilgan bo'lsa o'sha, aks holda so'rov tili.
	lang := req.Language
	if lang == "" {
		lang = requestLocale
	}

	user := &domain.User{
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hash,
		Role:     role,
		IsActive: true,
		Language: lang,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("AuthService.Register: create: %w", err)
	}

	return s.issueTokens(ctx, user)
}

// Login — login/parolni tekshirib token jufti qaytaradi.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	user, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("AuthService.Login: %w", err)
	}

	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}
	if !security.CheckPassword(user.Password, req.Password) {
		return nil, domain.ErrInvalidCredentials
	}

	if err := s.applyDevicePolicy(ctx, user.ID, req.Device, req.ForceDevice); err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user)
}

// applyDevicePolicy — bir qurilma cheklovi.
//
// Oqim:
//  1. Qurilma ma'lumoti berilmagan bo'lsa (eski mijoz, web) — cheklov yo'q.
//  2. Boshqa qurilmada faol sessiya bo'lsa va foydalanuvchi hali rozilik
//     bermagan bo'lsa — ErrDeviceConflict. Handler mijozga qaysi qurilma
//     ekanini aytadi.
//  3. Rozilik berilgan bo'lsa (force) — eski sessiyalar yopiladi va
//     refresh token bekor qilinadi, ya'ni eski qurilma chiqarib yuboriladi.
//
// Nima uchun ATAYLAB to'sqinlik: hisobni bo'lishish reytingni buzadi
// (ikki kishining qadami bitta hisobga tushadi).
func (s *AuthService) applyDevicePolicy(ctx context.Context, userID uuid.UUID, d *dto.DeviceInfo, force bool) error {
	if s.sessions == nil || d == nil || d.DeviceID == "" {
		return nil
	}

	other, err := s.sessions.ActiveOther(ctx, userID, d.DeviceID)
	if err != nil {
		return fmt.Errorf("AuthService.Login: qurilma: %w", err)
	}
	if other != nil && !force {
		// Handler bu xatoni ushlab, qurilma nomini javobga qo'shadi.
		return domain.ErrDeviceConflict
	}

	if other != nil {
		if _, err := s.sessions.RevokeOthers(ctx, userID, d.DeviceID, domain.RevokeNewDevice); err != nil {
			return fmt.Errorf("AuthService.Login: eski sessiya: %w", err)
		}
		// Eski qurilmaning refresh tokeni ham yaroqsiz bo'lsin.
		// (issueTokens uni baribir almashtiradi — bu qo'shimcha kafolat.)
		s.redis.Del(ctx, refreshKey(userID))
	}

	if _, err := s.sessions.Upsert(ctx, userID, domain.DeviceInfo{
		DeviceID:   d.DeviceID,
		DeviceName: d.DeviceName,
		Platform:   d.Platform,
		AppVersion: d.AppVersion,
		IP:         d.IP,
		UserAgent:  d.UserAgent,
	}); err != nil {
		return fmt.Errorf("AuthService.Login: sessiya: %w", err)
	}

	// Joriy qurilmani belgilaymiz — DeviceSession middleware shu kalitni
	// tekshirib eski qurilmani DARROV chiqaradi (access token muddatini
	// kutmasdan).
	s.redis.Set(ctx, middleware.SessionDeviceKey(userID.String()),
		d.DeviceID, s.jwt.RefreshTTL())
	return nil
}

// ConflictingDevice — login rad etilganda qaysi qurilma bandligini aytadi.
func (s *AuthService) ConflictingDevice(ctx context.Context, email, deviceID string) *domain.UserSession {
	if s.sessions == nil {
		return nil
	}
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}
	return s.ConflictingDeviceFor(ctx, user.ID, deviceID)
}

// ConflictingDeviceFor — o'sha narsa, lekin foydalanuvchi ID si bo'yicha
// (HEMIS oqimida email emas, token ma'lum).
func (s *AuthService) ConflictingDeviceFor(ctx context.Context, userID uuid.UUID, deviceID string) *domain.UserSession {
	if s.sessions == nil {
		return nil
	}
	other, err := s.sessions.ActiveOther(ctx, userID, deviceID)
	if err != nil {
		return nil
	}
	return other
}

// PendingUserID — hali almashtirilmagan HEMIS kodi kimga tegishli.
//
// Konflikt javobida qaysi qurilma bandligini aytish uchun kerak: bu paytda
// foydalanuvchi hali kirmagan, shuning uchun uni faqat kod ichidagi
// tokendan aniqlash mumkin.
func (s *AuthService) PendingUserID(ctx context.Context, code string) (uuid.UUID, bool) {
	data, err := s.redis.Get(ctx, hemisCodeKey(code)).Bytes()
	if err != nil {
		return uuid.Nil, false
	}
	var tokens dto.TokenResponse
	if err := json.Unmarshal(data, &tokens); err != nil {
		return uuid.Nil, false
	}
	claims, err := s.jwt.ParseAccess(tokens.AccessToken)
	if err != nil {
		return uuid.Nil, false
	}
	return claims.UserID, true
}

// Sessions — foydalanuvchining faol qurilmalari.
func (s *AuthService) Sessions(ctx context.Context, userID uuid.UUID) ([]domain.UserSession, error) {
	if s.sessions == nil {
		return nil, nil
	}
	return s.sessions.ListActive(ctx, userID)
}

// RevokeSession — foydalanuvchi o'z qurilmasini o'chiradi.
//
// O'chirilgan qurilma JORIY bo'lsa Redis kaliti ham olib tashlanadi:
// aks holda u o'zini o'zi chiqara olmasdi.
func (s *AuthService) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	if s.sessions == nil {
		return domain.ErrNotFound
	}

	// Qaysi qurilma ekanini oldindan bilib olamiz.
	var deviceID string
	if list, err := s.sessions.ListActive(ctx, userID); err == nil {
		for _, x := range list {
			if x.ID == sessionID {
				deviceID = x.DeviceID
				break
			}
		}
	}

	if err := s.sessions.Revoke(ctx, userID, sessionID, domain.RevokeUser); err != nil {
		return err
	}

	if deviceID != "" {
		key := middleware.SessionDeviceKey(userID.String())
		if cur, err := s.redis.Get(ctx, key).Result(); err == nil && cur == deviceID {
			// Kalit O'CHIRILMAYDI, balki hech bir qurilmaga mos
			// kelmaydigan qiymatga qo'yiladi.
			//
			// O'chirilsa middleware uni "qurilma siyosati ishlatilmagan"
			// deb o'qib, so'rovni O'TKAZIB YUBORARDI — ya'ni foydalanuvchi
			// o'zini chiqara olmasdi.
			s.redis.Set(ctx, key, middleware.NoDevice, s.jwt.RefreshTTL())
			s.redis.Del(ctx, refreshKey(userID))
		}
	}
	return nil
}

// Refresh — yaroqli refresh token bilan yangi juft chiqaradi (rotatsiya bilan).
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	claims, err := s.jwt.ParseRefresh(refreshToken)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// Redis dagi joriy JTI bilan solishtirish (revoke/rotatsiya nazorati)
	stored, err := s.redis.Get(ctx, refreshKey(claims.UserID)).Result()
	if err != nil || stored != claims.ID {
		return nil, domain.ErrUnauthorized
	}

	user, err := s.users.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.issueTokens(ctx, user)
}

// HemisAuthURL — HEMIS authorize URL'ini qaytaradi va state'ni Redis'da saqlaydi (CSRF).
func (s *AuthService) HemisAuthURL(ctx context.Context, provider string) (string, error) {
	if s.oauth == nil {
		return "", domain.ErrUnauthorized
	}
	authURL, state, err := s.oauth.AuthorizationURL(provider)
	if err != nil {
		return "", err
	}
	if err := s.redis.Set(ctx, hemisStateKey(state), provider, s.stateTTL).Err(); err != nil {
		return "", fmt.Errorf("AuthService.HemisAuthURL: state saqlash: %w", err)
	}
	return authURL, nil
}

// HemisCallback — state'ni tekshirib, code'dan profil olib, userni topib token beradi.
func (s *AuthService) HemisCallback(ctx context.Context, provider, state, code string) (*dto.TokenResponse, error) {
	if s.oauth == nil {
		return nil, domain.ErrUnauthorized
	}
	if state == "" || code == "" {
		return nil, domain.ErrInvalidCredentials
	}

	// State CSRF tekshiruvi (Redis) — provayder bilan mos kelishi kerak.
	stored, err := s.redis.Get(ctx, hemisStateKey(state)).Result()
	if err != nil || stored != provider {
		return nil, domain.ErrUnauthorized
	}
	s.redis.Del(ctx, hemisStateKey(state))

	profile, err := s.oauth.FetchUser(ctx, provider, code)
	if err != nil {
		return nil, fmt.Errorf("AuthService.HemisCallback: %w", err)
	}

	var hemisID *int64
	if profile.ID != 0 {
		id := profile.ID
		hemisID = &id
	}

	user, err := s.users.GetByHemis(ctx, hemisID, profile.Login)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// HEMIS'da bor, lekin bizda sync qilinmagan / aktiv emas.
			return nil, domain.ErrUnauthorized
		}
		return nil, fmt.Errorf("AuthService.HemisCallback: %w", err)
	}
	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	return s.issueTokens(ctx, user)
}

// StashTokens — token juftini bir martalik qisqa muddatli code ostida Redis'da saqlaydi
// va o'sha code'ni qaytaradi (mobil deep link orqali uzatish uchun — token URL'ga tushmaydi).
func (s *AuthService) StashTokens(ctx context.Context, tokens *dto.TokenResponse) (string, error) {
	code := uuid.NewString()
	data, err := json.Marshal(tokens)
	if err != nil {
		return "", fmt.Errorf("AuthService.StashTokens: marshal: %w", err)
	}
	if err := s.redis.Set(ctx, hemisCodeKey(code), data, s.codeTTL).Err(); err != nil {
		return "", fmt.Errorf("AuthService.StashTokens: redis: %w", err)
	}
	return code, nil
}

// ExchangeTokens — bir martalik code'ni token juftiga almashtiradi (va code'ni o'chiradi).
func (s *AuthService) ExchangeTokens(ctx context.Context, code string, device *dto.DeviceInfo, force bool) (*dto.TokenResponse, error) {
	if code == "" {
		return nil, domain.ErrInvalidCredentials
	}
	data, err := s.redis.Get(ctx, hemisCodeKey(code)).Bytes()
	if err != nil {
		// Yo'q yoki muddati tugagan — bir martalik.
		return nil, domain.ErrUnauthorized
	}

	var tokens dto.TokenResponse
	if err := json.Unmarshal(data, &tokens); err != nil {
		s.redis.Del(ctx, hemisCodeKey(code))
		return nil, fmt.Errorf("AuthService.ExchangeTokens: unmarshal: %w", err)
	}

	// Qurilma cheklovi HEMIS oqimida ham qo'llanadi: haqiqiy foydalanuvchilar
	// aynan shu yo'l bilan kiradi (dev login emas).
	//
	// Kod konflikt holatida O'CHIRILMAYDI: foydalanuvchi rozilik berib
	// o'sha kod bilan qayta yuboradi. Kodning TTL i qisqa (HEMIS_OAUTH_CODE_TTL),
	// shuning uchun oyna uzoq ochiq qolsa qayta kirish kerak bo'ladi.
	claims, cerr := s.jwt.ParseAccess(tokens.AccessToken)
	if cerr == nil {
		if err := s.applyDevicePolicy(ctx, claims.UserID, device, force); err != nil {
			return nil, err
		}
	}

	s.redis.Del(ctx, hemisCodeKey(code))
	return &tokens, nil
}

// Logout — refresh tokenni bekor qiladi (Redis dan o'chiradi).
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.redis.Del(ctx, refreshKey(userID)).Err()
}

// issueTokens — access + refresh yaratadi va refresh JTI ni Redis ga saqlaydi.
func (s *AuthService) issueTokens(ctx context.Context, user *domain.User) (*dto.TokenResponse, error) {
	access, err := s.jwt.GenerateAccess(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("issueTokens: access: %w", err)
	}
	refresh, err := s.jwt.GenerateRefresh(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("issueTokens: refresh: %w", err)
	}

	// Refresh token JTI ni Redis ga TTL bilan saqlash
	claims, err := s.jwt.ParseRefresh(refresh)
	if err != nil {
		return nil, fmt.Errorf("issueTokens: parse: %w", err)
	}
	if err := s.redis.Set(ctx, refreshKey(user.ID), claims.ID, s.jwt.RefreshTTL()).Err(); err != nil {
		return nil, fmt.Errorf("issueTokens: redis: %w", err)
	}

	return &dto.TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		User: dto.UserInfo{
			ID:       user.ID.String(),
			FullName: user.FullName,
			Email:    user.Email,
			Role:     string(user.Role),
			Language: user.Language,
		},
	}, nil
}
