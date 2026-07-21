package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Bildirishnoma turlari. Mobil ilova shu bo'yicha ikonka/rang tanlaydi,
// shuning uchun ro'yxat kodda (kontent emas, ko'rinish qoidasi).
const (
	NotifyAchievement  = "achievement"   // yangi yutuq
	NotifyChallenge    = "challenge"     // chellenj bajarildi / tugayapti
	NotifyCompetition  = "competition"   // musobaqa eslatmasi
	NotifyRewardIssued = "reward_issued" // sovg'a topshirishga tayyor/berildi
	NotifyRewardCancel = "reward_cancel" // buyurtma bekor qilindi, coin qaytdi
	NotifyCoins        = "coins"         // FIT Coin harakati
	NotifyAnnouncement = "announcement"  // admin e'loni
)

// ValidNotificationType — tur ro'yxatda bormi.
func ValidNotificationType(t string) bool {
	switch t {
	case NotifyAchievement, NotifyChallenge, NotifyCompetition,
		NotifyRewardIssued, NotifyRewardCancel, NotifyCoins, NotifyAnnouncement:
		return true
	}
	return false
}

// Notification — bitta foydalanuvchiga bitta xabar.
type Notification struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`

	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Type   string    `json:"type" gorm:"not null"`

	Title string `json:"title" gorm:"not null"`
	Body  string `json:"body,omitempty"`

	RefType string     `json:"ref_type,omitempty"`
	RefID   *uuid.UUID `json:"ref_id,omitempty" gorm:"type:uuid"`

	// ReadAt — nil bo'lsa o'qilmagan.
	ReadAt *time.Time `json:"read_at,omitempty"`

	Metadata datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (Notification) TableName() string { return "notifications" }

// NotificationFilter — ro'yxat so'rovi.
type NotificationFilter struct {
	// UnreadOnly — faqat o'qilmaganlar.
	UnreadOnly bool
	Page       int
	Limit      int
}

// AnnouncementTarget — admin e'loni kimga boradi.
//
// nil maydonlar "cheklov yo'q" degani: uchalasi ham nil bo'lsa e'lon
// barcha aktiv foydalanuvchilarga yuboriladi.
type AnnouncementTarget struct {
	FacultyID *uuid.UUID
	GroupID   *uuid.UUID
	// Role — "student" | "employee". Bo'sh bo'lsa ikkalasi ham.
	Role string
}

// NotificationRepository — bildirishnomalar uchun port.
type NotificationRepository interface {
	// List — foydalanuvchining xabarlari (yangi -> eski).
	List(ctx context.Context, userID uuid.UUID, f NotificationFilter) ([]Notification, int64, error)
	// UnreadCount — o'qilmaganlar soni (qo'ng'iroq nishoni).
	UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)
	// MarkRead — bitta xabarni o'qilgan deb belgilaydi (egalik tekshiriladi).
	MarkRead(ctx context.Context, userID, id uuid.UUID) error
	// MarkAllRead — barchasini o'qilgan deb belgilaydi.
	MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error)

	// Create — bitta xabar. ref_id berilgan bo'lsa IDEMPOTENT: o'sha manba
	// uchun ikkinchi xabar qo'shilmaydi va ErrAlreadyExists qaytadi.
	Create(ctx context.Context, n *Notification) error

	// Broadcast — target bo'yicha topilgan har bir foydalanuvchiga xabar
	// yozadi va yozilgan qatorlar sonini qaytaradi.
	//
	// Sikl ichida INSERT qilinmaydi (§3.1): bitta
	// `INSERT ... SELECT` so'rovi bilan bajariladi.
	Broadcast(ctx context.Context, t AnnouncementTarget, n Notification) (int64, error)
}
