package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// RewardService — FIT Coin do'koni use-case qatlami.
type RewardService struct {
	repo domain.RewardRepository
}

func NewRewardService(repo domain.RewardRepository) *RewardService {
	return &RewardService{repo: repo}
}

// Categories — admin panel dinamik formasi uchun kategoriya ro'yxati (§16.2).
func (s *RewardService) Categories() []string {
	return []string{
		domain.RewardCategoryMerch,
		domain.RewardCategoryEquipment,
		domain.RewardCategoryCertificate,
		domain.RewardCategoryOther,
	}
}

// List — do'kon ro'yxati.
func (s *RewardService) List(ctx context.Context, f domain.RewardFilter) ([]domain.Reward, int64, error) {
	if f.Category != "" && !domain.ValidRewardCategory(f.Category) {
		return nil, 0, fmt.Errorf("%w: kategoriya", domain.ErrValidation)
	}
	return s.repo.List(ctx, f)
}

func (s *RewardService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Reward, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *RewardService) Create(ctx context.Context, r *domain.Reward) error {
	normalize(r)
	if err := s.validate(r); err != nil {
		return err
	}
	return s.repo.Create(ctx, r)
}

func (s *RewardService) Update(ctx context.Context, r *domain.Reward) error {
	normalize(r)
	if err := s.validate(r); err != nil {
		return err
	}
	return s.repo.Update(ctx, r)
}

// normalize — DB cheklovlariga mos holatga keltiradi.
//
// config: ustun NOT NULL DEFAULT '{}'. GORM nil JSON ni ANIQ NULL sifatida
// yuboradi va DB default'i ishlamaydi — shuning uchun bo'sh obyektni o'zimiz
// qo'yamiz. Bu servis qatlamida: seeder yoki boshqa chaqiruvchi ham
// shu yo'ldan o'tadi.
func normalize(r *domain.Reward) {
	if len(r.Config) == 0 {
		r.Config = []byte("{}")
	}
	if r.Category == "" {
		r.Category = domain.RewardCategoryOther
	}
}

func (s *RewardService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// validate — biznes qoidalari (DTO bind validatsiyasidan tashqari).
func (s *RewardService) validate(r *domain.Reward) error {
	if r.Title == "" {
		return fmt.Errorf("%w: nom bo'sh", domain.ErrValidation)
	}
	if r.CostCoins <= 0 {
		return fmt.Errorf("%w: narx musbat bo'lishi kerak", domain.ErrValidation)
	}
	if !domain.ValidRewardCategory(r.Category) {
		return fmt.Errorf("%w: kategoriya", domain.ErrValidation)
	}
	if r.Stock != nil && *r.Stock < 0 {
		return fmt.Errorf("%w: miqdor manfiy bo'la olmaydi", domain.ErrValidation)
	}
	if r.PerUserLimit != nil && *r.PerUserLimit <= 0 {
		return fmt.Errorf("%w: limit musbat bo'lishi kerak", domain.ErrValidation)
	}
	// Vaqt oynasi teskari bo'lsa sovg'a hech qachon ko'rinmaydi — bu
	// odatda admin xatosi, jimgina qabul qilmaymiz.
	if r.StartsAt != nil && r.EndsAt != nil && r.EndsAt.Before(*r.StartsAt) {
		return fmt.Errorf("%w: tugash sanasi boshlanishdan oldin", domain.ErrValidation)
	}
	return nil
}

// Redeem — sovg'ani almashtirish (mobil).
func (s *RewardService) Redeem(ctx context.Context, userID, rewardID uuid.UUID) (*domain.RewardRedemption, error) {
	red, err := s.repo.Redeem(ctx, userID, rewardID)
	if err != nil {
		return nil, fmt.Errorf("RewardService.Redeem: %w", err)
	}
	return red, nil
}

// MyRedemptions — foydalanuvchining buyurtmalari.
func (s *RewardService) MyRedemptions(ctx context.Context, userID uuid.UUID, f domain.RedemptionFilter) ([]domain.RedemptionDetail, int64, error) {
	// Egalik MAJBURAN o'rnatiladi: mijoz user_id yuborib boshqa odamning
	// buyurtmalarini ko'ra olmasin (§17.3 #26 IDOR).
	f.UserID = &userID
	return s.repo.ListRedemptions(ctx, f)
}

// ListRedemptions — admin ro'yxati (hamma foydalanuvchi).
func (s *RewardService) ListRedemptions(ctx context.Context, f domain.RedemptionFilter) ([]domain.RedemptionDetail, int64, error) {
	if f.Status != "" && !validRedemptionStatus(f.Status) {
		return nil, 0, fmt.Errorf("%w: holat", domain.ErrValidation)
	}
	return s.repo.ListRedemptions(ctx, f)
}

func (s *RewardService) MarkIssued(ctx context.Context, id, adminID uuid.UUID, note string) (*domain.RewardRedemption, error) {
	return s.repo.MarkIssued(ctx, id, adminID, note)
}

func (s *RewardService) Cancel(ctx context.Context, id, adminID uuid.UUID, note string) (*domain.RewardRedemption, error) {
	return s.repo.Cancel(ctx, id, adminID, note)
}

func validRedemptionStatus(s string) bool {
	switch s {
	case domain.RedemptionPending, domain.RedemptionIssued, domain.RedemptionCancelled:
		return true
	}
	return false
}
