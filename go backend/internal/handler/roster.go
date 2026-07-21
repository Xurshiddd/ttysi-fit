package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/i18n"
	"github.com/ttysi-fit/backend/internal/middleware"
	"github.com/ttysi-fit/backend/internal/service"
	"github.com/ttysi-fit/backend/pkg/security"
	"go.uber.org/zap"
)

// RosterHandler — guruhlar ro'yxati va roster (talaba/o'qituvchi) sync.
type RosterHandler struct {
	svc *service.RosterService
	jwt *security.JWTManager
	log *zap.Logger
}

func NewRosterHandler(svc *service.RosterService, jwt *security.JWTManager, log *zap.Logger) *RosterHandler {
	return &RosterHandler{svc: svc, jwt: jwt, log: log}
}

func (h *RosterHandler) Register(r gin.IRouter) {
	r.GET("/groups", h.listGroups)

	admin := r.Group("/admin/hemis")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.POST("/sync/groups", h.sync(h.svc.SyncGroups))
		admin.POST("/sync/students", h.sync(h.svc.SyncStudents))
		admin.POST("/sync/employees", h.sync(h.svc.SyncEmployees))
	}
}

func (h *RosterHandler) listGroups(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var facultyID *uuid.UUID
	if q := c.Query("faculty_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
			return
		}
		facultyID = &id
	}

	groups, err := h.svc.ListGroups(c.Request.Context(), facultyID)
	if err != nil {
		logServerError(h.log, "list groups", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	out := make([]structureBrief, 0, len(groups))
	for _, g := range groups {
		out = append(out, structureBrief{ID: g.ID.String(), Name: g.Name, Code: ""})
	}
	c.JSON(http.StatusOK, dto.OK(out))
}

// sync — admin sync endpointlari uchun umumiy o'rovchi.
func (h *RosterHandler) sync(fn func(ctx context.Context) (*domain.SyncStats, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		loc := middleware.GetLocale(c)
		stats, err := fn(c.Request.Context())
		if err != nil {
			logServerError(h.log, "hemis roster sync", err)
			c.JSON(http.StatusBadGateway, dto.ErrResponse(i18n.T(loc, i18n.MsgHemisError)))
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":    stats,
			"message": i18n.T(loc, i18n.MsgSyncSuccess),
		})
	}
}
