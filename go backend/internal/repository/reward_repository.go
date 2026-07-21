package repository

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rewardRepository struct {
	db *gorm.DB
}

func NewRewardRepository(db *gorm.DB) domain.RewardRepository {
	return &rewardRepository{db: db}
}

func (r *rewardRepository) List(ctx context.Context, f domain.RewardFilter) ([]domain.Reward, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.Reward{}).Where("deleted_at IS NULL")
	if f.Category != "" {
		q = q.Where("category = ?", f.Category)
	}
	if f.OnlyAvailable {
		now := time.Now()
		q = q.Where("is_active = TRUE").
			Where("stock IS NULL OR stock > 0").
			Where("starts_at IS NULL OR starts_at <= ?", now).
			Where("ends_at IS NULL OR ends_at >= ?", now)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("reward: count: %w", err)
	}

	var rows []domain.Reward
	err := q.Order("cost_coins, title").
		Limit(f.Limit).Offset((f.Page - 1) * f.Limit).
		Find(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("reward: list: %w", err)
	}
	return rows, total, nil
}

func (r *rewardRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Reward, error) {
	var out domain.Reward
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).First(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("reward: get: %w", err)
	}
	return &out, nil
}

func (r *rewardRepository) Create(ctx context.Context, rw *domain.Reward) error {
	return r.db.WithContext(ctx).Create(rw).Error
}

func (r *rewardRepository) Update(ctx context.Context, rw *domain.Reward) error {
	res := r.db.WithContext(ctx).Model(&domain.Reward{}).
		Where("id = ? AND deleted_at IS NULL", rw.ID).
		Select("title", "description", "image_url", "category", "cost_coins",
			"stock", "per_user_limit", "is_active", "starts_at", "ends_at",
			"config", "updated_at").
		Updates(rw)
	if res.Error != nil {
		return fmt.Errorf("reward: update: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete — soft delete. Buyurtmalar tarixi saqlanib qolishi kerak, shuning
// uchun qator o'chirilmaydi (foreign key ham buni taqiqlaydi).
func (r *rewardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&domain.Reward{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if res.Error != nil {
		return fmt.Errorf("reward: delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Redeem — sovg'ani coinga almashtiradi.
//
// BITTA TRANZAKSIYADA (CLAUDE.md §13.2), quyidagi tartibda:
//
//  1. Sovg'a qatorini FOR UPDATE bilan bloklaymiz. Bu MAJBURIY: aks holda
//     oxirgi bitta sovg'ani ikki foydalanuvchi bir vaqtda "ko'rib", ikkalasi
//     ham sotib olardi (stock -1 ga tushardi).
//  2. Vaqt oynasi, aktivligi va miqdorini tekshiramiz.
//  3. Foydalanuvchi limitini tekshiramiz.
//  4. Balansni SUM(amount) bilan hisoblaymiz — ledger modeli (§4.3): balans
//     ustuni yo'q, shuning uchun yangilanadigan qator ham yo'q.
//  5. Miqdorni kamaytiramiz.
//  6. Ledger'ga MANFIY yozuv (ref_id — buyurtma ID si).
//  7. Buyurtma yozuvi.
//
// ref_id sifatida REWARD emas, REDEMPTION ID ishlatiladi: fit_coins dagi
// unique indeks (user_id, reason, ref_id) reward bo'lganda foydalanuvchiga
// bitta sovg'ani ikkinchi marta olishga yo'l qo'ymasdi.
func (r *rewardRepository) Redeem(ctx context.Context, userID, rewardID uuid.UUID) (*domain.RewardRedemption, error) {
	var out *domain.RewardRedemption

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rw domain.Reward
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", rewardID).
			First(&rw).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}

		if !rw.Available(time.Now()) {
			return fmt.Errorf("%w: sovg'a hozir mavjud emas", domain.ErrValidation)
		}

		// Shaxsiy limit — bekor qilinganlar hisobga olinmaydi.
		if rw.PerUserLimit != nil {
			var taken int64
			if err := tx.Model(&domain.RewardRedemption{}).
				Where("user_id = ? AND reward_id = ? AND status <> ?",
					userID, rewardID, domain.RedemptionCancelled).
				Count(&taken).Error; err != nil {
				return err
			}
			if taken >= int64(*rw.PerUserLimit) {
				return fmt.Errorf("%w: bu sovg'a uchun limit tugagan", domain.ErrValidation)
			}
		}

		// Balans — ledger yig'indisi. Tranzaksiya ichida o'qiymiz, shuning
		// uchun parallel xarid oralab kirolmaydi (sovg'a qatori bloklangan;
		// bir foydalanuvchining ikki parallel xaridi ham shu yerda ketma-ket
		// bo'ladi, chunki ikkalasi ham o'sha qatorni kutadi).
		var balance int64
		if err := tx.Model(&domain.FitCoin{}).
			Where("user_id = ?", userID).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&balance).Error; err != nil {
			return err
		}
		if balance < int64(rw.CostCoins) {
			return domain.ErrInsufficientBalance
		}

		if rw.Stock != nil {
			// Shart bilan yangilash — qo'shimcha himoya qatlami.
			res := tx.Model(&domain.Reward{}).
				Where("id = ? AND stock > 0", rw.ID).
				UpdateColumn("stock", gorm.Expr("stock - 1"))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("%w: sovg'a tugagan", domain.ErrValidation)
			}
		}

		code, err := redemptionCode()
		if err != nil {
			return err
		}

		// ID oldindan yasaladi: ledger yozuvi shu buyurtmaga havola qiladi.
		redemptionID := uuid.New()
		coin := &domain.FitCoin{
			UserID:  userID,
			Amount:  -rw.CostCoins,
			Reason:  domain.CoinReasonPurchase,
			RefType: domain.CoinRefReward,
			RefID:   &redemptionID,
			Note:    rw.Title,
		}
		if err := tx.Create(coin).Error; err != nil {
			return err
		}

		red := &domain.RewardRedemption{
			ID:        redemptionID,
			UserID:    userID,
			RewardID:  rw.ID,
			CostCoins: rw.CostCoins,
			Status:    domain.RedemptionPending,
			Code:      code,
		}
		if err := tx.Create(red).Error; err != nil {
			return err
		}

		out = red
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *rewardRepository) ListRedemptions(ctx context.Context, f domain.RedemptionFilter) ([]domain.RedemptionDetail, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}

	where := "1 = 1"
	var args []any
	if f.Status != "" {
		where += " AND rr.status = ?"
		args = append(args, f.Status)
	}
	if f.UserID != nil {
		where += " AND rr.user_id = ?"
		args = append(args, *f.UserID)
	}

	var total int64
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM reward_redemptions rr WHERE %s`, where)
	if err := r.db.WithContext(ctx).Raw(countQ, args...).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("reward: redemptions count: %w", err)
	}

	// JOIN bilan bitta so'rov — sovg'a nomi va foydalanuvchi uchun alohida
	// so'rov qilinmaydi (§3.1 N+1).
	listQ := fmt.Sprintf(`
		SELECT rr.*,
		       rw.title      AS reward_title,
		       rw.image_url  AS reward_image_url,
		       u.full_name   AS user_full_name
		FROM reward_redemptions rr
		JOIN rewards rw ON rw.id = rr.reward_id
		JOIN users   u  ON u.id  = rr.user_id
		WHERE %s
		ORDER BY rr.created_at DESC
		LIMIT ? OFFSET ?`, where)

	all := append(append([]any{}, args...), f.Limit, (f.Page-1)*f.Limit)

	var rows []domain.RedemptionDetail
	if err := r.db.WithContext(ctx).Raw(listQ, all...).Scan(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("reward: redemptions: %w", err)
	}
	return rows, total, nil
}

// MarkIssued — buyurtmani topshirilgan deb belgilaydi.
func (r *rewardRepository) MarkIssued(ctx context.Context, id, adminID uuid.UUID, note string) (*domain.RewardRedemption, error) {
	var out *domain.RewardRedemption

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var red domain.RewardRedemption
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).First(&red).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		if red.Status != domain.RedemptionPending {
			return fmt.Errorf("%w: buyurtma holati %s", domain.ErrValidation, red.Status)
		}

		now := time.Now()
		red.Status = domain.RedemptionIssued
		red.IssuedAt = &now
		red.IssuedBy = &adminID
		if note != "" {
			red.Note = note
		}
		if err := tx.Model(&domain.RewardRedemption{}).Where("id = ?", id).
			Updates(map[string]any{
				"status": red.Status, "issued_at": red.IssuedAt,
				"issued_by": red.IssuedBy, "note": red.Note, "updated_at": now,
			}).Error; err != nil {
			return err
		}
		out = &red
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Cancel — buyurtmani bekor qiladi va coin'ni QAYTARADI.
//
// Qaytarish ledger'ga MUSBAT yozuv sifatida qo'shiladi (§4.3: yozuv
// o'chirilmaydi, teskari yozuv qo'shiladi). ref_id o'sha buyurtma —
// unique indeks tufayli ikki marta qaytarib bo'lmaydi.
//
// Miqdor ham tiklanadi: sovg'a hali topshirilmagan edi.
func (r *rewardRepository) Cancel(ctx context.Context, id, adminID uuid.UUID, note string) (*domain.RewardRedemption, error) {
	var out *domain.RewardRedemption

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var red domain.RewardRedemption
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).First(&red).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		if red.Status == domain.RedemptionCancelled {
			return domain.ErrAlreadyExists
		}

		refID := red.ID
		refund := &domain.FitCoin{
			UserID:  red.UserID,
			Amount:  red.CostCoins, // musbat — qaytarish
			Reason:  domain.CoinReasonPurchaseRefund,
			RefType: domain.CoinRefReward,
			RefID:   &refID,
			Note:    note,
		}
		if err := tx.Create(refund).Error; err != nil {
			if isUniqueViolation(err) {
				return domain.ErrAlreadyExists
			}
			return err
		}

		// Miqdorni tiklaymiz (faqat cheklangan sovg'ada).
		if err := tx.Model(&domain.Reward{}).
			Where("id = ? AND stock IS NOT NULL", red.RewardID).
			UpdateColumn("stock", gorm.Expr("stock + 1")).Error; err != nil {
			return err
		}

		now := time.Now()
		red.Status = domain.RedemptionCancelled
		red.Note = note
		if err := tx.Model(&domain.RewardRedemption{}).Where("id = ?", id).
			Updates(map[string]any{
				"status": red.Status, "note": red.Note, "updated_at": now,
			}).Error; err != nil {
			return err
		}
		out = &red
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// redemptionCode — buyurtma kodi.
//
// crypto/rand ishlatiladi: kod topshirishda shaxsni tasdiqlaydi, ya'ni
// taxmin qilib bo'lmasligi kerak (math/rand bashorat qilinadi).
// Chalkashadigan belgilar (0/O, 1/I) olib tashlangan — kod og'zaki
// aytiladi va qo'lda kiritiladi.
func redemptionCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	const n = 8

	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", fmt.Errorf("reward: kod yasash: %w", err)
		}
		b[i] = alphabet[idx.Int64()]
	}
	return string(b), nil
}
