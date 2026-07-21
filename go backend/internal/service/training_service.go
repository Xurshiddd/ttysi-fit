package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// TrainingService — mashg'ulotlar use-case qatlami.
type TrainingService struct {
	repo domain.TrainingRepository
}

func NewTrainingService(repo domain.TrainingRepository) *TrainingService {
	return &TrainingService{repo: repo}
}

func (s *TrainingService) Create(ctx context.Context, t *domain.Training) error {
	if err := s.prepare(t); err != nil {
		return err
	}
	return s.repo.Create(ctx, t)
}

func (s *TrainingService) Update(ctx context.Context, t *domain.Training) error {
	if err := s.prepare(t); err != nil {
		return err
	}
	return s.repo.Update(ctx, t)
}

func (s *TrainingService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *TrainingService) List(ctx context.Context, f domain.TrainingFilter) ([]domain.Training, int64, error) {
	return s.repo.List(ctx, f)
}

func (s *TrainingService) Categories(ctx context.Context, publishedOnly bool) ([]string, error) {
	return s.repo.Categories(ctx, publishedOnly)
}

// Get — to'liq mashg'ulot. countView true bo'lsa ko'rishlar oshiriladi.
func (s *TrainingService) Get(ctx context.Context, id uuid.UUID, countView bool) (*domain.Training, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if countView {
		// Hisob yordamchi ma'lumot — yiqilsa videoni ko'rsatmaslik mantiqsiz.
		_ = s.repo.IncrementViews(ctx, id)
		t.Views++
	}
	return t, nil
}

func (s *TrainingService) prepare(t *domain.Training) error {
	if !domain.ValidTrainingStatus(t.Status) {
		return fmt.Errorf("%w: noma'lum holat", domain.ErrValidation)
	}
	if !domain.ValidTrainingLevel(t.Level) {
		return fmt.Errorf("%w: noma'lum daraja", domain.ErrValidation)
	}

	// Kategoriya normallashtirish: "Kardio" va "kardio " ikki xil kategoriya
	// bo'lib ketmasin (DISTINCT ro'yxati ifloslanardi).
	t.Category = strings.TrimSpace(t.Category)

	if t.Status == domain.TrainingStatusPublished && t.PublishedAt == nil {
		now := time.Now()
		t.PublishedAt = &now
	}
	return nil
}
