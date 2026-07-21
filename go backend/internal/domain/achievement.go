package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Yutuq turlari (CLAUDE.md §16.1/§16.3). Chellenj registri bilan bir xil
// andoza: yangi tur qo'shish uchun shu yerga konstanta + registrga spec —
// migration ham, admin panel formasi ham o'zgarmaydi.
type AchievementType string

const (
	AchStepsTotal     AchievementType = "steps_total"     // jami N qadam
	AchDistanceTotal  AchievementType = "distance_total"  // jami N km
	AchActiveDays     AchievementType = "active_days"     // N kun faol bo'lish
	AchChallengeCount AchievementType = "challenge_count" // N ta chellenjni yakunlash
	AchManual         AchievementType = "manual"          // qo'lda beriladi (g'olib, ishtirokchi)
)

// Berish usuli.
const (
	AwardModeAuto   = "auto"   // mezon ma'lumotdan hisoblanadi
	AwardModeManual = "manual" // admin qo'lda beradi
)

// Yutuq holati.
const (
	AchStatusDraft    = "draft"
	AchStatusActive   = "active"
	AchStatusArchived = "archived"
)

// Achievement — yutuq shabloni. Umumiy maydonlar ustunda, turga xos mezon
// Criteria JSONB da.
type Achievement struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	Type        AchievementType `json:"type" gorm:"not null"`
	Title       string          `json:"title" gorm:"not null"`
	Description string          `json:"description,omitempty"`
	AwardMode   string          `json:"award_mode" gorm:"not null"`
	Status      string          `json:"status" gorm:"not null"`
	RewardCoins int             `json:"reward_coins" gorm:"not null"`

	Criteria datatypes.JSON `json:"criteria" gorm:"type:jsonb;not null;default:'{}'"`

	IconURL             string         `json:"icon_url,omitempty"`
	CoverURL            string         `json:"cover_url,omitempty"`
	CertificateTemplate string         `json:"certificate_template,omitempty"`
	Metadata            datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (Achievement) TableName() string { return "achievements" }

// UserAchievement — foydalanuvchiga berilgan yutuq (= sertifikat yozuvi).
type UserAchievement struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	AchievementID uuid.UUID `json:"achievement_id" gorm:"type:uuid;not null"`

	AwardedAt time.Time  `json:"awarded_at"`
	AwardedBy *uuid.UUID `json:"awarded_by,omitempty" gorm:"type:uuid"`

	// ProgressValue — berilgan paytdagi qiymat snapshot'i (sertifikatda turadi).
	ProgressValue  float64 `json:"progress_value"`
	Note           string  `json:"note,omitempty"`
	CertificateURL string  `json:"certificate_url,omitempty"`
	RewardGranted  bool    `json:"reward_granted" gorm:"not null"`

	Metadata datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (UserAchievement) TableName() string { return "user_achievements" }

// AchievementView — mobil ro'yxat uchun read-model: yutuq + shu foydalanuvchi
// holati (qozonilganmi, hozirgi progressi qancha).
type AchievementView struct {
	Achievement
	Earned      bool       `json:"earned"`
	EarnedAt    *time.Time `json:"earned_at,omitempty"`
	Progress    float64    `json:"progress"`     // hozirgi qiymat (metrika birligida)
	Target      float64    `json:"target"`       // mezondagi maqsad (0 — qo'lda beriladigan)
	ProgressPct float64    `json:"progress_pct"` // 0..100

	// AwardID — berilgan yutuq yozuvi (user_achievements.id). Mijoz sertifikat
	// havolasini shundan yasaydi: GET /achievements/awards/{award_id}/certificate.
	// Qozonilmagan yutuqda nil.
	AwardID *uuid.UUID `json:"award_id,omitempty"`

	// CertificateURL — qo'lda yuklangan sertifikat fayli bo'lsa (ixtiyoriy).
	// Odatda bo'sh: sertifikat AwardID orqali dinamik chiziladi.
	CertificateURL string `json:"certificate_url,omitempty"`
}

// ─────────────────────────── Tur registri (§16.2) ───────────────────────────

// AchievementTypeSpec — bitta yutuq turining to'liq ta'rifi.
type AchievementTypeSpec struct {
	Type   AchievementType  `json:"type"`
	Label  string           `json:"label"`
	Fields []ChallengeField `json:"fields"` // maydon ta'rifi chellenj bilan umumiy

	// AwardMode — bu tur qanday beriladi. Admin panel shunga qarab
	// "qo'lda berish" tugmasini ko'rsatadi yoki yashiradi.
	AwardMode string `json:"award_mode"`

	// Source — progress qayerdan hisoblanadi. Bo'sh — avtomatik hisoblanmaydi.
	// MUHIM: bu qiymat SQL yasashda ishlatiladi, shuning uchun faqat shu
	// registrdan keladi — mijoz kiritmasi bu yerga hech qachon tushmaydi (§3.2).
	Source AchievementSource `json:"-"`

	// ThresholdKey — criteria'dagi maqsad kaliti. Bo'sh — maqsadsiz tur.
	ThresholdKey string `json:"-"`

	// ThresholdUnitToMetric — maqsad birligini metrika birligiga o'giradi
	// (km -> metr: 1000).
	ThresholdUnitToMetric float64 `json:"-"`
}

// AchievementSource — progress manbai. Registrdan tashqarida yaratilmaydi.
type AchievementSource string

const (
	// SourceActivitySum — activities jadvalidagi ustun yig'indisi.
	SourceActivitySum AchievementSource = "activity_sum"
	// SourceActiveDays — steps > 0 bo'lgan kunlar soni.
	SourceActiveDays AchievementSource = "active_days"
	// SourceChallengeDone — yakunlangan chellenjlar soni.
	SourceChallengeDone AchievementSource = "challenge_done"
)

// achievementTypes — yagona haqiqat manbai: validatsiya, admin formasi va
// progress hisoblash shu yerdan oziqlanadi.
var achievementTypes = map[AchievementType]AchievementTypeSpec{
	AchStepsTotal: {
		Type:  AchStepsTotal,
		Label: "Jami qadam",
		Fields: []ChallengeField{
			{Key: "threshold", Label: "Maqsad", Type: FieldNumber, Required: true, Min: f64(1), Unit: "qadam"},
		},
		AwardMode:             AwardModeAuto,
		Source:                SourceActivitySum,
		ThresholdKey:          "threshold",
		ThresholdUnitToMetric: 1,
	},
	AchDistanceTotal: {
		Type:  AchDistanceTotal,
		Label: "Jami masofa",
		Fields: []ChallengeField{
			{Key: "threshold", Label: "Maqsad", Type: FieldNumber, Required: true, Min: f64(0.1), Unit: "km"},
		},
		AwardMode:             AwardModeAuto,
		Source:                SourceActivitySum,
		ThresholdKey:          "threshold",
		ThresholdUnitToMetric: 1000, // km -> metr
	},
	AchActiveDays: {
		Type:  AchActiveDays,
		Label: "Faol kunlar",
		Fields: []ChallengeField{
			{Key: "threshold", Label: "Maqsad", Type: FieldNumber, Required: true, Min: f64(1), Unit: "kun"},
		},
		AwardMode:             AwardModeAuto,
		Source:                SourceActiveDays,
		ThresholdKey:          "threshold",
		ThresholdUnitToMetric: 1,
	},
	AchChallengeCount: {
		Type:  AchChallengeCount,
		Label: "Yakunlangan chellenjlar",
		Fields: []ChallengeField{
			{Key: "threshold", Label: "Maqsad", Type: FieldNumber, Required: true, Min: f64(1), Unit: "ta"},
		},
		AwardMode:             AwardModeAuto,
		Source:                SourceChallengeDone,
		ThresholdKey:          "threshold",
		ThresholdUnitToMetric: 1,
	},
	AchManual: {
		Type:  AchManual,
		Label: "Qo'lda beriladigan (g'olib, ishtirokchi)",
		Fields: []ChallengeField{
			{Key: "note", Label: "Izoh", Type: FieldText, Required: false},
		},
		AwardMode: AwardModeManual,
		// Source bo'sh — avtomatik baholanmaydi.
	},
}

// activityMetricColumn — SourceActivitySum turlari uchun activities ustuni.
// Registr ichida qoladi: SQL'ga faqat shu yerdan ustun nomi tushadi.
func activityMetricColumn(t AchievementType) string {
	switch t {
	case AchStepsTotal:
		return "steps"
	case AchDistanceTotal:
		return "distance_m"
	}
	return ""
}

// AchievementMetricColumn — repository uchun ustun nomi (bo'sh — mos emas).
func AchievementMetricColumn(t AchievementType) string { return activityMetricColumn(t) }

// AchievementTypeSpecs — admin panel dinamik formasi uchun
// (GET /admin/achievement-types). Tartib barqaror.
func AchievementTypeSpecs() []AchievementTypeSpec {
	order := []AchievementType{
		AchStepsTotal, AchDistanceTotal, AchActiveDays, AchChallengeCount, AchManual,
	}
	out := make([]AchievementTypeSpec, 0, len(order))
	for _, t := range order {
		out = append(out, achievementTypes[t])
	}
	return out
}

// AchievementSpec — tur ta'rifini qaytaradi.
func AchievementSpec(t AchievementType) (AchievementTypeSpec, bool) {
	s, ok := achievementTypes[t]
	return s, ok
}

// ValidAchievementType — tur registrda bormi.
func ValidAchievementType(t string) bool {
	_, ok := achievementTypes[AchievementType(t)]
	return ok
}

// ValidAchievementStatus — holat qiymati to'g'rimi.
func ValidAchievementStatus(s string) bool {
	switch s {
	case AchStatusDraft, AchStatusActive, AchStatusArchived:
		return true
	}
	return false
}

// ValidateAchievementCriteria — mezonni turga qarab tekshiradi (§16.2).
// Registrda e'lon qilinmagan kalitlar rad etiladi: admin xato yozgan kalit
// ("treshold") jimgina saqlanib, yutuq hech qachon berilmasdi.
func ValidateAchievementCriteria(t AchievementType, raw []byte) error {
	spec, ok := achievementTypes[t]
	if !ok {
		return &ErrChallengeConfig{Field: "type", Reason: "noma'lum tur"}
	}
	return validateFields(spec.Fields, raw)
}

// AchievementThreshold — criteria'dan maqsadni metrika birligida qaytaradi.
// 0 — maqsadsiz tur (manual) yoki maqsad ko'rsatilmagan.
func AchievementThreshold(t AchievementType, raw []byte) float64 {
	spec, ok := achievementTypes[t]
	if !ok || spec.ThresholdKey == "" || len(raw) == 0 {
		return 0
	}
	cfg := map[string]any{}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return 0
	}
	n, ok := cfg[spec.ThresholdKey].(float64)
	if !ok {
		return 0
	}
	return n * spec.ThresholdUnitToMetric
}

// AchievementProgressPct — progressni 0..100 oralig'ida qaytaradi.
// Maqsadsiz turda (manual) 0 qaytadi — mijoz progress ko'rsatmaydi.
func AchievementProgressPct(progress, target float64) float64 {
	if target <= 0 {
		return 0
	}
	pct := progress / target * 100
	if pct > 100 {
		return 100
	}
	if pct < 0 {
		return 0
	}
	return pct
}

// AwardDetail — sertifikat chizish uchun kerakli to'liq ma'lumot
// (berilgan yutuq + yutuq shabloni + foydalanuvchi ismi, bitta JOIN'da).
type AwardDetail struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	UserFullName  string
	Type          AchievementType
	Title         string
	Description   string
	Criteria      datatypes.JSON
	ProgressValue float64
	Note          string
	AwardedAt     time.Time
}

// AchievementValueLabel — yutuq o'lchovini o'qiladigan matnga aylantiradi
// (sertifikatda ko'rsatish uchun). Qo'lda beriladigan turda bo'sh.
func AchievementValueLabel(t AchievementType, value float64) string {
	if value <= 0 {
		return ""
	}
	switch t {
	case AchStepsTotal:
		return groupDigits(int64(value)) + " qadam"
	case AchDistanceTotal:
		// value metrda saqlanadi (§ AchievementThreshold: km -> metr).
		return strings.TrimSuffix(strconv.FormatFloat(value/1000, 'f', 1, 64), ".0") + " km"
	case AchActiveDays:
		return groupDigits(int64(value)) + " kun"
	case AchChallengeCount:
		return groupDigits(int64(value)) + " ta chellenj"
	}
	return ""
}

// groupDigits — 4238 -> "4 238" (uch xonali guruhlar, bo'sh joy bilan).
func groupDigits(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	lead := len(s) % 3
	if lead > 0 {
		b.WriteString(s[:lead])
	}
	for i := lead; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// ErrAchievementAwardMode — turga mos kelmaydigan berish usuli.
type ErrAchievementAwardMode struct {
	Type   AchievementType
	Reason string
}

func (e *ErrAchievementAwardMode) Error() string {
	return fmt.Sprintf("achievement award mode: %s: %s", e.Type, e.Reason)
}

// AchievementFilter — ro'yxat so'rovi.
type AchievementFilter struct {
	Status    string // bo'sh — hammasi (admin); mobil "active" so'raydi
	Type      string
	AwardMode string
	Page      int
	Limit     int
}

// AchievementRepository — yutuq ma'lumotlari uchun port.
type AchievementRepository interface {
	Create(ctx context.Context, a *Achievement) error
	Update(ctx context.Context, a *Achievement) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Achievement, error)

	// List — admin ro'yxati (filtr + paginatsiya).
	List(ctx context.Context, f AchievementFilter) ([]Achievement, int64, error)

	// ListForUser — mobil ro'yxat: yutuq + foydalanuvchi holati va progressi.
	// Bitta so'rovda (N+1 yo'q, §3.1): progress LEFT JOIN agregatlaridan olinadi.
	ListForUser(ctx context.Context, userID uuid.UUID, f AchievementFilter) ([]AchievementView, int64, error)

	// ListEarned — foydalanuvchi qozongan yutuqlar (profil/sertifikatlar bo'limi).
	ListEarned(ctx context.Context, userID uuid.UUID, f AchievementFilter) ([]AchievementView, int64, error)

	// Award — yutuqni beradi (idempotent: takror chaqiruvda ErrAlreadyExists).
	// awardedBy nil bo'lsa avtomatik berilgan hisoblanadi.
	Award(ctx context.Context, userID, achievementID uuid.UUID, awardedBy *uuid.UUID, value float64, note string) (*UserAchievement, error)

	// EvaluateAuto — foydalanuvchi uchun avtomatik yutuqlarni baholaydi va
	// mezonga yetganlarini beradi. Yangi berilganlarini qaytaradi.
	EvaluateAuto(ctx context.Context, userID uuid.UUID) ([]UserAchievement, error)

	// GetAward — berilgan yutuq tafsiloti (sertifikat chizish uchun).
	GetAward(ctx context.Context, userAchievementID uuid.UUID) (*AwardDetail, error)
}
