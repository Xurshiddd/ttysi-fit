package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Sessiya bekor qilinish sabablari.
const (
	RevokeNewDevice = "new_device" // boshqa qurilmada kirildi
	RevokeUser      = "user"       // foydalanuvchi o'zi o'chirdi
	RevokeAdmin     = "admin"
	RevokeLogout    = "logout"
)

// ErrDeviceConflict — boshqa qurilmada faol sessiya bor.
//
// Login shu xato bilan to'xtaydi va mijozga qaysi qurilma ekanligi
// aytiladi. Foydalanuvchi rozi bo'lsa `force_device: true` bilan qayta
// yuboradi — o'shanda eski sessiya bekor qilinadi.
var ErrDeviceConflict = errors.New("boshqa qurilmada faol sessiya bor")

// UserSession — bitta qurilmadagi sessiya.
type UserSession struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`

	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`

	DeviceID   string `json:"device_id" gorm:"not null"`
	DeviceName string `json:"device_name,omitempty"`
	Platform   string `json:"platform,omitempty"`
	AppVersion string `json:"app_version,omitempty"`

	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"-"` // javobga chiqmaydi: uzun va foydasiz

	LastSeenAt time.Time `json:"last_seen_at"`

	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	RevokedReason string     `json:"revoked_reason,omitempty"`
}

func (UserSession) TableName() string { return "user_sessions" }

// Active — sessiya hali faolmi.
func (s UserSession) Active() bool { return s.RevokedAt == nil }

// DeviceInfo — mijoz login paytida yuboradigan qurilma ma'lumoti.
type DeviceInfo struct {
	DeviceID   string
	DeviceName string
	Platform   string
	AppVersion string
	IP         string
	UserAgent  string
}

// SessionRepository — qurilma sessiyalari uchun port.
type SessionRepository interface {
	// ActiveOther — shu foydalanuvchining BOSHQA qurilmadagi faol sessiyasi.
	// Yo'q bo'lsa (nil, nil).
	ActiveOther(ctx context.Context, userID uuid.UUID, deviceID string) (*UserSession, error)
	// ListActive — "Mening qurilmalarim".
	ListActive(ctx context.Context, userID uuid.UUID) ([]UserSession, error)
	// Upsert — qurilma uchun sessiya yaratadi yoki mavjudini yangilaydi.
	Upsert(ctx context.Context, userID uuid.UUID, d DeviceInfo) (*UserSession, error)
	// RevokeOthers — shu qurilmadan boshqa barcha faol sessiyalarni yopadi.
	RevokeOthers(ctx context.Context, userID uuid.UUID, deviceID, reason string) (int64, error)
	// Revoke — bitta sessiyani yopadi (egalik tekshiriladi).
	Revoke(ctx context.Context, userID, sessionID uuid.UUID, reason string) error
	// Touch — oxirgi faollik vaqtini yangilaydi.
	Touch(ctx context.Context, userID uuid.UUID, deviceID string) error
}
