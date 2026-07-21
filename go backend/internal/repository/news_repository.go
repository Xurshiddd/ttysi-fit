package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
)

type newsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) domain.NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) Create(ctx context.Context, n *domain.News) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *newsRepository) Update(ctx context.Context, n *domain.News) error {
	// Select bilan aniq ro'yxat: `views` bu yerda YO'Q — aks holda admin
	// tahrirlaganda ko'rishlar soni nolga tushib ketardi (n.Views bo'sh keladi).
	res := r.db.WithContext(ctx).
		Model(&domain.News{}).
		Where("id = ? AND deleted_at IS NULL", n.ID).
		Select("title", "excerpt", "body", "cover_url", "status",
			"published_at", "pinned", "updated_at").
		Updates(n)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *newsRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&domain.News{}).
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

func (r *newsRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.News, error) {
	var n domain.News
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&n).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepository) List(ctx context.Context, f domain.NewsFilter) ([]domain.NewsListItem, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.News{}).Where("deleted_at IS NULL")

	if f.PublishedOnly {
		// Rejalashtirilgan e'lon: published_at kelajakda bo'lsa hali ko'rsatilmaydi.
		q = q.Where("status = ?", domain.NewsStatusPublished).
			Where("published_at IS NULL OR published_at <= ?", time.Now())
	} else if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}

	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("title ILIKE ? OR excerpt ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.NewsListItem
	err := q.
		Select("id, title, COALESCE(excerpt, '') AS excerpt, COALESCE(cover_url, '') AS cover_url, status, published_at, pinned, views").
		// Muhim yangilik yuqorida, keyin yangi -> eski.
		Order("pinned DESC, COALESCE(published_at, created_at) DESC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// IncrementViews — atomik oshirish.
//
// `SET views = views + 1` — o'qib, +1 qilib, yozish emas: ikki parallel o'qish
// bir xil qiymatni ko'rib, bittasining hisobi yo'qolardi (§11.3 race).
func (r *newsRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.News{}).
		Where("id = ? AND deleted_at IS NULL", id).
		UpdateColumn("views", gorm.Expr("views + 1")).Error
}
