package security

import "golang.org/x/crypto/bcrypt"

// bcrypt cost — CLAUDE.md 3.2: cost >= 12.
const bcryptCost = 12

// HashPassword — parolni bcrypt bilan xeshlaydi.
func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword — ochiq parol xeshga mosligini tekshiradi.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
