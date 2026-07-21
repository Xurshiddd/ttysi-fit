package dto

import "encoding/json"

// AchievementRequest — admin yutuq shablonini yaratadi/yangilaydi (§16.3).
//
// `criteria` ataylab `json.RawMessage`: ichki tuzilishi TURGA bog'liq va bu
// yerda ma'lum emas. Uni domain registri tekshiradi
// (ValidateAchievementCriteria), shu sababli yangi tur qo'shilganda bu DTO
// o'zgarmaydi.
//
// `award_mode` bu yerda YO'Q — u turdan kelib chiqadi (service.validate).
// Aks holda admin avtomatik yutuqni "manual" deb belgilab, mezonni chetlab
// o'tgan bo'lardi.
type AchievementRequest struct {
	Type                string          `json:"type" binding:"required"`
	Title               string          `json:"title" binding:"required,min=3,max=255"`
	Description         string          `json:"description" binding:"omitempty,max=5000"`
	Status              string          `json:"status" binding:"omitempty,oneof=draft active archived"`
	RewardCoins         int             `json:"reward_coins" binding:"gte=0"`
	IconURL             string          `json:"icon_url" binding:"omitempty,url,max=1000"`
	CoverURL            string          `json:"cover_url" binding:"omitempty,url,max=1000"`
	CertificateTemplate string          `json:"certificate_template" binding:"omitempty,max=5000"`
	Criteria            json.RawMessage `json:"criteria"`
}

// AwardRequest — admin yutuqni qo'lda beradi (musobaqa g'olibi, tadbir).
type AwardRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Note   string `json:"note" binding:"omitempty,max=500"`
}
