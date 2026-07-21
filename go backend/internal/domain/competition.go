package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Musobaqa turlari (§16.3). Yangi tur qo'shish: konstanta + registrga spec.
type CompetitionType string

const (
	CompetitionIndividual CompetitionType = "individual" // shaxsiy
	CompetitionTeam       CompetitionType = "team"       // jamoaviy
	CompetitionFacultyVs  CompetitionType = "faculty_vs" // fakultetlararo
	CompetitionCustom     CompetitionType = "custom"     // boshqa
)

// Musobaqa holati. Chellenjdan farqli: `registration` bosqichi bor.
const (
	CompStatusDraft        = "draft"
	CompStatusRegistration = "registration" // ro'yxatdan o'tish ochiq
	CompStatusOngoing      = "ongoing"      // ketmoqda
	CompStatusFinished     = "finished"
)

// Ro'yxatdan o'tish holati.
const (
	RegStatusRegistered = "registered"
	RegStatusCancelled  = "cancelled"
)

// Competition — umumiy maydonlar ustunda, turga xos parametrlar Config JSONB da.
type Competition struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	Type        CompetitionType `json:"type" gorm:"not null"`
	Title       string          `json:"title" gorm:"not null"`
	Description string          `json:"description,omitempty"`
	Scope       string          `json:"scope" gorm:"not null"`
	Status      string          `json:"status" gorm:"not null"`

	StartsAt  *time.Time `json:"starts_at,omitempty"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
	RegEndsAt *time.Time `json:"reg_ends_at,omitempty"`

	Location        string `json:"location,omitempty"`
	MaxParticipants *int   `json:"max_participants,omitempty"`
	RewardCoins     int    `json:"reward_coins" gorm:"not null"`

	Config   datatypes.JSON `json:"config" gorm:"type:jsonb;not null;default:'{}'"`
	CoverURL string         `json:"cover_url,omitempty"`
	Metadata datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (Competition) TableName() string { return "competitions" }

// RegOpen — ro'yxatdan o'tish hozir mumkinmi. Uchta shart birga:
// holat `registration`, muddat o'tmagan, joy bor.
//
// Bu biznes qoidasi, shuning uchun domenda: repozitoriy ham (yozishdan oldin),
// read-model ham (tugmani ko'rsatish uchun) shu yagona manbadan foydalanadi —
// ikki joyda ikki xil talqin bo'lib qolmasin.
func (c *Competition) RegOpen(participants int, now time.Time) bool {
	if c.Status != CompStatusRegistration {
		return false
	}
	if c.RegEndsAt != nil && c.RegEndsAt.Before(now) {
		return false
	}
	if c.MaxParticipants != nil && *c.MaxParticipants > 0 && participants >= *c.MaxParticipants {
		return false
	}
	return true
}

// CompetitionRegistration — ishtirok yozuvi.
type CompetitionRegistration struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	CompetitionID uuid.UUID `json:"competition_id" gorm:"type:uuid;not null"`

	Status       string    `json:"status" gorm:"not null"`
	RegisteredAt time.Time `json:"registered_at"`

	Result datatypes.JSON `json:"result,omitempty" gorm:"type:jsonb"`
	Place  *int16         `json:"place,omitempty"`

	RewardGranted bool `json:"reward_granted" gorm:"not null"`
}

func (CompetitionRegistration) TableName() string { return "competition_registrations" }

// CompetitionView — mobil ro'yxat: musobaqa + shu foydalanuvchi holati.
type CompetitionView struct {
	Competition
	Registered       bool   `json:"registered"`
	Place            *int16 `json:"place,omitempty"`
	RewardGranted    bool   `json:"reward_granted"`
	ParticipantCount int    `json:"participant_count"`
	// RegOpen — hozir ro'yxatdan o'tish mumkinmi (holat + muddat + joy soni).
	RegOpen bool `json:"reg_open"`
}

// ─────────────────────── Tur registri (§16.2 andozasi) ───────────────────────

// CompetitionTypeSpec — turning ta'rifi: admin formasi shu yerdan yasaladi.
type CompetitionTypeSpec struct {
	Type   CompetitionType  `json:"type"`
	Label  string           `json:"label"`
	Fields []ChallengeField `json:"fields"` // maydon ta'rifi chellenj bilan bir xil shakl
}

var competitionTypes = map[CompetitionType]CompetitionTypeSpec{
	CompetitionIndividual: {
		Type:  CompetitionIndividual,
		Label: "Shaxsiy musobaqa",
		Fields: []ChallengeField{
			{Key: "sport", Label: "Sport turi", Type: FieldText, Required: true},
		},
	},
	CompetitionTeam: {
		Type:  CompetitionTeam,
		Label: "Jamoaviy musobaqa",
		Fields: []ChallengeField{
			{Key: "sport", Label: "Sport turi", Type: FieldText, Required: true},
			{Key: "team_size", Label: "Jamoa a'zolari", Type: FieldNumber, Required: true, Min: f64(2), Unit: "kishi"},
		},
	},
	CompetitionFacultyVs: {
		Type:  CompetitionFacultyVs,
		Label: "Fakultetlararo",
		Fields: []ChallengeField{
			{Key: "sport", Label: "Sport turi", Type: FieldText, Required: true},
			{Key: "metric", Label: "Hisob mezoni", Type: FieldSelect, Required: true,
				Options: []string{"steps", "distance", "active_min", "score"}},
		},
	},
	CompetitionCustom: {
		Type:  CompetitionCustom,
		Label: "Boshqa",
		Fields: []ChallengeField{
			{Key: "note", Label: "Izoh", Type: FieldText, Required: false},
		},
	},
}

// CompetitionTypeSpecs — admin dinamik formasi uchun (GET /competition-types).
func CompetitionTypeSpecs() []CompetitionTypeSpec {
	order := []CompetitionType{CompetitionIndividual, CompetitionTeam, CompetitionFacultyVs, CompetitionCustom}
	out := make([]CompetitionTypeSpec, 0, len(order))
	for _, t := range order {
		out = append(out, competitionTypes[t])
	}
	return out
}

func ValidCompetitionType(t string) bool {
	_, ok := competitionTypes[CompetitionType(t)]
	return ok
}

func ValidCompetitionStatus(s string) bool {
	switch s {
	case CompStatusDraft, CompStatusRegistration, CompStatusOngoing, CompStatusFinished:
		return true
	}
	return false
}

// ValidateCompetitionConfig — turga qarab config tekshiruvi.
// Chellenjdagi maydon validatsiyasini qayta ishlatadi (validateFields).
func ValidateCompetitionConfig(t CompetitionType, raw []byte) error {
	spec, ok := competitionTypes[t]
	if !ok {
		return &ErrChallengeConfig{Field: "type", Reason: "noma'lum tur"}
	}
	return validateFields(spec.Fields, raw)
}

// validateFields — maydon ta'riflariga qarab JSON config'ni tekshiradi.
// Chellenj va musobaqa uchun umumiy: ikkalasida ham "tur -> maydonlar" andozasi.
func validateFields(fields []ChallengeField, raw []byte) error {
	cfg := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &cfg); err != nil {
			return &ErrChallengeConfig{Field: "config", Reason: "JSON o'qib bo'lmadi"}
		}
	}

	known := make(map[string]ChallengeField, len(fields))
	for _, f := range fields {
		known[f.Key] = f
	}
	// Registrda e'lon qilinmagan kalit — xato (jimgina saqlanib qolmasin).
	for key := range cfg {
		if _, ok := known[key]; !ok {
			return &ErrChallengeConfig{Field: key, Reason: "bu tur uchun noma'lum maydon"}
		}
	}

	for _, f := range fields {
		v, present := cfg[f.Key]
		if !present || v == nil {
			if f.Required {
				return &ErrChallengeConfig{Field: f.Key, Reason: "majburiy maydon"}
			}
			continue
		}
		if err := validateFieldValue(f, v); err != nil {
			return err
		}
	}
	return nil
}

// CompetitionFilter — ro'yxat so'rovi.
type CompetitionFilter struct {
	Status string
	Type   string
	Page   int
	Limit  int
}

// CompetitionRepository — musobaqa ma'lumotlari uchun port.
type CompetitionRepository interface {
	Create(ctx context.Context, c *Competition) error
	Update(ctx context.Context, c *Competition) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Competition, error)
	List(ctx context.Context, f CompetitionFilter) ([]Competition, int64, error)
	// ListForUser — musobaqa + foydalanuvchi holati + ishtirokchilar soni.
	ListForUser(ctx context.Context, userID uuid.UUID, f CompetitionFilter) ([]CompetitionView, int64, error)
	// Register — ro'yxatdan o'tadi. Joy soni tekshiruvi tranzaksiya ichida.
	Register(ctx context.Context, userID, competitionID uuid.UUID) (*CompetitionRegistration, error)
	// Cancel — ishtirokni bekor qiladi (yozuv o'chirilmaydi, status o'zgaradi).
	Cancel(ctx context.Context, userID, competitionID uuid.UUID) error
	// Participants — admin uchun ishtirokchilar ro'yxati.
	Participants(ctx context.Context, competitionID uuid.UUID, page, limit int) ([]CompetitionRegistration, int64, error)
}
