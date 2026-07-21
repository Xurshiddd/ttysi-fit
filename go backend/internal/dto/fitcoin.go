package dto

// AdminGrantCoinsRequest — admin qo'lda FIT Coin beradi/oladi.
// Manfiy `amount` — olish (balans manfiy bo'lib ketmaydi, servis tekshiradi).
type AdminGrantCoinsRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Amount int    `json:"amount" binding:"required,ne=0"`
	Note   string `json:"note" binding:"omitempty,max=500"`
}
