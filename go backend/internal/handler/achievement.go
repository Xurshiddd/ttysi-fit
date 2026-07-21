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
	"gorm.io/datatypes"
)

// AchievementHandler — yutuqlar va sertifikatlar (§16.3).
//
//	GET    /achievements              — aktiv yutuqlar + mening progressim (mobil)
//	GET    /achievements/me           — men qozonganlarim (profil, sertifikatlar)
//	GET    /achievement-types         — tur ta'riflari (dinamik forma)
//	GET    /admin/achievements        — hammasi (draft ham)
//	POST   /admin/achievements
//	PUT    /admin/achievements/:id
//	DELETE /admin/achievements/:id
//	POST   /admin/achievements/:id/award — qo'lda berish (g'olib, ishtirokchi)
type AchievementHandler struct {
	svc       *service.AchievementService
	jwt       *security.JWTManager
	log       *zap.Logger
	mediaBase string
}

func NewAchievementHandler(svc *service.AchievementService, jwt *security.JWTManager, log *zap.Logger, mediaBase string) *AchievementHandler {
	return &AchievementHandler{svc: svc, jwt: jwt, log: log, mediaBase: mediaBase}
}

func (h *AchievementHandler) Register(r gin.IRouter) {
	types := r.Group("/achievement-types")
	types.Use(middleware.Auth(h.jwt))
	{
		types.GET("", h.types)
	}

	g := r.Group("/achievements")
	g.Use(middleware.Auth(h.jwt))
	{
		g.GET("", h.list)
		g.GET("/me", h.mine)
		// :id — berilgan yutuq (user_achievements.id), yutuq shabloni emas.
		g.GET("/awards/:id/certificate", h.certificate)
	}

	admin := r.Group("/admin/achievements")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.GET("", h.adminList)
		admin.POST("", h.create)
		admin.PUT("/:id", h.update)
		admin.DELETE("/:id", h.remove)
		admin.POST("/:id/award", h.award)
	}
}

// types — admin panel dinamik formasi uchun (§16.2).
func (h *AchievementHandler) types(c *gin.Context) {
	c.JSON(http.StatusOK, dto.OK(h.svc.Types()))
}

// list — aktiv yutuqlar + shu foydalanuvchi progressi.
func (h *AchievementHandler) list(c *gin.Context) {
	loc := middleware.GetLocale(c)

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	f := domain.AchievementFilter{
		Status: domain.AchStatusActive, // mobil draft/arxivni ko'rmaydi
		Type:   c.Query("type"),
		Page:   atoiDefault(c.Query("page"), 1),
		Limit:  atoiDefault(c.Query("limit"), 20),
	}
	if f.Type != "" && !domain.ValidAchievementType(f.Type) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	items, total, err := h.svc.ListForUser(c.Request.Context(), userID, f)
	if err != nil {
		logServerError(h.log, "achievement list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	h.absolutize(items)
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

// mine — foydalanuvchi qozongan yutuqlar (profil bo'limi, sertifikatlar).
func (h *AchievementHandler) mine(c *gin.Context) {
	loc := middleware.GetLocale(c)

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	f := domain.AchievementFilter{
		Type:  c.Query("type"),
		Page:  atoiDefault(c.Query("page"), 1),
		Limit: atoiDefault(c.Query("limit"), 20),
	}
	if f.Type != "" && !domain.ValidAchievementType(f.Type) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	items, total, err := h.svc.ListEarned(c.Request.Context(), userID, f)
	if err != nil {
		logServerError(h.log, "achievement mine", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	h.absolutize(items)
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

// certificate — sertifikat PDF sini qaytaradi.
//
// Fayl diskda saqlanmaydi, har so'rovda chiziladi — shablon o'zgarsa avval
// berilgan sertifikatlar ham darrov yangilanadi (pkg/certificate izohiga qarang).
func (h *AchievementHandler) certificate(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	isAdmin := middleware.GetRole(c) == string(domain.RoleAdmin)
	pdf, filename, err := h.svc.Certificate(c.Request.Context(), id, userID, isAdmin)
	if err != nil {
		h.achievementErr(c, loc, err, "achievement certificate")
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdf)
}

func (h *AchievementHandler) adminList(c *gin.Context) {
	loc := middleware.GetLocale(c)

	f := domain.AchievementFilter{
		Status:    c.Query("status"),
		Type:      c.Query("type"),
		AwardMode: c.Query("award_mode"),
		Page:      atoiDefault(c.Query("page"), 1),
		Limit:     atoiDefault(c.Query("limit"), 20),
	}
	if f.Status != "" && !domain.ValidAchievementStatus(f.Status) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if f.Type != "" && !domain.ValidAchievementType(f.Type) {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		logServerError(h.log, "achievement admin list", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}
	for i := range items {
		items[i].IconURL = absoluteMediaURL(h.mediaBase, items[i].IconURL)
		items[i].CoverURL = absoluteMediaURL(h.mediaBase, items[i].CoverURL)
	}
	c.JSON(http.StatusOK, dto.Paginated(items, dto.Meta{Page: f.Page, Limit: f.Limit, Total: total}))
}

func (h *AchievementHandler) create(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var req dto.AchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	a := achievementFromRequest(&req, uuid.Nil)
	if err := h.svc.Create(c.Request.Context(), a); err != nil {
		h.achievementErr(c, loc, err, "achievement create")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(a))
}

func (h *AchievementHandler) update(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	var req dto.AchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	a := achievementFromRequest(&req, id)
	if err := h.svc.Update(c.Request.Context(), a); err != nil {
		h.achievementErr(c, loc, err, "achievement update")
		return
	}
	c.JSON(http.StatusOK, dto.OK(a))
}

func (h *AchievementHandler) remove(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		h.achievementErr(c, loc, err, "achievement delete")
		return
	}
	c.Status(http.StatusNoContent)
}

// award — admin yutuqni qo'lda beradi (faqat manual turlar uchun).
func (h *AchievementHandler) award(c *gin.Context) {
	loc := middleware.GetLocale(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}
	adminID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrResponse(i18n.T(loc, i18n.MsgUnauthorized)))
		return
	}

	var req dto.AwardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, validationResponse(loc, err))
		return
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
		return
	}

	ua, err := h.svc.AwardManual(c.Request.Context(), id, userID, adminID, req.Note)
	if err != nil {
		h.achievementErr(c, loc, err, "achievement award")
		return
	}
	c.JSON(http.StatusCreated, dto.OK(ua))
}

func (h *AchievementHandler) absolutize(items []domain.AchievementView) {
	for i := range items {
		items[i].IconURL = absoluteMediaURL(h.mediaBase, items[i].IconURL)
		items[i].CoverURL = absoluteMediaURL(h.mediaBase, items[i].CoverURL)
		items[i].CertificateURL = absoluteMediaURL(h.mediaBase, items[i].CertificateURL)
	}
}

// achievementErr — domen xatolarini HTTP statuslarga o'giradi. Ichki tafsilot
// mijozga chiqmaydi (§3.4), mezon xatosi esa adminga foydali — maydon nomi
// bilan ko'rsatiladi (aks holda "nega yutuq saqlanmadi" bilinmasdi).
func (h *AchievementHandler) achievementErr(c *gin.Context, loc i18n.Locale, err error, op string) {
	var cfgErr *domain.ErrChallengeConfig
	switch {
	case errors.As(err, &cfgErr):
		c.JSON(http.StatusBadRequest, dto.ErrValidation(
			i18n.T(loc, i18n.MsgValidationFailed),
			map[string]string{cfgErr.Field: cfgErr.Reason},
		))
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, dto.ErrDetailed(i18n.T(loc, i18n.MsgValidationFailed), err.Error()))
	case errors.Is(err, domain.ErrAlreadyExists):
		c.JSON(http.StatusConflict, dto.ErrResponse(i18n.T(loc, i18n.MsgAchAlreadyAwarded)))
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrResponse(i18n.T(loc, i18n.MsgNotFound)))
	default:
		logServerError(h.log, op, err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
	}
}

func achievementFromRequest(req *dto.AchievementRequest, id uuid.UUID) *domain.Achievement {
	a := &domain.Achievement{
		ID:                  id,
		Type:                domain.AchievementType(req.Type),
		Title:               req.Title,
		Description:         req.Description,
		Status:              req.Status,
		RewardCoins:         req.RewardCoins,
		IconURL:             req.IconURL,
		CoverURL:            req.CoverURL,
		CertificateTemplate: req.CertificateTemplate,
	}
	if a.Status == "" {
		a.Status = domain.AchStatusDraft
	}
	if len(req.Criteria) > 0 {
		a.Criteria = datatypes.JSON(req.Criteria)
	} else {
		a.Criteria = datatypes.JSON([]byte("{}"))
	}
	// AwardMode service.validate da turdan kelib chiqib to'ldiriladi.
	return a
}
