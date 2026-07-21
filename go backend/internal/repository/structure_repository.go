package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type structureRepository struct {
	db *gorm.DB
}

func NewStructureRepository(db *gorm.DB) domain.StructureRepository {
	return &structureRepository{db: db}
}

// UpsertBatch — hemis_id konflikti bo'yicha ko'tarib-yangilaydi.
// Bitta INSERT ... ON CONFLICT so'rovi — for ichida DB so'rovi yo'q (N+1 yo'q).
func (r *structureRepository) UpsertBatch(ctx context.Context, items []domain.Structure) (*domain.SyncStats, error) {
	if len(items) == 0 {
		return &domain.SyncStats{}, nil
	}

	// Dedup: bitta batch ichida bir hemis_id ikki marta bo'lmasligi kerak.
	{
		seen := make(map[int64]struct{}, len(items))
		out := make([]domain.Structure, 0, len(items))
		for _, it := range items {
			if _, ok := seen[it.HemisID]; ok {
				continue
			}
			seen[it.HemisID] = struct{}{}
			out = append(out, it)
		}
		items = out
	}

	// Sync oldidan mavjud hemis_id larni aniqlash (created vs updated hisobi uchun).
	hemisIDs := make([]int64, len(items))
	for i := range items {
		hemisIDs[i] = items[i].HemisID
	}
	var existing int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Structure{}).
		Where("hemis_id IN ?", hemisIDs).
		Count(&existing).Error; err != nil {
		return nil, err
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "hemis_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "code",
				"structure_type_code", "structure_type_name",
				"locality_type_code", "locality_type_name",
				"parent_hemis_id", "active", "raw", "synced_at", "updated_at",
			}),
		}).
		Create(&items).Error
	if err != nil {
		return nil, err
	}

	return &domain.SyncStats{
		Total:   len(items),
		Created: len(items) - int(existing),
		Updated: int(existing),
	}, nil
}

// RelinkParents — parent_hemis_id asosida parent_id ni bitta so'rovda bog'laydi.
func (r *structureRepository) RelinkParents(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE structures s
		SET parent_id = p.id
		FROM structures p
		WHERE s.parent_hemis_id IS NOT NULL
		  AND p.hemis_id = s.parent_hemis_id
		  AND (s.parent_id IS DISTINCT FROM p.id)
	`).Error
}

func (r *structureRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Structure, error) {
	var s domain.Structure
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *structureRepository) ListByType(ctx context.Context, typeCode string) ([]domain.Structure, error) {
	return r.ListByTypeAndParent(ctx, typeCode, nil)
}

func (r *structureRepository) ListByTypeAndParent(ctx context.Context, typeCode string, parentID *uuid.UUID) ([]domain.Structure, error) {
	q := r.db.WithContext(ctx).
		Where("structure_type_code = ? AND active = TRUE", typeCode)
	if parentID != nil {
		q = q.Where("parent_id = ?", *parentID)
	}

	var items []domain.Structure
	if err := q.Order("name ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
