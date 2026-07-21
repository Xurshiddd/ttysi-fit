package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
)

type ratingRepository struct {
	db *gorm.DB
}

// NewRatingRepository — RatingRepository portini qaytaradi.
func NewRatingRepository(db *gorm.DB) domain.RatingRepository {
	return &ratingRepository{db: db}
}

// trustedSourceCond — reytingga FAQAT qurilmadan kelgan faollik kiradi.
//
// NEGA: `POST /activities` istalgan foydalanuvchiga o'z qadamini yozishga
// ruxsat beradi (qo'lda kiritish uchun mo'ljallangan edi), upsert esa
// GREATEST — ya'ni bir marta yozilgan katta qiymatni qaytarib bo'lmaydi.
// Shu ikkisi birga reytingni ochiq qoldirardi: "200000 qadam" yuborgan
// foydalanuvchi abadiy birinchi o'rinda turaverardi.
//
// Qo'lda kiritilgan faollik SHAXSIY statistikada qoladi (foydalanuvchi
// o'zi uchun yuritishi mumkin), lekin musobaqa reytingiga kirmaydi.
// Manba ro'yxati kodda: uni mijoz belgilaydi, admin siyosati emas.
const trustedSourceCond = " AND a.source IN ('health_connect', 'healthkit')"

// periodCond — davr va manba bo'yicha activities JOIN sharti
// (CLAUDE.md §13.4 namunasi).
//
// Davr qiymati handler'da enum bilan validatsiya qilinadi — SQL'ga user input
// tushmaydi (§3.2), bu yerda faqat oldindan yozilgan constant satrlar.
func periodCond(p domain.RatingPeriod) string {
	switch p {
	case domain.PeriodWeek:
		return " AND a.activity_date >= CURRENT_DATE - INTERVAL '6 days'" + trustedSourceCond
	case domain.PeriodMonth:
		return " AND a.activity_date >= CURRENT_DATE - INTERVAL '29 days'" + trustedSourceCond
	default: // PeriodAll
		return trustedSourceCond
	}
}

// ListIndividual — talaba/xodim shaxsiy reytingi: jami qadam bo'yicha,
// RANK() bilan bitta so'rov (N+1 yo'q, §3.1/§13.4).
func (r *ratingRepository) ListIndividual(ctx context.Context, f domain.RatingFilter) ([]domain.RatingEntry, int64, error) {
	role := string(domain.RoleStudent)
	if f.Type == domain.RatingEmployee {
		role = string(domain.RoleEmployee)
	}

	// Umumiy son (meta.total) — JOIN'siz yengil so'rov.
	var total int64
	countQ := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("role = ? AND deleted_at IS NULL AND is_active = TRUE", role)
	if f.FacultyID != nil {
		countQ = countQ.Where("faculty_id = ?", f.FacultyID)
	}
	if f.GroupID != nil {
		countQ = countQ.Where("group_id = ?", f.GroupID)
	}
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("rating: count: %w", err)
	}

	// Filtr shartlari — faqat parametr bilan (SQL injection yo'q).
	where := "u.role = ? AND u.deleted_at IS NULL AND u.is_active = TRUE"
	args := []any{role}
	if f.FacultyID != nil {
		where += " AND u.faculty_id = ?"
		args = append(args, f.FacultyID)
	}
	if f.GroupID != nil {
		where += " AND u.group_id = ?"
		args = append(args, f.GroupID)
	}
	args = append(args, f.Limit, (f.Page-1)*f.Limit)

	var rows []domain.RatingEntry
	err := r.db.WithContext(ctx).Raw(`
		WITH agg AS (
			SELECT u.id, u.full_name AS name, u.avatar_url,
			       COALESCE(s.name, '') AS faculty_name,
			       COALESCE(g.name, '') AS group_name,
			       COALESCE(SUM(a.steps), 0)      AS total_steps,
			       COALESCE(SUM(a.distance_m), 0) AS total_distance_m,
			       COALESCE(SUM(a.calories), 0)   AS total_calories,
			       COUNT(a.id) FILTER (WHERE a.steps > 0) AS active_days
			FROM users u
			LEFT JOIN activities a ON a.user_id = u.id AND a.deleted_at IS NULL`+periodCond(f.Period)+`
			LEFT JOIN structures s ON s.id = u.faculty_id
			LEFT JOIN groups g     ON g.id = u.group_id
			WHERE `+where+`
			GROUP BY u.id, u.full_name, u.avatar_url, s.name, g.name
		)
		SELECT *, RANK() OVER (ORDER BY total_steps DESC, id) AS rank
		FROM agg
		ORDER BY rank
		LIMIT ? OFFSET ?
	`, args...).Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("rating: individual: %w", err)
	}
	return rows, total, nil
}

// ListGroups — guruhlar reytingi: jon boshiga o'rtacha qadam bo'yicha
// (katta guruh adolatsiz ustunlik olmasligi uchun).
func (r *ratingRepository) ListGroups(ctx context.Context, f domain.RatingFilter) ([]domain.RatingEntry, int64, error) {
	var total int64
	countQ := r.db.WithContext(ctx).Model(&domain.Group{}).
		Where("deleted_at IS NULL AND active = TRUE")
	if f.FacultyID != nil {
		countQ = countQ.Where("faculty_id = ?", f.FacultyID)
	}
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("rating: groups count: %w", err)
	}

	where := "g.deleted_at IS NULL AND g.active = TRUE"
	args := []any{}
	if f.FacultyID != nil {
		where += " AND g.faculty_id = ?"
		args = append(args, f.FacultyID)
	}
	args = append(args, f.Limit, (f.Page-1)*f.Limit)

	var rows []domain.RatingEntry
	err := r.db.WithContext(ctx).Raw(`
		WITH agg AS (
			SELECT g.id, g.name,
			       COALESCE(s.name, '') AS faculty_name,
			       COUNT(DISTINCT u.id) AS member_count,
			       COALESCE(SUM(a.steps), 0)      AS total_steps,
			       COALESCE(SUM(a.distance_m), 0) AS total_distance_m,
			       COALESCE(SUM(a.calories), 0)   AS total_calories
			FROM groups g
			LEFT JOIN structures s ON s.id = g.faculty_id
			LEFT JOIN users u ON u.group_id = g.id AND u.deleted_at IS NULL AND u.is_active = TRUE
			LEFT JOIN activities a ON a.user_id = u.id AND a.deleted_at IS NULL`+periodCond(f.Period)+`
			WHERE `+where+`
			GROUP BY g.id, g.name, s.name
		)
		SELECT *,
		       total_steps::float / GREATEST(member_count, 1) AS avg_steps,
		       RANK() OVER (ORDER BY total_steps::float / GREATEST(member_count, 1) DESC, id) AS rank
		FROM agg
		ORDER BY rank
		LIMIT ? OFFSET ?
	`, args...).Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("rating: groups: %w", err)
	}
	return rows, total, nil
}

// ListFaculties — fakultetlar reytingi: jon boshiga o'rtacha qadam bo'yicha.
// Fakultet = users.faculty_id ko'rsatgan structures yozuvi.
func (r *ratingRepository) ListFaculties(ctx context.Context, f domain.RatingFilter) ([]domain.RatingEntry, int64, error) {
	var rows []domain.RatingEntry
	err := r.db.WithContext(ctx).Raw(`
		WITH agg AS (
			SELECT s.id, s.name,
			       COUNT(DISTINCT u.id) AS member_count,
			       COALESCE(SUM(a.steps), 0)      AS total_steps,
			       COALESCE(SUM(a.distance_m), 0) AS total_distance_m,
			       COALESCE(SUM(a.calories), 0)   AS total_calories
			FROM structures s
			JOIN users u ON u.faculty_id = s.id AND u.deleted_at IS NULL AND u.is_active = TRUE
			LEFT JOIN activities a ON a.user_id = u.id AND a.deleted_at IS NULL`+periodCond(f.Period)+`
			WHERE s.deleted_at IS NULL
			GROUP BY s.id, s.name
		)
		SELECT *,
		       total_steps::float / GREATEST(member_count, 1) AS avg_steps,
		       RANK() OVER (ORDER BY total_steps::float / GREATEST(member_count, 1) DESC, id) AS rank
		FROM agg
		ORDER BY rank
		LIMIT ? OFFSET ?
	`, f.Limit, (f.Page-1)*f.Limit).Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("rating: faculties: %w", err)
	}
	// Fakultetlar soni kichik — total = topilgan qatorlar bo'yicha alohida son.
	var total int64
	err = r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT u.faculty_id) FROM users u
		WHERE u.faculty_id IS NOT NULL AND u.deleted_at IS NULL AND u.is_active = TRUE
	`).Scan(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("rating: faculties count: %w", err)
	}
	return rows, total, nil
}

// MyRank — foydalanuvchining umumiy va fakultet ichidagi o'rni (bitta so'rov).
// Reyting o'z roli (student/employee) ichida hisoblanadi.
func (r *ratingRepository) MyRank(ctx context.Context, userID uuid.UUID, period domain.RatingPeriod) (*domain.MyRating, error) {
	var out domain.MyRating
	err := r.db.WithContext(ctx).Raw(`
		WITH agg AS (
			SELECT u.id, u.faculty_id,
			       COALESCE(SUM(a.steps), 0) AS total_steps
			FROM users u
			LEFT JOIN activities a ON a.user_id = u.id AND a.deleted_at IS NULL`+periodCond(period)+`
			WHERE u.deleted_at IS NULL AND u.is_active = TRUE
			  AND u.role = (SELECT role FROM users WHERE id = ?)
			GROUP BY u.id, u.faculty_id
		),
		ranked AS (
			SELECT id, total_steps,
			       RANK() OVER (ORDER BY total_steps DESC, id) AS global_rank,
			       RANK() OVER (PARTITION BY faculty_id ORDER BY total_steps DESC, id) AS faculty_rank,
			       COUNT(*) OVER () AS total_users
			FROM agg
		)
		SELECT global_rank, faculty_rank, total_users, total_steps
		FROM ranked WHERE id = ?
	`, userID, userID).Scan(&out).Error
	if err != nil {
		return nil, fmt.Errorf("rating: my rank: %w", err)
	}
	if out.TotalUsers == 0 {
		return nil, domain.ErrNotFound
	}
	return &out, nil
}
