package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/i18n"
	"github.com/ttysi-fit/backend/pkg/security"
)

const (
	ctxUserID = "user_id"
	ctxRole   = "role"
)

// Auth — access tokenni tekshiruvchi middleware.
func Auth(jwt *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		loc := GetLocale(c)

		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgMissingAuthHeader)))
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwt.ParseAccess(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgTokenInvalid)))
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// RequireRole — faqat ruxsat etilgan rollarga yo'l ochadi.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := c.GetString(ctxRole)
		if _, ok := allowed[role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrResponse(i18n.T(GetLocale(c), i18n.MsgForbidden)))
			return
		}
		c.Next()
	}
}

// GetRole — kontekstdan foydalanuvchi rolini oladi (Auth o'rnatadi).
// Auth ishlamagan bo'lsa bo'sh satr — ya'ni hech qaysi rolga teng emas.
func GetRole(c *gin.Context) string {
	return c.GetString(ctxRole)
}

// GetUserID — kontekstdan foydalanuvchi ID sini oladi.
func GetUserID(c *gin.Context) (uuid.UUID, error) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return uuid.Nil, security.ErrInvalidToken
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, security.ErrInvalidToken
	}
	return id, nil
}
