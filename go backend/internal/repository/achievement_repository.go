package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type achievementRepository struct {
	db *gorm.DB
}

func NewAchievementRepository(db *gorm.DB) domain.AchievementRepository {
	return &achievementRepository{db: db}
}

func (r *achievementRepository) Create(ctx context.Context, a *domain.Achievement) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *achievementRepository) Update(ctx context.Context, a *domain.Achievement) error {
	// Select bilan aniq ro'yxat: mijoz yubormagan maydon nolga tushib qolmasin
	// va created_at/deleted_at tegilmasin.
	res := r.db.WithContext(ctx).
		Model(&domain.Achievement{}).
		Where("id = ? AND deleted_at IS NULL", a.ID).
		Select("type", "title", "description", "award_mode", "status",
			"reward_coins", "criteria", "icon_url", "cover_url",
			"certificate_template", "updated_at").
		Updates(a)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *achievementRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).
		Model(&domain.Achievement{}).
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

func (r *achievementRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Achievement, error) {
	var a domain.Achievement
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *achievementRepository) List(ctx context.Context, f domain.AchievementFilter) ([]domain.Achievement, int64, error) {
	q := r.listQuery(f)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, limit := normalizePage(f.Page, f.Limit)
	var items []domain.Achievement
	err := q.Order("created_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *achievementRepository) listQuery(f domain.AchievementFilter) *gorm.DB {
	q := r.db.Model(&domain.Achievement{}).Where("deleted_at IS NULL")
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.AwardMode != "" {
		q = q.Where("award_mode = ?", f.AwardMode)
	}
	return q
}

// normalizePage — sahifa/limit chegaralari (§14.2: default 20, max 100).
func normalizePage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}

// userMetrics — foydalanuvchining barcha yutuq turlari uchun kerak bo'ladigan
// yig'ma ko'rsatkichlari. Hammasi BITTA so'rovda olinadi: yutuqlar soni
// qancha bo'lsa ham qo'shimcha so'rov ketmaydi (§3.1 — for ichida DB yo'q).
type userMetrics struct {
	Steps          float64
	DistanceM      float64
	ActiveDays     float64
	ChallengesDone float64
}

func (r *achievementRepository) metrics(ctx context.Context, userID uuid.UUID) (userMetrics, error) {
	var m userMetrics
	// Har bir ko'rsatkich mustaqil subquery: JOIN qilinsa satrlar ko'payib
	// SUM ikki barobar chiqib ketardi (fan-out).
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE((SELECT SUM(steps) FROM activities
			          WHERE user_id = ? AND deleted_at IS NULL), 0)      AS steps,
			COALESCE((SELECT SUM(distance_m) FROM activities
			          WHERE user_id = ? AND deleted_at IS NULL), 0)      AS distance_m,
			COALESCE((SELECT COUNT(*) FROM activities
			          WHERE user_id = ? AND deleted_at IS NULL
			            AND steps > 0), 0)                               AS active_days,
			COALESCE((SELECT COUNT(*) FROM user_challenges
			          WHERE user_id = ? AND completed_at IS NOT NULL), 0) AS challenges_done
	`, userID, userID, userID, userID).Scan(&m).Error
	return m, err
}

// progressFor — turga qarab mos ko'rsatkichni tanlaydi.
func progressFor(t domain.AchievementType, m userMetrics) float64 {
	spec, ok := domain.AchievementSpec(t)
	if !ok {
		return 0
	}
	switch spec.Source {
	case domain.SourceActivitySum:
		switch domain.AchievementMetricColumn(t) {
		case "steps":
			return m.Steps
		case "distance_m":
			return m.DistanceM
		}
	case domain.SourceActiveDays:
		return m.ActiveDays
	case domain.SourceChallengeDone:
		return m.ChallengesDone
	}
	return 0
}

// ListForUser — yutuqlar + shu foydalanuvchi holati va progressi.
//
// Uch so'rov, ro'yxat uzunligidan qat'i nazar: (1) yutuqlar sahifasi,
// (2) foydalanuvchi ko'rsatkichlari, (3) qozonilgan yozuvlar `IN` bilan.
func (r *achievementRepository) ListForUser(ctx context.Context, userID uuid.UUID, f domain.AchievementFilter) ([]domain.AchievementView, int64, error) {
	items, total, err := r.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	if len(items) == 0 {
		return []domain.AchievementView{}, total, nil
	}

	m, err := r.metrics(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	earned, err := r.earnedByAchievement(ctx, userID, ids(items))
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.AchievementView, 0, len(items))
	for _, a := range items {
		out = append(out, buildView(a, m, earned))
	}
	return out, total, nil
}

// ListEarned — faqat qozonilgan yutuqlar (profil "Yutuqlarim"/sertifikatlar).
func (r *achievementRepository) ListEarned(ctx context.Context, userID uuid.UUID, f domain.AchievementFilter) ([]domain.AchievementView, int64, error) {
	page, limit := normalizePage(f.Page, f.Limit)

	base := r.db.WithContext(ctx).
		Table("user_achievements ua").
		Joins("JOIN achievements a ON a.id = ua.achievement_id AND a.deleted_at IS NULL").
		Where("ua.user_id = ?", userID)
	if f.Type != "" {
		base = base.Where("a.type = ?", f.Type)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// AwardID alohida nom bilan tanlanadi: `a.*` ichidagi `id` yutuq shabloniniki,
	// sertifikat havolasi uchun esa berilgan yozuv (ua.id) kerak.
	var rows []struct {
		domain.Achievement
		AwardID        uuid.UUID
		AwardedAt      time.Time
		ProgressValue  float64
		CertificateURL string
	}
	err := base.
		Select("a.*, ua.id AS award_id, ua.awarded_at, ua.progress_value, ua.certificate_url").
		Order("ua.awarded_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]domain.AchievementView, 0, len(rows))
	for _, row := range rows {
		at, awardID := row.AwardedAt, row.AwardID
		out = append(out, domain.AchievementView{
			Achievement:    row.Achievement,
			Earned:         true,
			EarnedAt:       &at,
			AwardID:        &awardID,
			Progress:       row.ProgressValue,
			Target:         domain.AchievementThreshold(row.Type, row.Criteria),
			ProgressPct:    100,
			CertificateURL: row.CertificateURL,
		})
	}
	return out, total, nil
}

func ids(items []domain.Achievement) []uuid.UUID {
	out := make([]uuid.UUID, len(items))
	for i := range items {
		out[i] = items[i].ID
	}
	return out
}

func (r *achievementRepository) earnedByAchievement(ctx context.Context, userID uuid.UUID, achIDs []uuid.UUID) (map[uuid.UUID]domain.UserAchievement, error) {
	var uas []domain.UserAchievement
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND achievement_id IN ?", userID, achIDs).
		Find(&uas).Error; err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]domain.UserAchievement, len(uas))
	for _, ua := range uas {
		out[ua.AchievementID] = ua
	}
	return out, nil
}

func buildView(a domain.Achievement, m userMetrics, earned map[uuid.UUID]domain.UserAchievement) domain.AchievementView {
	v := domain.AchievementView{Achievement: a}
	v.Target = domain.AchievementThreshold(a.Type, a.Criteria)

	if ua, ok := earned[a.ID]; ok {
		at := ua.AwardedAt
		awardID := ua.ID
		v.Earned = true
		v.EarnedAt = &at
		v.AwardID = &awardID
		// Qozonilgach snapshot ko'rsatiladi: keyin faollik o'zgarsa ham
		// sertifikatdagi raqam o'zgarmasligi kerak.
		v.Progress = ua.ProgressValue
		v.ProgressPct = 100
		v.CertificateURL = ua.CertificateURL
		return v
	}

	v.Progress = progressFor(a.Type, m)
	v.ProgressPct = domain.AchievementProgressPct(v.Progress, v.Target)
	return v
}

// Award — yutuqni beradi. Unique indeks tufayli idempotent: takror berishda
// yangi yozuv qo'shilmaydi va ErrAlreadyExists qaytadi.
func (r *achievementRepository) Award(ctx context.Context, userID, achievementID uuid.UUID, awardedBy *uuid.UUID, value float64, note string) (*domain.UserAchievement, error) {
	ua := domain.UserAchievement{
		UserID:        userID,
		AchievementID: achievementID,
		AwardedAt:     time.Now(),
		AwardedBy:     awardedBy,
		ProgressValue: value,
		Note:          note,
	}

	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "achievement_id"}},
			DoNothing: true,
		}).
		Create(&ua)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrAlreadyExists
	}
	return &ua, nil
}

// GetAward — berilgan yutuq tafsiloti: yozuv + shablon + foydalanuvchi ismi.
// Bitta JOIN — sertifikat uchun uchta alohida so'rov kerak emas.
func (r *achievementRepository) GetAward(ctx context.Context, userAchievementID uuid.UUID) (*domain.AwardDetail, error) {
	var d domain.AwardDetail
	err := r.db.WithContext(ctx).
		Table("user_achievements ua").
		Select(`ua.id, ua.user_id, ua.progress_value, ua.note, ua.awarded_at,
		        u.full_name AS user_full_name,
		        a.type, a.title, a.description, a.criteria`).
		Joins("JOIN achievements a ON a.id = ua.achievement_id AND a.deleted_at IS NULL").
		Joins("JOIN users u ON u.id = ua.user_id AND u.deleted_at IS NULL").
		Where("ua.id = ?", userAchievementID).
		Scan(&d).Error
	if err != nil {
		return nil, err
	}
	if d.ID == uuid.Nil {
		return nil, domain.ErrNotFound
	}
	return &d, nil
}

// EvaluateAuto — foydalanuvchi uchun aktiv avtomatik yutuqlarni baholaydi.
//
// Faollik sinxronlangach chaqiriladi. Ko'rsatkichlar bir marta olinadi,
// so'ng har bir yutuq xotirada solishtiriladi — `for` ichida DB so'rovi
// faqat haqiqatan beriladigan yutuq uchun bo'ladi (§3.1).
func (r *achievementRepository) EvaluateAuto(ctx context.Context, userID uuid.UUID) ([]domain.UserAchievement, error) {
	var items []domain.Achievement
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND status = ? AND award_mode = ?",
			domain.AchStatusActive, domain.AwardModeAuto).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}

	m, err := r.metrics(ctx, userID)
	if err != nil {
		return nil, err
	}

	earned, err := r.earnedByAchievement(ctx, userID, ids(items))
	if err != nil {
		return nil, err
	}

	var granted []domain.UserAchievement
	for _, a := range items {
		if _, ok := earned[a.ID]; ok {
			continue // allaqachon berilgan
		}
		target := domain.AchievementThreshold(a.Type, a.Criteria)
		if target <= 0 {
			continue // mezonsiz — avtomatik berilmaydi
		}
		progress := progressFor(a.Type, m)
		if progress < target {
			continue
		}

		ua, err := r.Award(ctx, userID, a.ID, nil, progress, "")
		if errors.Is(err, domain.ErrAlreadyExists) {
			continue // parallel baholash ulgurgan — normal holat
		}
		if err != nil {
			return nil, err
		}
		granted = append(granted, *ua)
	}
	return granted, nil
}
