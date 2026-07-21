package hemis

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// RosterDTO — talaba yoki o'qituvchining kerakli maydonlari (users ga moslangan).
type RosterDTO struct {
	HemisID    int64
	HemisLogin string
	FullName   string
	Role       string // "student" | "teacher"
	Email      string

	FacultyHemisID    *int64 // talaba/o'qituvchi fakulteti
	DepartmentHemisID *int64 // o'qituvchi kafedrasi
	GroupHemisID      *int64 // talaba guruhi
	Course            *int16

	Gender    string
	BirthDate *time.Time
	Position  string
	Specialty string
	AvatarURL string
	IsActive  bool

	// EmploymentFormCode — xodim ish shakli ("11" = Asosiy ish joy). Dedup uchun.
	EmploymentFormCode string
}

// ── Talaba ─────────────────────────────────────────────

type studentItem struct {
	ID              int64          `json:"id"`
	FullName        string         `json:"full_name"`
	StudentIDNumber string         `json:"student_id_number"`
	Image           string         `json:"image"`
	ImageFull       string         `json:"image_full"`
	Email           string         `json:"email"`
	Gender          *typeRef       `json:"gender"`
	BirthDate       int64          `json:"birth_date"`
	Department      *departmentRef `json:"department"` // = Fakultet
	Specialty       *struct {
		Name string `json:"name"`
	} `json:"specialty"`
	Group *struct {
		ID int64 `json:"id"`
	} `json:"group"`
	Level         *typeRef `json:"level"` // name "2-kurs"
	StudentStatus *typeRef `json:"studentStatus"`
}

// FetchStudents — barcha talabalarni tortib oladi.
func (c *Client) FetchStudents(ctx context.Context) ([]RosterDTO, error) {
	items, err := c.fetchAll(ctx, c.cfg.StudentPath, nil)
	if err != nil {
		return nil, err
	}

	out := make([]RosterDTO, 0, len(items))
	for _, raw := range items {
		var it studentItem
		if err := json.Unmarshal(raw, &it); err != nil {
			return nil, fmt.Errorf("hemis: student tahlil xatosi: %w", err)
		}
		d := RosterDTO{
			HemisID:    it.ID,
			HemisLogin: it.StudentIDNumber,
			FullName:   it.FullName,
			Role:       "student",
			Email:      it.Email,
			Gender:     genderFromCode(it.Gender),
			BirthDate:  unixToTime(it.BirthDate),
			AvatarURL:  avatar(it.Image, it.ImageFull),
			IsActive:   it.StudentStatus != nil && it.StudentStatus.Code == "11",
		}
		if it.Department != nil && it.Department.ID != 0 {
			fid := it.Department.ID
			d.FacultyHemisID = &fid
		}
		if it.Group != nil && it.Group.ID != 0 {
			gid := it.Group.ID
			d.GroupHemisID = &gid
		}
		if it.Specialty != nil {
			d.Specialty = it.Specialty.Name
		}
		if it.Level != nil {
			d.Course = parseCourse(it.Level.Name)
		}
		out = append(out, d)
	}
	return out, nil
}

// ── O'qituvchi / xodim ─────────────────────────────────

type employeeItem struct {
	ID               int64          `json:"id"`
	FullName         string         `json:"full_name"`
	EmployeeIDNumber string         `json:"employee_id_number"`
	Image            string         `json:"image"`
	ImageFull        string         `json:"image_full"`
	Gender           *typeRef       `json:"gender"`
	BirthDate        int64          `json:"birth_date"`
	Specialty        string         `json:"specialty"` // xodimda matn (talabada obyekt)
	Department       *departmentRef `json:"department"` // = Kafedra
	StaffPosition    *typeRef       `json:"staffPosition"`
	EmploymentForm   *typeRef       `json:"employmentForm"` // "11" = Asosiy ish joy
	EmployeeStatus   *typeRef       `json:"employeeStatus"`
	Active           bool           `json:"active"`
}

// FetchEmployees — barcha xodim/o'qituvchilarni tortib oladi (?type=all).
func (c *Client) FetchEmployees(ctx context.Context) ([]RosterDTO, error) {
	extra := url.Values{}
	if c.cfg.EmployeeType != "" {
		extra.Set("type", c.cfg.EmployeeType)
	}

	items, err := c.fetchAll(ctx, c.cfg.EmployeePath, extra)
	if err != nil {
		return nil, err
	}

	out := make([]RosterDTO, 0, len(items))
	for _, raw := range items {
		var it employeeItem
		if err := json.Unmarshal(raw, &it); err != nil {
			return nil, fmt.Errorf("hemis: employee tahlil xatosi: %w", err)
		}
		d := RosterDTO{
			HemisID:    it.ID,
			HemisLogin: it.EmployeeIDNumber,
			FullName:   it.FullName,
			Role:       "employee",
			Specialty:  it.Specialty,
			Gender:     genderFromCode(it.Gender),
			BirthDate:  unixToTime(it.BirthDate),
			AvatarURL:  avatar(it.Image, it.ImageFull),
			IsActive:   it.Active,
		}
		if it.Department != nil && it.Department.ID != 0 {
			did := it.Department.ID
			d.DepartmentHemisID = &did
		}
		if it.StaffPosition != nil {
			d.Position = it.StaffPosition.Name
		}
		if it.EmploymentForm != nil {
			d.EmploymentFormCode = it.EmploymentForm.Code
		}
		out = append(out, d)
	}
	return out, nil
}

// ── Yordamchilar ───────────────────────────────────────

// genderFromCode — HEMIS gender kodi: 11=erkak, 12=ayol.
func genderFromCode(t *typeRef) string {
	if t == nil {
		return ""
	}
	switch t.Code {
	case "11":
		return "male"
	case "12":
		return "female"
	default:
		return ""
	}
}

func unixToTime(sec int64) *time.Time {
	if sec <= 0 {
		return nil
	}
	t := time.Unix(sec, 0).UTC()
	return &t
}

// avatar — HEMIS rasm URL'ini tanlaydi.
// HEMIS ikki maydon qaytaradi:
//   - image      → kichik crop/thumbnail (/static/crop/.../320__90_*.jpg) — avatar uchun MOS
//   - image_full → to'liq original (/static/uploads/pi/...) — og'irroq, bizga kerak emas
// Avatar sifatida crop afzal; bo'sh bo'lsa full'ga, u ham bo'sh bo'lsa "" ga tushadi.
// "" qaytsa, frontend default profil rasmni ko'rsatadi (UserAvatar komponenti).
func avatar(crop, full string) string {
	if crop != "" {
		return crop
	}
	return full
}

var courseRe = regexp.MustCompile(`\d+`)

// parseCourse — "2-kurs" dan 2 ni ajratib oladi.
func parseCourse(levelName string) *int16 {
	m := courseRe.FindString(levelName)
	if m == "" {
		return nil
	}
	n, err := strconv.Atoi(m)
	if err != nil {
		return nil
	}
	c := int16(n)
	return &c
}
