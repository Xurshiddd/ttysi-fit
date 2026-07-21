package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders — xavfsizlik header'lari (CLAUDE.md §17.3 #42/#43/#44).
//
// API va statik fayllar (avatarlar) uchun:
//   - nosniff — MIME sniffing bloklanadi (r.Static orqali xizmat qilinadigan
//     rasmlar brauzerda skript sifatida talqin qilinmasligi uchun muhim)
//   - X-Frame-Options + CSP frame-ancestors — clickjacking himoyasi
//   - qattiq CSP — API javoblari hech qanday resurs yuklamaydi
//
// HSTS faqat production'da (TLS ortida) qo'shiladi — local HTTP'ni buzmaslik
// uchun (§17.1 — bu yumshatish emas, HSTS HTTP'da ma'nosiz).
func SecurityHeaders(isProd bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Permitted-Cross-Domain-Policies", "none")
		if isProd {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}
