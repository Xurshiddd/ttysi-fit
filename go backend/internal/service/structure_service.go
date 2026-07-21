package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/pkg/hemis"
	"gorm.io/datatypes"
)

// StructureService — HEMIS strukturalari use-case qatlami.
type StructureService struct {
	repo               domain.StructureRepository
	hemis              *hemis.Client
	facultyTypeCode    string
	departmentTypeCode string
}

func NewStructureService(repo domain.StructureRepository, client *hemis.Client, facultyTypeCode, departmentTypeCode string) *StructureService {
	return &StructureService{
		repo:               repo,
		hemis:              client,
		facultyTypeCode:    facultyTypeCode,
		departmentTypeCode: departmentTypeCode,
	}
}

// SyncStructures — HEMIS dan barcha strukturalarni tortib, DB ga ko'tarib-yangilaydi.
func (s *StructureService) SyncStructures(ctx context.Context) (*domain.SyncStats, error) {
	dtos, err := s.hemis.FetchStructures(ctx)
	if err != nil {
		return nil, fmt.Errorf("SyncStructures: fetch: %w", err)
	}

	now := time.Now()
	items := make([]domain.Structure, 0, len(dtos))
	for _, d := range dtos {
		items = append(items, domain.Structure{
			HemisID:           d.HemisID,
			Name:              d.Name,
			Code:              d.Code,
			StructureTypeCode: d.StructureTypeCode,
			StructureTypeName: d.StructureTypeName,
			LocalityTypeCode:  d.LocalityTypeCode,
			LocalityTypeName:  d.LocalityTypeName,
			ParentHemisID:     d.ParentHemisID,
			Active:            d.Active,
			Raw:               datatypes.JSON(d.Raw),
			SyncedAt:          &now,
		})
	}

	stats, err := s.repo.UpsertBatch(ctx, items)
	if err != nil {
		return nil, fmt.Errorf("SyncStructures: upsert: %w", err)
	}

	// Daraxtni bog'lash (parent_hemis_id → parent_id), bitta SQL.
	if err := s.repo.RelinkParents(ctx); err != nil {
		return nil, fmt.Errorf("SyncStructures: relink: %w", err)
	}

	return stats, nil
}

// ListFaculties — fakultet tipidagi strukturalarni qaytaradi.
func (s *StructureService) ListFaculties(ctx context.Context) ([]domain.Structure, error) {
	return s.repo.ListByType(ctx, s.facultyTypeCode)
}

// ListDepartments — kafedralarni qaytaradi; facultyID berilsa o'sha fakultetnikini.
func (s *StructureService) ListDepartments(ctx context.Context, facultyID *uuid.UUID) ([]domain.Structure, error) {
	return s.repo.ListByTypeAndParent(ctx, s.departmentTypeCode, facultyID)
}
