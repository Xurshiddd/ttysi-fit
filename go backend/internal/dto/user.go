package dto

// UpdateProfileRequest — foydalanuvchi o'z profilini yangilaydi (PUT /users/me).
//
// Maydonlar pointer: nil — "tegilmasin", "" — "tozalansin". Shu sababli qisman
// yangilash mumkin (mijoz faqat o'zgargan maydonni yuboradi).
//
// Bu yerda ataylab faqat uchta maydon bor. full_name, email, role, gender,
// birth_date, course, position, specialty, avatar_url — bularni HEMIS sync
// qayta yozadi, shuning uchun tahrirlash ma'nosiz bo'lardi (o'zgarish keyingi
// syncda yo'qolardi). Entity to'g'ridan-to'g'ri bind qilinmaydi — DTO
// whitelisting (CLAUDE.md §17.3 #13).
type UpdateProfileRequest struct {
	// e164_opt — bo'sh satr ham qabul qilinadi (raqamni tozalash). Oddiy `e164`
	// bo'lsa {"phone":""} rad etilardi va raqamni o'chirib bo'lmasdi.
	Phone    *string `json:"phone" binding:"omitempty,e164_opt"`
	Bio      *string `json:"bio" binding:"omitempty,max=500"`
	Language *string `json:"language" binding:"omitempty,oneof=uz ru en"`
}
