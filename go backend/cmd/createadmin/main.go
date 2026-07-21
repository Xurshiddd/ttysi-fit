// cmd/createadmin — admin foydalanuvchi yaratadi (yoki mavjudini admin qiladi).
//
// Foydalanish:
//   go run cmd/createadmin/main.go <email> <parol> [to'liq_ism]
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ttysi-fit/backend/config"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/internal/repository"
	"github.com/ttysi-fit/backend/pkg/database"
	"github.com/ttysi-fit/backend/pkg/security"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Foydalanish: go run cmd/createadmin/main.go <email> <parol> [to'liq_ism]")
		os.Exit(1)
	}
	email, password := os.Args[1], os.Args[2]
	fullName := "Administrator"
	if len(os.Args) > 3 {
		fullName = os.Args[3]
	}

	cfg, err := config.Load(config.EnvFile())
	if err != nil {
		fmt.Println("config xato:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.NewPostgres(ctx, cfg.DB, cfg.IsProduction())
	if err != nil {
		fmt.Println("postgres ulanmadi:", err)
		os.Exit(1)
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		fmt.Println("parol xeshlash xato:", err)
		os.Exit(1)
	}

	repo := repository.NewUserRepository(db)

	existing, err := repo.GetByEmail(ctx, email)
	switch {
	case err == nil:
		// Mavjud foydalanuvchini admin qilish + parolni yangilash
		existing.Role = domain.RoleAdmin
		existing.Password = hash
		existing.IsActive = true
		if err := repo.Update(ctx, existing); err != nil {
			fmt.Println("yangilash xato:", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Mavjud foydalanuvchi admin qilindi: %s\n", email)
	case errors.Is(err, domain.ErrNotFound):
		u := &domain.User{
			FullName: fullName,
			Email:    email,
			Password: hash,
			Role:     domain.RoleAdmin,
			IsActive: true,
			Language: "uz",
		}
		if err := repo.Create(ctx, u); err != nil {
			fmt.Println("yaratish xato:", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Admin yaratildi: %s (%s)\n", email, u.ID)
	default:
		fmt.Println("xato:", err)
		os.Exit(1)
	}
}
