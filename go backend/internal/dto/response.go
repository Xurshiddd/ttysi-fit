package dto

// Meta — paginatsiya va qo'shimcha ma'lumotlar uchun.
type Meta struct {
	Page  int   `json:"page,omitempty"`
	Limit int   `json:"limit,omitempty"`
	Total int64 `json:"total,omitempty"`
}

// Response — standart muvaffaqiyatli javob qobig'i: { "data": ..., "meta": ... }
type Response struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

// ErrorResponse — standart xato javobi. Ichki tafsilotlar mijozga chiqmaydi.
// Error — foydalanuvchi tiliga tarjima qilingan xabar.
// Details — texnik tafsilot (validatsiya maydonlari kabi), ixtiyoriy.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details string            `json:"details,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// OK — ma'lumot bilan muvaffaqiyatli javob qaytaradi.
func OK(data any) Response {
	return Response{Data: data}
}

// Paginated — paginatsiyali javob qaytaradi.
func Paginated(data any, meta Meta) Response {
	return Response{Data: data, Meta: &meta}
}

// ErrResponse — xato xabari bilan javob qaytaradi.
func ErrResponse(msg string) ErrorResponse {
	return ErrorResponse{Error: msg}
}

// ErrDetailed — tarjima qilingan xabar + texnik tafsilot bilan javob.
func ErrDetailed(msg, details string) ErrorResponse {
	return ErrorResponse{Error: msg, Details: details}
}

// ErrValidation — tarjima qilingan umumiy xabar + maydon bo'yicha xatolar.
func ErrValidation(msg string, fields map[string]string) ErrorResponse {
	return ErrorResponse{Error: msg, Fields: fields}
}
