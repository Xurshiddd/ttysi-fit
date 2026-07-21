package dto

// TrainingRequest — admin mashg'ulot yaratadi/tahrirlaydi.
// `views` bu yerda yo'q — uni tizim hisoblaydi (§17.3 #13).
type TrainingRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"omitempty,max=5000"`
	// Category — erkin matn: yangi kategoriya qo'shish uchun kod o'zgarmaydi (§16).
	Category string `json:"category" binding:"omitempty,max=100"`
	Level    string `json:"level" binding:"omitempty,oneof=beginner intermediate advanced"`

	VideoURL     string `json:"video_url" binding:"required,url,max=1000"`
	ThumbnailURL string `json:"thumbnail_url" binding:"omitempty,url,max=1000"`
	DurationMin  *int16 `json:"duration_min" binding:"omitempty,gt=0"`

	Status      string  `json:"status" binding:"omitempty,oneof=draft published"`
	PublishedAt *string `json:"published_at" binding:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	SortOrder   int     `json:"sort_order"`
}
