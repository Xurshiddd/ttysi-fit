package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type fitCoinRepository struct {
	db *gorm.DB
}

func NewFitCoinRepository(db *gorm.DB) domain.FitCoinRepository {
	return &fitCoinRepository{db: db}
}

// Balance — bitta agregat so'rov (§3.1: alohida so'rovlar emas).
func (r *fitCoinRepository) Balance(ctx context.Context, userID uuid.UUID) (*domain.CoinBalance, error) {
	var b domain.CoinBalance
	err := r.db.WithContext(ctx).
		Table("fit_coins").
		Where("user_id = ?", userID).
		Select(`COALESCE(SUM(amount), 0) AS balance,
			COALESCE(SUM(amount) FILTER (WHERE amount > 0), 0) AS earned,
			COALESCE(-SUM(amount) FILTER (WHERE amount < 0), 0) AS spent`).
		Scan(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *fitCoinRepository) History(ctx context.Context, userID uuid.UUID, f domain.CoinFilter) ([]domain.FitCoin, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.FitCoin{}).Where("user_id = ?", userID)
	if f.Reason != "" {
		q = q.Where("reason = ?", f.Reason)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.FitCoin
	err := q.Order("created_at DESC").
		Limit(f.Limit).
		Offset((f.Page - 1) * f.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// isUniqueViolation — PostgreSQL 23505 (unique_violation).
// Idempotentlikni DB indeksi kafolatlaydi: oldindan SELECT bilan tekshirish
// poygaga ochiq bo'lardi (ikki parallel so'rov ikkalasi ham "yo'q" deb ko'rib
// ikkita yozuv qo'shardi). Shuning uchun yozib ko'ramiz va xatoni tutamiz.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}

func (r *fitCoinRepository) Grant(ctx context.Context, c *domain.FitCoin) error {
	err := r.db.WithContext(ctx).Create(c).Error
	if isUniqueViolation(err) {
		return domain.ErrAlreadyExists
	}
	return err
}

// GrantChallengeReward — chellenj mukofotini beradi.
//
// Bitta tranzaksiyada (CLAUDE.md §13.2):
//  1. user_challenges qatorini FOR UPDATE bilan bloklaymiz — parallel so'rov
//     kutadi va mukofot ikki marta berilmaydi;
//  2. chellenj yakunlanganini va mukofot hali berilmaganini tekshiramiz;
//  3. ledger'ga yozamiz;
//  4. reward_granted = true.
//
// Ikkinchi himoya qatlami — fit_coins dagi unique indeks (user_id, reason, ref_id).
func (r *fitCoinRepository) GrantChallengeReward(ctx context.Context, userID, challengeID uuid.UUID) (*domain.FitCoin, error) {
	var out *domain.FitCoin

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var uc domain.UserChallenge
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND challenge_id = ?", userID, challengeID).
			First(&uc).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}

		if uc.CompletedAt == nil {
			return fmt.Errorf("%w: chellenj hali yakunlanmagan", domain.ErrValidation)
		}
		if uc.RewardGranted {
			return domain.ErrAlreadyExists
		}

		var ch domain.Challenge
		if err := tx.Where("id = ? AND deleted_at IS NULL", challengeID).First(&ch).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}
		if ch.RewardCoins <= 0 {
			return fmt.Errorf("%w: bu chellenjda mukofot yo'q", domain.ErrValidation)
		}

		ref := challengeID
		coin := &domain.FitCoin{
			UserID:  userID,
			Amount:  ch.RewardCoins,
			Reason:  domain.CoinReasonChallengeReward,
			RefType: domain.CoinRefChallenge,
			RefID:   &ref,
			Note:    ch.Title,
		}
		if err := tx.Create(coin).Error; err != nil {
			if isUniqueViolation(err) {
				return domain.ErrAlreadyExists
			}
			return err
		}

		if err := tx.Model(&domain.UserChallenge{}).
			Where("id = ?", uc.ID).
			Update("reward_granted", true).Error; err != nil {
			return err
		}

		out = coin
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
