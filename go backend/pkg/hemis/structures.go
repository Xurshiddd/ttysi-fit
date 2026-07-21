package hemis

import (
	"context"
	"encoding/json"
	"fmt"
)

// StructureDTO — HEMIS strukturasining tahlil qilingan ko'rinishi.
type StructureDTO struct {
	HemisID           int64
	Name              string
	Code              string
	StructureTypeCode string
	StructureTypeName string
	LocalityTypeCode  string
	LocalityTypeName  string
	ParentHemisID     *int64
	Active            bool
	Raw               []byte
}

type structureItem struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Code          string   `json:"code"`
	StructureType *typeRef `json:"structureType"`
	LocalityType  *typeRef `json:"localityType"`
	// parent — HEMIS uni raqam (ota id) yoki null qilib qaytaradi.
	Parent *int64 `json:"parent"`
	Active bool   `json:"active"`
}

// FetchStructures — barcha strukturalarni (paginatsiya bo'ylab) tortib oladi.
func (c *Client) FetchStructures(ctx context.Context) ([]StructureDTO, error) {
	items, err := c.fetchAll(ctx, c.cfg.StructurePath, nil)
	if err != nil {
		return nil, err
	}

	out := make([]StructureDTO, 0, len(items))
	for _, raw := range items {
		var it structureItem
		if err := json.Unmarshal(raw, &it); err != nil {
			return nil, fmt.Errorf("hemis: structure tahlil xatosi: %w", err)
		}
		dto := StructureDTO{
			HemisID: it.ID,
			Name:    it.Name,
			Code:    it.Code,
			Active:  it.Active,
			Raw:     raw,
		}
		if it.StructureType != nil {
			dto.StructureTypeCode = it.StructureType.Code
			dto.StructureTypeName = it.StructureType.Name
		}
		if it.LocalityType != nil {
			dto.LocalityTypeCode = it.LocalityType.Code
			dto.LocalityTypeName = it.LocalityType.Name
		}
		if it.Parent != nil && *it.Parent != 0 {
			pid := *it.Parent
			dto.ParentHemisID = &pid
		}
		out = append(out, dto)
	}
	return out, nil
}
