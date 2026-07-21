package dto

// NewsRequest — admin yangilik yaratadi/tahrirlaydi.
//
// `views` va `author_id` bu yerda YO'Q: birinchisi tizim hisoblaydi,
// ikkinchisini token belgilaydi. Mijoz ularni yubora olmaydi (§17.3 #13).
type NewsRequest struct {
	Title    string `json:"title" binding:"required,min=3,max=255"`
	Excerpt  string `json:"excerpt" binding:"omitempty,max=500"`
	Body     string `json:"body" binding:"required,min=10"`
	CoverURL string `json:"cover_url" binding:"omitempty,url,max=1000"`
	Status   string `json:"status" binding:"omitempty,oneof=draft published"`
	// PublishedAt — bo'sh bo'lsa va status=published bo'lsa servis hozirgi
	// vaqtni qo'yadi. Kelajakdagi sana — rejalashtirilgan e'lon.
	PublishedAt *string `json:"published_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Pinned      bool    `json:"pinned"`
}
