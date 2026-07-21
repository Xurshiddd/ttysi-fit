// cmd/sync — HEMIS ma'lumotlarini DB ga sinxronlovchi CLI buyrug'i.
// Admin JWT shart emas — bootstrap uchun to'g'ridan-to'g'ri ishlatiladi.
//
// Foydalanish:
//
//	go run cmd/sync/main.go            # hammasi (structures→groups→students→employees)
//	go run cmd/sync/main.go structures
//	go run cmd/sync/main.go groups
//	go run cmd/sync/main.go students
//	go run cmd/sync/main.go employees
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ttysi-fit/backend/config"
	"github.com/ttysi-fit/backend/internal/domain"
	"github.com/ttysi-fit/backend/internal/repository"
	"github.com/ttysi-fit/backend/internal/service"
	"github.com/ttysi-fit/backend/pkg/database"
	"github.com/ttysi-fit/backend/pkg/hemis"
	"github.com/ttysi-fit/backend/pkg/logger"
	"github.com/ttysi-fit/backend/pkg/media"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load(config.EnvFile())
	if err != nil {
		fmt.Println("config xato:", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Flags.LogLevel, cfg.App.Env)
	if err != nil {
		fmt.Println("logger xato:", err)
		os.Exit(1)
	}
	defer func() { _ = log.Sync() }()

	if cfg.HEMIS.Token == "" {
		log.Fatal("HEMIS_TOKEN bo'sh — .env.local ga token kiriting")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	db, err := database.NewPostgres(ctx, cfg.DB, cfg.IsProduction())
	if err != nil {
		log.Fatal("postgres ulanmadi", zap.Error(err))
	}

	client := hemis.NewClient(cfg.HEMIS)
	structureSvc := service.NewStructureService(
		repository.NewStructureRepository(db), client,
		cfg.HEMIS.FacultyTypeCode, cfg.HEMIS.DepartmentTypeCode,
	)

	// Avatar yuklab oluvchi (API bilan bir xil mantiq). Sozlanmagan bo'lsa nil.
	var avatarStore service.AvatarStore
	if cfg.Media.DownloadAvatars {
		// publicBase "" — API bilan bir xil: DB ga nisbiy yo'l yoziladi.
		dl, err := media.NewDownloader(
			cfg.Media.Dir, cfg.Media.RoutePrefix, "",
			cfg.Media.MaxImageBytes, cfg.Media.DownloadTimeout,
			cfg.Media.AllowedHosts, // SSRF allowlist (§17.3 #7)
		)
		if err != nil {
			log.Fatal("media downloader yaratilmadi", zap.Error(err))
		}
		avatarStore = dl
	}

	rosterSvc := service.NewRosterService(
		repository.NewGroupRepository(db),
		repository.NewUserRepository(db),
		client,
		avatarStore,
		log,
		cfg.Media.DownloadWorkers,
	)

	// Sync bosqichlari (tartib muhim: structures → groups → students/employees)
	type step struct {
		name string
		fn   func(context.Context) (*domain.SyncStats, error)
	}
	steps := map[string][]step{
		"structures": {{"Strukturalar", structureSvc.SyncStructures}},
		"groups":     {{"Guruhlar", rosterSvc.SyncGroups}},
		"students":   {{"Talabalar", rosterSvc.SyncStudents}},
		"employees":  {{"O'qituvchilar", rosterSvc.SyncEmployees}},
	}

	target := "all"
	if len(os.Args) > 1 {
		target = os.Args[1]
	}

	var order []step
	if target == "all" {
		order = append(order, steps["structures"]...)
		order = append(order, steps["groups"]...)
		order = append(order, steps["students"]...)
		order = append(order, steps["employees"]...)
	} else if s, ok := steps[target]; ok {
		order = s
	} else {
		log.Fatal("noma'lum buyruq: " + target + " (all|structures|groups|students|employees)")
	}

	for _, st := range order {
		start := time.Now()
		log.Info("sync boshlandi", zap.String("bosqich", st.name))
		stats, err := st.fn(ctx)
		if err != nil {
			log.Fatal("sync xatosi", zap.String("bosqich", st.name), zap.Error(err))
		}
		fmt.Printf("✅ %-14s Jami: %-6d Yangi: %-6d Yangilangan: %-6d (%s)\n",
			st.name, stats.Total, stats.Created, stats.Updated,
			time.Since(start).Round(time.Millisecond))
	}
}
