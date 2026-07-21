package domain

import (
	"context"

	"github.com/google/uuid"
)

// RatingType — reyting kesimi.
type RatingType string

const (
	RatingStudent  RatingType = "student"  // talabalar (shaxsiy)
	RatingEmployee RatingType = "employee" // xodimlar (shaxsiy)
	RatingGroup    RatingType = "group"    // guruhlar (jon boshiga o'rtacha)
	RatingFaculty  RatingType = "faculty"  // fakultetlar (jon boshiga o'rtacha)
)

// ValidRatingType — ruxsat etilgan kesim tekshiruvi.
func ValidRatingType(t string) bool {
	switch RatingType(t) {
	case RatingStudent, RatingEmployee, RatingGroup, RatingFaculty:
		return true
	}
	return false
}

// RatingPeriod — reyting davri.
type RatingPeriod string

const (
	PeriodWeek  RatingPeriod = "week"  // oxirgi 7 kun
	PeriodMonth RatingPeriod = "month" // oxirgi 30 kun
	PeriodAll   RatingPeriod = "all"   // butun tarix
)

// ValidRatingPeriod — ruxsat etilgan davr tekshiruvi.
func ValidRatingPeriod(p string) bool {
	switch RatingPeriod(p) {
	case PeriodWeek, PeriodMonth, PeriodAll:
		return true
	}
	return false
}

// RatingEntry — reyting jadvalidagi bitta qator (shaxs, guruh yoki fakultet).
type RatingEntry struct {
	Rank int64     `json:"rank"`
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	// Shaxsiy reyting uchun
	AvatarURL   string `json:"avatar_url,omitempty"`
	FacultyName string `json:"faculty_name,omitempty"`
	GroupName   string `json:"group_name,omitempty"`

	// Guruh/fakultet reytingi uchun
	MemberCount int     `json:"member_count,omitempty"`
	AvgSteps    float64 `json:"avg_steps,omitempty"` // jon boshiga o'rtacha qadam

	TotalSteps     int64   `json:"total_steps"`
	TotalDistanceM float64 `json:"total_distance_m"`
	TotalCalories  float64 `json:"total_calories"`
	ActiveDays     int     `json:"active_days,omitempty"` // faqat shaxsiy
}

// MyRating — foydalanuvchining o'z o'rni (mobil bosh sahifa uchun).
type MyRating struct {
	GlobalRank  int64 `json:"global_rank"`
	FacultyRank int64 `json:"faculty_rank,omitempty"`
	TotalUsers  int64 `json:"total_users"`
	TotalSteps  int64 `json:"total_steps"`
}

// RatingFilter — reyting so'rovi parametrlari.
type RatingFilter struct {
	Type      RatingType
	Period    RatingPeriod
	FacultyID *uuid.UUID // student/employee/group kesimlarini fakultetga toraytirish
	GroupID   *uuid.UUID // student kesimini guruhga toraytirish
	Page      int
	Limit     int
}

// RatingRepository — reyting so'rovlari uchun port (interfeys).
type RatingRepository interface {
	// ListIndividual — talaba/xodim shaxsiy reytingi (jami qadam bo'yicha).
	ListIndividual(ctx context.Context, f RatingFilter) ([]RatingEntry, int64, error)
	// ListGroups — guruhlar reytingi (jon boshiga o'rtacha qadam bo'yicha).
	ListGroups(ctx context.Context, f RatingFilter) ([]RatingEntry, int64, error)
	// ListFaculties — fakultetlar reytingi (jon boshiga o'rtacha qadam bo'yicha).
	ListFaculties(ctx context.Context, f RatingFilter) ([]RatingEntry, int64, error)
	// MyRank — foydalanuvchining umumiy va fakultet ichidagi o'rni.
	MyRank(ctx context.Context, userID uuid.UUID, period RatingPeriod) (*MyRating, error)
}
