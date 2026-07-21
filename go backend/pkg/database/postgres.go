package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ttysi-fit/backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgres — gorm orqali PostgreSQL ulanishini ochadi va pool ni sozlaydi.
func NewPostgres(ctx context.Context, cfg config.DBConfig, isProd bool) (*gorm.DB, error) {
	logLevel := logger.Info
	if isProd {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("postgres: ulanib bo'lmadi: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("postgres: sql.DB olinmadi: %w", err)
	}

	// Connection pool (CLAUDE.md 13.1)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(1 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("postgres: ping muvaffaqiyatsiz: %w", err)
	}

	return db, nil
}
