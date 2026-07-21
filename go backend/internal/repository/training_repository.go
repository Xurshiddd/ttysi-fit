package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
)

type trainingRepository struct {
	db *gorm.DB
}

func NewTrainingRepository(db *gorm.DB) domain.TrainingRepository {
	return &trainingRepository{db: db}
}

func (r *trainingRepository) Create(ctx context.Context, t *domain.Training) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *trainingRepository) Update(ctx context.Context, t *domain.Training) error {
	// `views` ro'yxatda yo'q: admin tahrirlaganda hisob nolga tushmasin.
	res := r.db.WithContext(ctx).
		Model(&domain.Training{}).
		Where("id = ? AND deleted_at IS NULL", t.ID).
		Select("title", "description", "category", "level", "video_url",
			"thumbnail_url", "duration_min", "status", "published_at",
			"sort_order", "updated_at").
		Updates(t)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *trainingRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&domain.Training{}).
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

func (r *trainingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Training, error) {
	var t domain.Training
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *trainingRepository) List(ctx context.Context, f domain.TrainingFilter) ([]domain.Training, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.Training{}).Where("deleted_at IS NULL")

	if f.PublishedOnly {
		q = q.Where("status = ?", domain.TrainingStatusPublished).
			Where("published_at IS NULL OR published_at <= ?", time.Now())
	} else if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.Level != "" {
		q = q.Where("level = ?", f.Level)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.Training
	err := q.Order("sort_order, COALESCE(published_at, created_at) DESC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *trainingRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.Training{}).
		Where("id = ? AND deleted_at IS NULL", id).
		UpdateColumn("views", gorm.Expr("views + 1")).Error
}

// Categories — mavjud kategoriyalar. Kodda ro'yxat yo'q: admin yangi
// kategoriya yozsa u avtomatik shu ro'yxatga tushadi (§16).
func (r *trainingRepository) Categories(ctx context.Context, publishedOnly bool) ([]string, error) {
	q := r.db.WithContext(ctx).
		Model(&domain.Training{}).
		Where("deleted_at IS NULL").
		Where("category IS NOT NULL AND category <> ''")

	if publishedOnly {
		q = q.Where("status = ?", domain.TrainingStatusPublished)
	}

	var out []string
	err := q.Distinct().Order("category").Pluck("category", &out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}
