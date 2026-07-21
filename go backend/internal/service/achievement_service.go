package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/pkg/certificate"
)

// AchievementService — yutuq/sertifikat use-case qatlami.
type AchievementService struct {
	repo  domain.AchievementRepository
	coins domain.FitCoinRepository
	// notify — yangi yutuq haqida xabar. nil bo'lsa xabar yozilmaydi.
	notify Notifier
	// signing — muhr/imzo. Startupda bir marta yuklanadi (diskdan har
	// so'rovda o'qish keraksiz).
	signing certificate.Signing
}

func NewAchievementService(repo domain.AchievementRepository, coins domain.FitCoinRepository, signing certificate.Signing, notify Notifier) *AchievementService {
	return &AchievementService{repo: repo, coins: coins, signing: signing, notify: notify}
}

// Types — admin panel dinamik formasi uchun tur ta'riflari (§16.2).
func (s *AchievementService) Types() []domain.AchievementTypeSpec {
	return domain.AchievementTypeSpecs()
}

func (s *AchievementService) Create(ctx context.Context, a *domain.Achievement) error {
	if err := s.validate(a); err != nil {
		return err
	}
	return s.repo.Create(ctx, a)
}

func (s *AchievementService) Update(ctx context.Context, a *domain.Achievement) error {
	if err := s.validate(a); err != nil {
		return err
	}
	return s.repo.Update(ctx, a)
}

func (s *AchievementService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *AchievementService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Achievement, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AchievementService) List(ctx context.Context, f domain.AchievementFilter) ([]domain.Achievement, int64, error) {
	return s.repo.List(ctx, f)
}

// ListForUser — mobil ro'yxat: yutuq + foydalanuvchi progressi.
func (s *AchievementService) ListForUser(ctx context.Context, userID uuid.UUID, f domain.AchievementFilter) ([]domain.AchievementView, int64, error) {
	return s.repo.ListForUser(ctx, userID, f)
}

// ListEarned — foydalanuvchi qozongan yutuqlar (profil, sertifikatlar).
func (s *AchievementService) ListEarned(ctx context.Context, userID uuid.UUID, f domain.AchievementFilter) ([]domain.AchievementView, int64, error) {
	return s.repo.ListEarned(ctx, userID, f)
}

// AwardManual — admin yutuqni qo'lda beradi (musobaqa g'olibi, tadbir ishtiroki).
//
// Faqat award_mode='manual' turlar uchun: avtomatik yutuqni qo'lda berish
// mezonni chetlab o'tish bo'lardi va reyting adolatini buzardi.
func (s *AchievementService) AwardManual(ctx context.Context, achievementID, userID, adminID uuid.UUID, note string) (*domain.UserAchievement, error) {
	a, err := s.repo.GetByID(ctx, achievementID)
	if err != nil {
		return nil, err
	}
	if a.Status != domain.AchStatusActive {
		return nil, fmt.Errorf("%w: yutuq aktiv emas", domain.ErrValidation)
	}
	if a.AwardMode != domain.AwardModeManual {
		return nil, fmt.Errorf("%w: bu yutuq avtomatik beriladi, qo'lda berib bo'lmaydi",
			domain.ErrValidation)
	}

	ua, err := s.repo.Award(ctx, userID, achievementID, &adminID, 0, note)
	if err != nil {
		return nil, err
	}

	s.grantReward(ctx, a, ua)
	return ua, nil
}

// Evaluate — foydalanuvchi uchun avtomatik yutuqlarni baholaydi va
// mezonga yetganlarini beradi. Faollik sinxronlangach chaqiriladi.
func (s *AchievementService) Evaluate(ctx context.Context, userID uuid.UUID) ([]domain.UserAchievement, error) {
	granted, err := s.repo.EvaluateAuto(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("AchievementService.Evaluate: %w", err)
	}

	for i := range granted {
		a, err := s.repo.GetByID(ctx, granted[i].AchievementID)
		if err != nil {
			continue
		}
		s.grantReward(ctx, a, &granted[i])

		// Yangi yutuq haqida xabar. ref_id — berilgan yutuq yozuvi,
		// shuning uchun takroriy baholashda ikkinchi xabar chiqmaydi
		// (Evaluate har faollik yozuvida ishlaydi).
		if s.notify != nil {
			ref := granted[i].ID
			s.notify.Notify(ctx, domain.Notification{
				UserID:  userID,
				Type:    domain.NotifyAchievement,
				Title:   "Yangi yutuq: " + a.Title,
				Body:    a.Description,
				RefType: domain.CoinRefAchievement,
				RefID:   &ref,
			})
		}
	}
	return granted, nil
}

// Certificate — berilgan yutuq uchun sertifikat PDF sini chizadi.
//
// XAVFSIZLIK (§17.3 №26 IDOR): sertifikat boshqa odamning ismi va yutug'ini
// oshkor qiladi, shuning uchun egalik SHART tekshiriladi. Admin istalganini
// ola oladi (hisobot/tekshiruv uchun), oddiy foydalanuvchi — faqat o'zinikini.
func (s *AchievementService) Certificate(ctx context.Context, awardID, requesterID uuid.UUID, isAdmin bool) ([]byte, string, error) {
	d, err := s.repo.GetAward(ctx, awardID)
	if err != nil {
		return nil, "", err
	}
	if !isAdmin && d.UserID != requesterID {
		// Mavjud emas deymiz: "bor, lekin sizniki emas" javobi ham
		// ma'lumot sizdiradi (boshqa odamda shu yutuq borligini bildiradi).
		return nil, "", domain.ErrNotFound
	}

	pdf, err := certificate.Render(certificate.Data{
		FullName:    d.UserFullName,
		Title:       d.Title,
		Description: d.Description,
		ValueLabel:  domain.AchievementValueLabel(d.Type, d.ProgressValue),
		Note:        d.Note,
		Number:      strings.ToUpper(d.ID.String()[:8]),
		AwardedAt:   d.AwardedAt,
		Signing:     s.signing,
	})
	if err != nil {
		return nil, "", fmt.Errorf("AchievementService.Certificate: %w", err)
	}
	return pdf, certificateFileName(d), nil
}

// certificateFileName — yuklab olinadigan fayl nomi. ASCII'ga cheklanadi:
// HTTP sarlavhasiga foydalanuvchi ismi to'g'ridan-to'g'ri tushmasin
// (§17.3 №11 header injection).
func certificateFileName(d *domain.AwardDetail) string {
	return fmt.Sprintf("sertifikat-%s.pdf", strings.ToUpper(d.ID.String()[:8]))
}

// grantReward — yutuq uchun FIT Coin yozadi (mukofot belgilangan bo'lsa).
//
// Ledger ref_id = user_achievement.id, shuning uchun takror chaqiruvda
// ikkinchi yozuv qo'shilmaydi (Grant idempotent). Coin yozilmasa yutuqning
// o'zi bekor qilinmaydi: yutuq berilgani muhimroq, mukofotni admin keyin
// qo'lda to'g'rilashi mumkin.
func (s *AchievementService) grantReward(ctx context.Context, a *domain.Achievement, ua *domain.UserAchievement) {
	if a.RewardCoins <= 0 || s.coins == nil {
		return
	}
	refID := ua.ID
	_ = s.coins.Grant(ctx, &domain.FitCoin{
		UserID:  ua.UserID,
		Amount:  a.RewardCoins,
		Reason:  domain.CoinReasonAchievementReward,
		RefType: domain.CoinRefAchievement,
		RefID:   &refID,
		Note:    a.Title,
	})
}

// validate — umumiy maydonlar + turga xos mezon (§16.2).
func (s *AchievementService) validate(a *domain.Achievement) error {
	if !domain.ValidAchievementType(string(a.Type)) {
		return fmt.Errorf("%w: noma'lum tur", domain.ErrValidation)
	}
	if !domain.ValidAchievementStatus(a.Status) {
		return fmt.Errorf("%w: noma'lum holat", domain.ErrValidation)
	}
	if a.RewardCoins < 0 {
		return fmt.Errorf("%w: mukofot manfiy bo'lmasin", domain.ErrValidation)
	}

	// award_mode turdan kelib chiqadi: admin uni erkin tanlay olmaydi, aks
	// holda avtomatik yutuqni "manual" qilib mezonni chetlab o'tish mumkin edi.
	spec, ok := domain.AchievementSpec(a.Type)
	if !ok {
		return fmt.Errorf("%w: noma'lum tur", domain.ErrValidation)
	}
	a.AwardMode = spec.AwardMode

	return domain.ValidateAchievementCriteria(a.Type, a.Criteria)
}
