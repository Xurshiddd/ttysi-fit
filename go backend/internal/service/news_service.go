package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// excerptRunes — avtomatik yasaladigan qisqa matn uzunligi.
const excerptRunes = 160

// NewsService — yangiliklar use-case qatlami.
type NewsService struct {
	repo domain.NewsRepository
}

func NewNewsService(repo domain.NewsRepository) *NewsService {
	return &NewsService{repo: repo}
}

func (s *NewsService) Create(ctx context.Context, n *domain.News) error {
	if err := s.prepare(n); err != nil {
		return err
	}
	return s.repo.Create(ctx, n)
}

func (s *NewsService) Update(ctx context.Context, n *domain.News) error {
	if err := s.prepare(n); err != nil {
		return err
	}
	return s.repo.Update(ctx, n)
}

func (s *NewsService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *NewsService) List(ctx context.Context, f domain.NewsFilter) ([]domain.NewsListItem, int64, error) {
	return s.repo.List(ctx, f)
}

// Get — to'liq yangilik. `countView` true bo'lsa ko'rishlar oshiriladi
// (mobil o'qish; admin ko'rigi hisobni buzmasin).
func (s *NewsService) Get(ctx context.Context, id uuid.UUID, countView bool) (*domain.News, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if countView {
		// Ko'rish hisobi — yordamchi ma'lumot. U yiqilsa yangilikni
		// ko'rsatmaslik mantiqsiz, shuning uchun xatoni yutamiz.
		_ = s.repo.IncrementViews(ctx, id)
		n.Views++
	}
	return n, nil
}

// prepare — validatsiya va avtomatik to'ldirish.
func (s *NewsService) prepare(n *domain.News) error {
	if !domain.ValidNewsStatus(n.Status) {
		return fmt.Errorf("%w: noma'lum holat", domain.ErrValidation)
	}

	// Excerpt bo'sh bo'lsa body'dan yasaymiz — admin har safar qo'lda
	// yozishga majbur bo'lmasin, ro'yxat esa bo'sh ko'rinmasin.
	if n.Excerpt == "" {
		n.Excerpt = domain.MakeExcerpt(n.Body, excerptRunes)
	}

	// E'lon qilinsa-yu vaqti ko'rsatilmagan bo'lsa — hozir.
	// Aks holda published_at NULL bo'lib, ro'yxatda tartib buzilardi.
	if n.Status == domain.NewsStatusPublished && n.PublishedAt == nil {
		now := time.Now()
		n.PublishedAt = &now
	}
	return nil
}
