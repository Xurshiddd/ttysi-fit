package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Yangilik holati.
const (
	NewsStatusDraft     = "draft"
	NewsStatusPublished = "published"
)

func ValidNewsStatus(s string) bool {
	switch s {
	case NewsStatusDraft, NewsStatusPublished:
		return true
	}
	return false
}

// News — yangilik. Tur registri yo'q: turga qarab o'zgaradigan maydonlari yo'q.
type News struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	Title   string `json:"title" gorm:"not null"`
	Excerpt string `json:"excerpt,omitempty"`
	Body    string `json:"body" gorm:"not null"`

	CoverURL    string     `json:"cover_url,omitempty"`
	Status      string     `json:"status" gorm:"not null"`
	PublishedAt *time.Time `json:"published_at,omitempty"`

	AuthorID *uuid.UUID `json:"author_id,omitempty" gorm:"type:uuid"`
	Pinned   bool       `json:"pinned" gorm:"not null"`
	Views    int        `json:"views" gorm:"not null"`

	Metadata datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
}

func (News) TableName() string { return "news" }

// NewsListItem — ro'yxat uchun yengil read-model.
//
// `body` ataylab YO'Q: yangilik matni uzun bo'lishi mumkin va ro'yxatda kerak
// emas. 20 ta yangilikning to'liq matnini yuborish trafikni behuda sarflardi
// (§17.3 #37 — excessive data exposure).
type NewsListItem struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Excerpt     string     `json:"excerpt"`
	CoverURL    string     `json:"cover_url"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at"`
	Pinned      bool       `json:"pinned"`
	Views       int        `json:"views"`
}

// MakeExcerpt — excerpt bo'sh bo'lsa body'dan qisqa matn yasaydi.
//
// So'z chegarasida kesadi: o'rtasidan kesish ("Universitet musoba...") xunuk
// ko'rinadi. Kirill/lotin harflari uchun rune bo'yicha sanaymiz — bayt bo'yicha
// kesish ko'p baytli harfni buzardi.
func MakeExcerpt(body string, maxRunes int) string {
	body = strings.TrimSpace(strings.Join(strings.Fields(body), " "))
	if body == "" {
		return ""
	}
	runes := []rune(body)
	if len(runes) <= maxRunes {
		return body
	}

	cut := string(runes[:maxRunes])
	// Oxirgi bo'sh joygacha qaytamiz — so'z o'rtasida kesilmasin.
	if i := strings.LastIndex(cut, " "); i > 0 {
		cut = cut[:i]
	}
	return cut + "…"
}

// NewsFilter — ro'yxat so'rovi.
type NewsFilter struct {
	Status string
	Search string
	Page   int
	Limit  int
	// PublishedOnly — mobil ro'yxat: faqat e'lon qilingan va vaqti kelganlar.
	PublishedOnly bool
}

// NewsRepository — yangiliklar uchun port.
type NewsRepository interface {
	Create(ctx context.Context, n *News) error
	Update(ctx context.Context, n *News) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// GetByID — to'liq yozuv (body bilan).
	GetByID(ctx context.Context, id uuid.UUID) (*News, error)
	List(ctx context.Context, f NewsFilter) ([]NewsListItem, int64, error)
	// IncrementViews — ko'rishlar sonini oshiradi (atomik UPDATE).
	IncrementViews(ctx context.Context, id uuid.UUID) error
}
