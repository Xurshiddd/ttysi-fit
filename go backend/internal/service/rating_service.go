package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/ttysi-fit/backend/internal/domain"
	"go.uber.org/zap"
)

// ratingCacheTTL — reyting cache muddati (CLAUDE.md §12.1 — statistika cache 5 min).
const ratingCacheTTL = 5 * time.Minute

// RatingResult — paginatsiyali reyting javobi (cache'da butunligicha saqlanadi).
type RatingResult struct {
	Entries []domain.RatingEntry `json:"entries"`
	Total   int64                `json:"total"`
}

// RatingService — reyting use-case qatlami: cache-aside (§12.2) + repository.
type RatingService struct {
	repo  domain.RatingRepository
	redis *redis.Client
	log   *zap.Logger
}

func NewRatingService(repo domain.RatingRepository, rdb *redis.Client, log *zap.Logger) *RatingService {
	return &RatingService{repo: repo, redis: rdb, log: log}
}

// cacheKey — §12.3 konvensiyasi: rating:{type}:{period}:{faculty}:{group}:{page}:{limit}
func cacheKey(f domain.RatingFilter) string {
	fac, grp := "-", "-"
	if f.FacultyID != nil {
		fac = f.FacultyID.String()
	}
	if f.GroupID != nil {
		grp = f.GroupID.String()
	}
	return fmt.Sprintf("rating:%s:%s:%s:%s:%d:%d", f.Type, f.Period, fac, grp, f.Page, f.Limit)
}

// List — reytingni cache-aside pattern bilan qaytaradi.
// Redis xatosi cache miss sifatida qaraladi (§12.5) — DB'dan o'qiladi.
func (s *RatingService) List(ctx context.Context, f domain.RatingFilter) (*RatingResult, error) {
	// Default va chegaralar (§14.2: default 20, max 100).
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 {
		f.Limit = 20
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Period == "" {
		f.Period = domain.PeriodWeek
	}

	key := cacheKey(f)

	// 1. Cache'dan o'qish.
	if cached, err := s.redis.Get(ctx, key).Bytes(); err == nil {
		var res RatingResult
		if err := json.Unmarshal(cached, &res); err == nil {
			return &res, nil
		}
	}

	// 2. DB'dan o'qish (kesimga qarab strategiya — §16.2 uslubi).
	var (
		entries []domain.RatingEntry
		total   int64
		err     error
	)
	switch f.Type {
	case domain.RatingGroup:
		entries, total, err = s.repo.ListGroups(ctx, f)
	case domain.RatingFaculty:
		entries, total, err = s.repo.ListFaculties(ctx, f)
	default: // student / employee
		entries, total, err = s.repo.ListIndividual(ctx, f)
	}
	if err != nil {
		return nil, fmt.Errorf("RatingService.List: %w", err)
	}
	if entries == nil {
		entries = []domain.RatingEntry{}
	}
	res := &RatingResult{Entries: entries, Total: total}

	// 3. Cache'ga yozish (TTL bilan). Xato — jiddiy emas, faqat log.
	if data, err := json.Marshal(res); err == nil {
		if err := s.redis.Set(ctx, key, data, ratingCacheTTL).Err(); err != nil {
			s.log.Warn("rating cache yozilmadi", zap.Error(err))
		}
	}
	return res, nil
}

// MyRank — foydalanuvchining o'z o'rni (qisqa TTL cache bilan).
func (s *RatingService) MyRank(ctx context.Context, userID uuid.UUID, period domain.RatingPeriod) (*domain.MyRating, error) {
	if !domain.ValidRatingPeriod(string(period)) {
		period = domain.PeriodWeek
	}
	key := fmt.Sprintf("rating:me:%s:%s", period, userID)

	if cached, err := s.redis.Get(ctx, key).Bytes(); err == nil {
		var m domain.MyRating
		if err := json.Unmarshal(cached, &m); err == nil {
			return &m, nil
		}
	}

	m, err := s.repo.MyRank(ctx, userID, period)
	if err != nil {
		return nil, fmt.Errorf("RatingService.MyRank: %w", err)
	}

	if data, err := json.Marshal(m); err == nil {
		if err := s.redis.Set(ctx, key, data, ratingCacheTTL).Err(); err != nil {
			s.log.Warn("rating cache yozilmadi", zap.Error(err))
		}
	}
	return m, nil
}
