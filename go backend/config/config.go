package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	// tzdata — IANA vaqt mintaqalari bazasi binarga singdiriladi.
	// Ansiz `time.LoadLocation("Asia/Tashkent")` scratch/alpine konteynerda va
	// Windows'da xato beradi (tizimda zoneinfo yo'q). ~450 KB.
	_ "time/tzdata"

	"github.com/joho/godotenv"
)

// Config — ilovaning barcha sozlamalari (.env asosida).
type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	HEMIS    HEMISConfig
	Media    MediaConfig
	OAuth    HEMISOAuthConfig
	Flags    FeatureFlags
	Security SecurityConfig
	Cert     CertificateConfig
}

// CertificateConfig — sertifikat imzo bloki (muhr, imzo, imzolovchi).
//
// Rasmlar kodga singdirilmagan: muhr/imzo skani institutning rasmiy hujjati,
// u repozitoriyda emas, serverdagi faylda turishi kerak.
type CertificateConfig struct {
	// StampPath — muhr PNG yo'li. Bo'sh bo'lsa muhr chizilmaydi.
	StampPath string
	// SignaturePath — imzo PNG yo'li. Bo'sh bo'lsa faqat chiziq qoladi.
	SignaturePath string
	// SignerName — imzolovchining F.I.O. si (chiziq ostida chiqadi).
	SignerName string
	// SignerTitle — lavozimi (masalan "Rektor").
	SignerTitle string
	// SampleStamp — "NAMUNA" muhrini chizish (haqiqiy muhr o'rniga).
	//
	// Faqat APP_ENV=local da ruxsat: production'da sertifikat muhrsiz
	// chiqqani "NAMUNA" yozuvi bilan chiqqanidan yaxshi (§17.1).
	SampleStamp bool
}

// SecurityConfig — perimetr himoyasi sozlamalari (CLAUDE.md §17).
type SecurityConfig struct {
	// RateLimitEnabled — inbound rate limiting yoqilganmi.
	// Default: local'dan boshqa barcha muhitlarda yoqilgan (§17.1).
	RateLimitEnabled bool
	// RateLimitGlobal — IP boshiga daqiqasiga umumiy so'rov limiti.
	RateLimitGlobal int
	// RateLimitAuth — IP boshiga daqiqasiga /auth/* so'rov limiti (brute-force).
	RateLimitAuth int
	// MaxBodyBytes — kiruvchi request body maksimal hajmi (DoS himoyasi).
	MaxBodyBytes int64
}

type AppConfig struct {
	Env            string
	Port           string
	Name           string
	AllowedOrigins []string // CORS uchun ruxsat etilgan originlar

	// TimezoneName — .env dagi APP_TIMEZONE (default "Asia/Tashkent").
	TimezoneName string
	// Timezone — kunlik chegaralarni (qaysi qadam qaysi kunga tegishli)
	// hisoblash uchun mintaqa. UTC ishlatilsa O'zbekistonda (UTC+5) mahalliy
	// 00:00–05:00 oralig'idagi faollik KECHAGI kunga yozilib, o'sha kunning
	// yozuvini buzardi. Server soati emas, aynan shu mintaqa asos qilinadi.
	Timezone *time.Location
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	// Timezone — sessiya vaqt mintaqasi (APP_TIMEZONE dan). DSN orqali
	// beriladi, shuning uchun SQL dagi CURRENT_DATE / NOW() server soatiga
	// emas, shu mintaqaga bo'ysunadi — reyting davri (rating_repository) va
	// kunlik statistika (activity_repository) ayni bir kunni ko'radi.
	Timezone string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// HEMISConfig — HEMIS integratsiyasi sozlamalari.
type HEMISConfig struct {
	BaseURL string        // masalan https://student.hemis.uz/rest/v1
	Token   string        // .env dagi API token
	Timeout time.Duration // so'rov timeout
	// Endpointlar (BaseURL ga nisbatan).
	StructurePath string
	GroupPath     string
	StudentPath   string
	EmployeePath  string
	// EmployeeType — employee-list uchun ?type parametri (masalan "all").
	EmployeeType string
	// FacultyTypeCode — qaysi structureType kodi "fakultet" ekanligi.
	FacultyTypeCode string
	// DepartmentTypeCode — qaysi structureType kodi "kafedra" ekanligi.
	DepartmentTypeCode string
	// PageLimit — bitta so'rovdagi yozuvlar soni (paginatsiya).
	PageLimit int
	// RateLimit — sekundiga ruxsat etilgan maksimal so'rov soni.
	// HEMIS 10 req/sek dan oshganda bloklaydi.
	RateLimit int
}

// MediaConfig — yuklab olinadigan media (avatar) fayllari sozlamalari.
type MediaConfig struct {
	// Dir — fayllar saqlanadigan lokal papka (masalan "./uploads").
	Dir string
	// RoutePrefix — fayllar xizmat qilinadigan HTTP yo'l prefiksi (masalan "/static").
	RoutePrefix string
	// PublicBaseURL — javob qaytarishda avatar nisbiy yo'liga qo'shiladigan asos
	// (masalan "http://localhost:8090"). DB da faqat nisbiy yo'l saqlanadi, shuning
	// uchun bu qiymatni o'zgartirish (local → production) eski yozuvlarni buzmaydi.
	PublicBaseURL string
	// DownloadAvatars — sync paytida rasmlar yuklab olinadimi (default: true).
	DownloadAvatars bool
	// MaxImageBytes — bitta rasm uchun maksimal hajm (default: 5MB).
	MaxImageBytes int64
	// DownloadWorkers — parallel yuklab oluvchi goroutine soni (default: 8).
	DownloadWorkers int
	// DownloadTimeout — bitta rasmni yuklab olish timeout'i (default: 15s).
	DownloadTimeout time.Duration
	// AllowedHosts — SSRF himoyasi: rasm faqat shu hostlardan yuklab olinadi
	// (default: HEMIS domenlari). §17.3 #7.
	AllowedHosts []string
}

// HEMISOAuthProvider — bitta HEMIS OAuth provayderi (talaba yoki xodim).
type HEMISOAuthProvider struct {
	ClientID     string
	ClientSecret string
	AuthorizeURL string
	TokenURL     string
	ResourceURL  string
	RedirectURI  string
}

// Configured — provayder ishlatishga tayyor (asosiy maydonlar to'ldirilgan).
func (p HEMISOAuthProvider) Configured() bool {
	return p.ClientID != "" && p.AuthorizeURL != "" && p.TokenURL != "" && p.ResourceURL != ""
}

// HEMISOAuthConfig — HEMIS OAuth (talaba + xodim) sozlamalari.
type HEMISOAuthConfig struct {
	Student  HEMISOAuthProvider
	Employee HEMISOAuthProvider
	Scopes   []string
	StateTTL time.Duration
	// AppRedirect — muvaffaqiyatli callback'dan keyin mobil ilovaga deep link
	// (masalan "ttysifit://oauth/callback"). Bo'sh bo'lsa, JSON token qaytadi (web/test).
	AppRedirect string
	// CodeTTL — bir martalik exchange code'ning yashash muddati (mobil oqim uchun).
	CodeTTL time.Duration
}

type FeatureFlags struct {
	SeedFakeData  bool
	EnableSwagger bool
	LogLevel      string
	MockSMS       bool
	MockEmail     bool
}

// IsProduction — production muhitda ekanligini bildiradi.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// DSN — PostgreSQL ulanish satrini qaytaradi.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.Timezone,
	)
}

// Addr — Redis manzilini qaytaradi.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// EnvFile — yuklanadigan .env faylini tanlaydi (tizim env asosida):
//  1. ENV_FILE berilgan bo'lsa — o'sha fayl;
//  2. aks holda ".env.<APP_ENV>" (masalan .env.production);
//  3. APP_ENV berilmagan bo'lsa — ".env.local".
//
// Shu tufayli prod'da hardcode qilingan .env.local yuklanmaydi (§17.3 #46/#47).
func EnvFile() string {
	if f := os.Getenv("ENV_FILE"); f != "" {
		return f
	}
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	return ".env." + env
}

// Load — .env faylini o'qib, Config ni to'ldiradi.
// envFile bo'sh bo'lsa, faqat tizim env o'zgaruvchilari ishlatiladi.
func Load(envFile string) (*Config, error) {
	if envFile != "" {
		// Fayl bo'lmasa ham xato emas — CI/prod da env tizimdan keladi.
		_ = godotenv.Load(envFile)
	}

	appEnv := getEnv("APP_ENV", "local")

	tzName := getEnv("APP_TIMEZONE", "Asia/Tashkent")
	tz, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("config: APP_TIMEZONE=%q noto'g'ri: %w", tzName, err)
	}

	cfg := &Config{
		App: AppConfig{
			Env:            appEnv,
			Port:           getEnv("APP_PORT", "8090"),
			Name:           getEnv("APP_NAME", "ttysi_fit"),
			AllowedOrigins: splitCSV(getEnv("CORS_ORIGINS", "http://localhost:3000")),
			TimezoneName:   tzName,
			Timezone:       tz,
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "ttysi_fit_dev"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			Timezone: tzName,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6380"), // docker-compose host porti (6379 boshqa loyihada band)
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", ""),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", ""),
			AccessTTL:     getEnvDuration("JWT_ACCESS_TTL", 15*time.Minute),
			RefreshTTL:    getEnvDuration("JWT_REFRESH_TTL", 168*time.Hour),
		},
		HEMIS: HEMISConfig{
			// .env da faqat asosiy URL va token saqlanadi; qolganlar kod default'lari.
			BaseURL:         getEnv("HEMIS_BASE_URL", "https://student.ttyesi.uz/rest/v1"),
			Token:           getEnv("HEMIS_TOKEN", ""),
			Timeout:         getEnvDuration("HEMIS_TIMEOUT", 30*time.Second),
			StructurePath:   getEnv("HEMIS_STRUCTURE_PATH", "/data/department-list"),
			GroupPath:       getEnv("HEMIS_GROUP_PATH", "/data/group-list"),
			StudentPath:     getEnv("HEMIS_STUDENT_PATH", "/data/student-list"),
			EmployeePath:    getEnv("HEMIS_EMPLOYEE_PATH", "/data/employee-list"),
			EmployeeType:    getEnv("HEMIS_EMPLOYEE_TYPE", "all"),
			FacultyTypeCode:    getEnv("HEMIS_FACULTY_TYPE_CODE", "11"),
			DepartmentTypeCode: getEnv("HEMIS_DEPARTMENT_TYPE_CODE", "12"),
			PageLimit:       getEnvInt("HEMIS_PAGE_LIMIT", 200),
			RateLimit:       getEnvInt("HEMIS_RATE_LIMIT", 10),
		},
		Media: MediaConfig{
			Dir:             getEnv("MEDIA_DIR", "./uploads"),
			RoutePrefix:     getEnv("MEDIA_ROUTE_PREFIX", "/static"),
			PublicBaseURL:   strings.TrimRight(getEnv("MEDIA_PUBLIC_BASE_URL", "http://localhost:8090"), "/"),
			DownloadAvatars: getEnvBool("MEDIA_DOWNLOAD_AVATARS", true),
			MaxImageBytes:   int64(getEnvInt("MEDIA_MAX_IMAGE_BYTES", 5<<20)), // 5MB
			DownloadWorkers: getEnvInt("MEDIA_DOWNLOAD_WORKERS", 8),
			DownloadTimeout: getEnvDuration("MEDIA_DOWNLOAD_TIMEOUT", 15*time.Second),
			AllowedHosts:    splitCSV(getEnv("MEDIA_ALLOWED_HOSTS", "student.ttyesi.uz,hemis.ttyesi.uz")),
		},
		OAuth: HEMISOAuthConfig{
			// Talaba provayderi (student.ttyesi.uz). client_id/secret .env'dan.
			Student: HEMISOAuthProvider{
				ClientID:     getEnv("HEMIS_OAUTH_STUDENT_CLIENT_ID", ""),
				ClientSecret: getEnv("HEMIS_OAUTH_STUDENT_CLIENT_SECRET", ""),
				AuthorizeURL: getEnv("HEMIS_OAUTH_STUDENT_AUTHORIZE_URL", "https://student.ttyesi.uz/oauth/authorize"),
				TokenURL:     getEnv("HEMIS_OAUTH_STUDENT_TOKEN_URL", "https://student.ttyesi.uz/oauth/access-token"),
				ResourceURL:  getEnv("HEMIS_OAUTH_STUDENT_RESOURCE_URL", "https://student.ttyesi.uz/oauth/api/user?fields=id,uuid,type,name,login,picture,email,university_id,phone"),
				RedirectURI:  getEnv("HEMIS_OAUTH_STUDENT_REDIRECT_URI", "http://localhost:8090/api/v1/auth/hemis/student/callback"),
			},
			// Xodim provayderi (hemis.ttyesi.uz). client_id/secret .env'dan.
			Employee: HEMISOAuthProvider{
				ClientID:     getEnv("HEMIS_OAUTH_EMPLOYEE_CLIENT_ID", ""),
				ClientSecret: getEnv("HEMIS_OAUTH_EMPLOYEE_CLIENT_SECRET", ""),
				AuthorizeURL: getEnv("HEMIS_OAUTH_EMPLOYEE_AUTHORIZE_URL", "https://hemis.ttyesi.uz/oauth/authorize"),
				TokenURL:     getEnv("HEMIS_OAUTH_EMPLOYEE_TOKEN_URL", "https://hemis.ttyesi.uz/oauth/access-token"),
				ResourceURL:  getEnv("HEMIS_OAUTH_EMPLOYEE_RESOURCE_URL", "https://hemis.ttyesi.uz/oauth/api/user?fields=id,uuid,type,name,login,picture,email,university_id,phone"),
				RedirectURI:  getEnv("HEMIS_OAUTH_EMPLOYEE_REDIRECT_URI", "http://localhost:8090/api/v1/auth/hemis/employee/callback"),
			},
			Scopes:      splitCSV(getEnv("HEMIS_OAUTH_SCOPES", "")),
			StateTTL:    getEnvDuration("HEMIS_OAUTH_STATE_TTL", 10*time.Minute),
			// Default mobil deep link — web/test uchun bo'sh qilib qo'ying (JSON qaytadi).
			AppRedirect: getEnv("HEMIS_OAUTH_APP_REDIRECT", "ttysifit://oauth/callback"),
			CodeTTL:     getEnvDuration("HEMIS_OAUTH_CODE_TTL", 60*time.Second),
		},
		Flags: FeatureFlags{
			SeedFakeData:  getEnvBool("SEED_FAKE_DATA", false),
			EnableSwagger: getEnvBool("ENABLE_SWAGGER", false),
			LogLevel:      getEnv("LOG_LEVEL", "info"),
			MockSMS:       getEnvBool("MOCK_SMS", false),
			MockEmail:     getEnvBool("MOCK_EMAIL", false),
		},
		Cert: CertificateConfig{
			StampPath:     getEnv("CERT_STAMP_PATH", ""),
			SignaturePath: getEnv("CERT_SIGNATURE_PATH", ""),
			SignerName:    getEnv("CERT_SIGNER_NAME", ""),
			SignerTitle:   getEnv("CERT_SIGNER_TITLE", ""),
			// NAMUNA muhri default: faqat local'da yoqilgan (§17.1).
			SampleStamp: getEnvBool("CERT_SAMPLE_STAMP", appEnv == "local"),
		},
		Security: SecurityConfig{
			// Rate limiting default: faqat local'da o'chirilgan (§17.1).
			RateLimitEnabled: getEnvBool("RATE_LIMIT_ENABLED", appEnv != "local"),
			RateLimitGlobal:  getEnvInt("RATE_LIMIT_GLOBAL_PER_MIN", 120),
			RateLimitAuth:    getEnvInt("RATE_LIMIT_AUTH_PER_MIN", 10),
			MaxBodyBytes:     int64(getEnvInt("MAX_BODY_BYTES", 1<<20)), // 1MB
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate — kritik sozlamalar to'g'riligini tekshiradi (CLAUDE.md §17 —
// production'da xavfsizlik yumshatilgan bo'lsa, server umuman ishga tushmaydi).
func (c *Config) validate() error {
	if c.IsProduction() {
		// #19 — JWT secret kuchli bo'lishi shart (bo'sh yoki qisqa taqiqlanadi).
		if len(c.JWT.AccessSecret) < 32 || len(c.JWT.RefreshSecret) < 32 {
			return fmt.Errorf("config: production da JWT secret kamida 32 belgi bo'lishi kerak")
		}
		if c.JWT.AccessSecret == c.JWT.RefreshSecret {
			return fmt.Errorf("config: access va refresh secret bir xil bo'lishi mumkin emas")
		}
		// #47 — prod'da fake data taqiqlanadi.
		if c.Flags.SeedFakeData {
			return fmt.Errorf("config: production da SEED_FAKE_DATA yoqilishi mumkin emas")
		}
		// #46 — prod'da swagger yopiq.
		if c.Flags.EnableSwagger {
			return fmt.Errorf("config: production da ENABLE_SWAGGER yoqilishi mumkin emas")
		}
		// #34 — DB ulanishi TLS'siz bo'lmasligi kerak.
		if c.DB.SSLMode == "disable" || c.DB.SSLMode == "" {
			return fmt.Errorf("config: production da DB_SSLMODE=disable taqiqlanadi (require/verify-full ishlating)")
		}
		// #15/#16/#40 — prod'da rate limiting majburiy.
		if !c.Security.RateLimitEnabled {
			return fmt.Errorf("config: production da RATE_LIMIT_ENABLED=false bo'lishi mumkin emas")
		}
		// "NAMUNA" muhri — sinov belgisi. Production'da sertifikat muhrsiz
		// chiqqani "NAMUNA" yozuvi bilan chiqqanidan yaxshi.
		if c.Cert.SampleStamp {
			return fmt.Errorf("config: production da CERT_SAMPLE_STAMP yoqilishi mumkin emas")
		}
	}
	return nil
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
