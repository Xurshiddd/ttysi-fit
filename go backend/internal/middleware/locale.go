package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ttysi-fit/backend/internal/i18n"
)

const ctxLocale = "locale"

// Locale — har bir so'rov uchun tilni aniqlaydi.
// Ustuvorlik: ?lang query > Accept-Language header > default (uz).
// Foydalanuvchining DB dagi tili server-initiated xabarlarda alohida ishlatiladi.
func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		var loc i18n.Locale = i18n.Default

		if q := c.Query("lang"); q != "" {
			if l, ok := i18n.Parse(q); ok {
				loc = l
			}
		} else if h := c.GetHeader("Accept-Language"); h != "" {
			loc = i18n.ParseAcceptLanguage(h)
		}

		c.Set(ctxLocale, string(loc))
		c.Next()
	}
}

// GetLocale — kontekstdan so'rov tilini oladi.
func GetLocale(c *gin.Context) i18n.Locale {
	if v, ok := c.Get(ctxLocale); ok {
		if s, ok := v.(string); ok && s != "" {
			return i18n.Locale(s)
		}
	}
	return i18n.Default
}
