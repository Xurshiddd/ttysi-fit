package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/i18n"
)

// BodyLimit — kiruvchi request body hajmini cheklaydi (DoS himoyasi,
// CLAUDE.md §17.3 #39 — resurs sarfi).
//
// Ikki qatlam:
//  1. Content-Length ma'lum bo'lsa — darhol 413 (body o'qilmasdan).
//  2. Chunked yoki yolg'on Content-Length uchun — MaxBytesReader qattiq
//     chegara qo'yadi (oshsa o'qish xato beradi, bind 400 qaytaradi).
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge,
				dto.ErrResponse(i18n.T(GetLocale(c), i18n.MsgRequestTooLarge)))
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
