package dto

// RegisterRequest — ro'yxatdan o'tish so'rovi.
type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required,min=3,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Phone    string `json:"phone" binding:"omitempty,e164"`
	Role     string `json:"role" binding:"omitempty,oneof=student employee"`
	Language string `json:"language" binding:"omitempty,oneof=uz ru en"`
}

// LoginRequest — kirish so'rovi.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`

	// Device — qurilma ma'lumoti (bir qurilma cheklovi uchun).
	// Berilmasa cheklov qo'llanmaydi (eski mijozlar, web admin panel).
	Device *DeviceInfo `json:"device"`

	// ForceDevice — foydalanuvchi "boshqa qurilmadan chiqarilsin" deb
	// rozilik berdi. Mijoz buni FAQAT konflikt oynasidan keyin yuboradi.
	ForceDevice bool `json:"force_device"`
}

// DeviceInfo — mijoz yuboradigan qurilma ma'lumoti.
//
// IP va UserAgent mijozdan OLINMAYDI — ularni server so'rovdan o'zi
// aniqlaydi (aks holda foydalanuvchi ularni istalgancha yozardi).
type DeviceInfo struct {
	DeviceID   string `json:"device_id" binding:"omitempty,max=128"`
	DeviceName string `json:"device_name" binding:"omitempty,max=255"`
	Platform   string `json:"platform" binding:"omitempty,oneof=android ios web"`
	AppVersion string `json:"app_version" binding:"omitempty,max=32"`

	IP        string `json:"-"`
	UserAgent string `json:"-"`
}

// DeviceConflictResponse — boshqa qurilmada faol sessiya bor.
type DeviceConflictResponse struct {
	Error string `json:"error"`
	// Device — band qurilma (foydalanuvchi tanishi uchun).
	Device struct {
		Name       string `json:"name"`
		Platform   string `json:"platform"`
		LastSeenAt string `json:"last_seen_at"`
	} `json:"device"`
}

// RefreshRequest — token yangilash so'rovi.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// HemisExchangeRequest — mobil OAuth: bir martalik code'ni token'ga almashtirish.
type HemisExchangeRequest struct {
	Code string `json:"code" binding:"required"`

	// Device / ForceDevice — LoginRequest dagi bilan bir xil ma'noda:
	// haqiqiy foydalanuvchilar aynan shu oqim bilan kiradi, shuning uchun
	// qurilma cheklovi bu yerda ham qo'llanadi.
	Device      *DeviceInfo `json:"device"`
	ForceDevice bool        `json:"force_device"`
}

// TokenResponse — access + refresh token javobi.
type TokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	User         UserInfo `json:"user"`
}

// UserInfo — javobda qaytariladigan xavfsiz foydalanuvchi ma'lumotlari.
type UserInfo struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Language string `json:"language"`
}
