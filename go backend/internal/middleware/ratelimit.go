package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/i18n"
	"go.uber.org/zap"
)

// RateLimitOpts — bitta limiter sozlamalari.
type RateLimitOpts struct {
	Name   string        // kalit nomi qismi (masalan "global", "auth")
	Limit  int           // Window ichida IP boshiga ruxsat etilgan maksimal so'rov
	Window time.Duration // hisoblash oynasi (masalan 1 daqiqa)
}

// RateLimit — Redis-backed, IP asosidagi inbound rate limiting
// (CLAUDE.md §17.3 #15/#16/#40 — brute-force / credential stuffing himoyasi).
//
// Kalit: ratelimit:{ip}:{name} (§12.3 konvensiyasi). Algoritm: fixed-window —
// pipeline'da INCR + ExpireNX (TTL har doim o'rnatilishi kafolatlanadi,
// bitta roundtrip). Limit oshsa 429 + Retry-After qaytadi.
//
// Redis xatosi so'rovni BLOKLAMAYDI (fail-open, §12.5 — Redis xatosi cache
// miss sifatida qaraladi) — faqat loglanadi.
func RateLimit(rdb *redis.Client, log *zap.Logger, o RateLimitOpts) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := "ratelimit:" + c.ClientIP() + ":" + o.Name

		pipe := rdb.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.ExpireNX(ctx, key, o.Window)
		if _, err := pipe.Exec(ctx); err != nil {
			log.Warn("ratelimit: redis xatosi (fail-open)", zap.Error(err))
			c.Next()
			return
		}

		if incr.Val() > int64(o.Limit) {
			if ttl, err := rdb.TTL(ctx, key).Result(); err == nil && ttl > 0 {
				c.Header("Retry-After", strconv.Itoa(int(ttl.Seconds())+1))
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				dto.ErrResponse(i18n.T(GetLocale(c), i18n.MsgTooManyRequests)))
			return
		}
		c.Next()
	}
}
