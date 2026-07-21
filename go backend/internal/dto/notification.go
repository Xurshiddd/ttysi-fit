package dto

// AnnouncementRequest — admin e'lon yuboradi.
//
// Maqsad maydonlari bo'sh bo'lsa cheklov yo'q: uchalasi ham bo'sh bo'lsa
// e'lon BARCHA aktiv foydalanuvchilarga boradi. Shu sababli admin panel
// yuborishdan oldin qancha odamga ketishini ko'rsatishi kerak.
type AnnouncementRequest struct {
	Title string `json:"title" binding:"required,min=3,max=255"`
	Body  string `json:"body" binding:"omitempty,max=2000"`

	FacultyID string `json:"faculty_id" binding:"omitempty,uuid"`
	GroupID   string `json:"group_id" binding:"omitempty,uuid"`
	Role      string `json:"role" binding:"omitempty,oneof=student employee"`
}
