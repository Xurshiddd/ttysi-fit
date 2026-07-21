package dto

// RewardRequest — admin sovg'a yaratadi/tahrirlaydi.
//
// `id`, `created_at` bu yerda YO'Q: mijoz ularni yubora olmasin (§17.3 #13
// mass assignment). Sovg'a mavjudligi (stock) ham faqat admin qo'lida.
type RewardRequest struct {
	Title       string `json:"title" binding:"required,min=2,max=255"`
	Description string `json:"description" binding:"omitempty,max=2000"`
	ImageURL    string `json:"image_url" binding:"omitempty,url,max=1000"`
	Category    string `json:"category" binding:"required,oneof=merch equipment certificate other"`

	CostCoins int `json:"cost_coins" binding:"required,gt=0,lte=1000000"`

	// Stock — nil bo'lsa cheksiz. 0 — tugagan.
	Stock *int `json:"stock" binding:"omitempty,gte=0"`
	// PerUserLimit — nil bo'lsa cheklovsiz.
	PerUserLimit *int `json:"per_user_limit" binding:"omitempty,gt=0"`

	IsActive bool    `json:"is_active"`
	StartsAt *string `json:"starts_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndsAt   *string `json:"ends_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`

	// Config — turga xos qo'shimcha maydonlar (§16.1): o'lcham, rang va h.k.
	Config map[string]any `json:"config"`
}

// RedemptionActionRequest — buyurtmani topshirish/bekor qilish izohi.
type RedemptionActionRequest struct {
	Note string `json:"note" binding:"omitempty,max=500"`
}
