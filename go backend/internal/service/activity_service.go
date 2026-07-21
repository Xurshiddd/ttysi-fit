package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/internal/dto"
	"go.uber.org/zap"
)

// maxBatchDays — bitta batch so'rovida ruxsat etilgan maksimal kun soni.
// Mijoz odatda oxirgi 7 kunni yuboradi; 31 zaxira bilan (oy) cheklangan.
const maxBatchDays = 31

// maxDeleteRange — admin bir martada o'chira oladigan maksimal oraliq.
// Xato bosilgan tugma bilan butun yillik tarix yo'qolib ketmasin.
const maxDeleteRange = 92 * 24 * time.Hour

// AchievementEvaluator — faollik yozilgach avtomatik yutuqlarni baholaydi.
//
// Interfeys sifatida e'lon qilingan: ActivityService AchievementService ga
// to'g'ridan-to'g'ri bog'lanmasin (CLAUDE.md §1 — DI interfeys orqali).
type AchievementEvaluator interface {
	Evaluate(ctx context.Context, userID uuid.UUID) ([]domain.UserAchievement, error)
}

// ActivityService — kunlik faollik use-case qatlami.
type ActivityService struct {
	repo      domain.ActivityRepository
	evaluator AchievementEvaluator
	loc       *time.Location
	log       *zap.Logger
}

func NewActivityService(repo domain.ActivityRepository, evaluator AchievementEvaluator, loc *time.Location, log *zap.Logger) *ActivityService {
	if loc == nil {
		loc = time.UTC
	}
	return &ActivityService{repo: repo, evaluator: evaluator, loc: loc, log: log}
}

// Record — kunlik faollikni yozadi/yangilaydi (bir kun bir yozuv).
func (s *ActivityService) Record(ctx context.Context, userID uuid.UUID, req dto.RecordActivityRequest) (*domain.Activity, error) {
	a, err := s.toActivity(userID, req)
	if err != nil {
		return nil, fmt.Errorf("ActivityService.Record: %w", err)
	}
	if err := s.repo.Upsert(ctx, a); err != nil {
		return nil, fmt.Errorf("ActivityService.Record: %w", err)
	}
	s.evaluate(ctx, userID)
	return a, nil
}

// RecordBatch — bir necha kunlik faollikni BITTA so'rovda yozadi.
//
// Mijoz ilova ochilganda telefondagi oxirgi kunlarni qayta yuboradi
// ("backfill"): foydalanuvchi bir necha kun ilovani ochmasa ham o'sha
// kunlar yo'qolmaydi. Har kunni alohida POST qilish — CLAUDE.md §3.1
// bo'yicha taqiqlangan sikl ichidagi so'rov, shuning uchun bitta bulk upsert.
func (s *ActivityService) RecordBatch(ctx context.Context, userID uuid.UUID, items []dto.RecordActivityRequest) ([]domain.Activity, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("ActivityService.RecordBatch: %w", domain.ErrEmptyBatch)
	}
	if len(items) > maxBatchDays {
		return nil, fmt.Errorf("ActivityService.RecordBatch: %w", domain.ErrBatchTooLarge)
	}

	// Sana bo'yicha dedup: bitta so'rovda bir kun ikki marta kelsa
	// PostgreSQL "ON CONFLICT DO UPDATE ... cannot affect row a second time"
	// xatosini beradi. Oxirgisi (eng yangi o'qish) ustun.
	byDate := make(map[string]*domain.Activity, len(items))
	order := make([]string, 0, len(items))
	for i := range items {
		a, err := s.toActivity(userID, items[i])
		if err != nil {
			return nil, fmt.Errorf("ActivityService.RecordBatch: %w", err)
		}
		key := a.ActivityDate.Format("2006-01-02")
		if _, seen := byDate[key]; !seen {
			order = append(order, key)
		}
		byDate[key] = a
	}

	rows := make([]domain.Activity, 0, len(order))
	for _, key := range order {
		rows = append(rows, *byDate[key])
	}

	if err := s.repo.UpsertMany(ctx, rows); err != nil {
		return nil, fmt.Errorf("ActivityService.RecordBatch: %w", err)
	}
	s.evaluate(ctx, userID)
	return rows, nil
}

// toActivity — DTO ni domain yozuviga o'giradi (sana tekshiruvi bilan).
func (s *ActivityService) toActivity(userID uuid.UUID, req dto.RecordActivityRequest) (*domain.Activity, error) {
	date := s.today()
	if req.Date != "" {
		// Mahalliy mintaqada o'qiymiz: mijoz yuborgan "2026-07-21" o'sha
		// kunning mahalliy 00:00 i, UTC dagi boshqa kun emas.
		parsed, err := time.ParseInLocation("2006-01-02", req.Date, s.loc)
		if err != nil {
			return nil, fmt.Errorf("sana: %w", err)
		}
		// Kelajakdagi sana — mijoz soati noto'g'ri yoki suiiste'mol
		// (bir kunni oldindan "to'ldirib" reytingda oldinga chiqish).
		if parsed.After(s.today()) {
			return nil, domain.ErrFutureDate
		}
		date = parsed
	}

	return &domain.Activity{
		UserID:       userID,
		ActivityDate: date,
		Steps:        req.Steps,
		Calories:     req.Calories,
		DistanceM:    req.DistanceM,
		ActiveMin:    req.ActiveMin,
		Source:       req.Source,
	}, nil
}

// evaluate — avtomatik yutuqlarni baholaydi.
// Xato bo'lsa faollik yozuvi BEKOR QILINMAYDI: qadam ma'lumoti yo'qolishi
// yutuq kechikishidan ancha yomon.
func (s *ActivityService) evaluate(ctx context.Context, userID uuid.UUID) {
	if s.evaluator == nil {
		return
	}
	if _, err := s.evaluator.Evaluate(ctx, userID); err != nil && s.log != nil {
		s.log.Warn("yutuq baholash", zap.Error(err))
	}
}

// List — oxirgi N kun (yoki [from,to]) faollik yozuvlari.
func (s *ActivityService) List(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]domain.Activity, error) {
	return s.repo.ListByUser(ctx, userID, from, to, limit)
}

// Stats — bugun/hafta/oy/jami yig'ma.
func (s *ActivityService) Stats(ctx context.Context, userID uuid.UUID) (*domain.ActivityStats, error) {
	return s.repo.Stats(ctx, userID, s.today())
}

// DeleteRange — foydalanuvchining oraliqdagi faolligini o'chiradi (admin).
//
// NEGA KERAK: upsert GREATEST bilan ishlaydi — bir marta yozilgan katta
// qiymatni qayta sinxron TUZATMAYDI. Xato (yoki soxta) yozuv paydo
// bo'lsa uni o'chirishdan boshqa yo'l yo'q edi; DB'ga qo'l bilan kirmasdan
// buni qilish imkoni bo'lishi kerak.
//
// O'chirilgandan keyin telefon oxirgi 7 kunni qayta yuboradi va haqiqiy
// qiymatlar o'z-o'zidan tiklanadi.
func (s *ActivityService) DeleteRange(ctx context.Context, userID uuid.UUID, fromStr, toStr string) (int64, error) {
	from, err := time.ParseInLocation("2006-01-02", fromStr, s.loc)
	if err != nil {
		return 0, fmt.Errorf("%w: from sanasi", domain.ErrValidation)
	}
	to, err := time.ParseInLocation("2006-01-02", toStr, s.loc)
	if err != nil {
		return 0, fmt.Errorf("%w: to sanasi", domain.ErrValidation)
	}
	if to.Before(from) {
		return 0, fmt.Errorf("%w: oraliq teskari", domain.ErrValidation)
	}
	// Bir martada butun tarixni o'chirib yuborishdan himoya: xato bosilgan
	// tugma bilan yillik ma'lumot yo'qolmasin.
	if to.Sub(from) > maxDeleteRange {
		return 0, fmt.Errorf("%w: oraliq %d kundan oshmasin",
			domain.ErrValidation, int(maxDeleteRange.Hours()/24))
	}

	n, err := s.repo.DeleteRange(ctx, userID, from, to)
	if err != nil {
		return 0, fmt.Errorf("ActivityService.DeleteRange: %w", err)
	}
	return n, nil
}

// Today — mahalliy mintaqadagi bugungi sana (handler ham foydalanadi).
func (s *ActivityService) Today() time.Time { return s.today() }

// Location — ilova vaqt mintaqasi (handler query sanalarini shu mintaqada o'qiydi).
func (s *ActivityService) Location() *time.Location { return s.loc }

// today — APP_TIMEZONE dagi bugungi kunning 00:00 i.
func (s *ActivityService) today() time.Time {
	n := time.Now().In(s.loc)
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, s.loc)
}
