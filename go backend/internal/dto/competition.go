package dto

import "encoding/json"

// CompetitionRequest — admin musobaqa yaratadi/yangilaydi (§16.3).
// `config` turga bog'liq — uni domain registri tekshiradi (ValidateCompetitionConfig).
type CompetitionRequest struct {
	Type        string          `json:"type" binding:"required"`
	Title       string          `json:"title" binding:"required,min=3,max=255"`
	Description string          `json:"description" binding:"omitempty,max=5000"`
	Scope       string          `json:"scope" binding:"omitempty,oneof=university faculty group"`
	Status      string          `json:"status" binding:"omitempty,oneof=draft registration ongoing finished"`

	StartsAt  *string `json:"starts_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndsAt    *string `json:"ends_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	RegEndsAt *string `json:"reg_ends_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`

	Location string `json:"location" binding:"omitempty,max=255"`
	// MaxParticipants — nil yoki 0: cheklovsiz.
	MaxParticipants *int            `json:"max_participants" binding:"omitempty,gte=0"`
	RewardCoins     int             `json:"reward_coins" binding:"gte=0"`
	CoverURL        string          `json:"cover_url" binding:"omitempty,url,max=1000"`
	Config          json.RawMessage `json:"config"`
}
