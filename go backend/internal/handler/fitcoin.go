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

// FitCoinHandler — FIT Coin ledger.
//
//	GET  /fit-coins/balance                — o'z balansi
//	GET  /fit-coins?page=1&limit=20        — o'z tarixi
//	POST /challenges/:id/claim-reward      — yakunlangan chellenj mukofotini olish
//	POST /admin/fit-coins/grant            — admin qo'lda berish/olish
type FitCoinHandler struct {
	svc *service.FitCoinService
	jwt *security.JWTManager
	log *zap.Logger
}

func NewFitCoinHandler(svc *service.FitCoinService, jwt *security.JWTManager, log *zap.Logger) *FitCoinHandler {
	return &FitCoinHandler{svc: svc, jwt: jwt, log: log}
}

func (h *FitCoinHandler) Register(r gin.IRouter) {
	g := r.Group("/fit-coins")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("/balance", h.balance)
		g.GET("", h.history)
	}

	// Mukofot olish chellenj resursiga tegishli — shu sababli /challenges ostida.
	ch := r.Group("/challenges")
	ch.Use(middleware.Auth(h.jwt))
	{
		ch.POST("/:id/claim-reward", h.claimReward)
	}

	admin := r.Group("/admin/fit-coins")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.POST("/grant", h.adminGrant)
		admin.GET("/:user_id", h.adminUserCoins)
	}
}

// adminUserCoins — admin uchun: tanlangan foydalanuvchi balansi + tarixi.
// Faqat admin (RequireRole) — oddiy foydalanuvchi boshqaning balansini
// ko'ra olmaydi (§17.3 #26 IDOR).
func (h *FitCoinHandler) adminUserCoins(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	bal, err := h.svc.Balance(c.Request.Context(), uid)
	if err != nil {
		logServerError(h.log, "admin user coins: balance", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	rows, total, err := h.svc.History(c.Request.Context(), uid, domain.CoinFilter{
		Page:  atoiDefault(c.Query("page"), 1),
		Limit: atoiDefault(c.Query("limit"), 20),
	})
	if err != nil {
		logServerError(h.log, "admin user coins: history", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{
		"balance": bal,
		"history": rows,
		"total":   total,
	}))
}

func (h *FitCoinHandler) balance(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	b, err := h.svc.Balance(c.Request.Context(), uid)
	if err != nil {
		logServerError(h.log, "coin balance", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(b))
}

// history — foydalanuvchi FAQAT o'z tarixini ko'radi: user_id token'dan olinadi,
// query'dan emas (§17.3 #26 IDOR).
func (h *FitCoinHandler) history(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	f := domain.CoinFilter{
		Reason: c.Query("reason"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	if f.Reason != "" && !domain.ValidCoinReason(f.Reason) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	rows, total, err := h.svc.History(c.Request.Context(), uid, f)
	if err != nil {
		logServerError(h.log, "coin history", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.Paginated(rows, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *FitCoinHandler) claimReward(c *gin.Context) {
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

	coin, err := h.svc.ClaimChallengeReward(c.Request.Context(), uid, id)
	if err != nil {
		h.coinErr(c, loc, err, "claim reward")
		return
	}
	c.JSON(http.StatusOK, dto.OK(coin))
}

func (h *FitCoinHandler) adminGrant(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.AdminGrantCoinsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	coin, err := h.svc.AdminGrant(c.Request.Context(), uid, req.Amount, req.Note)
	if err != nil {
		h.coinErr(c, loc, err, "admin grant")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(coin))
}

// coinErr — domen xatolarini HTTP statusga o'giradi.
func (h *FitCoinHandler) coinErr(c *gin.Context, loc i18n.Locale, err error, op string) {
	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		// Mukofot allaqachon olingan — bu xato emas, takroriy so'rov.
		c.JSON(http.StatusConflict, dto.ErrResponse(i18n.T(loc, i18n.MsgCoinAlreadyClaimed)))
	case errors.Is(err, domain.ErrInsufficientBalance):
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgInsufficientBalance)))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
	default:
		logServerError(h.log, op, err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
	}
}
