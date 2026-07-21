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
	"gorm.io/datatypes"
)

// CompetitionHandler — musobaqalar (§16.3).
//
//	GET    /competition-types
//	GET    /competitions?status=registration
//	GET    /competitions/:id
//	POST   /competitions/:id/register
//	DELETE /competitions/:id/register      — ishtirokni bekor qilish
//	GET    /admin/competitions
//	POST   /admin/competitions
//	PUT    /admin/competitions/:id
//	DELETE /admin/competitions/:id
//	GET    /admin/competitions/:id/participants
type CompetitionHandler struct {
	svc *service.CompetitionService
	jwt *security.JWTManager
	log *zap.Logger
}

func NewCompetitionHandler(svc *service.CompetitionService, jwt *security.JWTManager, log *zap.Logger) *CompetitionHandler {
	return &CompetitionHandler{svc: svc, jwt: jwt, log: log}
}

func (h *CompetitionHandler) Register(r gin.IRouter) {
	types := r.Group("/competition-types")
	types.Use(middleware.Auth(h.jwt))
	{
		types.GET("", h.types)
	}

	g := r.Group("/competitions")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("", h.list)
		g.GET("/:id", h.get)
		g.POST("/:id/register", h.register)
		g.DELETE("/:id/register", h.cancel)
	}

	admin := r.Group("/admin/competitions")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.GET("", h.adminList)
		admin.POST("", h.create)
		admin.PUT("/:id", h.update)
		admin.DELETE("/:id", h.remove)
		admin.GET("/:id/participants", h.participants)
	}
}

func (h *CompetitionHandler) types(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(h.svc.Types()))
}

// list — mobil ro'yxat. Draft mobil foydalanuvchiga ko'rinmaydi.
func (h *CompetitionHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	f := domain.CompetitionFilter{
		Status: c.Query("status"),
		Type:   c.Query("type"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	if (f.Status != "" && !domain.ValidCompetitionStatus(f.Status)) ||
		(f.Type != "" && !domain.ValidCompetitionType(f.Type)) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if f.Status == domain.CompStatusDraft {
		c.JSON(http.StatusForbidden, dto.ErrResponse(i18n.T(loc, i18n.MsgForbidden)))
		return
	}

	items, total, err := h.svc.ListForUser(c.Request.Context(), uid, f)
	if err != nil {
		logServerError(h.log, "competition list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	// Mobil foydalanuvchi draft'ni ko'rmasin (status bo'sh so'ralganda ham).
	if f.Status == "" && middleware.GetRole(c) != string(domain.RoleAdmin) {
		filtered := items[:0]
		for _, v := range items {
			if v.Status != domain.CompStatusDraft {
				filtered = append(filtered, v)
			}
		}
		items = filtered
	}

	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *CompetitionHandler) get(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	comp, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.compErr(c, loc, err, "competition get")
		return
	}
	if comp.Status == domain.CompStatusDraft && middleware.GetRole(c) != string(domain.RoleAdmin) {
		c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(comp))
}

func (h *CompetitionHandler) register(c *gin.Context) {
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

	reg, err := h.svc.Register(c.Request.Context(), uid, id)
	if err != nil {
		h.compErr(c, loc, err, "competition register")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(reg))
}

func (h *CompetitionHandler) cancel(c *gin.Context) {
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

	if err := h.svc.Cancel(c.Request.Context(), uid, id); err != nil {
		h.compErr(c, loc, err, "competition cancel")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CompetitionHandler) adminList(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.CompetitionFilter{
		Status: c.Query("status"),
		Type:   c.Query("type"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	if (f.Status != "" && !domain.ValidCompetitionStatus(f.Status)) ||
		(f.Type != "" && !domain.ValidCompetitionType(f.Type)) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "competition admin list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *CompetitionHandler) create(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.CompetitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	comp, err := competitionFromRequest(&req, uuid.Nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
		return
	}
	if err := h.svc.Create(c.Request.Context(), comp); err != nil {
		h.compErr(c, loc, err, "competition create")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(comp))
}

func (h *CompetitionHandler) update(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	var req dto.CompetitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	comp, err := competitionFromRequest(&req, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
		return
	}
	if err := h.svc.Update(c.Request.Context(), comp); err != nil {
		h.compErr(c, loc, err, "competition update")
		return
	}
	c.JSON(http.StatusOK, dto.OK(comp))
}

func (h *CompetitionHandler) remove(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.compErr(c, loc, err, "competition delete")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CompetitionHandler) participants(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	page := atoiDefault(c.Query("page"), 1)
	limit := atoiDefault(c.Query("limit"), 20)

	rows, total, err := h.svc.Participants(c.Request.Context(), id, page, limit)
	if err != nil {
		logServerError(h.log, "competition participants", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.Paginated(rows, dto.Meta{Page: page, Limit: limit, Total: total}))
}

func (h *CompetitionHandler) compErr(c *gin.Context, loc i18n.Locale, err error, op string) {
	var cfgErr *domain.ErrChallengeConfig
	switch {
	case errors.As(err, &cfgErr):
		c.JSON(http.StatusBadRequest, dto.ErrValidation(
			i18n.T(loc, i18n.MsgValidationFailed),
			map[string]string{cfgErr.Field: cfgErr.Reason},
		))
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, dto.ErrResponse(i18n.T(loc, i18n.MsgCompAlreadyRegistered)))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
	default:
		logServerError(h.log, op, err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
	}
}

// competitionFromRequest — DTO -> domen. Mijoz id/created_at yubora olmaydi (§17.3 #13).
func competitionFromRequest(req *dto.CompetitionRequest, id uuid.UUID) (*domain.Competition, error) {
	comp := &domain.Competition{
		ID:              id,
		Type:            domain.CompetitionType(req.Type),
		Title:           req.Title,
		Description:     req.Description,
		Scope:           req.Scope,
		Status:          req.Status,
		Location:        req.Location,
		MaxParticipants: req.MaxParticipants,
		RewardCoins:     req.RewardCoins,
		CoverURL:        req.CoverURL,
	}
	if comp.Scope == "" {
		comp.Scope = domain.ChallengeScopeUniversity
	}
	if comp.Status == "" {
		comp.Status = domain.CompStatusDraft
	}
	if len(req.Config) == 0 {
		comp.Config = datatypes.JSON([]byte("{}"))
	} else {
		comp.Config = datatypes.JSON(req.Config)
	}

	for _, p := range []struct {
		src *string
		dst **time.Time
	}{
		{req.StartsAt, &comp.StartsAt},
		{req.EndsAt, &comp.EndsAt},
		{req.RegEndsAt, &comp.RegEndsAt},
	} {
		if p.src == nil || *p.src == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, *p.src)
		if err != nil {
			return nil, err
		}
		*p.dst = &t
	}
	return comp, nil
}
