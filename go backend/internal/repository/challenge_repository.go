package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type challengeRepository struct {
	db *gorm.DB
}

func NewChallengeRepository(db *gorm.DB) domain.ChallengeRepository {
	return &challengeRepository{db: db}
}

func (r *challengeRepository) Create(ctx context.Context, c *domain.Challenge) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *challengeRepository) Update(ctx context.Context, c *domain.Challenge) error {
	// Select bilan aniq ro'yxat: mijoz yubormagan maydon nolga tushib qolmasin
	// va created_at/deleted_at tegilmasin.
	res := r.db.WithContext(ctx).
		Model(&domain.Challenge{}).
		Where("id = ? AND deleted_at IS NULL", c.ID).
		Select("type", "title", "description", "scope", "starts_at", "ends_at",
			"status", "reward_coins", "config", "cover_url", "updated_at").
		Updates(c)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *challengeRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&domain.Challenge{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *challengeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Challenge, error) {
	var c domain.Challenge
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *challengeRepository) List(ctx context.Context, f domain.ChallengeFilter) ([]domain.Challenge, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.Challenge{}).Where("deleted_at IS NULL")
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.Challenge
	err := q.Order("COALESCE(starts_at, created_at) DESC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListForUser — chellenjlar + shu foydalanuvchi holati.
//
// Ikkita so'rov: (1) chellenjlar sahifasi, (2) o'sha id lar bo'yicha ishtirok
// yozuvlari `IN` bilan. Ro'yxat uzunligidan qat'i nazar 2 ta so'rov — `for`
// ichida so'rov yo'q (§3.1).
func (r *challengeRepository) ListForUser(ctx context.Context, userID uuid.UUID, f domain.ChallengeFilter) ([]domain.ChallengeView, int64, error) {
	items, total, err := r.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		return []domain.ChallengeView{}, total, nil
	}

	ids := make([]uuid.UUID, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}

	var ucs []domain.UserChallenge
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND challenge_id IN ?", userID, ids).
		Find(&ucs).Error; err != nil {
		return nil, 0, err
	}
	byChallenge := make(map[uuid.UUID]domain.UserChallenge, len(ucs))
	for _, uc := range ucs {
		byChallenge[uc.ChallengeID] = uc
	}

	out := make([]domain.ChallengeView, 0, len(items))
	for _, c := range items {
		v := domain.ChallengeView{Challenge: c}
		v.Target = domain.ChallengeTarget(c.Type, c.Config)
		if uc, ok := byChallenge[c.ID]; ok {
			v.Joined = true
			v.Progress = uc.Progress
			v.Completed = uc.CompletedAt != nil
			v.RewardGranted = uc.RewardGranted
		}
		if v.Target > 0 {
			pct := v.Progress / v.Target * 100
			if pct > 100 {
				pct = 100
			}
			v.ProgressPct = pct
		}
		out = append(out, v)
	}
	return out, total, nil
}

// Join — idempotent: takroriy qo'shilish xato bermaydi (unique indeks bo'yicha).
func (r *challengeRepository) Join(ctx context.Context, userID, challengeID uuid.UUID) error {
	uc := domain.UserChallenge{
		UserID:      userID,
		ChallengeID: challengeID,
		JoinedAt:    time.Now(),
	}
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "challenge_id"}},
			DoNothing: true,
		}).
		Create(&uc).Error
}

// RecalcProgress — ishtirokchi progressini `activities` dan qayta yig'adi.
//
// Metrika ustuni (steps/distance_m/active_min) SQL'ga matn sifatida tushadi,
// shuning uchun u FAQAT domain registridan olinadi — mijoz kiritmasi bu yerga
// hech qachon yetib kelmaydi (§3.2 SQL injection).
func (r *challengeRepository) RecalcProgress(ctx context.Context, userID, challengeID uuid.UUID) (*domain.UserChallenge, error) {
	ch, err := r.GetByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	spec, ok := domain.ChallengeSpec(ch.Type)
	if !ok {
		return nil, domain.ErrNotFound
	}

	var uc domain.UserChallenge
	err = r.db.WithContext(ctx).
		Where("user_id = ? AND challenge_id = ?", userID, challengeID).
		First(&uc).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Metrikasiz tur (custom) — progress avtomatik hisoblanmaydi.
	if spec.Metric == "" {
		return &uc, nil
	}

	q := r.db.WithContext(ctx).
		Table("activities").
		Where("user_id = ? AND deleted_at IS NULL", userID)
	if ch.StartsAt != nil {
		q = q.Where("activity_date >= ?", ch.StartsAt)
	}
	if ch.EndsAt != nil {
		q = q.Where("activity_date <= ?", ch.EndsAt)
	}

	var progress float64
	if err := q.Select("COALESCE(SUM(" + spec.Metric + "), 0)").Scan(&progress).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	fields := map[string]any{"progress": progress, "updated_at": now}

	target := domain.ChallengeTarget(ch.Type, ch.Config)
	if target > 0 && progress >= target && uc.CompletedAt == nil {
		fields["completed_at"] = now
		uc.CompletedAt = &now
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.UserChallenge{}).
		Where("id = ?", uc.ID).
		Updates(fields).Error; err != nil {
		return nil, err
	}

	uc.Progress = progress
	uc.UpdatedAt = now
	return &uc, nil
}
