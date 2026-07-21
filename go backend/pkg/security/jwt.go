package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token yaroqsiz")
	ErrExpiredToken = errors.New("token muddati o'tgan")
)

// TokenType — access yoki refresh tokenni ajratish uchun.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims — JWT ichidagi ma'lumotlar.
type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Role   string    `json:"role"`
	Type   TokenType `json:"typ"`
	jwt.RegisteredClaims
}

// JWTManager — token imzolash va tekshirish (HS256).
type JWTManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

// GenerateAccess — access token yaratadi.
func (m *JWTManager) GenerateAccess(userID uuid.UUID, role string) (string, error) {
	return m.generate(userID, role, AccessToken, m.accessSecret, m.accessTTL)
}

// GenerateRefresh — refresh token yaratadi.
func (m *JWTManager) GenerateRefresh(userID uuid.UUID, role string) (string, error) {
	return m.generate(userID, role, RefreshToken, m.refreshSecret, m.refreshTTL)
}

func (m *JWTManager) generate(userID uuid.UUID, role string, typ TokenType, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		Type:   typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseAccess — access tokenni tekshirib, claimlarni qaytaradi.
func (m *JWTManager) ParseAccess(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.accessSecret, AccessToken)
}

// ParseRefresh — refresh tokenni tekshirib, claimlarni qaytaradi.
func (m *JWTManager) ParseRefresh(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.refreshSecret, RefreshToken)
}

// RefreshTTL — refresh token amal qilish muddatini qaytaradi (Redis TTL uchun).
func (m *JWTManager) RefreshTTL() time.Duration { return m.refreshTTL }

func (m *JWTManager) parse(tokenStr string, secret []byte, expected TokenType) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		// Algoritm almashtirish hujumidan himoya.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if !token.Valid || claims.Type != expected {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
