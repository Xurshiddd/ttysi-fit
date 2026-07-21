package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// fakeAnalyticsRepo — DB'siz sinov uchun. Har metod nechta marta va qaysi
// filtr bilan chaqirilganini yozib boradi.
//
// MUTEX SHART: `Get` uchala so'rovni errgroup bilan PARALLEL bajaradi
// (§11.2), ya'ni bu maydonlarga uchta goroutine bir vaqtda yozadi.
// Himoyasiz `go test -race` ni yiqitardi.
type fakeAnalyticsRepo struct {
	mu        sync.Mutex
	gotFilter domain.AnalyticsFilter
	calls     int
	err       error

	// Eksport semaforini sinash uchun: repo "ishlab turgan" holatda
	// ushlab turiladi.
	streamStarted chan struct{}
	streamRelease chan struct{}
}

func newStreamRepo() *fakeAnalyticsRepo {
	return &fakeAnalyticsRepo{
		streamStarted: make(chan struct{}, 8),
		streamRelease: make(chan struct{}),
	}
}

func (f *fakeAnalyticsRepo) Overview(_ context.Context, flt domain.AnalyticsFilter) (*domain.AnalyticsOverview, error) {
	f.mu.Lock()
	f.calls++
	f.gotFilter = flt
	f.mu.Unlock()
	return &domain.AnalyticsOverview{TotalSteps: 100}, f.err
}

func (f *fakeAnalyticsRepo) Timeseries(_ context.Context, _ domain.AnalyticsFilter) ([]domain.AnalyticsPoint, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	return []domain.AnalyticsPoint{{Date: "2026-07-21", Steps: 100}}, f.err
}

func (f *fakeAnalyticsRepo) ByFaculty(_ context.Context, _ domain.AnalyticsFilter) ([]domain.FacultyStat, error) {
	f.mu.Lock()
	f.calls++
	f.mu.Unlock()
	return []domain.FacultyStat{{Name: "Iqtisodiyot"}}, f.err
}

func (f *fakeAnalyticsRepo) StreamUserActivity(_ context.Context, _ domain.AnalyticsFilter, fn func(domain.UserActivityRow) error) error {
	if f.err != nil {
		return f.err
	}
	f.streamStarted <- struct{}{} // "boshladim"
	<-f.streamRelease             // testdan "tugat" signalini kutamiz
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
	repo.mu.Lock()
	got := repo.gotFilter.FacultyID
	repo.mu.Unlock()
	if got == nil || *got != id {
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
	repo.mu.Lock()
	calls := repo.calls
	repo.mu.Unlock()
	if calls != 3 {
		t.Errorf("3 ta so'rov kutilgandi, olingan %d", calls)
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

// Parallel eksport CHEKLANGAN bo'lishi kerak.
//
// NEGA: eksport uzoq davom etadi va DB ulanishini band qiladi. Pool 25 ta
// ulanishdan iborat — cheklovsiz bir necha admin bir vaqtda eksport bossa
// butun ilova javob bermay qolardi (§17.3 #39).
func TestAnalyticsService_LimitsConcurrentExports(t *testing.T) {
	repo := newStreamRepo()
	svc := NewAnalyticsService(repo, time.UTC)
	f, _ := svc.Filter("week", nil)

	noop := func(domain.UserActivityRow) error { return nil }

	// Barcha joyni band qilamiz va ushlab turamiz.
	var wg sync.WaitGroup
	for i := 0; i < maxConcurrentExports; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = svc.StreamUserActivity(context.Background(), f, noop)
		}()
		<-repo.streamStarted // shu goroutine haqiqatan slot olganiga ishonch
	}

	// Navbatdagisi KUTMASDAN ErrBusy qaytarishi kerak: admin "yuklanmoqda"
	// holatida osilib qolmasin.
	if err := svc.StreamUserActivity(context.Background(), f, noop); !errors.Is(err, domain.ErrBusy) {
		close(repo.streamRelease)
		t.Fatalf("ErrBusy kutilgandi, olingan: %v", err)
	}

	// Hammasini bir yo'la bo'shatamiz va tugashini kutamiz.
	close(repo.streamRelease)
	wg.Wait()

	// Slot qaytarilgan bo'lishi kerak.
	if err := svc.StreamUserActivity(context.Background(), f, noop); err != nil {
		t.Errorf("eksport tugagach joy bo'shamadi: %v", err)
	}
}
