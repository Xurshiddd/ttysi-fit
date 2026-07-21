package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// CompetitionService — musobaqa use-case qatlami.
type CompetitionService struct {
	repo domain.CompetitionRepository
}

func NewCompetitionService(repo domain.CompetitionRepository) *CompetitionService {
	return &CompetitionService{repo: repo}
}

// Types — admin dinamik formasi uchun tur ta'riflari (§16.2).
func (s *CompetitionService) Types() []domain.CompetitionTypeSpec {
	return domain.CompetitionTypeSpecs()
}

func (s *CompetitionService) Create(ctx context.Context, c *domain.Competition) error {
	if err := s.validate(c); err != nil {
		return err
	}
	return s.repo.Create(ctx, c)
}

func (s *CompetitionService) Update(ctx context.Context, c *domain.Competition) error {
	if err := s.validate(c); err != nil {
		return err
	}
	return s.repo.Update(ctx, c)
}

func (s *CompetitionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *CompetitionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Competition, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CompetitionService) List(ctx context.Context, f domain.CompetitionFilter) ([]domain.Competition, int64, error) {
	return s.repo.List(ctx, f)
}

func (s *CompetitionService) ListForUser(ctx context.Context, userID uuid.UUID, f domain.CompetitionFilter) ([]domain.CompetitionView, int64, error) {
	return s.repo.ListForUser(ctx, userID, f)
}

// Register — ro'yxatdan o'tish. Holat/muddat/joy tekshiruvi repozitoriyda,
// tranzaksiya va blok ostida (poygaga qarshi).
func (s *CompetitionService) Register(ctx context.Context, userID, competitionID uuid.UUID) (*domain.CompetitionRegistration, error) {
	return s.repo.Register(ctx, userID, competitionID)
}

func (s *CompetitionService) Cancel(ctx context.Context, userID, competitionID uuid.UUID) error {
	return s.repo.Cancel(ctx, userID, competitionID)
}

func (s *CompetitionService) Participants(ctx context.Context, competitionID uuid.UUID, page, limit int) ([]domain.CompetitionRegistration, int64, error) {
	return s.repo.Participants(ctx, competitionID, page, limit)
}

func (s *CompetitionService) validate(c *domain.Competition) error {
	if !domain.ValidCompetitionType(string(c.Type)) {
		return fmt.Errorf("%w: noma'lum tur", domain.ErrValidation)
	}
	if !domain.ValidCompetitionStatus(c.Status) {
		return fmt.Errorf("%w: noma'lum holat", domain.ErrValidation)
	}
	if !domain.ValidChallengeScope(c.Scope) {
		return fmt.Errorf("%w: noma'lum qamrov", domain.ErrValidation)
	}
	if c.StartsAt != nil && c.EndsAt != nil && c.EndsAt.Before(*c.StartsAt) {
		return fmt.Errorf("%w: tugash sanasi boshlanishdan oldin", domain.ErrValidation)
	}
	// Ro'yxatdan o'tish musobaqa tugagandan keyin yopilishi mantiqsiz.
	if c.RegEndsAt != nil && c.StartsAt != nil && c.RegEndsAt.After(*c.StartsAt) {
		return fmt.Errorf("%w: ro'yxat muddati musobaqa boshlanishidan keyin", domain.ErrValidation)
	}
	if c.RewardCoins < 0 {
		return fmt.Errorf("%w: mukofot manfiy bo'lmasin", domain.ErrValidation)
	}
	if c.MaxParticipants != nil && *c.MaxParticipants < 0 {
		return fmt.Errorf("%w: joy soni manfiy bo'lmasin", domain.ErrValidation)
	}
	return domain.ValidateCompetitionConfig(c.Type, c.Config)
}
