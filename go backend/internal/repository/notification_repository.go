package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) List(ctx context.Context, userID uuid.UUID, f domain.NotificationFilter) ([]domain.Notification, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userID)
	if f.UnreadOnly {
		q = q.Where("read_at IS NULL")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("notification: count: %w", err)
	}

	var rows []domain.Notification
	err := q.Order("created_at DESC").
		Limit(f.Limit).Offset((f.Page - 1) * f.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("notification: list: %w", err)
	}
	return rows, total, nil
}

func (r *notificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&n).Error
	if err != nil {
		return 0, fmt.Errorf("notification: unread count: %w", err)
	}
	return n, nil
}

// MarkRead — egalik SHART tekshiriladi (§17.3 #26 IDOR): user_id shart
// ichida, ya'ni boshqaning xabarini o'qilgan qilib bo'lmaydi.
//
// Allaqachon o'qilgan bo'lsa `read_at` qayta yozilmaydi — birinchi
// o'qilgan vaqt saqlanib qoladi.
func (r *notificationRepository) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	res := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", id, userID).
		Update("read_at", time.Now())
	if res.Error != nil {
		return fmt.Errorf("notification: mark read: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		// Yo'q, boshqaniki yoki allaqachon o'qilgan — uchalasi ham
		// mijoz uchun bir xil: qo'shimcha ma'lumot sizdirmaymiz.
		return nil
	}
	return nil
}

func (r *notificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	res := r.db.WithContext(ctx).Model(&domain.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", time.Now())
	if res.Error != nil {
		return 0, fmt.Errorf("notification: mark all read: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// Create — bitta xabar. ref_id bo'lsa idempotent (unique indeks orqali).
func (r *notificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	err := r.db.WithContext(ctx).Create(n).Error
	if isUniqueViolation(err) {
		// Shu manba uchun xabar allaqachon yuborilgan.
		return domain.ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("notification: create: %w", err)
	}
	return nil
}

// Broadcast — target bo'yicha barcha foydalanuvchiga BITTA so'rovda yozadi.
//
// `INSERT ... SELECT`: 9 600 foydalanuvchiga e'lon yuborishda 9 600 ta
// alohida INSERT (§3.1 taqiqlaydi) o'rniga bitta so'rov ketadi.
func (r *notificationRepository) Broadcast(ctx context.Context, t domain.AnnouncementTarget, n domain.Notification) (int64, error) {
	where := "deleted_at IS NULL AND is_active = TRUE"
	var args []any

	if t.FacultyID != nil {
		where += " AND faculty_id = ?"
		args = append(args, *t.FacultyID)
	}
	if t.GroupID != nil {
		where += " AND group_id = ?"
		args = append(args, *t.GroupID)
	}
	if t.Role != "" {
		where += " AND role = ?"
		args = append(args, t.Role)
	}

	q := fmt.Sprintf(`
		INSERT INTO notifications (user_id, type, title, body, ref_type, metadata)
		SELECT id, ?, ?, ?, ?, ?
		FROM users
		WHERE %s
	`, where)

	all := []any{n.Type, n.Title, n.Body, n.RefType, n.Metadata}
	all = append(all, args...)

	res := r.db.WithContext(ctx).Exec(q, all...)
	if res.Error != nil {
		return 0, fmt.Errorf("notification: broadcast: %w", res.Error)
	}
	return res.RowsAffected, nil
}
