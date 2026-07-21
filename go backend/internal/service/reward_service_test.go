package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

type fakeRewardRepo struct {
	created *domain.Reward
	filter  domain.RedemptionFilter
	err     error
}

func (f *fakeRewardRepo) List(context.Context, domain.RewardFilter) ([]domain.Reward, int64, error) {
	return nil, 0, f.err
}
func (f *fakeRewardRepo) GetByID(context.Context, uuid.UUID) (*domain.Reward, error) {
	return &domain.Reward{}, f.err
}
func (f *fakeRewardRepo) Create(_ context.Context, r *domain.Reward) error {
	f.created = r
	return f.err
}
func (f *fakeRewardRepo) Update(_ context.Context, r *domain.Reward) error {
	f.created = r
	return f.err
}
func (f *fakeRewardRepo) Delete(context.Context, uuid.UUID) error { return f.err }
func (f *fakeRewardRepo) Redeem(context.Context, uuid.UUID, uuid.UUID) (*domain.RewardRedemption, error) {
	return &domain.RewardRedemption{Code: "TESTCODE"}, f.err
}
func (f *fakeRewardRepo) ListRedemptions(_ context.Context, flt domain.RedemptionFilter) ([]domain.RedemptionDetail, int64, error) {
	f.filter = flt
	return nil, 0, f.err
}
func (f *fakeRewardRepo) MarkIssued(context.Context, uuid.UUID, uuid.UUID, string) (*domain.RewardRedemption, error) {
	return &domain.RewardRedemption{}, f.err
}
func (f *fakeRewardRepo) Cancel(context.Context, uuid.UUID, uuid.UUID, string) (*domain.RewardRedemption, error) {
	return &domain.RewardRedemption{}, f.err
}

func validReward() *domain.Reward {
	return &domain.Reward{
		Title:     "Futbolka",
		Category:  domain.RewardCategoryMerch,
		CostCoins: 50,
	}
}

// config ustuni NOT NULL: GORM nil JSON ni ANIQ NULL qilib yuboradi va
// DB default'i ('{}') ishlamaydi. Bu haqiqiy 500 xatosi edi.
func TestRewardService_NormalizesConfig(t *testing.T) {
	repo := &fakeRewardRepo{}
	svc := NewRewardService(repo)

	if err := svc.Create(context.Background(), validReward()); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(repo.created.Config) == 0 {
		t.Error("config bo'sh qoldi — DB NOT NULL cheklovini buzadi")
	}
	if string(repo.created.Config) != "{}" {
		t.Errorf("config: kutilgan {}, olingan %s", repo.created.Config)
	}
}

func TestRewardService_ValidationRules(t *testing.T) {
	svc := NewRewardService(&fakeRewardRepo{})

	cases := []struct {
		name   string
		mutate func(*domain.Reward)
	}{
		{"nomsiz", func(r *domain.Reward) { r.Title = "" }},
		{"narx nol", func(r *domain.Reward) { r.CostCoins = 0 }},
		{"narx manfiy", func(r *domain.Reward) { r.CostCoins = -10 }},
		{"noma'lum kategoriya", func(r *domain.Reward) { r.Category = "yolgon" }},
		{"manfiy miqdor", func(r *domain.Reward) { n := -1; r.Stock = &n }},
		{"nol limit", func(r *domain.Reward) { n := 0; r.PerUserLimit = &n }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := validReward()
			c.mutate(r)
			if err := svc.Create(context.Background(), r); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("ErrValidation kutilgandi, olingan: %v", err)
			}
		})
	}
}

// Teskari vaqt oynasi — sovg'a hech qachon ko'rinmaydi. Bu odatda admin
// xatosi, jimgina qabul qilmaymiz.
func TestRewardService_RejectsInvertedWindow(t *testing.T) {
	svc := NewRewardService(&fakeRewardRepo{})

	r := validReward()
	start := time.Now().AddDate(0, 0, 10)
	end := time.Now()
	r.StartsAt, r.EndsAt = &start, &end

	if err := svc.Create(context.Background(), r); !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ErrValidation kutilgandi, olingan: %v", err)
	}
}

// IDOR (§17.3 #26): foydalanuvchi boshqa odamning buyurtmalarini ko'rmasin.
// MyRedemptions filtrni MAJBURAN o'zinikiga qo'yishi kerak.
func TestRewardService_MyRedemptionsForcesOwnership(t *testing.T) {
	repo := &fakeRewardRepo{}
	svc := NewRewardService(repo)

	me := uuid.New()
	someoneElse := uuid.New()

	// Mijoz boshqa odamning ID sini "yuborishga" urinsa ham.
	_, _, err := svc.MyRedemptions(context.Background(), me,
		domain.RedemptionFilter{UserID: &someoneElse})
	if err != nil {
		t.Fatalf("MyRedemptions: %v", err)
	}
	if repo.filter.UserID == nil || *repo.filter.UserID != me {
		t.Error("filtr o'zgartirilmadi — boshqa odamning buyurtmalari ochilardi")
	}
}

func TestRewardService_RejectsUnknownStatusFilter(t *testing.T) {
	svc := NewRewardService(&fakeRewardRepo{})

	_, _, err := svc.ListRedemptions(context.Background(),
		domain.RedemptionFilter{Status: "'; DROP TABLE rewards; --"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ErrValidation kutilgandi, olingan: %v", err)
	}
}

func TestRewardService_RejectsUnknownCategoryFilter(t *testing.T) {
	svc := NewRewardService(&fakeRewardRepo{})

	_, _, err := svc.List(context.Background(), domain.RewardFilter{Category: "yolgon"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("ErrValidation kutilgandi, olingan: %v", err)
	}
}

// Reward.Available — do'konda nima ko'rinishini hal qiladi.
func TestReward_Available(t *testing.T) {
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	past := now.AddDate(0, 0, -1)
	future := now.AddDate(0, 0, 1)
	zero, some := 0, 5

	cases := []struct {
		name string
		r    domain.Reward
		want bool
	}{
		{"oddiy aktiv", domain.Reward{IsActive: true}, true},
		{"nofaol", domain.Reward{IsActive: false}, false},
		{"miqdor tugagan", domain.Reward{IsActive: true, Stock: &zero}, false},
		{"miqdor bor", domain.Reward{IsActive: true, Stock: &some}, true},
		{"hali boshlanmagan", domain.Reward{IsActive: true, StartsAt: &future}, false},
		{"tugagan", domain.Reward{IsActive: true, EndsAt: &past}, false},
		{"oyna ichida", domain.Reward{IsActive: true, StartsAt: &past, EndsAt: &future}, true},
		{"o'chirilgan", domain.Reward{IsActive: true, DeletedAt: &past}, false},
	}
	for _, c := range cases {
		if got := c.r.Available(now); got != c.want {
			t.Errorf("%s: kutilgan %v, olingan %v", c.name, c.want, got)
		}
	}
}
