package handler

import (
	"errors"
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

// NotificationHandler — ilova ichidagi bildirishnomalar.
//
//	GET  /notifications?unread=1&page=&limit=
//	GET  /notifications/unread-count
//	POST /notifications/:id/read
//	POST /notifications/read-all
//
//	POST /admin/notifications        — e'lon yuborish
type NotificationHandler struct {
	svc *service.NotificationService
	jwt *security.JWTManager
	log *zap.Logger
}

func NewNotificationHandler(svc *service.NotificationService, jwt *security.JWTManager, log *zap.Logger) *NotificationHandler {
	return &NotificationHandler{svc: svc, jwt: jwt, log: log}
}

func (h *NotificationHandler) Register(r gin.IRouter) {
	g := r.Group("/notifications")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("", h.list)
		g.GET("/unread-count", h.unreadCount)
		g.POST("/:id/read", h.markRead)
		g.POST("/read-all", h.markAllRead)
	}

	admin := r.Group("/admin/notifications")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.POST("", h.announce)
	}
}

func (h *NotificationHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	f := domain.NotificationFilter{
		UnreadOnly: c.Query("unread") == "1",
		Page:       atoiDefault(c.Query("page"), 1),
		Limit:      atoiDefault(c.Query("limit"), 20),
	}

	rows, total, err := h.svc.List(c.Request.Context(), uid, f)
	if err != nil {
		logServerError(h.log, "list notifications", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.Paginated(rows, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

// unreadCount — qo'ng'iroq nishoni. Mobil ilova buni tez-tez so'raydi,
// shuning uchun javob iloji boricha yengil (faqat son).
func (h *NotificationHandler) unreadCount(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	n, err := h.svc.UnreadCount(c.Request.Context(), uid)
	if err != nil {
		logServerError(h.log, "unread count", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"unread": n}))
}

func (h *NotificationHandler) markRead(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	if err := h.svc.MarkRead(c.Request.Context(), uid, id); err != nil {
		logServerError(h.log, "mark read", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NotificationHandler) markAllRead(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	n, err := h.svc.MarkAllRead(c.Request.Context(), uid)
	if err != nil {
		logServerError(h.log, "mark all read", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"read": n}))
}

// announce — admin e'loni (fakultet/guruh/rol bo'yicha yoki hammaga).
func (h *NotificationHandler) announce(c *gin.Context) {
	loc := middleware.GetLocale(c)

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	var req dto.AnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	t := domain.AnnouncementTarget{Role: req.Role}
	if req.FacultyID != "" {
		id, err := uuid.Parse(req.FacultyID)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
			return
		}
		t.FacultyID = &id
	}
	if req.GroupID != "" {
		id, err := uuid.Parse(req.GroupID)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
			return
		}
		t.GroupID = &id
	}

	sent, err := h.svc.Announce(c.Request.Context(), t, req.Title, req.Body)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, dto.ErrDetailed(
				i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
			return
		}
		logServerError(h.log, "announce", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	// Audit (§17.3 #50): ommaviy xabar — kim, kimlarga, nechta.
	if h.log != nil {
		h.log.Info("admin e'lon yubordi",
			zap.String("admin_id", adminID.String()),
			zap.String("title", req.Title),
			zap.Int64("sent", sent),
		)
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{"sent": sent}))
}
