package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Role — foydalanuvchi roli.
type Role string

const (
	RoleStudent Role = "student"
	// RoleEmployee — barcha xodimlar (o'qituvchilar va o'qituvchi bo'lmaganlar).
	RoleEmployee Role = "employee"
	// RoleTeacher — o'qituvchi (xodimlar ichidan lavozim bo'yicha ajratilishi mumkin).
	RoleTeacher Role = "teacher"
	RoleAdmin   Role = "admin"
)

// User — platformaning barcha foydalanuvchilari (role bilan ajratiladi).
type User struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	FullName string `json:"full_name" gorm:"not null"`
	Email    string `json:"email" gorm:"not null"`
	Phone    string `json:"phone,omitempty"`
	Password string `json:"-"` // sinxron foydalanuvchilarda bo'sh bo'lishi mumkin
	Role     Role   `json:"role" gorm:"not null;default:student"`

	// HEMIS sync
	HemisID    *int64 `json:"hemis_id,omitempty" gorm:"uniqueIndex"`
	HemisLogin string `json:"hemis_login,omitempty"` // student_id_number / employee_id_number

	// Tashkiliy bog'lanish (FK lar)
	FacultyID    *uuid.UUID `json:"faculty_id,omitempty" gorm:"type:uuid"`    // talaba/o'qituvchi fakulteti
	DepartmentID *uuid.UUID `json:"department_id,omitempty" gorm:"type:uuid"` // o'qituvchi kafedrasi
	GroupID      *uuid.UUID `json:"group_id,omitempty" gorm:"type:uuid"`      // talaba guruhi
	Course       *int16     `json:"course,omitempty"`

	// Relink uchun HEMIS id lari
	FacultyHemisID    *int64 `json:"-"`
	DepartmentHemisID *int64 `json:"-"`
	GroupHemisID      *int64 `json:"-"`

	// Profil
	Gender    string     `json:"gender,omitempty" gorm:"type:varchar(10)"`
	BirthDate *time.Time `json:"birth_date,omitempty" gorm:"type:date"`
	Position  string     `json:"position,omitempty"`  // xodim lavozimi
	Specialty string     `json:"specialty,omitempty"` // mutaxassislik
	AvatarURL string     `json:"avatar_url,omitempty"`
	Bio       string     `json:"bio,omitempty"`
	IsActive  bool       `json:"is_active" gorm:"not null"`

	// Language — foydalanuvchi tili (server-initiated xabarlar uchun: email, SMS, push).
	Language string `json:"language" gorm:"type:varchar(5);default:'uz'"`
}

func (User) TableName() string { return "users" }

// UserFilter — admin foydalanuvchilar ro'yxati uchun filtr.
type UserFilter struct {
	Role      string
	FacultyID *uuid.UUID
	GroupID   *uuid.UUID
	Search    string // full_name / email / hemis_login bo'yicha
	Page      int
	Limit     int
}

// UserListItem — admin ro'yxati uchun yengil read-model (fakultet/guruh nomi bilan).
type UserListItem struct {
	ID          uuid.UUID `json:"id"`
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	Gender      string    `json:"gender"`
	Course      *int16    `json:"course"`
	IsActive    bool      `json:"is_active"`
	AvatarURL   string    `json:"avatar_url"`
	HemisLogin  string    `json:"hemis_login"`
	Position    string    `json:"position"`
	FacultyName string    `json:"faculty_name"`
	GroupName   string    `json:"group_name"`
}

// UserProfile — foydalanuvchining o'z profili uchun read-model.
// Fakultet/kafedra/guruh nomlari JOIN bilan keladi (N+1 yo'q, §3.1).
// Parol va ichki HEMIS id lari bu yerda yo'q (§17.3 #37 — excessive data exposure).
type UserProfile struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      Role      `json:"role"`
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	Language  string    `json:"language"`

	Gender     string     `json:"gender"`
	BirthDate  *time.Time `json:"birth_date"`
	Course     *int16     `json:"course"`
	Position   string     `json:"position"`
	Specialty  string     `json:"specialty"`
	HemisLogin string     `json:"hemis_login"`

	FacultyName    string `json:"faculty_name"`
	DepartmentName string `json:"department_name"`
	GroupName      string `json:"group_name"`
}

// ProfileUpdate — foydalanuvchi o'zi o'zgartira oladigan maydonlar (nil — tegilmaydi).
//
// MUHIM: bu ro'yxat ataylab qisqa. full_name, email, role, gender, birth_date,
// course, position, specialty, avatar_url — bularning hammasini HEMIS sync
// har safar qayta yozadi (UserRepository.UpsertRoster dagi DoUpdates ro'yxati).
// Ularni tahrirlashga ruxsat berilsa, o'zgarish keyingi syncda jimgina yo'qolardi.
// Shu bilan birga bu DTO whitelisting — mass assignment himoyasi (§17.3 #13).
type ProfileUpdate struct {
	Phone    *string
	Bio      *string
	Language *string
}

// UserRepository — foydalanuvchi ma'lumotlari uchun port (interfeys).
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	// GetProfile — o'z profili (fakultet/kafedra/guruh nomi bilan).
	GetProfile(ctx context.Context, id uuid.UUID) (*UserProfile, error)
	// UpdateProfile — faqat ProfileUpdate dagi maydonlarni yangilaydi.
	UpdateProfile(ctx context.Context, id uuid.UUID, p ProfileUpdate) (*UserProfile, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	// GetByHemis — hemis_id yoki hemis_login bo'yicha (HEMIS OAuth uchun).
	GetByHemis(ctx context.Context, hemisID *int64, hemisLogin string) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	// List — admin uchun paginatsiyali, filtrlangan ro'yxat + umumiy son.
	List(ctx context.Context, f UserFilter) ([]UserListItem, int64, error)
	// UpsertRoster — HEMIS roster (talaba/o'qituvchi) ni hemis_id bo'yicha ko'tarib-yangilaydi.
	UpsertRoster(ctx context.Context, users []User) (*SyncStats, error)
	// RelinkRoster — faculty_hemis_id/department_hemis_id/group_hemis_id ni FK larga bog'laydi.
	RelinkRoster(ctx context.Context) error
}
