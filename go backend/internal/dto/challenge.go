package dto

import "encoding/json"

// ChallengeRequest — admin chellenj yaratadi/yangilaydi (§16).
//
// `config` ataylab `json.RawMessage`: uning ichki tuzilishi TURGA bog'liq va
// bu yerda ma'lum emas. Uni domain registri tekshiradi (ValidateChallengeConfig),
// shu sababli yangi tur qo'shilganda bu DTO o'zgarmaydi.
type ChallengeRequest struct {
	Type        string          `json:"type" binding:"required"`
	Title       string          `json:"title" binding:"required,min=3,max=255"`
	Description string          `json:"description" binding:"omitempty,max=5000"`
	Scope       string          `json:"scope" binding:"omitempty,oneof=university faculty group"`
	StartsAt    *string         `json:"starts_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndsAt      *string         `json:"ends_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Status      string          `json:"status" binding:"omitempty,oneof=draft active finished"`
	RewardCoins int             `json:"reward_coins" binding:"gte=0"`
	CoverURL    string          `json:"cover_url" binding:"omitempty,url,max=1000"`
	Config      json.RawMessage `json:"config"`
}
