package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Activity — foydalanuvchining bir kunlik faolligi (qadam, kaloriya, masofa).
// Bir kun — bir yozuv (user_id + activity_date noyob), upsert qilinadi.
type Activity struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	UserID       uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	ActivityDate time.Time `json:"activity_date" gorm:"type:date;not null"`

	Steps     int     `json:"steps" gorm:"not null"`
	Calories  float64 `json:"calories" gorm:"not null"`
	DistanceM float64 `json:"distance_m" gorm:"not null"`
	ActiveMin int     `json:"active_min" gorm:"not null"`
	Source    string  `json:"source,omitempty"`
}

func (Activity) TableName() string { return "activities" }

// ActivityStats — profil/bosh sahifa uchun yig'ma statistika.
type ActivityStats struct {
	TodaySteps     int     `json:"today_steps"`
	TodayCalories  float64 `json:"today_calories"`
	TodayDistanceM float64 `json:"today_distance_m"`
	TodayActiveMin int     `json:"today_active_min"`
	WeekSteps      int     `json:"week_steps"`
	MonthSteps     int     `json:"month_steps"`
	TotalSteps     int     `json:"total_steps"`
}

// ActivityRepository — faollik ma'lumotlari uchun port (interfeys).
type ActivityRepository interface {
	// Upsert — bir kunlik yozuvni ko'tarib-yangilaydi (user_id+activity_date bo'yicha).
	Upsert(ctx context.Context, a *Activity) error
	// UpsertMany — bir necha kunlik yozuvni BITTA so'rovda ko'tarib-yangilaydi.
	UpsertMany(ctx context.Context, rows []Activity) error
	// ListByUser — [from, to] oralig'idagi yozuvlar (sana kamayish tartibida).
	ListByUser(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]Activity, error)
	// Stats — bugun/hafta/oy/jami yig'ma (bitta so'rovda).
	// today — APP_TIMEZONE dagi bugungi sana (server soati emas).
	Stats(ctx context.Context, userID uuid.UUID, today time.Time) (*ActivityStats, error)
	// DeleteRange — [from, to] oralig'idagi yozuvlarni O'CHIRADI (admin).
	// O'chirilgan qatorlar sonini qaytaradi.
	DeleteRange(ctx context.Context, userID uuid.UUID, from, to time.Time) (int64, error)
}
