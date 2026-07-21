package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type activityRepository struct {
	db *gorm.DB
}

// NewActivityRepository — ActivityRepository portini qaytaradi.
func NewActivityRepository(db *gorm.DB) domain.ActivityRepository {
	return &activityRepository{db: db}
}

// mergeConflict — (user_id, activity_date) bo'yicha ko'tarib-yangilash qoidasi.
//
// Ustiga yozish (oddiy SET) EMAS, balki GREATEST: kunlik hisoblagich kun
// davomida faqat o'sadi, kamaymaydi. Ustiga yozilsa quyidagilar ma'lumotni
// yo'q qilardi:
//   - foydalanuvchi ikkita qurilmadan sinxron qilsa (planshetda 200 qadam
//     telefondagi 12 000 ni o'chirib yuborardi);
//   - Health Connect ruxsati qayta berilganda vaqtincha kichik qiymat qaytarsa;
//   - mijoz eski (keshlangan) o'qishni kech yuborsa.
//
// source — oxirgi yozuvchiniki: qiymat qaysi qiymat ustun kelganidan qat'i
// nazar, oxirgi sinxron manbasini bilish diagnostika uchun foydali.
func mergeConflict() clause.OnConflict {
	return clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "activity_date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"steps":      gorm.Expr("GREATEST(activities.steps, EXCLUDED.steps)"),
			"calories":   gorm.Expr("GREATEST(activities.calories, EXCLUDED.calories)"),
			"distance_m": gorm.Expr("GREATEST(activities.distance_m, EXCLUDED.distance_m)"),
			"active_min": gorm.Expr("GREATEST(activities.active_min, EXCLUDED.active_min)"),
			"source":     gorm.Expr("EXCLUDED.source"),
			"updated_at": gorm.Expr("NOW()"),
		}),
	}
}

// Upsert — (user_id, activity_date) bo'yicha ko'tarib-yangilaydi.
func (r *activityRepository) Upsert(ctx context.Context, a *domain.Activity) error {
	return r.db.WithContext(ctx).Clauses(mergeConflict()).Create(a).Error
}

// UpsertMany — bir necha kunlik yozuvni BITTA INSERT ... ON CONFLICT so'rovida
// yozadi (CLAUDE.md §3.1 — sikl ichida DB so'rovi bo'lmasin).
//
// Chaqiruvchi rows ichida sana takrorlanmasligini kafolatlashi shart:
// PostgreSQL bir so'rovda bitta qatorni ikki marta yangilashga yo'l qo'ymaydi.
func (r *activityRepository) UpsertMany(ctx context.Context, rows []domain.Activity) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(mergeConflict()).Create(&rows).Error
}

// ListByUser — [from, to] oralig'idagi yozuvlar, sana kamayish tartibida.
func (r *activityRepository) ListByUser(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]domain.Activity, error) {
	if limit <= 0 || limit > 366 {
		limit = 31
	}
	var rows []domain.Activity
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL AND activity_date BETWEEN ? AND ?", userID, from, to).
		Order("activity_date DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

// DeleteRange — oraliqdagi faollik yozuvlarini butunlay o'chiradi.
//
// Soft delete EMAS: upsert (user_id, activity_date) bo'yicha noyob indeksga
// tayanadi, "o'chirilgan" qator qolsa telefon o'sha kunni qayta yubora
// olmasdi va tuzatishning ma'nosi yo'qolardi.
//
// Ma'lumot yo'qolmaydi: telefon oxirgi 7 kunni backfill bilan qayta
// yuboradi (kBackfillDays), ya'ni haqiqiy qiymatlar o'z-o'zidan tiklanadi.
func (r *activityRepository) DeleteRange(ctx context.Context, userID uuid.UUID, from, to time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("user_id = ? AND activity_date BETWEEN ? AND ?", userID, from, to).
		Delete(&domain.Activity{})
	if res.Error != nil {
		return 0, fmt.Errorf("activity: delete range: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// Stats — bugun/hafta/oy/jami yig'ma (bitta so'rov, GROUP BY'siz FILTER bilan).
//
// "Bugun" DB serverining CURRENT_DATE i emas, chaqiruvchi bergan today:
// mahalliy mintaqada (APP_TIMEZONE) kun 00:00 da almashadi, UTC da esa
// O'zbekiston uchun mahalliy 05:00 da — natijada tunda statistika
// noto'g'ri kunni ko'rsatardi.
func (r *activityRepository) Stats(ctx context.Context, userID uuid.UUID, today time.Time) (*domain.ActivityStats, error) {
	d := today.Format("2006-01-02")
	weekStart := today.AddDate(0, 0, -6).Format("2006-01-02")
	monthStart := today.AddDate(0, 0, -29).Format("2006-01-02")

	var s domain.ActivityStats
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(steps)      FILTER (WHERE activity_date = ?), 0) AS today_steps,
			COALESCE(SUM(calories)   FILTER (WHERE activity_date = ?), 0) AS today_calories,
			COALESCE(SUM(distance_m) FILTER (WHERE activity_date = ?), 0) AS today_distance_m,
			COALESCE(SUM(active_min) FILTER (WHERE activity_date = ?), 0) AS today_active_min,
			COALESCE(SUM(steps) FILTER (WHERE activity_date >= ?), 0) AS week_steps,
			COALESCE(SUM(steps) FILTER (WHERE activity_date >= ?), 0) AS month_steps,
			COALESCE(SUM(steps), 0) AS total_steps
		FROM activities
		WHERE user_id = ? AND deleted_at IS NULL
	`, d, d, d, d, weekStart, monthStart, userID).Scan(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}
