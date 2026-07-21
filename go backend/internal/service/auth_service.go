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
}

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

	return s.issueTokens(ctx, user)
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
func (s *AuthService) ExchangeTokens(ctx context.Context, code string) (*dto.TokenResponse, error) {
	if code == "" {
		return nil, domain.ErrInvalidCredentials
	}
	data, err := s.redis.Get(ctx, hemisCodeKey(code)).Bytes()
	if err != nil {
		// Yo'q yoki muddati tugagan — bir martalik.
		return nil, domain.ErrUnauthorized
	}
	s.redis.Del(ctx, hemisCodeKey(code))

	var tokens dto.TokenResponse
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("AuthService.ExchangeTokens: unmarshal: %w", err)
	}
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
