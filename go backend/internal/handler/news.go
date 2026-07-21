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

// NewsHandler — yangiliklar (§16.3: kontent admin panel orqali kiritiladi).
//
//	GET    /news?page=1&limit=20   — e'lon qilinganlar (mobil)
//	GET    /news/:id               — to'liq matn (+ko'rish hisobi)
//	GET    /admin/news             — hammasi (draft'lar ham)
//	POST   /admin/news
//	PUT    /admin/news/:id
//	DELETE /admin/news/:id
type NewsHandler struct {
	svc       *service.NewsService
	jwt       *security.JWTManager
	log       *zap.Logger
	mediaBase string
}

func NewNewsHandler(svc *service.NewsService, jwt *security.JWTManager, log *zap.Logger, mediaBase string) *NewsHandler {
	return &NewsHandler{svc: svc, jwt: jwt, log: log, mediaBase: mediaBase}
}

func (h *NewsHandler) Register(r gin.IRouter) {
	g := r.Group("/news")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("", h.list)
		g.GET("/:id", h.get)
	}

	admin := r.Group("/admin/news")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.GET("", h.adminList)
		admin.POST("", h.create)
		admin.PUT("/:id", h.update)
		admin.DELETE("/:id", h.remove)
	}
}

// list — mobil ro'yxat: faqat e'lon qilingan va vaqti kelganlar.
func (h *NewsHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.NewsFilter{
		PublishedOnly: true,
		Page:          atoiDefault(c.Query("page"), 1),
		Limit:         atoiDefault(c.Query("limit"), 20),
	}

	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "news list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	for i := range items {
		items[i].CoverURL = absoluteMediaURL(h.mediaBase, items[i].CoverURL)
	}
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *NewsHandler) get(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	isAdmin := middleware.GetRole(c) == string(domain.RoleAdmin)
	// Ko'rish hisobi faqat oddiy foydalanuvchida — admin ko'rigi statistikani buzmasin.
	n, err := h.svc.Get(c.Request.Context(), id, !isAdmin)
	if err != nil {
		h.newsErr(c, loc, err, "news get")
		return
	}

	// Draft yoki hali e'lon qilinmagani faqat adminga ko'rinadi (§17.3 #26/#29).
	if !isAdmin {
		notReady := n.Status != domain.NewsStatusPublished ||
			(n.PublishedAt != nil && n.PublishedAt.After(time.Now()))
		if notReady {
			c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
			return
		}
	}

	n.CoverURL = absoluteMediaURL(h.mediaBase, n.CoverURL)
	c.JSON(http.StatusOK, dto.OK(n))
}

func (h *NewsHandler) adminList(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.NewsFilter{
		Status: c.Query("status"),
		Search: c.Query("search"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	if f.Status != "" && !domain.ValidNewsStatus(f.Status) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "news admin list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	for i := range items {
		items[i].CoverURL = absoluteMediaURL(h.mediaBase, items[i].CoverURL)
	}
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *NewsHandler) create(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.NewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	n, err := newsFromRequest(&req, uuid.Nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
		return
	}
	// Muallif — token egasi (mijoz yubora olmaydi).
	if uid, err := middleware.GetUserID(c); err == nil {
		n.AuthorID = &uid
	}

	if err := h.svc.Create(c.Request.Context(), n); err != nil {
		h.newsErr(c, loc, err, "news create")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(n))
}

func (h *NewsHandler) update(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	var req dto.NewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	n, err := newsFromRequest(&req, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
		return
	}
	if err := h.svc.Update(c.Request.Context(), n); err != nil {
		h.newsErr(c, loc, err, "news update")
		return
	}
	c.JSON(http.StatusOK, dto.OK(n))
}

func (h *NewsHandler) remove(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.newsErr(c, loc, err, "news delete")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *NewsHandler) newsErr(c *gin.Context, loc i18n.Locale, err error, op string) {
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

func newsFromRequest(req *dto.NewsRequest, id uuid.UUID) (*domain.News, error) {
	n := &domain.News{
		ID:       id,
		Title:    req.Title,
		Excerpt:  req.Excerpt,
		Body:     req.Body,
		CoverURL: req.CoverURL,
		Status:   req.Status,
		Pinned:   req.Pinned,
	}
	if n.Status == "" {
		n.Status = domain.NewsStatusDraft
	}
	if req.PublishedAt != nil && *req.PublishedAt != "" {
		t, err := time.Parse(time.RFC3339, *req.PublishedAt)
		if err != nil {
			return nil, err
		}
		n.PublishedAt = &t
	}
	return n, nil
}
