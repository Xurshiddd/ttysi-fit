package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) domain.SessionRepository {
	return &sessionRepository{db: db}
}

// ActiveOther — boshqa qurilmadagi faol sessiya (eng yangisi).
func (r *sessionRepository) ActiveOther(ctx context.Context, userID uuid.UUID, deviceID string) (*domain.UserSession, error) {
	var s domain.UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_id <> ? AND revoked_at IS NULL", userID, deviceID).
		Order("last_seen_at DESC").
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // konflikt yo'q
	}
	if err != nil {
		return nil, fmt.Errorf("session: active other: %w", err)
	}
	return &s, nil
}

func (r *sessionRepository) ListActive(ctx context.Context, userID uuid.UUID) ([]domain.UserSession, error) {
	var rows []domain.UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Order("last_seen_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("session: list: %w", err)
	}
	return rows, nil
}

// Upsert — qurilma uchun sessiya.
//
// (user_id, device_id) bo'yicha unique indeks bor (faqat faol qatorlarda),
// shuning uchun bir qurilmadan qayta kirishda yangi qator emas, o'shaning
// o'zi yangilanadi — "Mening qurilmalarim" ro'yxati takrorlanmaydi.
func (r *sessionRepository) Upsert(ctx context.Context, userID uuid.UUID, d domain.DeviceInfo) (*domain.UserSession, error) {
	now := time.Now()
	s := &domain.UserSession{
		UserID:     userID,
		DeviceID:   d.DeviceID,
		DeviceName: d.DeviceName,
		Platform:   d.Platform,
		AppVersion: d.AppVersion,
		IP:         d.IP,
		UserAgent:  d.UserAgent,
		LastSeenAt: now,
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "device_id"}},
			// Partial unique indeks uchun shart: faqat faol qatorlar.
			TargetWhere: clause.Where{Exprs: []clause.Expression{
				clause.Expr{SQL: "revoked_at IS NULL"},
			}},
			DoUpdates: clause.AssignmentColumns([]string{
				"device_name", "platform", "app_version", "ip", "user_agent", "last_seen_at",
			}),
		}).
		Create(s).Error
	if err != nil {
		return nil, fmt.Errorf("session: upsert: %w", err)
	}
	return s, nil
}

func (r *sessionRepository) RevokeOthers(ctx context.Context, userID uuid.UUID, deviceID, reason string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&domain.UserSession{}).
		Where("user_id = ? AND device_id <> ? AND revoked_at IS NULL", userID, deviceID).
		Updates(map[string]any{"revoked_at": time.Now(), "revoked_reason": reason})
	if res.Error != nil {
		return 0, fmt.Errorf("session: revoke others: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// Revoke — bitta sessiyani yopadi. Egalik SHART tekshiriladi
// (§17.3 #26 IDOR): boshqaning sessiyasini yopib bo'lmaydi.
func (r *sessionRepository) Revoke(ctx context.Context, userID, sessionID uuid.UUID, reason string) error {
	res := r.db.WithContext(ctx).Model(&domain.UserSession{}).
		Where("id = ? AND user_id = ? AND revoked_at IS NULL", sessionID, userID).
		Updates(map[string]any{"revoked_at": time.Now(), "revoked_reason": reason})
	if res.Error != nil {
		return fmt.Errorf("session: revoke: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *sessionRepository) Touch(ctx context.Context, userID uuid.UUID, deviceID string) error {
	return r.db.WithContext(ctx).Model(&domain.UserSession{}).
		Where("user_id = ? AND device_id = ? AND revoked_at IS NULL", userID, deviceID).
		Update("last_seen_at", time.Now()).Error
}
