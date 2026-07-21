package handler

import (
	"encoding/json"
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

// RewardHandler — FIT Coin do'koni.
//
//	GET  /rewards                     — do'kon ro'yxati (mobil)
//	GET  /rewards/:id
//	POST /rewards/:id/redeem          — coinga almashtirish
//	GET  /rewards/my-redemptions      — mening buyurtmalarim
//	GET  /reward-categories           — dinamik forma uchun
//
//	GET    /admin/rewards             — hammasi (nofaol ham)
//	POST   /admin/rewards
//	PUT    /admin/rewards/:id
//	DELETE /admin/rewards/:id
//	GET    /admin/redemptions
//	POST   /admin/redemptions/:id/issue
//	POST   /admin/redemptions/:id/cancel
type RewardHandler struct {
	svc       *service.RewardService
	jwt       *security.JWTManager
	log       *zap.Logger
	mediaBase string
}

func NewRewardHandler(svc *service.RewardService, jwt *security.JWTManager, log *zap.Logger, mediaBase string) *RewardHandler {
	return &RewardHandler{svc: svc, jwt: jwt, log: log, mediaBase: mediaBase}
}

func (h *RewardHandler) Register(r gin.IRouter) {
	g := r.Group("")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("/reward-categories", h.categories)
		g.GET("/rewards", h.list)
		g.GET("/rewards/:id", h.get)
		g.POST("/rewards/:id/redeem", h.redeem)
		g.GET("/my-redemptions", h.myRedemptions)
	}

	admin := r.Group("/admin")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.GET("/rewards", h.adminList)
		admin.POST("/rewards", h.create)
		admin.PUT("/rewards/:id", h.update)
		admin.DELETE("/rewards/:id", h.remove)

		admin.GET("/redemptions", h.adminRedemptions)
		admin.POST("/redemptions/:id/issue", h.issue)
		admin.POST("/redemptions/:id/cancel", h.cancel)
	}
}

func (h *RewardHandler) categories(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(h.svc.Categories()))
}

// list — mobil do'kon: faqat ayni damda olinishi mumkin bo'lganlar.
func (h *RewardHandler) list(c *gin.Context) {
	h.listWith(c, true)
}

// adminList — admin: nofaol va tugaganlar ham ko'rinadi.
func (h *RewardHandler) adminList(c *gin.Context) {
	h.listWith(c, false)
}

func (h *RewardHandler) listWith(c *gin.Context, onlyAvailable bool) {
	loc := middleware.GetLocale(c)

	f := domain.RewardFilter{
		Category:      c.Query("category"),
		OnlyAvailable: onlyAvailable,
		Page:          atoiDefault(c.Query("page"), 1),
		Limit:         atoiDefault(c.Query("limit"), 20),
	}

	rows, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		h.respondErr(c, loc, err, "list rewards")
		return
	}
	for i := range rows {
		rows[i].ImageURL = absoluteMediaURL(h.mediaBase, rows[i].ImageURL)
	}
	c.JSON(http.StatusOK, dto.Paginated(rows, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *RewardHandler) get(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	rw, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		h.respondErr(c, loc, err, "get reward")
		return
	}
	rw.ImageURL = absoluteMediaURL(h.mediaBase, rw.ImageURL)
	c.JSON(http.StatusOK, dto.OK(rw))
}

// redeem — sovg'ani coinga almashtirish.
func (h *RewardHandler) redeem(c *gin.Context) {
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

	red, err := h.svc.Redeem(c.Request.Context(), uid, id)
	if err != nil {
		h.respondErr(c, loc, err, "redeem reward")
		return
	}
	c.JSON(http.StatusOK, dto.OK(red))
}

func (h *RewardHandler) myRedemptions(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	f := domain.RedemptionFilter{
		Status: c.Query("status"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	rows, total, err := h.svc.MyRedemptions(c.Request.Context(), uid, f)
	if err != nil {
		h.respondErr(c, loc, err, "my redemptions")
		return
	}
	for i := range rows {
		rows[i].RewardImageURL = absoluteMediaURL(h.mediaBase, rows[i].RewardImageURL)
	}
	c.JSON(http.StatusOK, dto.Paginated(rows, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *RewardHandler) adminRedemptions(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.RedemptionFilter{
		Status: c.Query("status"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	rows, total, err := h.svc.ListRedemptions(c.Request.Context(), f)
	if err != nil {
		h.respondErr(c, loc, err, "admin redemptions")
		return
	}
	for i := range rows {
		rows[i].RewardImageURL = absoluteMediaURL(h.mediaBase, rows[i].RewardImageURL)
	}
	c.JSON(http.StatusOK, dto.Paginated(rows, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *RewardHandler) create(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.RewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	rw, err := rewardFromRequest(&req, uuid.Nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if err := h.svc.Create(c.Request.Context(), rw); err != nil {
		h.respondErr(c, loc, err, "create reward")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(rw))
}

func (h *RewardHandler) update(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	var req dto.RewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	rw, err := rewardFromRequest(&req, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if err := h.svc.Update(c.Request.Context(), rw); err != nil {
		h.respondErr(c, loc, err, "update reward")
		return
	}
	c.JSON(http.StatusOK, dto.OK(rw))
}

func (h *RewardHandler) remove(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.respondErr(c, loc, err, "delete reward")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *RewardHandler) issue(c *gin.Context) {
	h.redemptionAction(c, true)
}

func (h *RewardHandler) cancel(c *gin.Context) {
	h.redemptionAction(c, false)
}

func (h *RewardHandler) redemptionAction(c *gin.Context, markIssued bool) {
	loc := middleware.GetLocale(c)

	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	var req dto.RedemptionActionRequest
	// Body ixtiyoriy — izohsiz ham topshirish mumkin.
	_ = c.ShouldBindJSON(&req)

	var red *domain.RewardRedemption
	op := "cancel redemption"
	if markIssued {
		op = "issue redemption"
		red, err = h.svc.MarkIssued(c.Request.Context(), id, adminID, req.Note)
	} else {
		red, err = h.svc.Cancel(c.Request.Context(), id, adminID, req.Note)
	}
	if err != nil {
		h.respondErr(c, loc, err, op)
		return
	}
	c.JSON(http.StatusOK, dto.OK(red))
}

// respondErr — domen xatolarini HTTP statuslarga o'giradi (§3.4).
func (h *RewardHandler) respondErr(c *gin.Context, loc i18n.Locale, err error, op string) {
	switch {
	case errors.Is(err, domain.ErrInsufficientBalance):
		// Mijozga aytiladi: bu uning tuzata oladigan holati.
		c.JSON(http.StatusConflict,
			dto.ErrResponse(i18n.T(loc, i18n.MsgInsufficientBalance)))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(
			i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
	default:
		logServerError(h.log, op, err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
	}
}

// rewardFromRequest — DTO'dan domen obyektiga (§17.3 #13: mijoz `id` va
// tizim maydonlarini yubora olmaydi — ular bu yerda o'rnatiladi).
func rewardFromRequest(req *dto.RewardRequest, id uuid.UUID) (*domain.Reward, error) {
	rw := &domain.Reward{
		ID:           id,
		Title:        req.Title,
		Description:  req.Description,
		ImageURL:     req.ImageURL,
		Category:     req.Category,
		CostCoins:    req.CostCoins,
		Stock:        req.Stock,
		PerUserLimit: req.PerUserLimit,
		IsActive:     req.IsActive,
	}

	if req.StartsAt != nil && *req.StartsAt != "" {
		t, err := time.Parse(time.RFC3339, *req.StartsAt)
		if err != nil {
			return nil, err
		}
		rw.StartsAt = &t
	}
	if req.EndsAt != nil && *req.EndsAt != "" {
		t, err := time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			return nil, err
		}
		rw.EndsAt = &t
	}

	if req.Config != nil {
		b, err := json.Marshal(req.Config)
		if err != nil {
			return nil, err
		}
		rw.Config = datatypes.JSON(b)
	}
	return rw, nil
}
