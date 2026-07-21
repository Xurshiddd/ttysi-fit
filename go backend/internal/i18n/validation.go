package i18n

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FieldLabel — maydonning tilga moslangan yorlig'ini qaytaradi.
// Katalogda topilmasa, json nomining o'zini qaytaradi.
func FieldLabel(l Locale, jsonName string) string {
	code := "field." + jsonName
	if _, ok := catalog[code]; ok {
		return T(l, code)
	}
	return jsonName
}

// ValidationFields — gin/validator xatosini maydon → tarjima qilingan xabar
// ko'rinishidagi mapga aylantiradi. Xato validator turidan bo'lmasa, nil qaytaradi.
func ValidationFields(l Locale, err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}

	out := make(map[string]string, len(ve))
	for _, fe := range ve {
		jsonName := fe.Field() // json tag nomi (RegisterTagNameFunc orqali)
		label := FieldLabel(l, jsonName)
		out[jsonName] = formatFieldError(l, fe, label)
	}
	return out
}

// formatFieldError — bitta maydon xatosini qoidaga qarab tarjima qiladi.
func formatFieldError(l Locale, fe validator.FieldError, label string) string {
	var code string
	switch fe.Tag() {
	case "required":
		code = ValRequired
	case "email":
		code = ValEmail
	case "min":
		code = ValMin
	case "max":
		code = ValMax
	case "oneof":
		code = ValOneOf
	case "e164", "e164_opt":
		code = ValE164
	default:
		code = ValInvalid
	}

	msg := T(l, code)
	msg = strings.ReplaceAll(msg, "{field}", label)

	param := fe.Param()
	if code == ValOneOf {
		// "student teacher" → "student, teacher"
		param = strings.ReplaceAll(param, " ", ", ")
	}
	msg = strings.ReplaceAll(msg, "{param}", param)
	return msg
}
