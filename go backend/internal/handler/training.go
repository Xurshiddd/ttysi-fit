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

// TrainingHandler — video mashg'ulotlar (§16.3).
//
//	GET    /trainings?category=&level=   — e'lon qilinganlar (mobil)
//	GET    /trainings/:id                — to'liq (+ko'rish hisobi)
//	GET    /training-categories          — mavjud kategoriyalar (filtr/forma uchun)
//	GET    /admin/trainings              — hammasi (draft ham)
//	POST   /admin/trainings
//	PUT    /admin/trainings/:id
//	DELETE /admin/trainings/:id
type TrainingHandler struct {
	svc       *service.TrainingService
	jwt       *security.JWTManager
	log       *zap.Logger
	mediaBase string
}

func NewTrainingHandler(svc *service.TrainingService, jwt *security.JWTManager, log *zap.Logger, mediaBase string) *TrainingHandler {
	return &TrainingHandler{svc: svc, jwt: jwt, log: log, mediaBase: mediaBase}
}

func (h *TrainingHandler) Register(r gin.IRouter) {
	cats := r.Group("/training-categories")
	cats.Use(middleware.Auth(h.jwt))
	{
		cats.GET("", h.categories)
	}

	g := r.Group("/trainings")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("", h.list)
		g.GET("/:id", h.get)
	}

	admin := r.Group("/admin/trainings")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.GET("", h.adminList)
		admin.POST("", h.create)
		admin.PUT("/:id", h.update)
		admin.DELETE("/:id", h.remove)
	}
}

// categories — admin uchun hamma, oddiy foydalanuvchi uchun faqat e'lon
// qilingan mashg'ulotlar kategoriyalari (draft kategoriyasi sizib chiqmasin).
func (h *TrainingHandler) categories(c *gin.Context) {
	loc := middleware.GetLocale(c)

	isAdmin := middleware.GetRole(c) == string(domain.RoleAdmin)
	list, err := h.svc.Categories(c.Request.Context(), !isAdmin)
	if err != nil {
		logServerError(h.log, "training categories", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	c.JSON(http.StatusOK, dto.OK(list))
}

func (h *TrainingHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.TrainingFilter{
		PublishedOnly: true,
		Category:      c.Query("category"),
		Level:         c.Query("level"),
		Search:        c.Query("search"),
		Page:          atoiDefault(c.Query("page"), 1),
		Limit:         atoiDefault(c.Query("limit"), 20),
	}
	// Enum validatsiya — SQL'ga faqat tekshirilgan qiymat boradi (§3.2).
	// Category erkin matn, lekin u parametrli so'rovga tushadi.
	if f.Level != "" && !domain.ValidTrainingLevel(f.Level) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "training list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	for i := range items {
		items[i].ThumbnailURL = absoluteMediaURL(h.mediaBase, items[i].ThumbnailURL)
	}
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *TrainingHandler) get(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	isAdmin := middleware.GetRole(c) == string(domain.RoleAdmin)
	t, err := h.svc.Get(c.Request.Context(), id, !isAdmin)
	if err != nil {
		h.trainingErr(c, loc, err, "training get")
		return
	}

	if !isAdmin {
		notReady := t.Status != domain.TrainingStatusPublished ||
			(t.PublishedAt != nil && t.PublishedAt.After(time.Now()))
		if notReady {
			c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
			return
		}
	}

	t.ThumbnailURL = absoluteMediaURL(h.mediaBase, t.ThumbnailURL)
	c.JSON(http.StatusOK, dto.OK(t))
}

func (h *TrainingHandler) adminList(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.TrainingFilter{
		Status:   c.Query("status"),
		Category: c.Query("category"),
		Level:    c.Query("level"),
		Search:   c.Query("search"),
		Page:     atoiDefault(c.Query("page"), 1),
		Limit:    atoiDefault(c.Query("limit"), 20),
	}
	if (f.Status != "" && !domain.ValidTrainingStatus(f.Status)) ||
		(f.Level != "" && !domain.ValidTrainingLevel(f.Level)) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "training admin list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	for i := range items {
		items[i].ThumbnailURL = absoluteMediaURL(h.mediaBase, items[i].ThumbnailURL)
	}
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *TrainingHandler) create(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.TrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	t, err := trainingFromRequest(&req, uuid.Nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
		return
	}
	if err := h.svc.Create(c.Request.Context(), t); err != nil {
		h.trainingErr(c, loc, err, "training create")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(t))
}

func (h *TrainingHandler) update(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	var req dto.TrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	t, err := trainingFromRequest(&req, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
		return
	}
	if err := h.svc.Update(c.Request.Context(), t); err != nil {
		h.trainingErr(c, loc, err, "training update")
		return
	}
	c.JSON(http.StatusOK, dto.OK(t))
}

func (h *TrainingHandler) remove(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.trainingErr(c, loc, err, "training delete")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TrainingHandler) trainingErr(c *gin.Context, loc i18n.Locale, err error, op string) {
	switch {
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
	default:
		logServerError(h.log, op, err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
	}
}

func trainingFromRequest(req *dto.TrainingRequest, id uuid.UUID) (*domain.Training, error) {
	t := &domain.Training{
		ID:           id,
		Title:        req.Title,
		Description:  req.Description,
		Category:     req.Category,
		Level:        req.Level,
		VideoURL:     req.VideoURL,
		ThumbnailURL: req.ThumbnailURL,
		DurationMin:  req.DurationMin,
		Status:       req.Status,
		SortOrder:    req.SortOrder,
	}
	if t.Level == "" {
		t.Level = domain.TrainingBeginner
	}
	if t.Status == "" {
		t.Status = domain.TrainingStatusDraft
	}
	if req.PublishedAt != nil && *req.PublishedAt != "" {
		p, err := time.Parse(time.RFC3339, *req.PublishedAt)
		if err != nil {
			return nil, err
		}
		t.PublishedAt = &p
	}
	return t, nil
}
