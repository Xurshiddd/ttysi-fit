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

// userRepository — domain.UserRepository ning gorm implementatsiyasi.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository — UserRepository portini qaytaradi (DI uchun interfeys).
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// GetByHemis — hemis_id yoki hemis_login bo'yicha (HEMIS OAuth profilidan keladi).
func (r *userRepository) GetByHemis(ctx context.Context, hemisID *int64, hemisLogin string) (*domain.User, error) {
	if hemisID == nil && hemisLogin == "" {
		return nil, domain.ErrNotFound
	}

	var u domain.User
	db := r.db.WithContext(ctx)

	var err error
	switch {
	case hemisID != nil && hemisLogin != "":
		err = db.Where("hemis_id = ? OR hemis_login = ?", *hemisID, hemisLogin).First(&u).Error
	case hemisID != nil:
		err = db.Where("hemis_id = ?", *hemisID).First(&u).Error
	default:
		err = db.Where("hemis_login = ?", hemisLogin).First(&u).Error
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepository) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// profileSelect — UserProfile uchun ustunlar. NULL bo'lishi mumkin bo'lgan
// matn ustunlari COALESCE bilan bo'sh satrga tushiriladi (Go string non-nullable).
const profileSelect = `u.id, u.full_name, u.email, u.role, u.language,
	COALESCE(u.phone, '') AS phone,
	COALESCE(u.avatar_url, '') AS avatar_url,
	COALESCE(u.bio, '') AS bio,
	COALESCE(u.gender, '') AS gender,
	u.birth_date, u.course,
	COALESCE(u.position, '') AS position,
	COALESCE(u.specialty, '') AS specialty,
	COALESCE(u.hemis_login, '') AS hemis_login,
	COALESCE(fac.name, '') AS faculty_name,
	COALESCE(dep.name, '') AS department_name,
	COALESCE(grp.name, '') AS group_name`

// GetProfile — o'z profili: fakultet/kafedra/guruh nomlari bitta JOIN so'rovda.
func (r *userRepository) GetProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	var rows []domain.UserProfile
	err := r.db.WithContext(ctx).
		Table("users u").
		Joins("LEFT JOIN structures fac ON fac.id = u.faculty_id").
		Joins("LEFT JOIN structures dep ON dep.id = u.department_id").
		Joins("LEFT JOIN groups grp ON grp.id = u.group_id").
		Where("u.id = ? AND u.deleted_at IS NULL", id).
		Select(profileSelect).
		Limit(1).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrNotFound
	}
	return &rows[0], nil
}

// UpdateProfile — faqat foydalanuvchi o'zgartira oladigan ustunlar yangilanadi.
// Ustun nomlari kodda qat'iy (map kaliti sifatida) — mijoz kiritmasi ustun
// nomiga aylanmaydi (§3.2, §17.3 #13).
func (r *userRepository) UpdateProfile(ctx context.Context, id uuid.UUID, p domain.ProfileUpdate) (*domain.UserProfile, error) {
	fields := map[string]any{}
	if p.Phone != nil {
		fields["phone"] = *p.Phone
	}
	if p.Bio != nil {
		fields["bio"] = *p.Bio
	}
	if p.Language != nil {
		fields["language"] = *p.Language
	}

	// Bo'sh so'rov — yozmaymiz, shunchaki joriy profilni qaytaramiz.
	if len(fields) > 0 {
		fields["updated_at"] = time.Now()
		res := r.db.WithContext(ctx).
			Model(&domain.User{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(fields)
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			return nil, domain.ErrNotFound
		}
	}
	return r.GetProfile(ctx, id)
}

// List — admin uchun paginatsiyali ro'yxat. Fakultet/guruh nomi JOIN bilan (N+1 yo'q).
func (r *userRepository) List(ctx context.Context, f domain.UserFilter) ([]domain.UserListItem, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	q := r.db.WithContext(ctx).
		Table("users u").
		Joins("LEFT JOIN structures fac ON fac.id = u.faculty_id").
		Joins("LEFT JOIN groups grp ON grp.id = u.group_id").
		Where("u.deleted_at IS NULL")

	if f.Role != "" {
		q = q.Where("u.role = ?", f.Role)
	}
	if f.FacultyID != nil {
		q = q.Where("u.faculty_id = ?", *f.FacultyID)
	}
	if f.GroupID != nil {
		q = q.Where("u.group_id = ?", *f.GroupID)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("u.full_name ILIKE ? OR u.email ILIKE ? OR u.hemis_login ILIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.UserListItem
	err := q.
		Select(`u.id, u.full_name, u.email, u.role, u.gender, u.course,
			u.is_active, u.avatar_url, u.hemis_login, u.position,
			COALESCE(fac.name, '') AS faculty_name,
			COALESCE(grp.name, '') AS group_name`).
		Order("u.full_name ASC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Scan(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// UpsertRoster — HEMIS roster ni hemis_id bo'yicha ko'tarib-yangilaydi.
// Faqat HEMIS dan keladigan maydonlar yangilanadi (parol, til kabilar tegilmaydi).
func (r *userRepository) UpsertRoster(ctx context.Context, users []domain.User) (*domain.SyncStats, error) {
	if len(users) == 0 {
		return &domain.SyncStats{}, nil
	}

	// Dedup: HEMIS employee-list bir xodimni bir nechta lavozim bilan
	// takror qaytarishi mumkin — bitta hemis_id faqat bir marta (birinchisi qoladi).
	seen := make(map[int64]struct{}, len(users))
	deduped := make([]domain.User, 0, len(users))
	for _, u := range users {
		if u.HemisID != nil {
			if _, ok := seen[*u.HemisID]; ok {
				continue
			}
			seen[*u.HemisID] = struct{}{}
		}
		deduped = append(deduped, u)
	}
	users = deduped

	hemisIDs := make([]int64, 0, len(users))
	for i := range users {
		if users[i].HemisID != nil {
			hemisIDs = append(hemisIDs, *users[i].HemisID)
		}
	}
	var existing int64
	if err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("hemis_id IN ?", hemisIDs).
		Count(&existing).Error; err != nil {
		return nil, err
	}

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "hemis_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"full_name", "hemis_login", "role", "email",
				"faculty_hemis_id", "department_hemis_id", "group_hemis_id",
				"course", "gender", "birth_date", "position", "specialty",
				"avatar_url", "is_active", "updated_at",
			}),
		}).
		CreateInBatches(&users, 500).Error
	if err != nil {
		return nil, err
	}

	return &domain.SyncStats{
		Total:   len(users),
		Created: len(users) - int(existing),
		Updated: int(existing),
	}, nil
}

// RelinkRoster — HEMIS id larini FK larga bog'laydi (uchta bitta-bitta SQL).
func (r *userRepository) RelinkRoster(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Exec(`
		UPDATE users u SET faculty_id = s.id
		FROM structures s
		WHERE u.faculty_hemis_id IS NOT NULL
		  AND s.hemis_id = u.faculty_hemis_id
		  AND (u.faculty_id IS DISTINCT FROM s.id)
	`).Error; err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Exec(`
		UPDATE users u SET department_id = s.id
		FROM structures s
		WHERE u.department_hemis_id IS NOT NULL
		  AND s.hemis_id = u.department_hemis_id
		  AND (u.department_id IS DISTINCT FROM s.id)
	`).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Exec(`
		UPDATE users u SET group_id = g.id
		FROM groups g
		WHERE u.group_hemis_id IS NOT NULL
		  AND g.hemis_id = u.group_hemis_id
		  AND (u.group_id IS DISTINCT FROM g.id)
	`).Error
}
