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
}

// RefreshRequest — token yangilash so'rovi.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// HemisExchangeRequest — mobil OAuth: bir martalik code'ni token'ga almashtirish.
type HemisExchangeRequest struct {
	Code string `json:"code" binding:"required"`
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
