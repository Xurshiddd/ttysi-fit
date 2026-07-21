package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// FIT Coin berish sabablari. Kodda faqat shu konstantalar ishlatiladi —
// `reason` DB'ga matn sifatida tushadi, shuning uchun u tekshirilgan bo'lishi shart.
const (
	CoinReasonChallengeReward   = "challenge_reward"
	CoinReasonCompetitionReward = "competition_reward"
	CoinReasonAchievementReward = "achievement_reward"
	CoinReasonAdminGrant        = "admin_grant"
	CoinReasonAdminRevoke       = "admin_revoke"
)

// Manba turlari (ref_type).
const (
	CoinRefChallenge   = "challenge"
	CoinRefCompetition = "competition"
	CoinRefAchievement = "achievement"
)

// ValidCoinReason — sabab ro'yxatda bormi.
func ValidCoinReason(r string) bool {
	switch r {
	case CoinReasonChallengeReward, CoinReasonCompetitionReward,
		CoinReasonAchievementReward, CoinReasonAdminGrant, CoinReasonAdminRevoke:
		return true
	}
	return false
}

// FitCoin — ledger yozuvi. O'zgarmas (append-only): yaratilgach tahrirlanmaydi.
// Xato bo'lsa teskari yozuv qo'shiladi (admin_revoke), yozuv o'chirilmaydi.
type FitCoin struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`

	UserID uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`

	// Amount — musbat tushum, manfiy chiqim. Nol bo'lmaydi (DB CHECK).
	Amount int    `json:"amount" gorm:"not null"`
	Reason string `json:"reason" gorm:"not null"`

	RefType string     `json:"ref_type,omitempty"`
	RefID   *uuid.UUID `json:"ref_id,omitempty" gorm:"type:uuid"`

	Note     string         `json:"note,omitempty"`
	Metadata datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (FitCoin) TableName() string { return "fit_coins" }

// CoinBalance — foydalanuvchi balansi va yig'ma ko'rsatkichlar.
type CoinBalance struct {
	Balance int `json:"balance"`
	Earned  int `json:"earned"` // jami tushum
	Spent   int `json:"spent"`  // jami chiqim (musbat son sifatida)
}

// CoinFilter — tarix so'rovi.
type CoinFilter struct {
	Reason string
	Page   int
	Limit  int
}

// FitCoinRepository — FIT Coin ledger uchun port.
type FitCoinRepository interface {
	// Balance — SUM(amount) bo'yicha joriy balans.
	Balance(ctx context.Context, userID uuid.UUID) (*CoinBalance, error)
	// History — foydalanuvchi tranzaksiyalari (yangi -> eski).
	History(ctx context.Context, userID uuid.UUID, f CoinFilter) ([]FitCoin, int64, error)
	// Grant — ledger'ga yozuv qo'shadi.
	// ref_id berilgan bo'lsa idempotent: o'sha manba uchun ikkinchi yozuv
	// qo'shilmaydi va ErrAlreadyExists qaytadi.
	Grant(ctx context.Context, c *FitCoin) error
	// GrantChallengeReward — chellenj mukofotini bitta tranzaksiyada beradi:
	// ishtirok yozuvini FOR UPDATE bilan bloklaydi, yakunlanganini va mukofot
	// hali berilmaganini tekshiradi, ledger'ga yozadi va reward_granted=true qiladi.
	GrantChallengeReward(ctx context.Context, userID, challengeID uuid.UUID) (*FitCoin, error)
}
