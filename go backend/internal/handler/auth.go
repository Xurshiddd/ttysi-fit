package handler

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/internal/dto"
	"github.com/ttysi-fit/backend/internal/i18n"
	"github.com/ttysi-fit/backend/internal/middleware"
	"github.com/ttysi-fit/backend/internal/service"
	"github.com/ttysi-fit/backend/pkg/security"
	"go.uber.org/zap"
)

// AuthHandler — autentifikatsiya HTTP endpointlari.
type AuthHandler struct {
	auth        *service.AuthService
	jwt         *security.JWTManager
	log         *zap.Logger
	appRedirect string // HEMIS callback'dan keyin mobil deep link (bo'sh bo'lsa JSON)
}

func NewAuthHandler(auth *service.AuthService, jwt *security.JWTManager, log *zap.Logger, appRedirect string) *AuthHandler {
	return &AuthHandler{auth: auth, jwt: jwt, log: log, appRedirect: appRedirect}
}

// Register — ochiq va himoyalangan auth route'larini ulaydi.
// mw — ixtiyoriy qo'shimcha middleware (masalan qattiq rate limiter,
// brute-force himoyasi uchun — CLAUDE.md §17.3 #15/#16).
func (h *AuthHandler) Register(r gin.IRouter, mw ...gin.HandlerFunc) {
	g := r.Group("/auth", mw...)
	{
		g.POST("/register", h.register)
		g.POST("/login", h.login)
		g.POST("/refresh", h.refresh)
		g.POST("/logout", middleware.Auth(h.jwt), h.logout)

		// HEMIS OAuth — :provider = "student" | "employee"
		g.GET("/hemis/:provider/login", h.hemisLogin)
		g.GET("/hemis/:provider/callback", h.hemisCallback)
		// Mobil: bir martalik code'ni token'ga almashtirish
		g.POST("/hemis/exchange", h.hemisExchange)
	}
}

func validHemisProvider(p string) bool {
	return p == "student" || p == "employee"
}

// hemisLogin — HEMIS authorize sahifasiga yo'naltiradi (state Redis'da saqlanadi).
func (h *AuthHandler) hemisLogin(c *gin.Context) {
	loc := middleware.GetLocale(c)
	provider := c.Param("provider")
	if !validHemisProvider(provider) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	authURL, err := h.auth.HemisAuthURL(c.Request.Context(), provider)
	if err != nil {
		h.handleError(c, "hemisLogin", err)
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

// hemisCallback — HEMIS'dan qaytadi: code+state ni tekshirib token beradi.
// appRedirect sozlangan bo'lsa mobil deep link'ga (fragment'da token), aks holda JSON.
func (h *AuthHandler) hemisCallback(c *gin.Context) {
	loc := middleware.GetLocale(c)
	provider := c.Param("provider")
	if !validHemisProvider(provider) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	tokens, err := h.auth.HemisCallback(
		c.Request.Context(), provider,
		c.Query("state"), c.Query("code"),
	)
	if err != nil {
		h.handleError(c, "hemisCallback", err)
		return
	}

	if h.appRedirect != "" {
		// Bir martalik code — token URL'ga tushmaydi (eng xavfsiz mobil oqim).
		code, err := h.auth.StashTokens(c.Request.Context(), tokens)
		if err != nil {
			h.handleError(c, "hemisCallback", err)
			return
		}
		c.Redirect(http.StatusFound, h.appRedirect+"?code="+url.QueryEscape(code))
		return
	}
	c.JSON(http.StatusOK, dto.OK(tokens))
}

// hemisExchange — mobil: deep link'dan kelgan bir martalik code'ni token'ga almashtiradi.
func (h *AuthHandler) hemisExchange(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.HemisExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	tokens, err := h.auth.ExchangeTokens(c.Request.Context(), req.Code)
	if err != nil {
		h.handleError(c, "hemisExchange", err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(tokens))
}

func (h *AuthHandler) register(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	res, err := h.auth.Register(c.Request.Context(), req, string(loc))
	if err != nil {
		h.handleError(c, "register", err)
		return
	}
	c.JSON(http.StatusCreated, dto.OK(res))
}

func (h *AuthHandler) login(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	res, err := h.auth.Login(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, "login", err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (h *AuthHandler) refresh(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	res, err := h.auth.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.handleError(c, "refresh", err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(res))
}

func (h *AuthHandler) logout(c *gin.Context) {
	loc := middleware.GetLocale(c)

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}
	if err := h.auth.Logout(c.Request.Context(), userID); err != nil {
		h.handleError(c, "logout", err)
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"message": i18n.T(loc, i18n.MsgLogoutSuccess)}))
}

// handleError — domain xatosini mos HTTP status va tarjima qilingan xabarga aylantiradi.
// Ichki tafsilotlar mijozga chiqmaydi (CLAUDE.md 3.4).
func (h *AuthHandler) handleError(c *gin.Context, op string, err error) {
	loc := middleware.GetLocale(c)
	switch {
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, dto.ErrResponse(i18n.T(loc, i18n.MsgEmailTaken)))
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgInvalidCredentials)))
	case errors.Is(err, domain.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
	default:
		logServerError(h.log, "auth handler", err, zap.String("op", op))
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
	}
}
