package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// fakeAnalyticsRepo — DB'siz sinov uchun. Har metod nechta marta va qaysi
// filtr bilan chaqirilganini yozib boradi.
type fakeAnalyticsRepo struct {
	gotFilter domain.AnalyticsFilter
	calls     int
	err       error
}

func (f *fakeAnalyticsRepo) Overview(_ context.Context, flt domain.AnalyticsFilter) (*domain.AnalyticsOverview, error) {
	f.calls++
	f.gotFilter = flt
	return &domain.AnalyticsOverview{TotalSteps: 100}, f.err
}

func (f *fakeAnalyticsRepo) Timeseries(_ context.Context, _ domain.AnalyticsFilter) ([]domain.AnalyticsPoint, error) {
	f.calls++
	return []domain.AnalyticsPoint{{Date: "2026-07-21", Steps: 100}}, f.err
}

func (f *fakeAnalyticsRepo) ByFaculty(_ context.Context, _ domain.AnalyticsFilter) ([]domain.FacultyStat, error) {
	f.calls++
	return []domain.FacultyStat{{Name: "Iqtisodiyot"}}, f.err
}

func (f *fakeAnalyticsRepo) StreamUserActivity(_ context.Context, _ domain.AnalyticsFilter, fn func(domain.UserActivityRow) error) error {
	if f.err != nil {
		return f.err
	}
	return fn(domain.UserActivityRow{FullName: "Test", TotalSteps: 10})
}

// Davr → sana oralig'i MAHALLIY mintaqada hisoblanishi kerak.
//
// NEGA: admin "hafta" deganda o'zining kalendar haftasini kutadi. UTC bilan
// hisoblansa O'zbekistonda tunda hisobot bir kun orqada qolardi — bu qadam
// sinxronidagi tuzatilgan xatoning aynan o'zi (qarang: activity_service_test).
func TestAnalyticsService_FilterUsesLocalTimezone(t *testing.T) {
	east := NewAnalyticsService(&fakeAnalyticsRepo{}, time.FixedZone("EAST", 12*3600))
	west := NewAnalyticsService(&fakeAnalyticsRepo{}, time.FixedZone("WEST", -12*3600))

	fe, err := east.Filter("week", nil)
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	fw, err := west.Filter("week", nil)
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}

	if fe.To.Format("2006-01-02") == fw.To.Format("2006-01-02") {
		t.Error("mintaqa e'tiborga olinmadi: ikkala mintaqada bir xil sana")
	}
}

// Davr oraliqlari to'g'ri kun sonini qamrashi kerak.
func TestAnalyticsService_FilterPeriods(t *testing.T) {
	svc := NewAnalyticsService(&fakeAnalyticsRepo{}, time.FixedZone("TEST", 5*3600))

	cases := []struct {
		period   string
		wantDays int
	}{
		{"week", 7},   // bugun + oldingi 6 kun
		{"month", 30}, // bugun + oldingi 29 kun
		{"all", maxRangeDays},
	}
	for _, c := range cases {
		f, err := svc.Filter(c.period, nil)
		if err != nil {
			t.Fatalf("%s: %v", c.period, err)
		}
		days := int(f.To.Sub(f.From).Hours()/24) + 1
		if days != c.wantDays {
			t.Errorf("%s: %d kun, kutilgan %d", c.period, days, c.wantDays)
		}
	}
}

// "all" ham cheklangan bo'lishi kerak: cheksiz oraliq generate_series'da
// millionlab qator yasab serverni cho'ktirishi mumkin (§17.3 #39).
func TestAnalyticsService_AllPeriodIsBounded(t *testing.T) {
	svc := NewAnalyticsService(&fakeAnalyticsRepo{}, time.UTC)

	f, err := svc.Filter("all", nil)
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	days := int(f.To.Sub(f.From).Hours()/24) + 1
	if days > maxRangeDays {
		t.Errorf("oraliq cheklanmagan: %d kun", days)
	}
}

// Noto'g'ri davr nomi rad etilsin (SQL'ga tushmasin).
func TestAnalyticsService_RejectsInvalidPeriod(t *testing.T) {
	svc := NewAnalyticsService(&fakeAnalyticsRepo{}, time.UTC)

	if _, err := svc.Filter("'; DROP TABLE users; --", nil); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ErrValidation kutilgandi, olingan: %v", err)
	}
	if _, err := svc.Filter("year", nil); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ErrValidation kutilgandi, olingan: %v", err)
	}
}

// Fakultet filtri repository'ga o'zgarishsiz yetib borishi kerak.
func TestAnalyticsService_PassesFacultyFilter(t *testing.T) {
	repo := &fakeAnalyticsRepo{}
	svc := NewAnalyticsService(repo, time.UTC)

	id := uuid.New()
	f, err := svc.Filter("week", &id)
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if _, err := svc.Get(context.Background(), f); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if repo.gotFilter.FacultyID == nil || *repo.gotFilter.FacultyID != id {
		t.Error("fakultet filtri repository'ga yetib bormadi")
	}
}

// Uchala so'rov ham bajarilishi kerak (parallel — errgroup).
func TestAnalyticsService_GetRunsAllThree(t *testing.T) {
	repo := &fakeAnalyticsRepo{}
	svc := NewAnalyticsService(repo, time.UTC)

	f, _ := svc.Filter("week", nil)
	res, err := svc.Get(context.Background(), f)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if repo.calls != 3 {
		t.Errorf("3 ta so'rov kutilgandi, olingan %d", repo.calls)
	}
	if res.Overview.TotalSteps != 100 || len(res.Timeseries) != 1 || len(res.Faculties) != 1 {
		t.Error("natija to'liq yig'ilmadi")
	}
	if res.From == "" || res.To == "" {
		t.Error("davr chegaralari javobda yo'q")
	}
}

// Bitta so'rov yiqilsa butun natija xato qaytarishi kerak (chala
// ma'lumot ko'rsatilmasin).
func TestAnalyticsService_GetFailsOnRepoError(t *testing.T) {
	repo := &fakeAnalyticsRepo{err: errors.New("db yiqildi")}
	svc := NewAnalyticsService(repo, time.UTC)

	f, _ := svc.Filter("week", nil)
	if _, err := svc.Get(context.Background(), f); err == nil {
		t.Error("xato kutilgandi")
	}
}
