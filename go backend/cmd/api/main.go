package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ttysi-fit/backend/config"
	"github.com/ttysi-fit/backend/internal/handler"
	"github.com/ttysi-fit/backend/internal/middleware"
	"github.com/ttysi-fit/backend/internal/repository"
	"github.com/ttysi-fit/backend/internal/service"
	"github.com/ttysi-fit/backend/internal/validation"
	"github.com/ttysi-fit/backend/pkg/certificate"
	"github.com/ttysi-fit/backend/pkg/database"
	"github.com/ttysi-fit/backend/pkg/hemis"
	"github.com/ttysi-fit/backend/pkg/logger"
	"github.com/ttysi-fit/backend/pkg/media"
	"github.com/ttysi-fit/backend/pkg/security"
	"go.uber.org/zap"
)

func main() {
	// 1. Config — .env fayli APP_ENV bo'yicha tanlanadi (.env.local / .env.production).
	cfg, err := config.Load(config.EnvFile())
	if err != nil {
		panic(err)
	}

	// 2. Logger
	log, err := logger.New(cfg.Flags.LogLevel, cfg.App.Env)
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	ctx := context.Background()

	// 3. PostgreSQL
	db, err := database.NewPostgres(ctx, cfg.DB, cfg.IsProduction())
	if err != nil {
		log.Fatal("postgres ulanmadi", zap.Error(err))
	}
	log.Info("postgres ulandi", zap.String("db", cfg.DB.Name))

	// 4. Redis
	rdb, err := database.NewRedis(ctx, cfg.Redis)
	if err != nil {
		log.Fatal("redis ulanmadi", zap.Error(err))
	}
	defer func() { _ = rdb.Close() }()
	log.Info("redis ulandi", zap.String("addr", cfg.Redis.Addr()))

	// 5. Gin router
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(cfg.App.AllowedOrigins, !cfg.IsProduction()))
	r.Use(middleware.Locale())                             // har bir so'rov tilini aniqlaydi
	r.Use(middleware.SecurityHeaders(cfg.IsProduction()))  // HSTS/CSP/nosniff (§17.3 #42–44)
	r.Use(middleware.BodyLimit(cfg.Security.MaxBodyBytes)) // DoS himoyasi (§17.3 #39)

	// Inbound rate limiting (§17.3 #15/#16/#40): umumiy + auth uchun qattiqroq.
	// Local'da default o'chirilgan (§17.1), prod'da majburiy (config.validate).
	var authLimiter []gin.HandlerFunc
	if cfg.Security.RateLimitEnabled {
		r.Use(middleware.RateLimit(rdb, log, middleware.RateLimitOpts{
			Name: "global", Limit: cfg.Security.RateLimitGlobal, Window: time.Minute,
		}))
		authLimiter = append(authLimiter, middleware.RateLimit(rdb, log, middleware.RateLimitOpts{
			Name: "auth", Limit: cfg.Security.RateLimitAuth, Window: time.Minute,
		}))
		log.Info("rate limiting yoqilgan",
			zap.Int("global_per_min", cfg.Security.RateLimitGlobal),
			zap.Int("auth_per_min", cfg.Security.RateLimitAuth),
		)
	}

	// Validator maydon xatolarida json nomlaridan foydalanish (i18n uchun)
	validation.RegisterJSONTagNames()
	validation.RegisterCustomRules()

	// Statik fayllar (yuklab olingan avatarlar): RoutePrefix → Media.Dir.
	// Masalan GET /static/avatars/123.jpg → ./uploads/avatars/123.jpg
	r.Static(cfg.Media.RoutePrefix, cfg.Media.Dir)

	// 5.1 Dependency Injection (domain → repository → service → handler)
	jwtManager := security.NewJWTManager(
		cfg.JWT.AccessSecret, cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL,
	)
	// Bekor qilingan qurilmani DARROV chiqaradi (mijoz X-Device-Id yuborsa).
	// jwtManager dan keyin turishi shart, route'lardan oldin.
	r.Use(middleware.DeviceSession(jwtManager, rdb, log))

	userRepo := repository.NewUserRepository(db)
	structureRepo := repository.NewStructureRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	ratingRepo := repository.NewRatingRepository(db)
	analyticsRepo := repository.NewAnalyticsRepository(db)
	rewardRepo := repository.NewRewardRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	hemisOAuthClient := hemis.NewOAuthClient(cfg.OAuth)
	authService := service.NewAuthService(userRepo, jwtManager, rdb, hemisOAuthClient, cfg.OAuth.StateTTL, cfg.OAuth.CodeTTL)
	// Qurilma sessiyalari: bir hisob — bir qurilma (ikkinchisida
	// kirilganda foydalanuvchidan rozilik so'raladi).
	authService.SetSessions(sessionRepo)
	hemisClient := hemis.NewClient(cfg.HEMIS)
	structureService := service.NewStructureService(structureRepo, hemisClient, cfg.HEMIS.FacultyTypeCode, cfg.HEMIS.DepartmentTypeCode)

	// Avatar yuklab oluvchi (sozlangan bo'lsa). avatarStore nil interfeys bo'lib
	// qolsa, RosterService rasm yuklab olmaydi va HEMIS URL'ini saqlaydi.
	var avatarStore service.AvatarStore
	if cfg.Media.DownloadAvatars {
		// publicBase "" — DB ga nisbiy yo'l yoziladi ("/static/avatars/1.jpg").
		// To'liq URL javob qaytarishda yasaladi (handler.absoluteMediaURL).
		dl, err := media.NewDownloader(
			cfg.Media.Dir, cfg.Media.RoutePrefix, "",
			cfg.Media.MaxImageBytes, cfg.Media.DownloadTimeout,
			cfg.Media.AllowedHosts, // SSRF allowlist (§17.3 #7)
		)
		if err != nil {
			log.Fatal("media downloader yaratilmadi", zap.Error(err))
		}
		avatarStore = dl
		log.Info("avatar yuklab olish yoqilgan",
			zap.String("dir", cfg.Media.Dir),
			zap.Int("workers", cfg.Media.DownloadWorkers),
		)
	}

	rosterService := service.NewRosterService(groupRepo, userRepo, hemisClient, avatarStore, log, cfg.Media.DownloadWorkers)
	userService := service.NewUserService(userRepo)
	ratingService := service.NewRatingService(ratingRepo, rdb, log)
	analyticsService := service.NewAnalyticsService(analyticsRepo, cfg.App.Timezone)
	// notificationService boshqa servislardan OLDIN: do'kon va yutuq
	// unga xabar yozadi (Notifier interfeysi orqali).
	notificationService := service.NewNotificationService(notificationRepo, log)
	rewardService := service.NewRewardService(rewardRepo, notificationService)
	challengeService := service.NewChallengeService(repository.NewChallengeRepository(db))
	// fitCoinRepo alohida o'zgaruvchida: yutuq mukofoti ham shu ledger'ga yozadi.
	fitCoinRepo := repository.NewFitCoinRepository(db)
	fitCoinService := service.NewFitCoinService(fitCoinRepo, notificationService)
	competitionService := service.NewCompetitionService(repository.NewCompetitionRepository(db))
	newsService := service.NewNewsService(repository.NewNewsRepository(db))
	trainingService := service.NewTrainingService(repository.NewTrainingRepository(db))

	// achievementService activityService dan OLDIN: faollik yozilgach avtomatik
	// yutuqlar shu orqali baholanadi (§16.3).
	// Sertifikat muhri/imzosi — startupda bir marta yuklanadi va tekshiriladi.
	// Fayl yaroqsiz bo'lsa server ishga tushmaydi: buzuq muhr bilan sertifikat
	// bergandan ko'ra, xatoni shu yerda ko'rish yaxshi.
	certSigning, err := certificate.LoadSigning(
		cfg.Cert.StampPath, cfg.Cert.SignaturePath,
		cfg.Cert.SignerName, cfg.Cert.SignerTitle, cfg.Cert.SampleStamp,
	)
	if err != nil {
		log.Fatal("sertifikat aktivlari", zap.Error(err))
	}

	achievementService := service.NewAchievementService(repository.NewAchievementRepository(db), fitCoinRepo, certSigning, notificationService)
	activityService := service.NewActivityService(activityRepo, achievementService, cfg.App.Timezone, log)

	v1 := r.Group("/api/v1")
	handler.NewHealthHandler(db, rdb).Register(v1)
	handler.NewAuthHandler(authService, jwtManager, log, cfg.OAuth.AppRedirect).Register(v1, authLimiter...)
	handler.NewStructureHandler(structureService, jwtManager, log).Register(v1)
	handler.NewRosterHandler(rosterService, jwtManager, log).Register(v1)
	handler.NewUserHandler(userService, jwtManager, log, cfg.Media.PublicBaseURL).Register(v1)
	handler.NewActivityHandler(activityService, jwtManager, log).Register(v1)
	handler.NewRatingHandler(ratingService, jwtManager, log, cfg.Media.PublicBaseURL).Register(v1)
	handler.NewAnalyticsHandler(analyticsService, jwtManager, log).Register(v1)
	handler.NewRewardHandler(rewardService, jwtManager, log, cfg.Media.PublicBaseURL).Register(v1)
	handler.NewNotificationHandler(notificationService, jwtManager, log).Register(v1)
	handler.NewChallengeHandler(challengeService, jwtManager, log).Register(v1)
	handler.NewFitCoinHandler(fitCoinService, jwtManager, log).Register(v1)
	handler.NewCompetitionHandler(competitionService, jwtManager, log).Register(v1)
	handler.NewNewsHandler(newsService, jwtManager, log, cfg.Media.PublicBaseURL).Register(v1)
	handler.NewTrainingHandler(trainingService, jwtManager, log, cfg.Media.PublicBaseURL).Register(v1)
	handler.NewAchievementHandler(achievementService, jwtManager, log, cfg.Media.PublicBaseURL).Register(v1)

	// 6. HTTP server + graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: r,
	}

	go func() {
		log.Info("server ishga tushdi", zap.String("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server xatosi", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("server to'xtatilmoqda...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown xatosi", zap.Error(err))
	}
	log.Info("server to'xtadi")
}
