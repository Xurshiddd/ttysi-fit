package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"go.uber.org/zap"
)

// NotificationService — ilova ichidagi bildirishnomalar.
type NotificationService struct {
	repo domain.NotificationRepository
	log  *zap.Logger
}

func NewNotificationService(repo domain.NotificationRepository, log *zap.Logger) *NotificationService {
	return &NotificationService{repo: repo, log: log}
}

func (s *NotificationService) List(ctx context.Context, userID uuid.UUID, f domain.NotificationFilter) ([]domain.Notification, int64, error) {
	return s.repo.List(ctx, userID, f)
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.UnreadCount(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.MarkRead(ctx, userID, id)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.MarkAllRead(ctx, userID)
}

// Notify — tizim hodisasi uchun xabar yozadi.
//
// XATO QAYTARMAYDI (ataylab). Bildirishnoma — YON TA'SIR: sovg'a
// topshirilgani yoki yutuq berilgani xabar yozilmagani uchun BEKOR
// QILINMASLIGI kerak. Xato faqat loglanadi.
//
// ref berilgan bo'lsa idempotent: takroriy hodisada ikkinchi xabar
// qo'shilmaydi (DB unique indeksi).
func (s *NotificationService) Notify(ctx context.Context, n domain.Notification) {
	if !domain.ValidNotificationType(n.Type) {
		if s.log != nil {
			s.log.Warn("noma'lum bildirishnoma turi", zap.String("type", n.Type))
		}
		return
	}

	err := s.repo.Create(ctx, &n)
	if err == nil || errors.Is(err, domain.ErrAlreadyExists) {
		return // yozildi yoki allaqachon bor — ikkalasi ham normal
	}
	if s.log != nil {
		s.log.Warn("bildirishnoma yozilmadi",
			zap.String("type", n.Type),
			zap.String("user_id", n.UserID.String()),
			zap.Error(err))
	}
}

// Announce — admin e'loni. Yuborilgan xabarlar sonini qaytaradi.
func (s *NotificationService) Announce(ctx context.Context, t domain.AnnouncementTarget, title, body string) (int64, error) {
	if title == "" {
		return 0, fmt.Errorf("%w: sarlavha bo'sh", domain.ErrValidation)
	}
	if t.Role != "" && t.Role != string(domain.RoleStudent) && t.Role != string(domain.RoleEmployee) {
		return 0, fmt.Errorf("%w: rol", domain.ErrValidation)
	}

	n := domain.Notification{
		Type:  domain.NotifyAnnouncement,
		Title: title,
		Body:  body,
	}
	sent, err := s.repo.Broadcast(ctx, t, n)
	if err != nil {
		return 0, fmt.Errorf("NotificationService.Announce: %w", err)
	}
	return sent, nil
}
