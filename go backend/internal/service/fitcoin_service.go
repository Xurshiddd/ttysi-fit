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
	// notify — coin harakati haqida xabar. nil bo'lsa xabar yozilmaydi.
	notify Notifier
}

func NewFitCoinService(repo domain.FitCoinRepository, notify Notifier) *FitCoinService {
	return &FitCoinService{repo: repo, notify: notify}
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
	coin, err := s.repo.GrantChallengeReward(ctx, userID, challengeID)
	if err != nil {
		return nil, err
	}

	// Mukofot tushgani haqida xabar. ref_id — ledger yozuvi, shuning uchun
	// takroriy so'rovda ikkinchi xabar chiqmaydi.
	if s.notify != nil && coin != nil {
		ref := coin.ID
		s.notify.Notify(ctx, domain.Notification{
			UserID:  userID,
			Type:    domain.NotifyChallenge,
			Title:   "Chellenj mukofoti",
			Body:    fmt.Sprintf("%s — %d FIT Coin hisobingizga tushdi", coin.Note, coin.Amount),
			RefType: domain.CoinRefChallenge,
			RefID:   &ref,
		})
	}
	return coin, nil
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
