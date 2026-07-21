package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// FIT Coin do'koni sabablari (fit_coins.reason).
const (
	// CoinReasonPurchase — sovg'aga almashtirish (manfiy yozuv).
	CoinReasonPurchase = "purchase"
	// CoinReasonPurchaseRefund — buyurtma bekor qilinganda qaytarish.
	CoinReasonPurchaseRefund = "purchase_refund"
)

// CoinRefReward — ref_type qiymati: manba almashtirish yozuvi.
const CoinRefReward = "reward"

// Sovg'a kategoriyalari. Ro'yxat kodda: kategoriya UI da filtr sifatida
// ishlatiladi va tarjima qilinadi (kontentdan farqli o'laroq).
const (
	RewardCategoryMerch       = "merch"       // kiyim, sumka, buyum
	RewardCategoryEquipment   = "equipment"   // sport jihozi
	RewardCategoryCertificate = "certificate" // chegirma/sertifikat
	RewardCategoryOther       = "other"
)

// ValidRewardCategory — kategoriya ro'yxatda bormi.
func ValidRewardCategory(c string) bool {
	switch c {
	case RewardCategoryMerch, RewardCategoryEquipment,
		RewardCategoryCertificate, RewardCategoryOther:
		return true
	}
	return false
}

// Almashtirish holatlari.
const (
	RedemptionPending   = "pending"   // buyurtma qabul qilindi
	RedemptionIssued    = "issued"    // topshirildi
	RedemptionCancelled = "cancelled" // bekor qilindi, coin qaytarildi
)

// Reward — do'kondagi sovg'a (admin panel yaratadi, §16).
type Reward struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" gorm:"index"`

	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Category    string `json:"category" gorm:"not null"`

	CostCoins int `json:"cost_coins" gorm:"not null"`

	// Stock — qolgan miqdor. nil — cheksiz.
	Stock *int `json:"stock"`
	// PerUserLimit — bitta foydalanuvchi necha marta ola oladi. nil — cheklovsiz.
	PerUserLimit *int `json:"per_user_limit"`

	IsActive bool       `json:"is_active" gorm:"not null;default:true"`
	StartsAt *time.Time `json:"starts_at,omitempty"`
	EndsAt   *time.Time `json:"ends_at,omitempty"`

	Config datatypes.JSON `json:"config,omitempty" gorm:"type:jsonb"`
}

func (Reward) TableName() string { return "rewards" }

// Available — sovg'a ayni damda olinishi mumkinmi (miqdor va vaqt bo'yicha).
// Foydalanuvchi balansi va shaxsiy limiti bu yerda tekshirilmaydi.
func (r Reward) Available(now time.Time) bool {
	if !r.IsActive || r.DeletedAt != nil {
		return false
	}
	if r.StartsAt != nil && now.Before(*r.StartsAt) {
		return false
	}
	if r.EndsAt != nil && now.After(*r.EndsAt) {
		return false
	}
	if r.Stock != nil && *r.Stock <= 0 {
		return false
	}
	return true
}

// RewardRedemption — almashtirish yozuvi.
type RewardRedemption struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	RewardID uuid.UUID `json:"reward_id" gorm:"type:uuid;not null"`

	// CostCoins — buyurtma paytidagi narx (keyingi narx o'zgarishi ta'sir qilmaydi).
	CostCoins int    `json:"cost_coins" gorm:"not null"`
	Status    string `json:"status" gorm:"not null"`
	Code      string `json:"code" gorm:"not null"`

	IssuedAt *time.Time `json:"issued_at,omitempty"`
	IssuedBy *uuid.UUID `json:"issued_by,omitempty" gorm:"type:uuid"`
	Note     string     `json:"note,omitempty"`
}

func (RewardRedemption) TableName() string { return "reward_redemptions" }

// RedemptionDetail — ro'yxat uchun sovg'a nomi bilan birga (JOIN natijasi).
// Alohida so'rovlarsiz (§3.1).
type RedemptionDetail struct {
	RewardRedemption
	RewardTitle    string `json:"reward_title"`
	RewardImageURL string `json:"reward_image_url,omitempty"`
	UserFullName   string `json:"user_full_name,omitempty"` // admin ro'yxatida
}

// RewardFilter — do'kon ro'yxati so'rovi.
type RewardFilter struct {
	Category string
	// OnlyAvailable — faqat olinishi mumkin bo'lganlar (mobil do'kon).
	OnlyAvailable bool
	Page          int
	Limit         int
}

// RedemptionFilter — buyurtmalar ro'yxati.
type RedemptionFilter struct {
	Status string
	UserID *uuid.UUID // nil — hamma (admin)
	Page   int
	Limit  int
}

// RewardRepository — do'kon uchun port.
type RewardRepository interface {
	List(ctx context.Context, f RewardFilter) ([]Reward, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Reward, error)
	Create(ctx context.Context, r *Reward) error
	Update(ctx context.Context, r *Reward) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Redeem — sovg'ani almashtiradi. BITTA tranzaksiyada:
	// sovg'ani bloklaydi, miqdor/limit/balansni tekshiradi, coin yechadi
	// va buyurtma yaratadi. Balans yetmasa ErrInsufficientBalance.
	Redeem(ctx context.Context, userID, rewardID uuid.UUID) (*RewardRedemption, error)

	ListRedemptions(ctx context.Context, f RedemptionFilter) ([]RedemptionDetail, int64, error)
	// MarkIssued — buyurtmani topshirilgan deb belgilaydi (admin).
	MarkIssued(ctx context.Context, id, adminID uuid.UUID, note string) (*RewardRedemption, error)
	// Cancel — buyurtmani bekor qiladi va coin'ni QAYTARADI (admin).
	Cancel(ctx context.Context, id, adminID uuid.UUID, note string) (*RewardRedemption, error)
}
