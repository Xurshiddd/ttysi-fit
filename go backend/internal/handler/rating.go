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

// RatingHandler — elektron sport reytingi endpointlari.
//
//	GET /ratings?type=student&period=week&faculty_id=&group_id=&page=1&limit=20
//	GET /ratings/me?period=week
type RatingHandler struct {
	svc       *service.RatingService
	jwt       *security.JWTManager
	log       *zap.Logger
	mediaBase string // avatar nisbiy yo'liga qo'shiladigan ommaviy asos
}

func NewRatingHandler(svc *service.RatingService, jwt *security.JWTManager, log *zap.Logger, mediaBase string) *RatingHandler {
	return &RatingHandler{svc: svc, jwt: jwt, log: log, mediaBase: mediaBase}
}

func (h *RatingHandler) Register(r gin.IRouter) {
	g := r.Group("/ratings")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("", h.list)
		g.GET("/me", h.me)
	}
}

// list — reyting jadvali (kesim + davr + filtr + paginatsiya).
func (h *RatingHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.RatingFilter{
		Type:   domain.RatingType(c.DefaultQuery("type", string(domain.RatingStudent))),
		Period: domain.RatingPeriod(c.DefaultQuery("period", string(domain.PeriodWeek))),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}

	// Enum validatsiya — SQL'ga faqat tekshirilgan qiymat boradi (§3.2).
	if !domain.ValidRatingType(string(f.Type)) || !domain.ValidRatingPeriod(string(f.Period)) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	if v := c.Query("faculty_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
			return
		}
		f.FacultyID = &id
	}
	if v := c.Query("group_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
			return
		}
		f.GroupID = &id
	}

	res, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "rating list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	for i := range res.Entries {
		res.Entries[i].AvatarURL = absoluteMediaURL(h.mediaBase, res.Entries[i].AvatarURL)
	}

	c.JSON(http.StatusOK, dto.Paginated(res.Entries, dto.Meta{
		Page: f.Page, Limit: f.Limit, Total: res.Total,
	}))
}

// me — foydalanuvchining o'z o'rni (umumiy + fakultet ichida).
func (h *RatingHandler) me(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	period := domain.RatingPeriod(c.DefaultQuery("period", string(domain.PeriodWeek)))
	if !domain.ValidRatingPeriod(string(period)) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	m, err := h.svc.MyRank(c.Request.Context(), uid, period)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
			return
		}
		logServerError(h.log, "rating me", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(m))
}
