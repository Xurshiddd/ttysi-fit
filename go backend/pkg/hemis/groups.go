package hemis

import (
	"context"
	"encoding/json"
	"fmt"
)

// GroupDTO — HEMIS guruhining kerakli maydonlari.
type GroupDTO struct {
	HemisID        int64
	Name           string
	FacultyHemisID *int64 // group.department = Fakultet
	SpecialtyCode  string
	SpecialtyName  string
	EducationLang  string
	Active         bool
}

type groupItem struct {
	ID         int64          `json:"id"`
	Name       string         `json:"name"`
	Department *departmentRef `json:"department"` // = Fakultet
	Specialty  *struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"specialty"`
	EducationLang *typeRef `json:"educationLang"`
	Active        bool     `json:"active"`
}

// FetchGroups — barcha guruhlarni tortib oladi.
func (c *Client) FetchGroups(ctx context.Context) ([]GroupDTO, error) {
	items, err := c.fetchAll(ctx, c.cfg.GroupPath, nil)
	if err != nil {
		return nil, err
	}

	out := make([]GroupDTO, 0, len(items))
	for _, raw := range items {
		var it groupItem
		if err := json.Unmarshal(raw, &it); err != nil {
			return nil, fmt.Errorf("hemis: group tahlil xatosi: %w", err)
		}
		dto := GroupDTO{
			HemisID:       it.ID,
			Name:          it.Name,
			Active:        it.Active,
			EducationLang: langName(it.EducationLang),
		}
		if it.Department != nil && it.Department.ID != 0 {
			fid := it.Department.ID
			dto.FacultyHemisID = &fid
		}
		if it.Specialty != nil {
			dto.SpecialtyCode = it.Specialty.Code
			dto.SpecialtyName = it.Specialty.Name
		}
		out = append(out, dto)
	}
	return out, nil
}

func langName(t *typeRef) string {
	if t == nil {
		return ""
	}
	return t.Name
}
