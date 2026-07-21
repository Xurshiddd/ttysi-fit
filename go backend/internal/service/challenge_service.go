package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// ChallengeService — chellenj use-case qatlami.
type ChallengeService struct {
	repo domain.ChallengeRepository
}

func NewChallengeService(repo domain.ChallengeRepository) *ChallengeService {
	return &ChallengeService{repo: repo}
}

// Types — admin panel dinamik formasi uchun tur ta'riflari (§16.2).
func (s *ChallengeService) Types() []domain.ChallengeTypeSpec {
	return domain.ChallengeTypeSpecs()
}

// Create — admin yangi chellenj yaratadi. Config turga qarab tekshiriladi.
func (s *ChallengeService) Create(ctx context.Context, c *domain.Challenge) error {
	if err := s.validate(c); err != nil {
		return err
	}
	return s.repo.Create(ctx, c)
}

func (s *ChallengeService) Update(ctx context.Context, c *domain.Challenge) error {
	if err := s.validate(c); err != nil {
		return err
	}
	return s.repo.Update(ctx, c)
}

func (s *ChallengeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *ChallengeService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Challenge, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ChallengeService) List(ctx context.Context, f domain.ChallengeFilter) ([]domain.Challenge, int64, error) {
	return s.repo.List(ctx, f)
}

// ListForUser — mobil ro'yxat (foydalanuvchi holati bilan).
func (s *ChallengeService) ListForUser(ctx context.Context, userID uuid.UUID, f domain.ChallengeFilter) ([]domain.ChallengeView, int64, error) {
	return s.repo.ListForUser(ctx, userID, f)
}

// Join — foydalanuvchi chellenjga qo'shiladi va progressi darrov hisoblanadi
// (chellenj boshlanganidan beri yig'ilgan faollik ham hisobga olinadi).
func (s *ChallengeService) Join(ctx context.Context, userID, challengeID uuid.UUID) (*domain.UserChallenge, error) {
	ch, err := s.repo.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	// Faqat aktiv chellenjga qo'shilish mumkin — draft admin ko'rinishi,
	// finished esa yopilgan.
	if ch.Status != domain.ChallengeStatusActive {
		return nil, fmt.Errorf("%w: chellenj aktiv emas", domain.ErrValidation)
	}
	if ch.EndsAt != nil && ch.EndsAt.Before(time.Now()) {
		return nil, fmt.Errorf("%w: chellenj muddati tugagan", domain.ErrValidation)
	}

	if err := s.repo.Join(ctx, userID, challengeID); err != nil {
		return nil, err
	}
	return s.repo.RecalcProgress(ctx, userID, challengeID)
}

// Progress — ishtirokchi progressini qayta hisoblab qaytaradi.
func (s *ChallengeService) Progress(ctx context.Context, userID, challengeID uuid.UUID) (*domain.UserChallenge, error) {
	return s.repo.RecalcProgress(ctx, userID, challengeID)
}

// validate — umumiy maydonlar + turga xos config (§16.2).
func (s *ChallengeService) validate(c *domain.Challenge) error {
	if !domain.ValidChallengeType(string(c.Type)) {
		return fmt.Errorf("%w: noma'lum tur", domain.ErrValidation)
	}
	if !domain.ValidChallengeStatus(c.Status) {
		return fmt.Errorf("%w: noma'lum holat", domain.ErrValidation)
	}
	if !domain.ValidChallengeScope(c.Scope) {
		return fmt.Errorf("%w: noma'lum qamrov", domain.ErrValidation)
	}
	if c.StartsAt != nil && c.EndsAt != nil && c.EndsAt.Before(*c.StartsAt) {
		return fmt.Errorf("%w: tugash sanasi boshlanishdan oldin", domain.ErrValidation)
	}
	if c.RewardCoins < 0 {
		return fmt.Errorf("%w: mukofot manfiy bo'lmasin", domain.ErrValidation)
	}
	return domain.ValidateChallengeConfig(c.Type, c.Config)
}
