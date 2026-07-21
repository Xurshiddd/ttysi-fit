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

type competitionRepository struct {
	db *gorm.DB
}

func NewCompetitionRepository(db *gorm.DB) domain.CompetitionRepository {
	return &competitionRepository{db: db}
}

func (r *competitionRepository) Create(ctx context.Context, c *domain.Competition) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *competitionRepository) Update(ctx context.Context, c *domain.Competition) error {
	res := r.db.WithContext(ctx).
		Model(&domain.Competition{}).
		Where("id = ? AND deleted_at IS NULL", c.ID).
		Select("type", "title", "description", "scope", "status", "starts_at", "ends_at",
			"reg_ends_at", "location", "max_participants", "reward_coins",
			"config", "cover_url", "updated_at").
		Updates(c)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *competitionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&domain.Competition{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]any{"deleted_at": now, "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *competitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Competition, error) {
	var c domain.Competition
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *competitionRepository) List(ctx context.Context, f domain.CompetitionFilter) ([]domain.Competition, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.Competition{}).Where("deleted_at IS NULL")
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.Competition
	err := q.Order("COALESCE(starts_at, created_at) DESC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ListForUser — musobaqalar + foydalanuvchi holati + ishtirokchilar soni.
//
// Uchta so'rov, ro'yxat uzunligidan qat'i nazar (§3.1 — `for` ichida so'rov yo'q):
// (1) sahifa, (2) shu id lar bo'yicha foydalanuvchi yozuvlari, (3) GROUP BY sanoq.
func (r *competitionRepository) ListForUser(ctx context.Context, userID uuid.UUID, f domain.CompetitionFilter) ([]domain.CompetitionView, int64, error) {
	items, total, err := r.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		return []domain.CompetitionView{}, total, nil
	}

	ids := make([]uuid.UUID, len(items))
	for i := range items {
		ids[i] = items[i].ID
	}

	var regs []domain.CompetitionRegistration
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND competition_id IN ?", userID, ids).
		Find(&regs).Error; err != nil {
		return nil, 0, err
	}
	mine := make(map[uuid.UUID]domain.CompetitionRegistration, len(regs))
	for _, g := range regs {
		mine[g.CompetitionID] = g
	}

	type countRow struct {
		CompetitionID uuid.UUID
		Cnt           int
	}
	var counts []countRow
	if err := r.db.WithContext(ctx).
		Table("competition_registrations").
		Select("competition_id, COUNT(*) AS cnt").
		Where("competition_id IN ? AND status = ?", ids, domain.RegStatusRegistered).
		Group("competition_id").
		Scan(&counts).Error; err != nil {
		return nil, 0, err
	}
	countBy := make(map[uuid.UUID]int, len(counts))
	for _, c := range counts {
		countBy[c.CompetitionID] = c.Cnt
	}

	now := time.Now()
	out := make([]domain.CompetitionView, 0, len(items))
	for _, c := range items {
		v := domain.CompetitionView{Competition: c}
		v.ParticipantCount = countBy[c.ID]

		if g, ok := mine[c.ID]; ok && g.Status == domain.RegStatusRegistered {
			v.Registered = true
			v.Place = g.Place
			v.RewardGranted = g.RewardGranted
		}
		v.RegOpen = c.RegOpen(v.ParticipantCount, now)
		out = append(out, v)
	}
	return out, total, nil
}

// Register — ro'yxatdan o'tish.
//
// Joy soni tekshiruvi POYGAGA ochiq: ikki kishi bir vaqtda oxirgi joyga
// yozilsa, ikkalasi ham "joy bor" deb ko'rishi mumkin. Shuning uchun musobaqa
// qatorini FOR UPDATE bilan bloklaymiz — shu musobaqa bo'yicha yozilishlar
// ketma-ket bajariladi (CLAUDE.md §13.2).
func (r *competitionRepository) Register(ctx context.Context, userID, competitionID uuid.UUID) (*domain.CompetitionRegistration, error) {
	var out *domain.CompetitionRegistration

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var c domain.Competition
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", competitionID).
			First(&c).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}

		if c.Status != domain.CompStatusRegistration {
			return fmt.Errorf("%w: ro'yxatdan o'tish yopiq", domain.ErrValidation)
		}
		if c.RegEndsAt != nil && c.RegEndsAt.Before(time.Now()) {
			return fmt.Errorf("%w: ro'yxatdan o'tish muddati tugagan", domain.ErrValidation)
		}

		// Mavjud yozuv bormi (bekor qilingan bo'lsa qayta tiklaymiz).
		var existing domain.CompetitionRegistration
		err = tx.Where("user_id = ? AND competition_id = ?", userID, competitionID).
			First(&existing).Error
		switch {
		case err == nil:
			if existing.Status == domain.RegStatusRegistered {
				return domain.ErrAlreadyExists
			}
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}

		// Joy sonini blok ostida sanaymiz.
		if c.MaxParticipants != nil && *c.MaxParticipants > 0 {
			var cnt int64
			if err := tx.Model(&domain.CompetitionRegistration{}).
				Where("competition_id = ? AND status = ?", competitionID, domain.RegStatusRegistered).
				Count(&cnt).Error; err != nil {
				return err
			}
			if cnt >= int64(*c.MaxParticipants) {
				return fmt.Errorf("%w: joylar to'lgan", domain.ErrValidation)
			}
		}

		now := time.Now()
		if existing.ID != uuid.Nil {
			// Bekor qilingan yozuvni tiklaymiz — yangi qator qo'shmaymiz
			// (unique indeks user_id+competition_id).
			if err := tx.Model(&domain.CompetitionRegistration{}).
				Where("id = ?", existing.ID).
				Updates(map[string]any{
					"status":        domain.RegStatusRegistered,
					"registered_at": now,
					"updated_at":    now,
				}).Error; err != nil {
				return err
			}
			existing.Status = domain.RegStatusRegistered
			existing.RegisteredAt = now
			out = &existing
			return nil
		}

		reg := &domain.CompetitionRegistration{
			UserID:        userID,
			CompetitionID: competitionID,
			Status:        domain.RegStatusRegistered,
			RegisteredAt:  now,
		}
		if err := tx.Create(reg).Error; err != nil {
			if isUniqueViolation(err) {
				return domain.ErrAlreadyExists
			}
			return err
		}
		out = reg
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Cancel — ishtirokni bekor qiladi. Yozuv o'chirilmaydi: kim qachon
// yozilgani/bekor qilgani tarixi qoladi.
func (r *competitionRepository) Cancel(ctx context.Context, userID, competitionID uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&domain.CompetitionRegistration{}).
		Where("user_id = ? AND competition_id = ? AND status = ?",
			userID, competitionID, domain.RegStatusRegistered).
		Updates(map[string]any{"status": domain.RegStatusCancelled, "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *competitionRepository) Participants(ctx context.Context, competitionID uuid.UUID, page, limit int) ([]domain.CompetitionRegistration, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	q := r.db.WithContext(ctx).
		Model(&domain.CompetitionRegistration{}).
		Where("competition_id = ? AND status = ?", competitionID, domain.RegStatusRegistered)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.CompetitionRegistration
	err := q.Order("COALESCE(place, 32767), registered_at").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
