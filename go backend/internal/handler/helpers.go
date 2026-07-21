package handler

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/i18n"
)

// logServerError — handler'dagi kutilmagan xatoni loglaydi.
//
// Mijoz so'rovni uzib yuborsa (sahifadan chiqdi, filtr tez almashdi, tarmoq
// uzildi) `context.Canceled` keladi — bu SERVER xatosi emas. Uni ERROR sifatida
// yozish loglarni ifloslantiradi va monitoringda soxta ogohlantirish beradi
// (CLAUDE.md §17.3 №50), shuning uchun bunday holat Debug darajasida qoladi.
//
// `DeadlineExceeded` esa aksincha — bu bizning timeout'imiz ishlagani, ya'ni
// so'rov haqiqatan sekin. U Warn sifatida ko'rinib turishi kerak.
func logServerError(log *zap.Logger, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	switch {
	case errors.Is(err, context.Canceled):
		log.Debug(msg+": mijoz so'rovni uzdi", fields...)
	case errors.Is(err, context.DeadlineExceeded):
		log.Warn(msg+": timeout", fields...)
	default:
		log.Error(msg, fields...)
	}
}

// absoluteMediaURL — DB dagi nisbiy media yo'liga ("/static/avatars/1.jpg")
// ommaviy asosni qo'shadi. DB da host saqlanmaydi — shuning uchun port/domen
// o'zgarsa (local → production) eski yozuvlar buzilmaydi.
//
// Absolyut URL (masalan HEMIS'niki, avatar yuklab olish o'chirilganda saqlanadi)
// va bo'sh qiymat o'zgarishsiz qaytadi.
func absoluteMediaURL(base, u string) string {
	if u == "" || !strings.HasPrefix(u, "/") {
		return u
	}
	return base + u
}

// validationResponse — bind/validatsiya xatosini tilga moslangan javobga aylantiradi.
// Validator xatosi bo'lsa maydon-bo'yicha tarjima, aks holda umumiy xabar + texnik tafsilot.
func validationResponse(loc i18n.Locale, err error) dto.ErrorResponse {
	if fields := i18n.ValidationFields(loc, err); len(fields) > 0 {
		return dto.ErrValidation(i18n.T(loc, i18n.MsgValidationFailed), fields)
	}
	return dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error())
}
