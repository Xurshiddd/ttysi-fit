package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ttysi-fit/backend/internal/domain"
)

// UserService — foydalanuvchi use-case qatlami (admin va profil uchun).
type UserService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{users: users}
}

// ListUsers — admin uchun paginatsiyali ro'yxat.
func (s *UserService) ListUsers(ctx context.Context, f domain.UserFilter) ([]domain.UserListItem, int64, error) {
	return s.users.List(ctx, f)
}

// GetProfile — foydalanuvchining o'z profili (token'dagi ID bo'yicha).
func (s *UserService) GetProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	return s.users.GetProfile(ctx, id)
}

// UpdateProfile — o'z profilini yangilash. Faqat domain.ProfileUpdate dagi
// maydonlar o'zgaradi; qolgani HEMIS sync ixtiyorida.
func (s *UserService) UpdateProfile(ctx context.Context, id uuid.UUID, p domain.ProfileUpdate) (*domain.UserProfile, error) {
	return s.users.UpdateProfile(ctx, id, p)
}
