package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Chellenj turlari (CLAUDE.md §16.1). Yangi tur qo'shish uchun:
//   1) shu yerga konstanta,
//   2) challengeTypes registriga spec,
//   3) tamom — migration ham, admin panel formasi ham o'zgarmaydi
//      (forma GET /challenge-types dan dinamik yasaladi).
type ChallengeType string

const (
	ChallengeSteps     ChallengeType = "steps"      // N qadam yig'ish
	ChallengeDistance  ChallengeType = "distance"   // N km yugurish/yurish
	ChallengeActiveMin ChallengeType = "active_min" // N faol daqiqa
	ChallengeCustom    ChallengeType = "custom"     // maqsadsiz aksiya (qo'lda yopiladi)
)

// Chellenj holati.
const (
	ChallengeStatusDraft    = "draft"
	ChallengeStatusActive   = "active"
	ChallengeStatusFinished = "finished"
)

// Chellenj qamrovi.
const (
	ChallengeScopeUniversity = "university"
	ChallengeScopeFaculty    = "faculty"
	ChallengeScopeGroup      = "group"
)

// Challenge — umumiy maydonlar ustunda, turga xos parametrlar Config JSONB da.
type Challenge struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	Type        ChallengeType `json:"type" gorm:"not null"`
	Title       string        `json:"title" gorm:"not null"`
	Description string        `json:"description,omitempty"`
	Scope       string        `json:"scope" gorm:"not null"`
	StartsAt    *time.Time    `json:"starts_at,omitempty"`
	EndsAt      *time.Time    `json:"ends_at,omitempty"`
	Status      string        `json:"status" gorm:"not null"`
	RewardCoins int           `json:"reward_coins" gorm:"not null"`

	Config   datatypes.JSON `json:"config" gorm:"type:jsonb;not null;default:'{}'"`
	CoverURL string         `json:"cover_url,omitempty"`
	Metadata datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (Challenge) TableName() string { return "challenges" }

// UserChallenge — foydalanuvchining chellenjdagi ishtiroki va progressi.
type UserChallenge struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	ChallengeID uuid.UUID `json:"challenge_id" gorm:"type:uuid;not null"`

	JoinedAt    time.Time  `json:"joined_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	Progress      float64 `json:"progress" gorm:"not null"`
	RewardGranted bool    `json:"reward_granted" gorm:"not null"`
}

func (UserChallenge) TableName() string { return "user_challenges" }

// ChallengeView — mobil ro'yxat uchun read-model: chellenj + shu foydalanuvchi
// holati (qo'shilganmi, progressi qancha).
type ChallengeView struct {
	Challenge
	Joined    bool    `json:"joined"`
	Progress  float64 `json:"progress"`
	Completed bool    `json:"completed"`
	// RewardGranted — mukofot allaqachon olinganmi. Mijoz shu bo'yicha
	// "Mukofotni olish" tugmasini ko'rsatadi yoki yashiradi.
	RewardGranted bool    `json:"reward_granted"`
	Target        float64 `json:"target"`       // config'dan olingan maqsad (0 — maqsadsiz)
	ProgressPct   float64 `json:"progress_pct"` // 0..100
}

// ─────────────────────────── Tur registri (§16.2) ───────────────────────────

// ChallengeFieldType — admin formasidagi maydon turi.
type ChallengeFieldType string

const (
	FieldNumber ChallengeFieldType = "number"
	FieldText   ChallengeFieldType = "text"
	FieldSelect ChallengeFieldType = "select"
)

// ChallengeField — turga xos maydon ta'rifi. Admin panel shu ro'yxatdan
// dinamik forma yasaydi — yangi tur qo'shilsa frontend o'zgarmaydi.
type ChallengeField struct {
	Key      string             `json:"key"`
	Label    string             `json:"label"`
	Type     ChallengeFieldType `json:"type"`
	Required bool               `json:"required"`
	Min      *float64           `json:"min,omitempty"`
	Options  []string           `json:"options,omitempty"`
	Unit     string             `json:"unit,omitempty"`
}

// ChallengeTypeSpec — bitta turning to'liq ta'rifi.
type ChallengeTypeSpec struct {
	Type   ChallengeType    `json:"type"`
	Label  string           `json:"label"`
	Fields []ChallengeField `json:"fields"`

	// Metric — progress `activities` jadvalining qaysi ustunidan yig'iladi.
	// Bo'sh — avtomatik hisoblanmaydi (custom).
	// MUHIM: bu qiymat SQL'ga ustun nomi sifatida tushadi, shuning uchun u
	// faqat shu registrdan keladi — mijoz kiritmasi hech qachon bu yerga
	// tushmaydi (§3.2 SQL injection).
	Metric string `json:"-"`

	// TargetKey — config'dagi maqsad kaliti. Bo'sh — maqsadsiz tur.
	TargetKey string `json:"-"`

	// TargetUnitToMetric — maqsad birligini metrika birligiga o'giradi
	// (masalan km -> metr: 1000).
	TargetUnitToMetric float64 `json:"-"`
}

func f64(v float64) *float64 { return &v }

// challengeTypes — yagona haqiqat manbai: validatsiya, admin formasi va
// progress hisoblash — hammasi shu yerdan oziqlanadi.
var challengeTypes = map[ChallengeType]ChallengeTypeSpec{
	ChallengeSteps: {
		Type:  ChallengeSteps,
		Label: "Qadam yig'ish",
		Fields: []ChallengeField{
			{Key: "target_steps", Label: "Maqsad", Type: FieldNumber, Required: true, Min: f64(1), Unit: "qadam"},
		},
		Metric:             "steps",
		TargetKey:          "target_steps",
		TargetUnitToMetric: 1,
	},
	ChallengeDistance: {
		Type:  ChallengeDistance,
		Label: "Masofa bosib o'tish",
		Fields: []ChallengeField{
			{Key: "target_km", Label: "Maqsad", Type: FieldNumber, Required: true, Min: f64(0.1), Unit: "km"},
		},
		Metric:             "distance_m",
		TargetKey:          "target_km",
		TargetUnitToMetric: 1000, // km -> metr
	},
	ChallengeActiveMin: {
		Type:  ChallengeActiveMin,
		Label: "Faol daqiqalar",
		Fields: []ChallengeField{
			{Key: "target_min", Label: "Maqsad", Type: FieldNumber, Required: true, Min: f64(1), Unit: "daqiqa"},
		},
		Metric:             "active_min",
		TargetKey:          "target_min",
		TargetUnitToMetric: 1,
	},
	ChallengeCustom: {
		Type:  ChallengeCustom,
		Label: "Boshqa aksiya",
		Fields: []ChallengeField{
			{Key: "note", Label: "Izoh", Type: FieldText, Required: false},
		},
		// Metric bo'sh — progress avtomatik hisoblanmaydi.
	},
}

// ChallengeTypeSpecs — admin panel dinamik formasi uchun (GET /challenge-types).
// Tartib barqaror bo'lishi uchun konstanta ro'yxat bo'yicha yig'iladi.
func ChallengeTypeSpecs() []ChallengeTypeSpec {
	order := []ChallengeType{ChallengeSteps, ChallengeDistance, ChallengeActiveMin, ChallengeCustom}
	out := make([]ChallengeTypeSpec, 0, len(order))
	for _, t := range order {
		out = append(out, challengeTypes[t])
	}
	return out
}

// ChallengeSpec — tur ta'rifini qaytaradi.
func ChallengeSpec(t ChallengeType) (ChallengeTypeSpec, bool) {
	s, ok := challengeTypes[t]
	return s, ok
}

// ValidChallengeType — tur registrda bormi.
func ValidChallengeType(t string) bool {
	_, ok := challengeTypes[ChallengeType(t)]
	return ok
}

// ValidChallengeStatus — holat qiymati to'g'rimi.
func ValidChallengeStatus(s string) bool {
	switch s {
	case ChallengeStatusDraft, ChallengeStatusActive, ChallengeStatusFinished:
		return true
	}
	return false
}

// ValidChallengeScope — qamrov qiymati to'g'rimi.
func ValidChallengeScope(s string) bool {
	switch s {
	case ChallengeScopeUniversity, ChallengeScopeFaculty, ChallengeScopeGroup:
		return true
	}
	return false
}

// ErrChallengeConfig — config turga mos emas (400 ga aylantiriladi).
type ErrChallengeConfig struct {
	Field  string
	Reason string
}

func (e *ErrChallengeConfig) Error() string {
	return fmt.Sprintf("challenge config: %s: %s", e.Field, e.Reason)
}

// ValidateChallengeConfig — config'ni turga qarab tekshiradi (§16.2).
//
// Registrda e'lon qilinmagan kalitlar rad etiladi: aks holda admin xato yozgan
// kalit ("target_step") jimgina saqlanib, chellenj hech qachon yakunlanmasdi.
func ValidateChallengeConfig(t ChallengeType, raw []byte) error {
	spec, ok := challengeTypes[t]
	if !ok {
		return &ErrChallengeConfig{Field: "type", Reason: "noma'lum tur"}
	}
	return validateFields(spec.Fields, raw)
}

// validateFieldValue — bitta maydon qiymatini ta'rifiga solishtiradi.
// Chellenj va musobaqa uchun umumiy (ikkalasi ham "tur -> maydonlar" andozasi).
func validateFieldValue(f ChallengeField, v any) error {
	switch f.Type {
	case FieldNumber:
		n, ok := v.(float64) // encoding/json barcha raqamni float64 qiladi
		if !ok {
			return &ErrChallengeConfig{Field: f.Key, Reason: "raqam bo'lishi kerak"}
		}
		if f.Min != nil && n < *f.Min {
			return &ErrChallengeConfig{Field: f.Key, Reason: fmt.Sprintf("eng kami %g", *f.Min)}
		}
	case FieldText:
		s, ok := v.(string)
		if !ok {
			return &ErrChallengeConfig{Field: f.Key, Reason: "matn bo'lishi kerak"}
		}
		if len(s) > 500 {
			return &ErrChallengeConfig{Field: f.Key, Reason: "ko'pi bilan 500 belgi"}
		}
	case FieldSelect:
		s, ok := v.(string)
		if !ok {
			return &ErrChallengeConfig{Field: f.Key, Reason: "matn bo'lishi kerak"}
		}
		for _, o := range f.Options {
			if o == s {
				return nil
			}
		}
		return &ErrChallengeConfig{Field: f.Key, Reason: "ruxsat etilmagan qiymat"}
	}
	return nil
}

// ChallengeTarget — config'dan maqsadni metrika birligida qaytaradi.
// 0 — maqsadsiz tur (custom) yoki maqsad ko'rsatilmagan.
func ChallengeTarget(t ChallengeType, raw []byte) float64 {
	spec, ok := challengeTypes[t]
	if !ok || spec.TargetKey == "" || len(raw) == 0 {
		return 0
	}
	cfg := map[string]any{}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return 0
	}
	n, ok := cfg[spec.TargetKey].(float64)
	if !ok {
		return 0
	}
	return n * spec.TargetUnitToMetric
}

// ChallengeFilter — ro'yxat so'rovi.
type ChallengeFilter struct {
	Status string // bo'sh — hammasi (admin); mobil "active" so'raydi
	Type   string
	Page   int
	Limit  int
}

// ChallengeRepository — chellenj ma'lumotlari uchun port.
type ChallengeRepository interface {
	Create(ctx context.Context, c *Challenge) error
	Update(ctx context.Context, c *Challenge) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Challenge, error)
	// List — admin ro'yxati (filtr + paginatsiya).
	List(ctx context.Context, f ChallengeFilter) ([]Challenge, int64, error)
	// ListForUser — mobil ro'yxat: chellenj + foydalanuvchi holati (bitta JOIN, N+1 yo'q).
	ListForUser(ctx context.Context, userID uuid.UUID, f ChallengeFilter) ([]ChallengeView, int64, error)
	// Join — foydalanuvchini chellenjga qo'shadi (idempotent).
	Join(ctx context.Context, userID, challengeID uuid.UUID) error
	// RecalcProgress — bitta ishtirokchining progressini qayta hisoblaydi.
	RecalcProgress(ctx context.Context, userID, challengeID uuid.UUID) (*UserChallenge, error)
}
