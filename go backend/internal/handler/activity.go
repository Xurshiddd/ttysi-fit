package handler

import (
	"errors"
	"net/http"
	"time"

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

// ActivityHandler — foydalanuvchining kunlik faolligi (qadam/kaloriya/masofa).
type ActivityHandler struct {
	svc *service.ActivityService
	jwt *security.JWTManager
	log *zap.Logger
}

func NewActivityHandler(svc *service.ActivityService, jwt *security.JWTManager, log *zap.Logger) *ActivityHandler {
	return &ActivityHandler{svc: svc, jwt: jwt, log: log}
}

func (h *ActivityHandler) Register(r gin.IRouter) {
	g := r.Group("/activities")
	g.Use(middleware.Auth(h.jwt))
	{
		g.POST("", h.record)
		g.POST("/batch", h.recordBatch)
		g.GET("", h.list)
		g.GET("/stats", h.stats)
	}

	// Faollikni tuzatish — faqat admin (§17.3 #28).
	admin := r.Group("/admin")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.DELETE("/users/:id/activities", h.deleteRange)
	}
}

// deleteRange — foydalanuvchining oraliqdagi faolligini o'chiradi.
//
//	DELETE /admin/users/:id/activities?from=YYYY-MM-DD&to=YYYY-MM-DD
//
// Xato yoki soxta yozuvni tuzatish uchun: upsert GREATEST bilan ishlagani
// sababli katta qiymat qayta sinxron bilan tuzalmaydi.
func (h *ActivityHandler) deleteRange(c *gin.Context) {
	loc := middleware.GetLocale(c)

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	n, err := h.svc.DeleteRange(c.Request.Context(), userID, c.Query("from"), c.Query("to"))
	if err != nil {
		h.respondErr(c, loc, err, "delete activity range")
		return
	}

	// Audit (§17.3 #50): ma'lumot o'chirish — kim, kimga, qaysi oraliq.
	if h.log != nil {
		h.log.Info("admin faollikni o'chirdi",
			zap.String("admin_id", adminID.String()),
			zap.String("user_id", userID.String()),
			zap.String("from", c.Query("from")),
			zap.String("to", c.Query("to")),
			zap.Int64("rows", n),
		)
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{"deleted": n}))
}

// record — kunlik faollikni yozadi/yangilaydi.
func (h *ActivityHandler) record(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	var req dto.RecordActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	a, err := h.svc.Record(c.Request.Context(), uid, req)
	if err != nil {
		h.respondErr(c, loc, err, "record activity")
		return
	}
	c.JSON(http.StatusOK, dto.OK(a))
}

// recordBatch — bir necha kunlik faollikni bitta so'rovda yozadi (backfill).
// Mijoz ilova ochilganda telefondagi oxirgi kunlarni qayta yuboradi.
func (h *ActivityHandler) recordBatch(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	var req dto.RecordActivityBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	rows, err := h.svc.RecordBatch(c.Request.Context(), uid, req.Items)
	if err != nil {
		h.respondErr(c, loc, err, "record activity batch")
		return
	}
	c.JSON(http.StatusOK, dto.OK(rows))
}

// respondErr — domen xatolarini HTTP statuslarga o'giradi (§3.4: ichki
// tafsilot mijozga chiqmaydi, lekin mijoz tuzata oladigan xato aytiladi).
func (h *ActivityHandler) respondErr(c *gin.Context, loc i18n.Locale, err error, op string) {
	switch {
	case errors.Is(err, domain.ErrFutureDate),
		errors.Is(err, domain.ErrEmptyBatch),
		errors.Is(err, domain.ErrBatchTooLarge),
		errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
	default:
		logServerError(h.log, op, err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
	}
}

// list — oxirgi kunlar faolligi (?limit, ?from=YYYY-MM-DD, ?to=YYYY-MM-DD).
func (h *ActivityHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	limit := atoiDefault(c.Query("limit"), 31)
	// Chegaralar mahalliy mintaqada (APP_TIMEZONE) — UTC da hisoblansa
	// O'zbekistonda tunda "bugun" bir kun orqada qolardi.
	loc2 := h.svc.Location()
	to := h.svc.Today()
	from := to.AddDate(0, 0, -30)
	if v := c.Query("to"); v != "" {
		if t, e := time.ParseInLocation("2006-01-02", v, loc2); e == nil {
			to = t
		}
	}
	if v := c.Query("from"); v != "" {
		if t, e := time.ParseInLocation("2006-01-02", v, loc2); e == nil {
			from = t
		}
	}

	items, err := h.svc.List(c.Request.Context(), uid, from, to, limit)
	if err != nil {
		logServerError(h.log, "list activities", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(items))
}

// stats — bugun/hafta/oy/jami yig'ma.
func (h *ActivityHandler) stats(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	s, err := h.svc.Stats(c.Request.Context(), uid)
	if err != nil {
		logServerError(h.log, "activity stats", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(s))
}
