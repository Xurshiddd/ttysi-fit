package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/internal/dto"
)

// fakeActivityRepo — DB'siz sinov uchun ActivityRepository.
// Nechta chaqiruv bo'lganini sanaydi: batch bitta so'rovda ketishini
// tekshirish uchun (CLAUDE.md §3.1 — sikl ichida DB so'rovi bo'lmasin).
type fakeActivityRepo struct {
	upsertCalls int
	manyCalls   int
	lastRows    []domain.Activity
	err         error
}

func (f *fakeActivityRepo) Upsert(_ context.Context, a *domain.Activity) error {
	f.upsertCalls++
	f.lastRows = []domain.Activity{*a}
	return f.err
}

func (f *fakeActivityRepo) UpsertMany(_ context.Context, rows []domain.Activity) error {
	f.manyCalls++
	f.lastRows = rows
	return f.err
}

func (f *fakeActivityRepo) ListByUser(context.Context, uuid.UUID, time.Time, time.Time, int) ([]domain.Activity, error) {
	return nil, nil
}

func (f *fakeActivityRepo) Stats(context.Context, uuid.UUID, time.Time) (*domain.ActivityStats, error) {
	return &domain.ActivityStats{}, nil
}

func newTestService(repo domain.ActivityRepository, loc *time.Location) *ActivityService {
	return NewActivityService(repo, nil, loc, nil)
}

// Sana sozlangan mintaqada hisoblanishi kerak, UTC da emas.
//
// Regressiya: avval `time.Now().UTC()` ishlatilardi. O'zbekiston UTC+5
// bo'lgani uchun mahalliy 00:00–05:00 oralig'ida "bugun" KECHAGI kunga
// tushib, o'sha kunning yozuvi ustiga yozilardi.
//
// Test barqaror bo'lishi uchun +12 va -12 mintaqalari olingan: ular doim
// 24 soat farq qiladi, ya'ni kalendar sanasi HECH QACHON teng bo'lmaydi.
// Eski (UTC'ga qotirilgan) kod ikkalasida bir xil sana berardi.
func TestActivityService_TodayUsesConfiguredTimezone(t *testing.T) {
	east := newTestService(&fakeActivityRepo{}, time.FixedZone("EAST", 12*3600))
	west := newTestService(&fakeActivityRepo{}, time.FixedZone("WEST", -12*3600))

	if e, w := east.Today().Format("2006-01-02"), west.Today().Format("2006-01-02"); e == w {
		t.Fatalf("mintaqa e'tiborga olinmadi: har ikkalasi ham %s", e)
	}
}

// Mijoz yuborgan sana o'sha mintaqada o'qilishi kerak — kun surilib ketmasin.
func TestActivityService_RecordParsesDateInLocation(t *testing.T) {
	tz, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		t.Fatalf("mintaqa yuklanmadi: %v", err)
	}
	repo := &fakeActivityRepo{}
	svc := newTestService(repo, tz)

	// Kechagi kun — o'tgan sana, ErrFutureDate ishga tushmasin.
	yesterday := time.Now().In(tz).AddDate(0, 0, -1).Format("2006-01-02")

	a, err := svc.Record(context.Background(), uuid.New(), dto.RecordActivityRequest{
		Date: yesterday, Steps: 5000, Source: "health_connect",
	})
	if err != nil {
		t.Fatalf("Record xato: %v", err)
	}
	if got := a.ActivityDate.Format("2006-01-02"); got != yesterday {
		t.Errorf("sana surilib ketdi: kutilgan %s, olingan %s", yesterday, got)
	}
	if h := a.ActivityDate.Hour(); h != 0 {
		t.Errorf("sana kun boshi bo'lishi kerak, soat: %d", h)
	}
}

// Kelajakdagi sana rad etilsin (mijoz soati noto'g'ri yoki reytingni
// sun'iy ko'tarish urinishi).
func TestActivityService_RejectsFutureDate(t *testing.T) {
	tz := time.FixedZone("TEST", 5*3600)
	repo := &fakeActivityRepo{}
	svc := newTestService(repo, tz)

	future := time.Now().In(tz).AddDate(0, 0, 2).Format("2006-01-02")
	_, err := svc.Record(context.Background(), uuid.New(), dto.RecordActivityRequest{
		Date: future, Steps: 99000,
	})
	if !errors.Is(err, domain.ErrFutureDate) {
		t.Fatalf("ErrFutureDate kutilgandi, olingan: %v", err)
	}
	if repo.upsertCalls != 0 {
		t.Errorf("rad etilgan yozuv DB ga borib qolibdi (%d chaqiruv)", repo.upsertCalls)
	}
}

// Backfill: bir necha kun BITTA so'rovda yozilishi kerak.
func TestActivityService_RecordBatchSingleQuery(t *testing.T) {
	tz := time.FixedZone("TEST", 5*3600)
	repo := &fakeActivityRepo{}
	svc := newTestService(repo, tz)

	now := time.Now().In(tz)
	items := make([]dto.RecordActivityRequest, 0, 7)
	for i := 6; i >= 0; i-- {
		items = append(items, dto.RecordActivityRequest{
			Date:   now.AddDate(0, 0, -i).Format("2006-01-02"),
			Steps:  1000 * (i + 1),
			Source: "health_connect",
		})
	}

	rows, err := svc.RecordBatch(context.Background(), uuid.New(), items)
	if err != nil {
		t.Fatalf("RecordBatch xato: %v", err)
	}
	if len(rows) != 7 {
		t.Errorf("7 kun kutilgandi, olingan %d", len(rows))
	}
	if repo.manyCalls != 1 {
		t.Errorf("bitta bulk so'rov kutilgandi, olingan %d", repo.manyCalls)
	}
	if repo.upsertCalls != 0 {
		t.Errorf("sikl ichida bitta-bitta Upsert qilinibdi (%d)", repo.upsertCalls)
	}
}

// Bir xil sana ikki marta kelsa dedup bo'lsin: PostgreSQL bitta so'rovda
// bir qatorni ikki marta yangilay olmaydi ("cannot affect row a second time").
func TestActivityService_RecordBatchDeduplicatesDates(t *testing.T) {
	tz := time.FixedZone("TEST", 5*3600)
	repo := &fakeActivityRepo{}
	svc := newTestService(repo, tz)

	day := time.Now().In(tz).AddDate(0, 0, -1).Format("2006-01-02")
	rows, err := svc.RecordBatch(context.Background(), uuid.New(), []dto.RecordActivityRequest{
		{Date: day, Steps: 3000},
		{Date: day, Steps: 4200}, // oxirgisi — eng yangi o'qish
	})
	if err != nil {
		t.Fatalf("RecordBatch xato: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("dedup ishlamadi: %d yozuv", len(rows))
	}
	if rows[0].Steps != 4200 {
		t.Errorf("oxirgi qiymat saqlanishi kerak edi, olingan %d", rows[0].Steps)
	}
}

func TestActivityService_RecordBatchLimits(t *testing.T) {
	tz := time.FixedZone("TEST", 5*3600)
	svc := newTestService(&fakeActivityRepo{}, tz)

	if _, err := svc.RecordBatch(context.Background(), uuid.New(), nil); !errors.Is(err, domain.ErrEmptyBatch) {
		t.Errorf("ErrEmptyBatch kutilgandi, olingan: %v", err)
	}

	too := make([]dto.RecordActivityRequest, maxBatchDays+1)
	if _, err := svc.RecordBatch(context.Background(), uuid.New(), too); !errors.Is(err, domain.ErrBatchTooLarge) {
		t.Errorf("ErrBatchTooLarge kutilgandi, olingan: %v", err)
	}
}
