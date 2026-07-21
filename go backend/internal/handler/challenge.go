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

// ChallengeHandler — chellenjlar (§16: kontent admin panel orqali boshqariladi).
//
//	GET    /challenge-types                 — tur ta'riflari (dinamik forma uchun)
//	GET    /challenges?status=active&page=1 — mobil ro'yxat (foydalanuvchi holati bilan)
//	GET    /challenges/:id
//	POST   /challenges/:id/join
//	GET    /admin/challenges                — admin ro'yxati (draft'lar ham)
//	POST   /admin/challenges
//	PUT    /admin/challenges/:id
//	DELETE /admin/challenges/:id
type ChallengeHandler struct {
	svc *service.ChallengeService
	jwt *security.JWTManager
	log *zap.Logger
}

func NewChallengeHandler(svc *service.ChallengeService, jwt *security.JWTManager, log *zap.Logger) *ChallengeHandler {
	return &ChallengeHandler{svc: svc, jwt: jwt, log: log}
}

func (h *ChallengeHandler) Register(r gin.IRouter) {
	// Tur ta'riflari — admin panel formasi shundan yasaladi.
	types := r.Group("/challenge-types")
	types.Use(middleware.Auth(h.jwt))
	{
		types.GET("", h.types)
	}

	g := r.Group("/challenges")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("", h.list)
		g.GET("/:id", h.get)
		g.POST("/:id/join", h.join)
	}

	admin := r.Group("/admin/challenges")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.GET("", h.adminList)
		admin.POST("", h.create)
		admin.PUT("/:id", h.update)
		admin.DELETE("/:id", h.remove)
	}
}

func (h *ChallengeHandler) types(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(h.svc.Types()))
}

// list — mobil ro'yxat. Default `active`: mobil foydalanuvchi draft'ni ko'rmaydi.
func (h *ChallengeHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	f := domain.ChallengeFilter{
		Status: c.DefaultQuery("status", domain.ChallengeStatusActive),
		Type:   c.Query("type"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	// Enum validatsiya — SQL'ga faqat tekshirilgan qiymat boradi (§3.2).
	if !domain.ValidChallengeStatus(f.Status) ||
		(f.Type != "" && !domain.ValidChallengeType(f.Type)) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	// Mobil foydalanuvchi draft ko'rmasin (§17.3 #26/#29).
	if f.Status == domain.ChallengeStatusDraft {
		c.JSON(http.StatusForbidden, dto.ErrResponse(i18n.T(loc, i18n.MsgForbidden)))
		return
	}

	items, total, err := h.svc.ListForUser(c.Request.Context(), uid, f)
	if err != nil {
		logServerError(h.log, "challenge list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *ChallengeHandler) get(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	ch, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.respondErr(c, loc, err, "challenge get")
		return
	}
	// Draft faqat admin uchun.
	if ch.Status == domain.ChallengeStatusDraft && middleware.GetRole(c) != string(domain.RoleAdmin) {
		c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(ch))
}

func (h *ChallengeHandler) join(c *gin.Context) {
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

	uc, err := h.svc.Join(c.Request.Context(), uid, id)
	if err != nil {
		h.respondErr(c, loc, err, "challenge join")
		return
	}
	c.JSON(http.StatusOK, dto.OK(uc))
}

func (h *ChallengeHandler) adminList(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.ChallengeFilter{
		Status: c.Query("status"), // bo'sh — hammasi (draft'lar ham)
		Type:   c.Query("type"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	if (f.Status != "" && !domain.ValidChallengeStatus(f.Status)) ||
		(f.Type != "" && !domain.ValidChallengeType(f.Type)) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "challenge admin list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *ChallengeHandler) create(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.ChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	ch, err := challengeFromRequest(&req, uuid.Nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
		return
	}

	if err := h.svc.Create(c.Request.Context(), ch); err != nil {
		h.respondErr(c, loc, err, "challenge create")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(ch))
}

func (h *ChallengeHandler) update(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	var req dto.ChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	ch, err := challengeFromRequest(&req, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
		return
	}

	if err := h.svc.Update(c.Request.Context(), ch); err != nil {
		h.respondErr(c, loc, err, "challenge update")
		return
	}
	c.JSON(http.StatusOK, dto.OK(ch))
}

func (h *ChallengeHandler) remove(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.respondErr(c, loc, err, "challenge delete")
		return
	}
	c.Status(http.StatusNoContent)
}

// respondErr — domen xatolarini HTTP statuslarga o'giradi. Ichki tafsilot
// mijozga chiqmaydi (§3.4), config xatosi esa adminga foydali — u ko'rsatiladi.
func (h *ChallengeHandler) respondErr(c *gin.Context, loc i18n.Locale, err error, op string) {
	var cfgErr *domain.ErrChallengeConfig
	switch {
	case errors.As(err, &cfgErr):
		c.JSON(http.StatusBadRequest, dto.ErrValidation(
			i18n.T(loc, i18n.MsgValidationFailed),
			map[string]string{cfgErr.Field: cfgErr.Reason},
		))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
	default:
		logServerError(h.log, op, err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
	}
}

// challengeFromRequest — DTO'dan domen obyektiga. Mijoz `id`, `created_at`
// kabilarni yubora olmaydi — ular bu yerda o'rnatiladi (§17.3 #13).
func challengeFromRequest(req *dto.ChallengeRequest, id uuid.UUID) (*domain.Challenge, error) {
	ch := &domain.Challenge{
		ID:          id,
		Type:        domain.ChallengeType(req.Type),
		Title:       req.Title,
		Description: req.Description,
		Scope:       req.Scope,
		Status:      req.Status,
		RewardCoins: req.RewardCoins,
		CoverURL:    req.CoverURL,
	}
	if ch.Scope == "" {
		ch.Scope = domain.ChallengeScopeUniversity
	}
	if ch.Status == "" {
		ch.Status = domain.ChallengeStatusDraft
	}
	if len(req.Config) == 0 {
		ch.Config = datatypes.JSON([]byte("{}"))
	} else {
		ch.Config = datatypes.JSON(req.Config)
	}

	if req.StartsAt != nil && *req.StartsAt != "" {
		t, err := time.Parse(time.RFC3339, *req.StartsAt)
		if err != nil {
			return nil, err
		}
		ch.StartsAt = &t
	}
	if req.EndsAt != nil && *req.EndsAt != "" {
		t, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			return nil, err
		}
		ch.EndsAt = &t
	}
	return ch, nil
}
