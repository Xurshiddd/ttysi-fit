package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ttysi-fit/backend/internal/dto"
	"gorm.io/gorm"
)

// HealthHandler — servis va bog'liqliklar sog'ligini tekshiradi.
type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: rdb}
}

// Register — health route'larini ro'yxatdan o'tkazadi.
func (h *HealthHandler) Register(r gin.IRouter) {
	r.GET("/health", h.Check)
}

// Check — Postgres va Redis holatini qaytaradi.
func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	status := gin.H{
		"status":   "ok",
		"postgres": "up",
		"redis":    "up",
	}
	httpStatus := http.StatusOK

	if sqlDB, err := h.db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
		status["postgres"] = "down"
		status["status"] = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		status["redis"] = "down"
		status["status"] = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, dto.OK(status))
}
