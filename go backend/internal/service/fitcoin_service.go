package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// FitCoinService — FIT Coin use-case qatlami.
type FitCoinService struct {
	repo domain.FitCoinRepository
}

func NewFitCoinService(repo domain.FitCoinRepository) *FitCoinService {
	return &FitCoinService{repo: repo}
}

func (s *FitCoinService) Balance(ctx context.Context, userID uuid.UUID) (*domain.CoinBalance, error) {
	return s.repo.Balance(ctx, userID)
}

func (s *FitCoinService) History(ctx context.Context, userID uuid.UUID, f domain.CoinFilter) ([]domain.FitCoin, int64, error) {
	return s.repo.History(ctx, userID, f)
}

// ClaimChallengeReward — foydalanuvchi yakunlangan chellenj mukofotini oladi.
// Idempotent: ikkinchi marta ErrAlreadyExists.
func (s *FitCoinService) ClaimChallengeReward(ctx context.Context, userID, challengeID uuid.UUID) (*domain.FitCoin, error) {
	return s.repo.GrantChallengeReward(ctx, userID, challengeID)
}

// AdminGrant — admin qo'lda coin beradi yoki oladi (manfiy amount).
//
// ref_id berilmaydi, ya'ni idempotentlik indeksi qo'llanmaydi: admin bir
// foydalanuvchiga bir necha marta coin bera olishi kerak. Har bir amal ledger'da
// alohida qator bo'lib qoladi.
func (s *FitCoinService) AdminGrant(ctx context.Context, userID uuid.UUID, amount int, note string) (*domain.FitCoin, error) {
	if amount == 0 {
		return nil, fmt.Errorf("%w: miqdor nol bo'lmasin", domain.ErrValidation)
	}

	reason := domain.CoinReasonAdminGrant
	if amount < 0 {
		reason = domain.CoinReasonAdminRevoke

		// Manfiy balansga tushirmaymiz: coin — hisob birligi, qarz emas.
		bal, err := s.repo.Balance(ctx, userID)
		if err != nil {
			return nil, err
		}
		if bal.Balance+amount < 0 {
			return nil, domain.ErrInsufficientBalance
		}
	}

	c := &domain.FitCoin{
		UserID: userID,
		Amount: amount,
		Reason: reason,
		Note:   note,
	}
	if err := s.repo.Grant(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
