package domain

import "errors"

// Domain-specific xatolar (CLAUDE.md 3.4).
var (
	ErrNotFound            = errors.New("topilmadi")
	ErrAlreadyExists       = errors.New("allaqachon mavjud")
	ErrInvalidCredentials  = errors.New("login yoki parol noto'g'ri")
	ErrUnauthorized        = errors.New("ruxsat yo'q")
	ErrInsufficientBalance = errors.New("balans yetarli emas")
	// ErrValidation — biznes qoidasi buzildi (handler'da 400 ga aylanadi).
	// DTO bind validatsiyasidan farq qiladi: bu domen qoidasi.
	ErrValidation = errors.New("ma'lumot noto'g'ri")

	// Faollik (activity) batch yozuvi xatolari.
	ErrEmptyBatch    = errors.New("bo'sh ro'yxat")
	ErrBatchTooLarge = errors.New("ro'yxat juda katta")
	// ErrFutureDate — kelajakdagi kunga faollik yozib bo'lmaydi (mijoz soati
	// noto'g'ri yoki reytingni sun'iy ko'tarish urinishi).
	ErrFutureDate = errors.New("kelajakdagi sana")

	// ErrBusy — resurs band (masalan bir vaqtda ruxsat etilgan eksport soni
	// tugagan). Mijoz birozdan keyin qayta urinishi mumkin.
	ErrBusy = errors.New("hozir band")
)
