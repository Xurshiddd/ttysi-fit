package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/i18n"
	"github.com/ttysi-fit/backend/pkg/security"
	"go.uber.org/zap"
)

// DeviceHeader — mijoz har so'rovda yuboradigan qurilma identifikatori.
const DeviceHeader = "X-Device-Id"

// NoDevice — "faol qurilma yo'q" belgisi.
//
// Sessiya o'chirilganda kalit O'CHIRILMAYDI, shu qiymatga qo'yiladi:
// yo'q kalit "siyosat ishlatilmagan" degani va so'rov o'tkazib
// yuborilardi. Bo'sh joy va `/` qurilma ID sida uchramaydi (ID —
// faqat harf va raqam).
const NoDevice = "-"

// SessionDeviceKey — foydalanuvchining JORIY qurilmasi.
//
// Bitta kalit yetarli, chunki siyosat "bir hisob — bir qurilma": boshqa
// qurilmada kirilganda bu qiymat almashadi va eski qurilma darrov chiqadi.
func SessionDeviceKey(userID string) string { return "session_device:" + userID }

// DeviceSession — bekor qilingan qurilmani DARROV chiqaradi.
//
// NEGA KERAK: JWT stateless — sessiya yopilganda ham access token imzosi
// yaroqli qolaveradi va eski qurilma yana 15 daqiqa (access TTL) ishlar edi.
// Foydalanuvchi esa "chiqarib yuborildi" deb kutadi.
//
// Global middleware sifatida ishlaydi va Auth dan MUSTAQIL: tokenni o'zi
// o'qiydi, shuning uchun mavjud route ro'yxatlarini o'zgartirish shart emas.
//
// Tekshiruv FAQAT `X-Device-Id` yuborilganda ishlaydi:
//   - admin panel (brauzer) qurilma cheklovidan tashqarida;
//   - eski mobil versiyalar ishlashda davom etadi.
//
// Redis xatosida O'TKAZIB YUBORADI (fail-open): bu asosiy auth nazorati
// emas (u — JWT imzosi), Redis uzilganda esa butun ilova qulashi
// bekor qilingan bitta sessiyadan ko'ra yomonroq.
func DeviceSession(jwt *security.JWTManager, rdb *redis.Client, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID := c.GetHeader(DeviceHeader)
		if deviceID == "" || rdb == nil {
			c.Next()
			return
		}

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.Next() // himoyalanmagan endpoint — Auth o'z ishini qiladi
			return
		}

		claims, err := jwt.ParseAccess(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.Next() // yaroqsiz token — Auth 401 qaytaradi
			return
		}

		current, err := rdb.Get(c.Request.Context(),
			SessionDeviceKey(claims.UserID.String())).Result()
		if err != nil {
			// Kalit yo'q (qurilma siyosati ishlatilmagan) yoki Redis xatosi.
			if err != redis.Nil && log != nil {
				log.Warn("qurilma sessiyasi tekshirilmadi", zap.Error(err))
			}
			c.Next()
			return
		}

		if current != deviceID {
			loc := GetLocale(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				dto.ErrResponse(i18n.T(loc, i18n.MsgSessionRevoked)))
			return
		}
		c.Next()
	}
}
