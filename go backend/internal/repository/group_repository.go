package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) domain.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) UpsertBatch(ctx context.Context, items []domain.Group) (*domain.SyncStats, error) {
	if len(items) == 0 {
		return &domain.SyncStats{}, nil
	}
	items = dedupGroups(items)

	hemisIDs := make([]int64, len(items))
	for i := range items {
		hemisIDs[i] = items[i].HemisID
	}
	var existing int64
	if err := r.db.WithContext(ctx).
		Model(&domain.Group{}).
		Where("hemis_id IN ?", hemisIDs).
		Count(&existing).Error; err != nil {
		return nil, err
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "hemis_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "faculty_hemis_id",
				"specialty_code", "specialty_name", "education_lang",
				"active", "synced_at", "updated_at",
			}),
		}).
		CreateInBatches(&items, 500).Error
	if err != nil {
		return nil, err
	}

	return &domain.SyncStats{
		Total:   len(items),
		Created: len(items) - int(existing),
		Updated: int(existing),
	}, nil
}

// dedupGroups — hemis_id bo'yicha takrorlarni olib tashlaydi (birinchisi qoladi).
func dedupGroups(items []domain.Group) []domain.Group {
	seen := make(map[int64]struct{}, len(items))
	out := make([]domain.Group, 0, len(items))
	for _, it := range items {
		if _, ok := seen[it.HemisID]; ok {
			continue
		}
		seen[it.HemisID] = struct{}{}
		out = append(out, it)
	}
	return out
}

func (r *groupRepository) RelinkFaculties(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE groups g
		SET faculty_id = s.id
		FROM structures s
		WHERE g.faculty_hemis_id IS NOT NULL
		  AND s.hemis_id = g.faculty_hemis_id
		  AND (g.faculty_id IS DISTINCT FROM s.id)
	`).Error
}

func (r *groupRepository) ListByFaculty(ctx context.Context, facultyID *uuid.UUID) ([]domain.Group, error) {
	q := r.db.WithContext(ctx).Where("active = TRUE")
	if facultyID != nil {
		q = q.Where("faculty_id = ?", *facultyID)
	}
	var items []domain.Group
	if err := q.Order("name ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
