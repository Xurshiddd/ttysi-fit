package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS — ruxsat etilgan originlarga ruxsat beradi (CLAUDE.md 3.2).
// allowLocalhostDev=true bo'lsa (dev rejim), har qanday localhost/127.0.0.1
// origin'ga ruxsat beriladi — Flutter web random portda ishlaganda qulay.
func CORS(allowed []string, allowLocalhostDev bool) gin.HandlerFunc {
	allowSet := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		allowSet[o] = struct{}{}
	}

	isLocalhost := func(origin string) bool {
		return strings.HasPrefix(origin, "http://localhost:") ||
			strings.HasPrefix(origin, "http://127.0.0.1:") ||
			origin == "http://localhost" || origin == "http://127.0.0.1"
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		_, allowedExact := allowSet[origin]
		allow := origin != "" && (allowedExact || (allowLocalhostDev && isLocalhost(origin)))

		if allow {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept-Language")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
