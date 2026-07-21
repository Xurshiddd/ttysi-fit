package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Structure — HEMIS tashkiliy birligi (fakultet, kafedra, bo'lim...).
// Butun tashkilot daraxti shu jadvalda; parent_id orqali bog'lanadi.
type Structure struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	// HEMIS sync kaliti
	HemisID int64 `json:"hemis_id" gorm:"uniqueIndex;not null"`

	Name string `json:"name" gorm:"not null"`
	Code string `json:"code"`

	StructureTypeCode string `json:"structure_type_code"`
	StructureTypeName string `json:"structure_type_name"`
	LocalityTypeCode  string `json:"locality_type_code"`
	LocalityTypeName  string `json:"locality_type_name"`

	ParentID      *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid"`
	ParentHemisID *int64     `json:"parent_hemis_id,omitempty"`

	Active   bool           `json:"active" gorm:"not null"`
	Raw      datatypes.JSON `json:"-" gorm:"type:jsonb"`
	SyncedAt *time.Time     `json:"synced_at,omitempty"`
}

func (Structure) TableName() string { return "structures" }

// SyncStats — sync natijasi xulosasi.
type SyncStats struct {
	Total   int `json:"total"`
	Created int `json:"created"`
	Updated int `json:"updated"`
}

// StructureRepository — strukturalar uchun port (interfeys).
type StructureRepository interface {
	// UpsertBatch — hemis_id bo'yicha ko'tarib-yangilaydi (bitta so'rovda, N+1 yo'q).
	UpsertBatch(ctx context.Context, items []Structure) (*SyncStats, error)
	// RelinkParents — parent_hemis_id bo'yicha parent_id ni bitta SQL bilan bog'laydi.
	RelinkParents(ctx context.Context) error
	GetByID(ctx context.Context, id uuid.UUID) (*Structure, error)
	// ListByType — berilgan structure_type_code bo'yicha ro'yxat (masalan fakultetlar).
	ListByType(ctx context.Context, typeCode string) ([]Structure, error)
	// ListByTypeAndParent — type bo'yicha, ixtiyoriy parent filtri bilan
	// (masalan fakultetga tegishli kafedralar). parentID nil bo'lsa hammasi.
	ListByTypeAndParent(ctx context.Context, typeCode string, parentID *uuid.UUID) ([]Structure, error)
}
