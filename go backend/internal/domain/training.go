package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Mashg'ulot darajasi — chegaralangan shkala (kontent emas), shuning uchun enum.
const (
	TrainingBeginner     = "beginner"
	TrainingIntermediate = "intermediate"
	TrainingAdvanced     = "advanced"
)

// Mashg'ulot holati.
const (
	TrainingStatusDraft     = "draft"
	TrainingStatusPublished = "published"
)

func ValidTrainingLevel(l string) bool {
	switch l {
	case TrainingBeginner, TrainingIntermediate, TrainingAdvanced:
		return true
	}
	return false
}

func ValidTrainingStatus(s string) bool {
	switch s {
	case TrainingStatusDraft, TrainingStatusPublished:
		return true
	}
	return false
}

// Training — video mashg'ulot.
type Training struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description,omitempty"`

	// Category — erkin matn (§16: yangi kategoriya qo'shish uchun kod
	// o'zgartirish shart emas). Admin mavjudlaridan tanlaydi yoki yangisini yozadi.
	Category string `json:"category,omitempty"`
	Level    string `json:"level" gorm:"not null"`

	VideoURL     string `json:"video_url" gorm:"not null"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	DurationMin  *int16 `json:"duration_min,omitempty"`

	Status      string     `json:"status" gorm:"not null"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Views       int        `json:"views" gorm:"not null"`
	SortOrder   int        `json:"sort_order" gorm:"not null"`

	Metadata datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (Training) TableName() string { return "trainings" }

// TrainingFilter — ro'yxat so'rovi.
type TrainingFilter struct {
	Category string
	Level    string
	Search   string
	Status   string
	Page     int
	Limit    int
	// PublishedOnly — mobil ro'yxat.
	PublishedOnly bool
}

// TrainingRepository — mashg'ulotlar uchun port.
type TrainingRepository interface {
	Create(ctx context.Context, t *Training) error
	Update(ctx context.Context, t *Training) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Training, error)
	List(ctx context.Context, f TrainingFilter) ([]Training, int64, error)
	IncrementViews(ctx context.Context, id uuid.UUID) error
	// Categories — mavjud kategoriyalar (DISTINCT). Mobil filtr va admin
	// formasidagi tanlov ro'yxati shundan yasaladi — kodda ro'yxat yo'q.
	Categories(ctx context.Context, publishedOnly bool) ([]string, error)
}
