package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"golang.org/x/sync/errgroup"
)

// maxRangeDays — bitta hisobot qamrashi mumkin bo'lgan maksimal kun soni.
// "all" davri ham shu bilan cheklanadi: cheksiz oraliq generate_series'da
// millionlab qator yasab serverni cho'ktirishi mumkin (§17.3 #39).
const maxRangeDays = 366

// AnalyticsService — admin dashboard va hisobotlar use-case qatlami.
type AnalyticsService struct {
	repo domain.AnalyticsRepository
	loc  *time.Location
}

func NewAnalyticsService(repo domain.AnalyticsRepository, loc *time.Location) *AnalyticsService {
	if loc == nil {
		loc = time.UTC
	}
	return &AnalyticsService{repo: repo, loc: loc}
}

// Filter — davr nomi va fakultetdan AnalyticsFilter yasaydi.
//
// Sana chegarasi mahalliy mintaqada hisoblanadi (APP_TIMEZONE): admin
// "hafta" deganda o'zining kalendar haftasini kutadi, UTC kunini emas.
func (s *AnalyticsService) Filter(period string, facultyID *uuid.UUID) (domain.AnalyticsFilter, error) {
	if !domain.ValidRatingPeriod(period) {
		return domain.AnalyticsFilter{}, fmt.Errorf("%w: davr", domain.ErrValidation)
	}

	n := time.Now().In(s.loc)
	to := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, s.loc)

	var from time.Time
	switch domain.RatingPeriod(period) {
	case domain.PeriodWeek:
		from = to.AddDate(0, 0, -6)
	case domain.PeriodMonth:
		from = to.AddDate(0, 0, -29)
	default: // all — cheklangan oyna (yuqoridagi izohga qarang)
		from = to.AddDate(0, 0, -(maxRangeDays - 1))
	}

	return domain.AnalyticsFilter{From: from, To: to, FacultyID: facultyID}, nil
}

// Get — dashboard uchun to'liq to'plam.
//
// Uchala so'rov bir-biriga bog'liq emas — errgroup bilan parallel bajariladi
// (§11.2). Biri xato qilsa context bekor bo'lib qolganlari ham to'xtaydi,
// ya'ni yiqilgan so'rov uchun bekorga DB band qilinmaydi.
func (s *AnalyticsService) Get(ctx context.Context, f domain.AnalyticsFilter) (*domain.Analytics, error) {
	var (
		overview *domain.AnalyticsOverview
		series   []domain.AnalyticsPoint
		faculty  []domain.FacultyStat
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		var err error
		overview, err = s.repo.Overview(gctx, f)
		return err
	})
	g.Go(func() error {
		var err error
		series, err = s.repo.Timeseries(gctx, f)
		return err
	})
	g.Go(func() error {
		var err error
		faculty, err = s.repo.ByFaculty(gctx, f)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("AnalyticsService.Get: %w", err)
	}

	return &domain.Analytics{
		From:       f.From.Format("2006-01-02"),
		To:         f.To.Format("2006-01-02"),
		Overview:   *overview,
		Timeseries: series,
		Faculties:  faculty,
	}, nil
}

// StreamUserActivity — eksport qatorlarini birma-bir uzatadi.
func (s *AnalyticsService) StreamUserActivity(ctx context.Context, f domain.AnalyticsFilter, fn func(domain.UserActivityRow) error) error {
	return s.repo.StreamUserActivity(ctx, f, fn)
}
