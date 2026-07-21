package handler

import (
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

// StructureHandler — fakultet ro'yxati va HEMIS sync endpointlari.
type StructureHandler struct {
	svc *service.StructureService
	jwt *security.JWTManager
	log *zap.Logger
}

func NewStructureHandler(svc *service.StructureService, jwt *security.JWTManager, log *zap.Logger) *StructureHandler {
	return &StructureHandler{svc: svc, jwt: jwt, log: log}
}

// Register — ochiq va admin route'larini ulaydi.
func (h *StructureHandler) Register(r gin.IRouter) {
	// Ochiq: ro'yxatdan o'tishda fakultet/kafedra tanlash uchun
	r.GET("/faculties", h.listFaculties)
	r.GET("/departments", h.listDepartments)

	// Admin: HEMIS sinxronizatsiyasi (kelajakda admin panel chaqiradi)
	admin := r.Group("/admin/hemis")
	admin.Use(middleware.Auth(h.jwt), middleware.RequireRole(string(domain.RoleAdmin)))
	{
		admin.POST("/sync/structures", h.syncStructures)
	}
}

// structureBrief — fakultet/kafedra ro'yxati uchun yengil javob.
type structureBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func (h *StructureHandler) listFaculties(c *gin.Context) {
	loc := middleware.GetLocale(c)

	faculties, err := h.svc.ListFaculties(c.Request.Context())
	if err != nil {
		logServerError(h.log, "list faculties", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	out := make([]structureBrief, 0, len(faculties))
	for _, f := range faculties {
		out = append(out, structureBrief{ID: f.ID.String(), Name: f.Name, Code: f.Code})
	}
	c.JSON(http.StatusOK, dto.OK(out))
}

// listDepartments — kafedralar ro'yxati; ?faculty_id=<uuid> bilan filtrlash mumkin.
func (h *StructureHandler) listDepartments(c *gin.Context) {
	loc := middleware.GetLocale(c)

	var facultyID *uuid.UUID
	if q := c.Query("faculty_id"); q != "" {
		id, err := uuid.Parse(q)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrResponse(i18n.T(loc, i18n.MsgValidationFailed)))
			return
		}
		facultyID = &id
	}

	departments, err := h.svc.ListDepartments(c.Request.Context(), facultyID)
	if err != nil {
		logServerError(h.log, "list departments", err)
		c.JSON(http.StatusInternalServerError, dto.ErrResponse(i18n.T(loc, i18n.MsgInternalError)))
		return
	}

	out := make([]structureBrief, 0, len(departments))
	for _, d := range departments {
		out = append(out, structureBrief{ID: d.ID.String(), Name: d.Name, Code: d.Code})
	}
	c.JSON(http.StatusOK, dto.OK(out))
}

func (h *StructureHandler) syncStructures(c *gin.Context) {
	loc := middleware.GetLocale(c)

	stats, err := h.svc.SyncStructures(c.Request.Context())
	if err != nil {
		logServerError(h.log, "hemis sync", err)
		c.JSON(http.StatusBadGateway, dto.ErrResponse(i18n.T(loc, i18n.MsgHemisError)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    stats,
		"message": i18n.T(loc, i18n.MsgSyncSuccess),
	})
}
