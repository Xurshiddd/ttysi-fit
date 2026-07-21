package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Group — HEMIS o'quv guruhi (/data/group-list). Guruh fakultetga tegishli.
type Group struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	HemisID int64  `json:"hemis_id" gorm:"uniqueIndex;not null"`
	Name    string `json:"name" gorm:"not null"`

	FacultyID      *uuid.UUID `json:"faculty_id,omitempty" gorm:"type:uuid"`
	FacultyHemisID *int64     `json:"-"`

	SpecialtyCode string `json:"specialty_code,omitempty"`
	SpecialtyName string `json:"specialty_name,omitempty"`
	EducationLang string `json:"education_lang,omitempty"`

	Active   bool       `json:"active" gorm:"not null"`
	SyncedAt *time.Time `json:"synced_at,omitempty"`
}

func (Group) TableName() string { return "groups" }

// GroupRepository — guruhlar uchun port (interfeys).
type GroupRepository interface {
	UpsertBatch(ctx context.Context, items []Group) (*SyncStats, error)
	// RelinkFaculties — faculty_hemis_id → faculty_id (structures) bitta SQL bilan.
	RelinkFaculties(ctx context.Context) error
	ListByFaculty(ctx context.Context, facultyID *uuid.UUID) ([]Group, error)
}
