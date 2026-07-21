package repository

import (
	"context"
	"fmt"

	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
)

type analyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository — AnalyticsRepository portini qaytaradi.
func NewAnalyticsRepository(db *gorm.DB) domain.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// scope — analitikaga kiradigan foydalanuvchilar sharti va argumentlari.
//
// Faqat aktiv, o'chirilmagan foydalanuvchilar. Fakultet filtri ixtiyoriy.
// Shart satri kodda yozilgan constant, qiymatlar esa faqat parametr orqali
// beriladi — SQL injection yo'q (§3.2/§17.3 #1).
func scope(f domain.AnalyticsFilter) (string, []any) {
	where := "u.deleted_at IS NULL AND u.is_active = TRUE"
	var args []any
	if f.FacultyID != nil {
		where += " AND u.faculty_id = ?"
		args = append(args, *f.FacultyID)
	}
	return where, args
}

// Overview — umumiy raqamlar BITTA so'rovda (§13.4).
//
// TotalUsers alohida sanaladi: faollik yozmagan foydalanuvchi ham ro'yxatda
// bor, JOIN bilan sanalsa u yo'qolib qolardi.
func (r *analyticsRepository) Overview(ctx context.Context, f domain.AnalyticsFilter) (*domain.AnalyticsOverview, error) {
	where, args := scope(f)

	q := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(a.steps), 0)                    AS total_steps,
			COALESCE(SUM(a.distance_m), 0) / 1000.0      AS total_distance_km,
			COUNT(DISTINCT a.user_id)                    AS active_users,
			(SELECT COUNT(*) FROM users u WHERE %s)      AS total_users,
			COALESCE(
				SUM(a.steps) / NULLIF(COUNT(DISTINCT a.user_id), 0),
			0)                                           AS avg_steps_per_active
		FROM users u
		LEFT JOIN activities a
			ON a.user_id = u.id
			AND a.deleted_at IS NULL
			AND a.activity_date BETWEEN ? AND ?
		WHERE %s
	`, where, where)

	// Argument tartibi: avval ichki SELECT (total_users) sharti, keyin
	// JOIN sanalari, keyin tashqi WHERE sharti.
	all := make([]any, 0, len(args)*2+2)
	all = append(all, args...)
	all = append(all, f.From, f.To)
	all = append(all, args...)

	var out domain.AnalyticsOverview
	if err := r.db.WithContext(ctx).Raw(q, all...).Scan(&out).Error; err != nil {
		return nil, fmt.Errorf("analytics: overview: %w", err)
	}
	return &out, nil
}

// Timeseries — kunlik dinamika.
//
// generate_series bilan bo'sh kunlar ham 0 qiymat bilan qaytadi: aks holda
// grafikda kun tashlab ketilib, chiziq yolg'on ko'rinish berardi.
func (r *analyticsRepository) Timeseries(ctx context.Context, f domain.AnalyticsFilter) ([]domain.AnalyticsPoint, error) {
	where, args := scope(f)

	q := fmt.Sprintf(`
		WITH days AS (
			SELECT generate_series(?::date, ?::date, INTERVAL '1 day')::date AS d
		),
		act AS (
			SELECT a.activity_date AS d,
			       SUM(a.steps)              AS steps,
			       COUNT(DISTINCT a.user_id) AS active_users
			FROM activities a
			JOIN users u ON u.id = a.user_id
			WHERE a.deleted_at IS NULL
			  AND a.activity_date BETWEEN ?::date AND ?::date
			  AND %s
			GROUP BY a.activity_date
		)
		SELECT to_char(days.d, 'YYYY-MM-DD')    AS date,
		       COALESCE(act.steps, 0)           AS steps,
		       COALESCE(act.active_users, 0)    AS active_users
		FROM days
		LEFT JOIN act ON act.d = days.d
		ORDER BY days.d
	`, where)

	all := []any{f.From, f.To, f.From, f.To}
	all = append(all, args...)

	var rows []domain.AnalyticsPoint
	if err := r.db.WithContext(ctx).Raw(q, all...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("analytics: timeseries: %w", err)
	}
	return rows, nil
}

// ByFaculty — fakultetlar kesimi, jon boshiga o'rtacha bo'yicha saralangan.
//
// Fakultet = `users.faculty_id` ko'rsatgan `structures` yozuvi (eski
// `faculties` jadvali 00004-migratsiyada olib tashlangan; reyting moduli
// ham shu bog'lanishdan foydalanadi).
//
// Foydalanuvchisi bor, lekin faollik yozmagan fakultet ham ro'yxatda qoladi
// (0 qadam bilan) — "kim qatnashmayapti" degan savol ham hisobot uchun muhim.
func (r *analyticsRepository) ByFaculty(ctx context.Context, f domain.AnalyticsFilter) ([]domain.FacultyStat, error) {
	where, args := scope(f)

	q := fmt.Sprintf(`
		SELECT fa.id                                   AS faculty_id,
		       fa.name                                 AS name,
		       COALESCE(SUM(a.steps), 0)               AS total_steps,
		       COUNT(DISTINCT u.id)                    AS user_count,
		       COUNT(DISTINCT a.user_id)               AS active_users,
		       COALESCE(
		           SUM(a.steps) / NULLIF(COUNT(DISTINCT u.id), 0),
		       0)                                      AS avg_steps
		FROM structures fa
		JOIN users u ON u.faculty_id = fa.id
		LEFT JOIN activities a
			ON a.user_id = u.id
			AND a.deleted_at IS NULL
			AND a.activity_date BETWEEN ? AND ?
		WHERE fa.deleted_at IS NULL AND %s
		GROUP BY fa.id, fa.name
		ORDER BY avg_steps DESC, fa.name
	`, where)

	all := []any{f.From, f.To}
	all = append(all, args...)

	var rows []domain.FacultyStat
	if err := r.db.WithContext(ctx).Raw(q, all...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("analytics: by faculty: %w", err)
	}
	return rows, nil
}

// StreamUserActivity — eksport qatorlarini birma-bir beradi (xotirada
// butun hisobot yig'ilmaydi).
func (r *analyticsRepository) StreamUserActivity(ctx context.Context, f domain.AnalyticsFilter, fn func(domain.UserActivityRow) error) error {
	where, args := scope(f)

	q := fmt.Sprintf(`
		SELECT u.full_name,
		       u.email,
		       u.role,
		       COALESCE(fa.name, '')                     AS faculty,
		       COALESCE(de.name, '')                     AS department,
		       COALESCE(g.name, '')                      AS group_name,
		       COALESCE(SUM(a.steps), 0)                 AS total_steps,
		       COALESCE(SUM(a.distance_m), 0) / 1000.0   AS distance_km,
		       COUNT(a.id)                               AS active_days
		FROM users u
		LEFT JOIN structures fa ON fa.id = u.faculty_id
		LEFT JOIN structures de ON de.id = u.department_id
		LEFT JOIN groups     g  ON g.id  = u.group_id
		LEFT JOIN activities a
			ON a.user_id = u.id
			AND a.deleted_at IS NULL
			AND a.activity_date BETWEEN ? AND ?
		WHERE %s
		GROUP BY u.id, u.full_name, u.email, u.role, fa.name, de.name, g.name
		ORDER BY total_steps DESC, u.full_name
	`, where)

	all := []any{f.From, f.To}
	all = append(all, args...)

	rows, err := r.db.WithContext(ctx).Raw(q, all...).Rows()
	if err != nil {
		return fmt.Errorf("analytics: export: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		// Context bekor qilingan bo'lsa (mijoz uzildi) — yozishni to'xtatamiz.
		if err := ctx.Err(); err != nil {
			return err
		}
		var row domain.UserActivityRow
		if err := rows.Scan(
			&row.FullName, &row.Email, &row.Role,
			&row.Faculty, &row.Department, &row.GroupName,
			&row.TotalSteps, &row.DistanceKm, &row.ActiveDays,
		); err != nil {
			return fmt.Errorf("analytics: export: scan: %w", err)
		}
		if err := fn(row); err != nil {
			return err
		}
	}
	return rows.Err()
}
