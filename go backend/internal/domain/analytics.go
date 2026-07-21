package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AnalyticsFilter — analitika so'rovining chegaralari.
//
// From/To — MAHALLIY mintaqadagi kun chegaralari (service davr nomidan
// hisoblaydi). CURRENT_DATE o'rniga aniq sana beriladi: shunda hisobot
// qaysi kunlarni qamraganini test ham, admin ham aniq biladi.
type AnalyticsFilter struct {
	From      time.Time
	To        time.Time
	FacultyID *uuid.UUID // nil — butun universitet
}

// AnalyticsOverview — yuqoridagi umumiy raqamlar.
type AnalyticsOverview struct {
	TotalSteps      int64   `json:"total_steps"`
	TotalDistanceKm float64 `json:"total_distance_km"`
	// ActiveUsers — davr ichida kamida bitta faollik yozuvi bo'lgan foydalanuvchi.
	ActiveUsers int64 `json:"active_users"`
	// TotalUsers — umuman ro'yxatdagi aktiv foydalanuvchilar (faollik yozmaganlar ham).
	TotalUsers int64 `json:"total_users"`
	// AvgStepsPerActive — faol foydalanuvchi boshiga o'rtacha qadam.
	AvgStepsPerActive int64 `json:"avg_steps_per_active"`
}

// AnalyticsPoint — kunlik dinamika grafigining bitta nuqtasi.
type AnalyticsPoint struct {
	Date        string `json:"date"` // YYYY-MM-DD
	Steps       int64  `json:"steps"`
	ActiveUsers int64  `json:"active_users"`
}

// FacultyStat — fakultetlar kesimi.
//
// AvgSteps (jon boshiga) asosiy taqqoslash mezoni: jami qadam bo'yicha
// saralansa katta fakultet doim yutib chiqadi va taqqoslash ma'nosini
// yo'qotadi (LOYIHA_HOLATI: reyting ham jon boshiga o'rtacha bo'yicha).
type FacultyStat struct {
	FacultyID   uuid.UUID `json:"faculty_id"`
	Name        string    `json:"name"`
	TotalSteps  int64     `json:"total_steps"`
	UserCount   int64     `json:"user_count"`
	ActiveUsers int64     `json:"active_users"`
	AvgSteps    int64     `json:"avg_steps"`
}

// Analytics — dashboard uchun to'liq to'plam (bitta so'rovda qaytadi).
type Analytics struct {
	From       string             `json:"from"`
	To         string             `json:"to"`
	Overview   AnalyticsOverview  `json:"overview"`
	Timeseries []AnalyticsPoint   `json:"timeseries"`
	Faculties  []FacultyStat      `json:"faculties"`
}

// UserActivityRow — eksport (CSV) uchun bitta foydalanuvchi qatori.
type UserActivityRow struct {
	FullName    string
	Email       string
	Role        string
	Faculty     string
	Department  string
	GroupName   string
	TotalSteps  int64
	DistanceKm  float64
	ActiveDays  int64
}

// AnalyticsRepository — analitika o'qish portlari.
type AnalyticsRepository interface {
	Overview(ctx context.Context, f AnalyticsFilter) (*AnalyticsOverview, error)
	// Timeseries — kunlik dinamika. Faollik bo'lmagan kunlar ham 0 bilan
	// qaytadi (grafikda uzilish bo'lmasin).
	Timeseries(ctx context.Context, f AnalyticsFilter) ([]AnalyticsPoint, error)
	ByFaculty(ctx context.Context, f AnalyticsFilter) ([]FacultyStat, error)
	// StreamUserActivity — eksport uchun qatorlarni BIRMA-BIR beradi.
	//
	// Slice qaytarmaydi: hisobotda o'n minglab foydalanuvchi bo'lishi mumkin,
	// hammasini xotiraga yig'ish serverni cho'ktiradi (§17.3 #39). Callback
	// xato qaytarsa iteratsiya to'xtaydi.
	StreamUserActivity(ctx context.Context, f AnalyticsFilter, fn func(UserActivityRow) error) error
}
