package handler

import (
	"errors"
	"net/http"
	"strconv"

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

// UserHandler — admin foydalanuvchilar boshqaruvi.
type UserHandler struct {
	svc       *service.UserService
	jwt       *security.JWTManager
	log       *zap.Logger
	mediaBase string // avatar nisbiy yo'liga qo'shiladigan ommaviy asos
}

func NewUserHandler(svc *service.UserService, jwt *security.JWTManager, log *zap.Logger, mediaBase string) *UserHandler {
	return &UserHandler{svc: svc, jwt: jwt, log: log, mediaBase: mediaBase}
}

func (h *UserHandler) Register(r gin.IRouter) {
	admin := r.Group("/admin/users")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.GET("", h.list)
	}

	// Har qanday autentifikatsiyalangan foydalanuvchi — o'z profili.
	me := r.Group("/users")
	me.Use(middleware.Auth(h.jwt))
	{
		me.GET("/me", h.me)
		me.PUT("/me", h.updateMe)
	}
}

// me — token egasining profili.
func (h *UserHandler) me(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	user, err := h.svc.GetProfile(c.Request.Context(), uid)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
			return
		}
		logServerError(h.log, "get profile", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	user.AvatarURL = absoluteMediaURL(h.mediaBase, user.AvatarURL)
	c.JSON(http.StatusOK, dto.OK(user))
}

// updateMe — token egasi o'z profilini yangilaydi.
// Faqat phone/bio/language; qolgan maydonlar HEMIS sync ixtiyorida (dto izohiga qarang).
func (h *UserHandler) updateMe(c *gin.Context) {
	loc := middleware.GetLocale(c)

	uid, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}

	user, err := h.svc.UpdateProfile(c.Request.Context(), uid, domain.ProfileUpdate{
		Phone:    req.Phone,
		Bio:      req.Bio,
		Language: req.Language,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
			return
		}
		logServerError(h.log, "update profile", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	user.AvatarURL = absoluteMediaURL(h.mediaBase, user.AvatarURL)
	c.JSON(http.StatusOK, dto.OK(user))
}

func (h *UserHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.UserFilter{
		Role:   c.Query("role"),
		Search: c.Query("search"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	if v := c.Query("faculty_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.FacultyID = &id
		}
	}
	if v := c.Query("group_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.GroupID = &id
		}
	}

	items, total, err := h.svc.ListUsers(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "list users", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	for i := range items {
		items[i].AvatarURL = absoluteMediaURL(h.mediaBase, items[i].AvatarURL)
	}

	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{
		Page:  f.Page,
		Limit: f.Limit,
		Total: total,
	}))
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return def
}
