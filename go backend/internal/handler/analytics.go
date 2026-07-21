package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/i18n"
	"github.com/ttysi-fit/backend/internal/middleware"
	"github.com/ttysi-fit/backend/internal/service"
	"github.com/ttysi-fit/backend/pkg/report"
	"github.com/ttysi-fit/backend/pkg/security"
	"go.uber.org/zap"
)

// AnalyticsHandler — admin dashboard analitikasi va hisobot eksporti.
//
//	GET /admin/analytics?period=week|month|all&faculty_id=<uuid>
//	GET /admin/reports/users.csv?period=...&faculty_id=...
type AnalyticsHandler struct {
	svc *service.AnalyticsService
	jwt *security.JWTManager
	log *zap.Logger
}

func NewAnalyticsHandler(svc *service.AnalyticsService, jwt *security.JWTManager, log *zap.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc, jwt: jwt, log: log}
}

func (h *AnalyticsHandler) Register(r gin.IRouter) {
	// Analitika butun universitet ma'lumotini ochadi — faqat admin (§17.3 #28).
	g := r.Group("/admin")
	g.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		g.GET("/analytics", h.analytics)
		g.GET("/reports/users.csv", h.exportUsers)
	}
}

// filterFromQuery — query parametrlarini tekshirib AnalyticsFilter yasaydi.
func (h *AnalyticsHandler) filterFromQuery(c *gin.Context) (domain.AnalyticsFilter, error) {
	period := c.DefaultQuery("period", string(domain.PeriodWeek))

	var facultyID *uuid.UUID
	if v := c.Query("faculty_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return domain.AnalyticsFilter{}, fmt.Errorf("%w: faculty_id", domain.ErrValidation)
		}
		facultyID = &id
	}

	return h.svc.Filter(period, facultyID)
}

// analytics — dashboard uchun umumiy raqamlar, kunlik dinamika va
// fakultetlar kesimi (bitta so'rovda).
func (h *AnalyticsHandler) analytics(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f, err := h.filterFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	// Og'ir agregatsiya — timeout bilan (§3.3).
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	res, err := h.svc.Get(ctx, f)
	if err != nil {
		if errors.Is(err, domain.ErrValidation) {
			c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
			return
		}
		logServerError(h.log, "analytics", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

// exportUsers — foydalanuvchilar kesimidagi faollik hisoboti (CSV).
//
// Javob STREAM qilinadi: qatorlar DB dan kelgani sari yoziladi, butun
// hisobot xotirada yig'ilmaydi (§17.3 #39 — resurs sarfi).
func (h *AnalyticsHandler) exportUsers(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f, err := h.filterFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	filename := fmt.Sprintf("ttysi_fit_hisobot_%s_%s.csv",
		f.From.Format("2006-01-02"), f.To.Format("2006-01-02"))

	c.Header("Content-Type", "text/csv; charset=utf-8")
	// Fayl nomi kod ichida yasaladi (mijoz kiritmaydi) — header injection yo'q.
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Header("X-Content-Type-Options", "nosniff")

	w := report.NewWriter(c.Writer)
	if err := w.Write("F.I.O.", "Email", "Rol", "Fakultet", "Kafedra",
		"Guruh", "Jami qadam", "Masofa (km)", "Faol kunlar"); err != nil {
		logServerError(h.log, "export users: sarlavha", err)
		return
	}

	err = h.svc.StreamUserActivity(ctx, f, func(row domain.UserActivityRow) error {
		return w.Write(
			row.FullName,
			row.Email,
			row.Role,
			row.Faculty,
			row.Department,
			row.GroupName,
			strconv.FormatInt(row.TotalSteps, 10),
			strconv.FormatFloat(row.DistanceKm, 'f', 2, 64),
			strconv.FormatInt(row.ActiveDays, 10),
		)
	})
	if err != nil {
		// Sarlavhalar allaqachon yuborilgan — status kodini o'zgartirib
		// bo'lmaydi. Xatoni loglaymiz, fayl chala qoladi.
		logServerError(h.log, "export users", err)
		return
	}

	if err := w.Flush(); err != nil {
		logServerError(h.log, "export users: flush", err)
	}
}
