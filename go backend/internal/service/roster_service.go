package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/pkg/hemis"
	"go.uber.org/zap"
)

// AvatarStore — tashqi rasmni yuklab olib lokal saqlovga joylaydi va
// uning ommaviy URL'ini qaytaradi. srcURL bo'sh bo'lsa ("", nil) qaytaradi.
type AvatarStore interface {
	Save(ctx context.Context, srcURL, name string) (string, error)
}

// RosterService — guruh va talaba/o'qituvchi sync use-case qatlami.
type RosterService struct {
	groups  domain.GroupRepository
	users   domain.UserRepository
	hemis   *hemis.Client
	avatars AvatarStore // nil bo'lsa — rasm yuklab olinmaydi (HEMIS URL saqlanadi)
	log     *zap.Logger
	workers int
}

// NewRosterService — roster servisini yaratadi.
// avatars nil bo'lsa, avatar yuklab olish o'chiriladi va HEMIS URL'i saqlanadi.
func NewRosterService(
	groups domain.GroupRepository,
	users domain.UserRepository,
	client *hemis.Client,
	avatars AvatarStore,
	log *zap.Logger,
	workers int,
) *RosterService {
	if workers <= 0 {
		workers = 8
	}
	return &RosterService{
		groups:  groups,
		users:   users,
		hemis:   client,
		avatars: avatars,
		log:     log,
		workers: workers,
	}
}

// SyncGroups — HEMIS guruhlarini DB ga ko'tarib-yangilaydi va fakultetga bog'laydi.
func (s *RosterService) SyncGroups(ctx context.Context) (*domain.SyncStats, error) {
	dtos, err := s.hemis.FetchGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("SyncGroups: fetch: %w", err)
	}

	now := time.Now()
	items := make([]domain.Group, 0, len(dtos))
	for _, d := range dtos {
		items = append(items, domain.Group{
			HemisID:        d.HemisID,
			Name:           d.Name,
			FacultyHemisID: d.FacultyHemisID,
			SpecialtyCode:  d.SpecialtyCode,
			SpecialtyName:  d.SpecialtyName,
			EducationLang:  d.EducationLang,
			Active:         d.Active,
			SyncedAt:       &now,
		})
	}

	stats, err := s.groups.UpsertBatch(ctx, items)
	if err != nil {
		return nil, fmt.Errorf("SyncGroups: upsert: %w", err)
	}
	if err := s.groups.RelinkFaculties(ctx); err != nil {
		return nil, fmt.Errorf("SyncGroups: relink: %w", err)
	}
	return stats, nil
}

// SyncStudents — HEMIS talabalarini users ga sync qiladi.
func (s *RosterService) SyncStudents(ctx context.Context) (*domain.SyncStats, error) {
	dtos, err := s.hemis.FetchStudents(ctx)
	if err != nil {
		return nil, fmt.Errorf("SyncStudents: fetch: %w", err)
	}
	return s.upsertRoster(ctx, dtos, "SyncStudents")
}

// SyncEmployees — HEMIS xodim/o'qituvchilarini users ga sync qiladi.
func (s *RosterService) SyncEmployees(ctx context.Context) (*domain.SyncStats, error) {
	dtos, err := s.hemis.FetchEmployees(ctx)
	if err != nil {
		return nil, fmt.Errorf("SyncEmployees: fetch: %w", err)
	}
	return s.upsertRoster(ctx, dedupEmployees(dtos), "SyncEmployees")
}

// dedupEmployees — bir xodim bir nechta lavozimda bo'lsa, "Asosiy ish joy"
// (employmentForm = "11") yozuvini afzal ko'rib, hemis_id bo'yicha bittaga qoldiradi.
func dedupEmployees(dtos []hemis.RosterDTO) []hemis.RosterDTO {
	pos := make(map[int64]int, len(dtos))
	out := make([]hemis.RosterDTO, 0, len(dtos))
	for _, d := range dtos {
		if i, ok := pos[d.HemisID]; ok {
			// Mavjud yozuv asosiy emas, joriysi asosiy bo'lsa — almashtiramiz.
			if d.EmploymentFormCode == "11" && out[i].EmploymentFormCode != "11" {
				out[i] = d
			}
			continue
		}
		pos[d.HemisID] = len(out)
		out = append(out, d)
	}
	return out
}

func (s *RosterService) upsertRoster(ctx context.Context, dtos []hemis.RosterDTO, op string) (*domain.SyncStats, error) {
	// Avatarlarni lokal saqlovga yuklab olamiz (sozlangan bo'lsa).
	// Har bir DTO.AvatarURL HEMIS URL'idan lokal URL'ga almashtiriladi.
	s.downloadAvatars(ctx, dtos)

	users := make([]domain.User, 0, len(dtos))
	for _, d := range dtos {
		hid := d.HemisID
		users = append(users, domain.User{
			HemisID:           &hid,
			HemisLogin:        d.HemisLogin,
			FullName:          d.FullName,
			Email:             d.Email,
			Role:              domain.Role(d.Role),
			FacultyHemisID:    d.FacultyHemisID,
			DepartmentHemisID: d.DepartmentHemisID,
			GroupHemisID:      d.GroupHemisID,
			Course:            d.Course,
			Gender:            d.Gender,
			BirthDate:         d.BirthDate,
			Position:          d.Position,
			Specialty:         d.Specialty,
			AvatarURL:         d.AvatarURL,
			IsActive:          d.IsActive,
		})
	}

	stats, err := s.users.UpsertRoster(ctx, users)
	if err != nil {
		return nil, fmt.Errorf("%s: upsert: %w", op, err)
	}
	if err := s.users.RelinkRoster(ctx); err != nil {
		return nil, fmt.Errorf("%s: relink: %w", op, err)
	}
	return stats, nil
}

// downloadAvatars — har bir DTO ning HEMIS rasm URL'ini yuklab olib, lokal
// saqlovdagi URL'ga almashtiradi. Parallel ishlaydi (workers chegarasi bilan).
//
// Muhim xususiyatlar:
//   - avatars sozlanmagan bo'lsa — hech narsa qilmaydi (HEMIS URL qoladi).
//   - Bitta rasmni yuklash xato bersa — sync TO'XTAMAYDI; o'sha foydalanuvchining
//     HEMIS URL'i o'zgarmay qoladi (zaxira sifatida), xato faqat loglanadi.
//   - Har bir goroutine panic'dan himoyalangan (CLAUDE.md §11.1).
//   - Har bir goroutine dtos ning FARQLI indeksiga yozadi — race yo'q.
func (s *RosterService) downloadAvatars(ctx context.Context, dtos []hemis.RosterDTO) {
	if s.avatars == nil {
		return
	}

	sem := make(chan struct{}, s.workers)
	var wg sync.WaitGroup
	var okCount, failCount atomic.Int64

	for i := range dtos {
		if dtos[i].AvatarURL == "" {
			continue
		}
		// Context bekor qilingan bo'lsa — to'xtaymiz.
		select {
		case <-ctx.Done():
			wg.Wait()
			return
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			defer func() {
				if r := recover(); r != nil && s.log != nil {
					s.log.Error("avatar download panic", zap.Any("recover", r))
				}
			}()

			src := dtos[idx].AvatarURL
			name := fmt.Sprintf("%d", dtos[idx].HemisID)

			localURL, err := s.avatars.Save(ctx, src, name)
			if err != nil {
				failCount.Add(1)
				if s.log != nil {
					s.log.Warn("avatar yuklab olinmadi (HEMIS URL qoldi)",
						zap.Int64("hemis_id", dtos[idx].HemisID),
						zap.String("src", src),
						zap.Error(err),
					)
				}
				return // HEMIS URL zaxira sifatida qoladi
			}
			if localURL != "" {
				dtos[idx].AvatarURL = localURL
				okCount.Add(1)
			}
		}(i)
	}

	wg.Wait()

	if s.log != nil {
		s.log.Info("avatarlar yuklab olindi",
			zap.Int64("muvaffaqiyat", okCount.Load()),
			zap.Int64("xato", failCount.Load()),
		)
	}
}

// ListGroups — guruhlarni qaytaradi; facultyID berilsa o'sha fakultetnikini.
func (s *RosterService) ListGroups(ctx context.Context, facultyID *uuid.UUID) ([]domain.Group, error) {
	return s.groups.ListByFaculty(ctx, facultyID)
}
