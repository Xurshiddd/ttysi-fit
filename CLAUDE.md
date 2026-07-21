# TTYSI_FIT — Loyiha qoidalari (CLAUDE.md)

## Loyiha haqida

**TTYSI_FIT** — universitet talabalari va professor-o'qituvchilarining jismoniy faolligini raqamlashtiruvchi sport platformasi.

- **Flutter** — mobil ilova (iOS + Android)
- **Go (Golang)** — backend API
- **Nuxt 3** — web admin panel

---

## 1. Arxitektura

```
ttysi fit/
├── flutter apps/          # Mobil ilova
├── go backend/            # REST/gRPC API
│   ├── cmd/               # main.go entry pointlar
│   ├── internal/
│   │   ├── domain/        # Entity, interface, business logic
│   │   ├── repository/    # DB queries (PostgreSQL)
│   │   ├── service/       # Use-case layer
│   │   ├── handler/       # HTTP handlers (gin/chi)
│   │   ├── middleware/     # Auth, logging, rate-limit
│   │   └── dto/           # Request/Response strukturalar
│   ├── migrations/        # SQL migration fayllari
│   ├── config/            # Config loader (.env based)
│   └── pkg/               # Umumiy utility paketlar
└── nuxt admin-panel/      # Nuxt 3 + Tailwind CSS
    ├── pages/
    ├── components/
    ├── composables/
    └── server/
```

### Arxitektura qoidalari

- **Clean Architecture**: domain → repository → service → handler. Pastki qatlam yuqoriga bog'liq bo'lmasligi kerak.
- **Dependency Injection**: interfeys orqali inject qilish, concrete implementation emas.
- **One file, one responsibility**: har bir fayl bitta vazifani bajarishi kerak.
- **Har qanday o'zgarishdan oldin**: mavjud kod strukturasini o'rganib, pattern va konvensiyalarga mos yozish.

---

## 2. Kod yozishdan oldin

**Har doim** kod yozishdan avval:

1. Mavjud shunga o'xshash implementatsiyani qidirish (`grep`).
2. Kamida **2–3 variant** taklif qilish (trade-off bilan).
3. Eng optimal variantni asoslab tanlash.
4. Faqat keyin kod yozish.

```
# Misol taklif formati:
Variant A: [tavsif] — Afzallik: [...] | Kamchilik: [...]
Variant B: [tavsif] — Afzallik: [...] | Kamchilik: [...]
✅ Tavsiya: Variant A — sababi: [...]
```

---

## 3. Backend — Go

### 3.1 N+1 muammosidan himoya

**QOIDA:** `for` ichida hech qachon DB so'rovi bo'lmasligi kerak.

```go
// ❌ XATO — N+1
for _, user := range users {
    db.Where("user_id = ?", user.ID).Find(&activities)
}

// ✅ TO'G'RI — bitta so'rov
db.Where("user_id IN ?", userIDs).Find(&activities)

// ✅ TO'G'RI — JOIN bilan
db.Preload("Activities").Find(&users)
```

- Ko'p bog'liq ma'lumot kerak bo'lsa: `JOIN` yoki `Preload` ishlatish.
- Og'ir aggregatsiya: `SELECT ... GROUP BY` bilan bir so'rovda.
- Paginatsiya: `LIMIT/OFFSET` yoki cursor-based pagination.

### 3.2 Xavfsizlik

```go
// ❌ XATO — SQL injection
db.Raw("SELECT * FROM users WHERE name = '" + name + "'")

// ✅ TO'G'RI — parametr bilan
db.Where("name = ?", name).First(&user)
```

- Barcha input `validate` qilinishi shart (`go-playground/validator`).
- Password: `bcrypt` (cost >= 12).
- JWT: access token (15 min) + refresh token (7 kun), `RS256` algoritm.
- Rate limiting: middleware orqali (IP asosida, `redis` backend).
- CORS: faqat ruxsat etilgan originlar.
- Fayl yuklash: tип, o'lcham va kengaytma tekshiruvi.
- Barcha DB operatsiyalar `context` bilan (`ctx` timeout).

### 3.3 Handler pattern

```go
func (h *UserHandler) GetProfile(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    userID, err := middleware.GetUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, dto.ErrResponse("unauthorized"))
        return
    }

    user, err := h.userService.GetByID(ctx, userID)
    if err != nil {
        // Xatoni log qil, lekin tashqariga chiqarma
        h.log.Error("get profile", zap.Error(err))
        c.JSON(http.StatusInternalServerError, dto.ErrResponse("internal error"))
        return
    }

    c.JSON(http.StatusOK, dto.OK(user))
}
```

### 3.4 Error handling

- Domain-specific xatolar: `errors.go` da constant sifatida saqlash.
- Tashqi API xatolari: wrap qilib qaytarish (`fmt.Errorf("service: %w", err)`).
- HTTP response: ichki xato tafsilotlari mijozga ko'rinmasligi kerak.
- Barcha xatolar `zap` yoki `slog` orqali loglanishi shart.

### 3.5 API dizayn

```
GET    /api/v1/users/:id          # Profil
PUT    /api/v1/users/:id          # Yangilash
GET    /api/v1/users/:id/stats    # Statistika
POST   /api/v1/activities         # Faollik qo'shish
GET    /api/v1/ratings?type=student&faculty_id=1&page=1&limit=20
POST   /api/v1/competitions/:id/register
GET    /api/v1/admin/reports      # Admin only
```

- Versiyalash: `/api/v1/`
- Paginatsiya: `page`, `limit`, `cursor` parametrlari
- Filter: query string orqali
- Response: `{ "data": ..., "meta": { "page": 1, "total": 100 } }`

---

## 4. Database — PostgreSQL

### 4.1 Kengayuvchan jadval yaratish

**QOIDA:** Kelajakda qo'shilishi mumkin bo'lgan ustunlar `NULLABLE` bo'lishi kerak.

```sql
-- ✅ TO'G'RI — kengayuvchan
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,                        -- soft delete

    -- Asosiy ma'lumotlar
    full_name   VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL UNIQUE,
    phone       VARCHAR(20),                         -- nullable
    password    VARCHAR(255) NOT NULL,
    role        VARCHAR(50) NOT NULL DEFAULT 'student',

    -- Universitet ma'lumotlari
    faculty_id  UUID REFERENCES faculties(id),
    department  VARCHAR(255),                        -- nullable
    course      SMALLINT,                            -- nullable
    group_name  VARCHAR(100),                        -- nullable

    -- Kengayish uchun
    avatar_url  TEXT,                               -- nullable
    bio         TEXT,                               -- nullable
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    metadata    JSONB                               -- kelajakdagi qo'shimcha maydonlar
);

-- Indekslar
CREATE INDEX idx_users_faculty_id ON users(faculty_id);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
```

### 4.2 Migratsiya qoidalari

- Har bir migratsiya faylida faqat bitta mantiqiy o'zgarish.
- Fayl nomi: `000001_create_users_table.up.sql` / `.down.sql`
- Ustun qo'shishda: `ALTER TABLE ... ADD COLUMN ... NULL` — hech qachon `NOT NULL DEFAULT` emas (lock oladi).
- Production migratsiya: `golang-migrate` yoki `goose` ishlatish.

### 4.3 Asosiy jadvallar

```
users           — barcha foydalanuvchilar (role bilan ajratiladi)
faculties       — fakultetlar
activities      — kunlik faollik (qadam, kaloriya, masofa)
ratings         — reyting yozuvlari (snapshot asosida)
competitions    — musobaqalar
competition_registrations — ro'yxatdan o'tishlar
challenges      — chellenjlar
user_challenges — foydalanuvchi-chellenj bog'lanishi
fit_coins       — FIT Coin tranzaksiyalari (ledger model)
certificates    — sertifikatlar
notifications   — bildirishnomalar
```

---

## 5. Muhit ajratish (Local vs Production)

### 5.1 .env tuzilmasi

```bash
# .env.local (git'ga kiritilmaydi)
APP_ENV=local
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ttysi_fit_dev
SEED_FAKE_DATA=true          # local da fake data yuklansin
ENABLE_SWAGGER=true
LOG_LEVEL=debug
MOCK_SMS=true                # SMS yuborilmasin, consolega chiqsin
MOCK_EMAIL=true

# .env.production
APP_ENV=production
DB_HOST=prod-db.internal
SEED_FAKE_DATA=false         # Production da fake data YO'Q
ENABLE_SWAGGER=false
LOG_LEVEL=warn
MOCK_SMS=false
MOCK_EMAIL=false
```

### 5.2 Fake data (seeder)

```go
// internal/seeder/seeder.go
func Run(cfg *config.Config, db *gorm.DB) {
    if !cfg.SeedFakeData {
        return  // Production da ishlamaydi
    }
    seedUsers(db)
    seedActivities(db)
    seedCompetitions(db)
}
```

- Seeder faqat `APP_ENV != production` da ishlaydi.
- Fake data belgilangan: `is_test_data = true` ustuni yoki alohida `test_` prefix.
- `make seed` — local seeder ishga tushirish.
- `make migrate-fresh` — local DB ni tozalab qayta yaratish.

### 5.3 Makefile

```makefile
.PHONY: dev seed migrate-fresh test build

dev:
	APP_ENV=local go run cmd/api/main.go

seed:
	APP_ENV=local go run cmd/seed/main.go

migrate-fresh:
	goose down-to 0 && goose up

test:
	go test ./... -v -race

build:
	go build -o bin/api cmd/api/main.go
```

---

## 6. Flutter — Mobil ilova

### 6.1 Papka tuzilmasi

```
lib/
├── core/
│   ├── api/          # Dio client, interceptors
│   ├── auth/         # Token storage (flutter_secure_storage)
│   ├── theme/        # Rang, shrift, spacing
│   └── utils/
├── features/
│   ├── auth/         # Login, register
│   ├── profile/      # Profil, statistika
│   ├── activity/     # Qadam hisoblash, faollik
│   ├── rating/       # Reyting ekranlari
│   ├── competition/  # Musobaqalar
│   ├── challenge/    # Chellenjlar
│   ├── training/     # Video mashqlar
│   └── news/         # Yangiliklar
└── shared/           # Umumiy widget'lar
```

### 6.2 State management

- **Riverpod** (v2) — asosiy state management.
- Har bir feature o'zining `provider.dart` fayliga ega.
- `AsyncValue` bilan loading/error holatlari boshqariladi.

### 6.3 Dizayn qoidalari

- Design system: `lib/core/theme/` da markazlashtirilgan ranglar, fontlar, radiuslar.
- Komponentlar reusable bo'lishi — bitta joyda o'zgartirilsa hamma joyda o'zgaradi.
- Accessibility: minimum touch target 48x48dp.
- Responsive: `LayoutBuilder` / `MediaQuery` orqali tablet moslashuvi.
- Dark mode: `ThemeMode.system` default.

---

## 7. Nuxt 3 — Admin panel

### 7.1 Papka tuzilmasi

```
nuxt admin-panel/
├── pages/
│   ├── dashboard/
│   ├── users/
│   ├── competitions/
│   ├── ratings/
│   └── reports/
├── components/
│   ├── ui/           # Asosiy UI komponentlar
│   ├── charts/       # Grafiklar (Chart.js / ECharts)
│   └── tables/       # Data tablalar
├── composables/
│   ├── useApi.ts     # API calls
│   └── useAuth.ts    # Auth state
├── middleware/
│   └── auth.ts       # Route himoya
└── server/
    └── api/          # BFF agar kerak bo'lsa
```

### 7.2 Dizayn

- **Tailwind CSS** + **shadcn-vue** yoki **PrimeVue**.
- Admin panel dizaynida aniqlik va axborot zichligi asosiy.
- Jadvallar: filter, sort, export (CSV/Excel) funksiyalari.
- Grafiklar: faollik dinamikasi, reyting o'zgarishi.

### 7.3 Responsivlik (MAJBURIY)

**QOIDA:** Har bir sahifa va komponent mobil, planshet va desktopda to'g'ri ko'rinishi shart. Faqat desktopda tekshirib "tayyor" deb hisoblash taqiqlanadi.

Tailwind breakpoint'lar (mobile-first — avval mobil, keyin kattaroq ekran):

| Prefix | Min kenglik | Maqsad |
|--------|-------------|--------|
| (default) | 0px | Telefon |
| `sm:` | 640px | Katta telefon |
| `md:` | 768px | Planshet |
| `lg:` | 1024px | Laptop |
| `xl:` | 1280px | Desktop |

```vue
<!-- ✅ TO'G'RI — mobile-first grid -->
<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">...</div>

<!-- ✅ TO'G'RI — jadval mobilda gorizontal scroll -->
<div class="overflow-x-auto">
  <table class="w-full min-w-[640px]">...</table>
</div>

<!-- ❌ XATO — qattiq kenglik, mobilda buziladi -->
<div class="w-[1200px]">...</div>
```

- **Sidebar**: `lg` dan past ekranda yashirin, hamburger tugma + overlay bilan ochiladi.
- **Jadvallar**: keng jadval `overflow-x-auto` ichida; ustun ko'p bo'lsa mobilda eng muhim ustunlarni qoldirish (`hidden sm:table-cell`).
- **Qattiq piksel kenglik** (`w-[1200px]`) o'rniga `max-w-*` + `w-full`.
- **Touch target**: tugma/havola minimum 40×40px (mobil barmoq uchun).
- **Matn**: `text-sm` mobilda, `md:text-base` kattaroq ekranda — kerak bo'lsa.
- Tekshirish: DevTools'da kamida 375px (telefon), 768px (planshet), 1280px (desktop) kengliklarda ko'rish.

---

## 8. Dizayn yondashuv

### 8.1 Mobil (Flutter)

- **Ranglar**: asosiy — `#1E3A5F` (chuqur ko'k) + `#00C896` (yashil accent).
- **Font**: `Inter` yoki `Nunito` — sport ilovasiga mos zamonaviy.
- **Animatsiya**: `flutter_animate` — 200–300ms, natural easing.
- **Ikonlar**: `Phosphor Icons` yoki `Lucide`.
- **Komponent uslubi**: card-based layout, rounded corners (12–16px), subtle shadow.

### 8.2 Admin panel (Nuxt)

- Minimal, professional — foydalanuvchi ma'lumot topishiga to'sqinlik qilmaydigan.
- Sidebar navigatsiya, topbar da foydalanuvchi info.
- Jadvalda virtual scroll (ko'p qator uchun).

---

## 9. Hisobot formati

Har bir ish yakunlangach **SHUNAQA** hisobot beriladi:

```
## Bajarildi ✅
- [nima qilindi, qaysi fayl]

## Qanday ishlaydi
- [qisqacha texnik tushuntirish]

## Keyin qilinishi kerak 🔜
- [ ] [keyingi qadam 1]
- [ ] [keyingi qadam 2]

## E'tibor berish kerak ⚠️
- [muhim nuance yoki xavf]
```

---

## 10. Umumiy taqiqlar

| Taqiq | Sabab |
|-------|-------|
| `for` ichida DB so'rovi | N+1 muammosi |
| Raw SQL string concatenation | SQL injection |
| `panic` ishlatish (handler'da) | Serverning tushib qolishi |
| `interface{}` / `any` haddan tashqari ishlatish | Type safety yo'qoladi |
| Secret'larni koda yozish | Xavfsizlik muammosi |
| Migration'da `NOT NULL` ustun qo'shish (default'siz) | Production lock |
| `TODO` yozib qoldirish | Texnik qarz to'planadi |
| Test'siz kritik funksiya | Regress xatolar |

---

## 11. Goroutine va WaitGroup qoidalari

### 11.1 Goroutine xavfsizligi

**QOIDA:** Har bir goroutine panic'dan himoyalangan bo'lishi kerak — bitta goroutine serverni tushirmasligi kerak.

```go
// ❌ XATO — panic butun serverni o'ldiradi
go func() {
    riskyOperation()
}()

// ✅ TO'G'RI — recover bilan
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Error("goroutine panic", zap.Any("recover", r))
        }
    }()
    riskyOperation()
}()
```

### 11.2 WaitGroup pattern

```go
// ✅ TO'G'RI — WaitGroup + errgroup bilan parallel so'rovlar
func (s *RatingService) GetFullRating(ctx context.Context, facultyID uuid.UUID) (*RatingResult, error) {
    g, ctx := errgroup.WithContext(ctx)

    var students []User
    var teachers []User

    g.Go(func() error {
        var err error
        students, err = s.repo.GetStudentsByFaculty(ctx, facultyID)
        return err
    })

    g.Go(func() error {
        var err error
        teachers, err = s.repo.GetTeachersByFaculty(ctx, facultyID)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, fmt.Errorf("GetFullRating: %w", err)
    }

    return &RatingResult{Students: students, Teachers: teachers}, nil
}
```

- `errgroup.WithContext` ishlatish — bitta goroutine xato qilsa, qolganlar bekor qilinadi.
- `sync.WaitGroup` faqat xatosiz parallel ishlar uchun (fan-out/fan-in).
- Goroutine ichida **hech qachon** tashqi o'zgaruvchiga `mutex` siz yozma.

### 11.3 Race condition oldini olish

```go
// ❌ XATO — race condition
var count int
for i := 0; i < 10; i++ {
    go func() { count++ }()
}

// ✅ TO'G'RI — atomic yoki mutex
var count atomic.Int64
for i := 0; i < 10; i++ {
    go func() { count.Add(1) }()
}

// ✅ TO'G'RI — mutex bilan map
type SafeCache struct {
    mu   sync.RWMutex
    data map[string]any
}
func (c *SafeCache) Set(k string, v any) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[k] = v
}
```

- `go test -race` har doim ishga tushirilsin.
- Shared state uchun `sync.RWMutex` (read-heavy) yoki `sync/atomic`.
- Channel orqali kommunikatsiya — to'g'ridan-to'g'ri o'zgaruvchi ulashish emas.

### 11.4 Context va goroutine lifecycle

```go
// ✅ TO'G'RI — context bekor qilinganda goroutine to'xtaydi
func (s *ActivityService) StartStepSync(ctx context.Context, userID uuid.UUID) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return  // context bekor — goroutine chiqadi
        case <-ticker.C:
            s.syncSteps(ctx, userID)
        }
    }
}
```

- Background goroutine'lar doim `ctx.Done()` ni tinglashi kerak.
- Server shutdown'da `graceful shutdown` — `os.Signal` + context cancel.

---

## 12. Redis qoidalari

### 12.1 Qachon Redis ishlatish

| Holat | Redis | PostgreSQL |
|-------|-------|------------|
| JWT refresh token | ✅ | ❌ |
| Rate limiting counter | ✅ | ❌ |
| Session cache | ✅ | ❌ |
| Reyting leaderboard (real-time) | ✅ | Snapshot uchun ✅ |
| OTP kodi (5 min TTL) | ✅ | ❌ |
| Faollik statistika cache | ✅ (TTL 5 min) | Source of truth ✅ |
| Foydalanuvchi profil | Cache uchun ✅ | Source of truth ✅ |

### 12.2 Redis pattern

```go
// ✅ Cache-aside pattern
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", id)

    // 1. Cache dan o'qish
    cached, err := s.redis.Get(ctx, cacheKey).Bytes()
    if err == nil {
        var user User
        if err := json.Unmarshal(cached, &user); err == nil {
            return &user, nil
        }
    }

    // 2. DB dan o'qish
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 3. Cache ga yozish (TTL bilan)
    data, _ := json.Marshal(user)
    s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

    return user, nil
}

// ✅ Cache invalidation — user yangilanganda
func (s *UserService) Update(ctx context.Context, user *User) error {
    if err := s.repo.Update(ctx, user); err != nil {
        return err
    }
    cacheKey := fmt.Sprintf("user:%s", user.ID)
    s.redis.Del(ctx, cacheKey)
    return nil
}
```

### 12.3 Redis key naming convention

```
user:{uuid}                    — foydalanuvchi profil cache
rating:student:{faculty_id}    — talabalar reytingi cache
rating:teacher:{faculty_id}    — o'qituvchilar reytingi cache
otp:{phone}                    — OTP kodi (TTL: 5 min)
refresh:{user_id}              — refresh token (TTL: 7 kun)
ratelimit:{ip}:{endpoint}      — rate limit counter (TTL: 1 min)
fitcoin:balance:{user_id}      — FIT Coin balans cache
leaderboard:global             — umumiy reyting (Sorted Set)
```

### 12.4 Redis Sorted Set — leaderboard

```go
// ✅ Real-time leaderboard uchun Sorted Set
func (s *RatingService) UpdateScore(ctx context.Context, userID uuid.UUID, score float64) error {
    return s.redis.ZAdd(ctx, "leaderboard:global", &redis.Z{
        Score:  score,
        Member: userID.String(),
    }).Err()
}

func (s *RatingService) GetTopN(ctx context.Context, n int64) ([]redis.Z, error) {
    return s.redis.ZRevRangeWithScores(ctx, "leaderboard:global", 0, n-1).Result()
}
```

### 12.5 Redis xavfsizligi

- Har doim `context` bilan (`redis.Client` ning barcha metodlari).
- Connection pool: `PoolSize: 10`, `MinIdleConns: 5`.
- Redis xatosi — **cache miss** sifatida qaralsin, server xatosi sifatida emas.
- Sensitive ma'lumot (password hash) Redis ga **HECH QACHON** yozilmasin.

---

## 13. PostgreSQL qoidalari

### 13.1 Connection pool sozlash

```go
// ✅ TO'G'RI — production uchun pool
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
sqlDB.SetConnMaxIdleTime(1 * time.Minute)
```

### 13.2 Transaction pattern

```go
// ✅ TO'G'RI — transaction bilan (FIT Coin transfer)
func (s *FitCoinService) Transfer(ctx context.Context, from, to uuid.UUID, amount int) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 1. Balansni tekshirish
        var fromBalance int
        if err := tx.Raw("SELECT balance FROM fit_coins WHERE user_id = ? FOR UPDATE", from).
            Scan(&fromBalance).Error; err != nil {
            return err
        }
        if fromBalance < amount {
            return ErrInsufficientBalance
        }

        // 2. Chiqarish
        if err := tx.Exec("UPDATE fit_coins SET balance = balance - ? WHERE user_id = ?",
            amount, from).Error; err != nil {
            return err
        }

        // 3. Kiritish
        if err := tx.Exec("UPDATE fit_coins SET balance = balance + ? WHERE user_id = ?",
            amount, to).Error; err != nil {
            return err
        }

        // 4. Ledger yozuvi
        return tx.Create(&FitCoinLedger{From: from, To: to, Amount: amount}).Error
    })
}
```

### 13.3 Index qoidalari

```sql
-- Tez-tez filter qilinadigan ustunlarga index
CREATE INDEX idx_activities_user_date ON activities(user_id, created_at DESC);
CREATE INDEX idx_ratings_faculty_score ON ratings(faculty_id, score DESC);
CREATE INDEX idx_competitions_status ON competitions(status) WHERE deleted_at IS NULL;

-- Partial index — faqat aktiv yozuvlar
CREATE INDEX idx_users_active ON users(role, faculty_id) WHERE deleted_at IS NULL AND is_active = TRUE;
```

### 13.4 Og'ir query'larni optimallashtirish

```go
// ❌ XATO — hamma ustunlarni yuklash
db.Find(&users)

// ✅ TO'G'RI — faqat kerakli ustunlar
db.Select("id, full_name, avatar_url, faculty_id").Find(&users)

// ✅ TO'G'RI — reyting hisoblash (bitta so'rov)
db.Raw(`
    SELECT u.id, u.full_name, u.faculty_id,
           COALESCE(SUM(a.steps), 0) as total_steps,
           RANK() OVER (PARTITION BY u.faculty_id ORDER BY SUM(a.steps) DESC) as rank
    FROM users u
    LEFT JOIN activities a ON a.user_id = u.id
        AND a.created_at >= NOW() - INTERVAL '30 days'
    WHERE u.role = 'student' AND u.deleted_at IS NULL
    GROUP BY u.id, u.full_name, u.faculty_id
`, facultyID).Scan(&ratings)
```

---

## 14. Paginatsiya qoidalari

**ASOSIY QOIDA — chegara 1000 yozuv:**

| Ma'lumotlar soni | Yondashuv | Sabab |
|------------------|-----------|-------|
| **≤ 1000 yozuv** | **Client-side** paginatsiya | Hammasi bir so'rovda yuklanadi, frontend sahifalarga bo'ladi. Filter/sort tez (server'siz). Kam so'rov. |
| **> 1000 yozuv** | **Server-side** paginatsiya | `?page&limit` bilan har sahifa alohida. Brauzer xotirasi va trafik tejaladi. |

### 14.1 Client-side (≤ 1000)

- Backend hamma yozuvni bitta `GET` da qaytaradi (masalan `/groups`, `/faculties`).
- Frontend `computed` bilan kerakli bo'lakni kesadi:

```ts
const page = ref(1)
const perPage = 20
const pageCount = computed(() => Math.max(1, Math.ceil(filtered.value.length / perPage)))
const paged = computed(() =>
  filtered.value.slice((page.value - 1) * perPage, page.value * perPage)
)
```

- Qidiruv/filter ham client-side (`computed`).
- **Eslatma:** ro'yxat 1000 dan oshib ketsa — server-side'ga o'tish SHART (yuqoridagi jadval).

### 14.2 Server-side (> 1000)

- Backend `?page` va `?limit` ni qabul qiladi, `meta.total` qaytaradi:
  - Response: `{ "data": [...], "meta": { "page": 1, "limit": 20, "total": 5000 } }`
- SQL: `LIMIT/OFFSET` yoki katta jadvalda **cursor-based** (keyset) pagination (OFFSET sekin bo'lib qoladi).
- `COUNT(*)` og'ir bo'lsa: alohida cache yoki taxminiy hisob.
- Frontend har sahifa o'zgarganda yangi so'rov yuboradi (`watch(page, load)`), `meta.total` dan `pageCount` hisoblaydi. Namuna: `pages/users.vue`.

```go
// ✅ Server-side — repository
func (r *groupRepo) List(ctx context.Context, f Filter) ([]Group, int64, error) {
    var total int64
    q := r.db.WithContext(ctx).Model(&Group{}).Where("deleted_at IS NULL")
    if err := q.Count(&total).Error; err != nil { return nil, 0, err }
    var rows []Group
    err := q.Order("name").Limit(f.Limit).Offset((f.Page - 1) * f.Limit).Find(&rows).Error
    return rows, total, err
}
```

- **Default limit**: 20, **max limit**: 100 (mijoz 100 dan oshiq so'rasa cheklab qo'yish).

---

## 15. Testlash

```go
// Har bir service metodi uchun unit test
func TestUserService_GetByID(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewUserRepository(t)
    svc := service.NewUserService(mockRepo)
    
    mockRepo.On("GetByID", ctx, userID).Return(fakeUser, nil)
    
    // Act
    result, err := svc.GetByID(ctx, userID)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, fakeUser.ID, result.ID)
}
```

- Unit test: service layer (mock repository).
- Integration test: `testcontainers-go` — haqiqiy PostgreSQL container.
- `go test -race` — race condition tekshiruvi.
- Coverage: asosiy service'lar uchun >= 70%.

---

## 16. Dinamik kontent va moslashuvchan sxema (challenge / musobaqa / yutuq)

**ASOSIY QOIDA:** Chellenj, musobaqa, yutuq/sertifikat, mashg'ulot, yangilik va shunga o'xshash kontent **kodga hardcode qilinmaydi** — hammasi **admin panel orqali** yaratiladi, tahrirlanadi, o'chiriladi (DB'da saqlanadi). Yangi chellenj turi yoki musobaqa qo'shish uchun kod o'zgartirish/redeploy SHART EMAS bo'lishi kerak.

```go
// ❌ XATO — kontent kodda
var challenges = []Challenge{{Name: "10 000 qadam", Target: 10000}}

// ✅ TO'G'RI — admin panel CRUD qiladi, DB dan o'qiladi
challenges, _ := repo.ListActiveChallenges(ctx)
```

### 16.1 Har xil tur — har xil maydonlar (moslashuvchan sxema)

Chellenjlar (va musobaqalar) **har xil turda** bo'ladi va har bir turning **parametrlari har xil**:

| Tur (`type`) | O'ziga xos maydonlar (config) |
|--------------|-------------------------------|
| `steps` (10 000 qadam) | `target_steps` |
| `distance` (yugurish marafoni) | `target_km`, `period` |
| `faculty_vs` (fakultetlararo) | `metric`, `faculty_ids` |
| `group_vs` (guruhlararo) | `metric`, `group_ids` |
| `custom` (boshqa aksiya) | ixtiyoriy kalitlar |

**QOIDA:** Umumiy maydonlar (id, title, description, type, scope, start/end, status, reward) — **ustun**; turga xos o'zgaruvchan maydonlar — **`config JSONB`** ustunida saqlanadi (CLAUDE.md §4.1 — `metadata JSONB` g'oyasi). Bu yangi tur qo'shishda migration shart qilmaydi.

```sql
-- ✅ TO'G'RI — kengayuvchan, turga xos maydonlar JSONB da
CREATE TABLE challenges (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,

    type        VARCHAR(50) NOT NULL,        -- steps | distance | faculty_vs | ...
    title       VARCHAR(255) NOT NULL,
    description TEXT,
    scope       VARCHAR(50) NOT NULL DEFAULT 'university', -- university|faculty|group
    starts_at   TIMESTAMPTZ,
    ends_at     TIMESTAMPTZ,
    status      VARCHAR(20) NOT NULL DEFAULT 'draft',      -- draft|active|finished
    reward_coins INTEGER NOT NULL DEFAULT 0,
    config      JSONB NOT NULL DEFAULT '{}'   -- turga xos: {"target_steps":10000} ...
);
CREATE INDEX idx_challenges_active ON challenges(status, ends_at) WHERE deleted_at IS NULL;
```

```go
// Go tomonda: umumiy maydonlar struct, turga xos qism map/JSON
type Challenge struct {
    ID     uuid.UUID
    Type   string
    Title  string
    Status string
    Config datatypes.JSON  // gorm.io/datatypes — JSONB
    // ...
}
```

### 16.2 Tur bo'yicha validatsiya va forma

- Backend `config` ni **turga qarab validatsiya** qiladi (masalan `type=steps` bo'lsa `target_steps` > 0 shart). Har tur uchun validator funksiyasi (registry/`switch type`).
- Admin panelda **forma dinamik**: avval `type` tanlanadi → o'sha turga mos maydonlar ko'rsatiladi. Maydon ta'riflari (schema) frontendda `type → fields` xaritasi yoki backenddan `GET /admin/challenge-types` orqali olinadi (yangi tur qo'shilsa frontend ham moslashadi).
- Progress/natija hisoblash ham turga bog'liq (`steps` → qadam yig'indisi, `distance` → masofa) — strategiya/`switch type` bilan, alohida-alohida funksiya.

### 16.3 Shu qoida tegishli bo'lgan modullar

Bir xil yondashuv (umumiy ustunlar + `JSONB config`/`metadata`, admin CRUD, turga qarab validatsiya) quyidagilarga ham qo'llanadi:

- **Musobaqalar** (`competitions`) — tur, bosqichlar, natija formati har xil
- **Yutuq/Sertifikat** (`achievements`, `certificates`) — shablon + berish mezoni (`criteria JSONB`)
- **Rag'batlantirish qoidalari** (FIT Coin qancha beriladi) — admin sozlaydi, kodga yozilmaydi
- **Mashg'ulotlar** (`trainings`) va **Yangiliklar** (`news`) — admin kontent kiritadi

**Xulosa:** kontent ham, uning tuzilishi (turlar/maydonlar) ham imkon qadar **ma'lumot** (data) bo'lsin, **kod** emas. Yangi tur/qoida qo'shish — admin amali, dasturchi ishi emas.

---

## 17. Kiberxavfsizlik (IMPORTANT)

> **IMPORTANT — eng yuqori ustuvorlik.** Tizim eng keng tarqalgan **TOP-50 hujum**ga dosh bera olishi SHART. Xavfsizlik standart holatda **qattiq (strict)** bo'ladi. Yumshatish (mitigation/relaxation) **faqat lokal tekshiruv uchun** va **faqat** `.env` da `APP_ENV=local` bo'lganda ruxsat etiladi. `APP_ENV` boshqa har qanday qiymat (`production`, `staging`, ...) bo'lsa — to'liq qattiq xavfsizlik.

```go
// ✅ Xavfsizlikni APP_ENV bo'yicha gate qilish
if cfg.IsProduction() { /* strict: TLS, CSP, rate limit, swagger off ... */ }
// Yumshatish faqat: cfg.App.Env == "local"
```

### 17.1 APP_ENV=local da YUMSHATISH MUMKIN (faqat qulaylik uchun)

Bular faqat ishlab chiqish qulayligi — **xavfsizlik yadrosi emas**:

- CORS: `http://localhost:*` ga ruxsat (prod'da faqat aniq originlar)
- Rate limiting: o'chirilgan yoki yumshoq (prod'da yoqilgan)
- `MOCK_SMS=true`, `MOCK_EMAIL=true` (haqiqiy yuborilmaydi)
- `ENABLE_SWAGGER=true` (prod'da `false`)
- TLS/HTTPS ixtiyoriy (prod'da MAJBURIY)
- Batafsil (debug) loglar va xato tafsilotlari (prod'da yashirin)
- `SEED_FAKE_DATA=true` (prod'da TAQIQ — config validate bilan)

### 17.2 HECH QACHON yumshatilmaydi (local'da ham qattiq)

- Parametrli SQL (injection himoyasi), input validatsiya
- Parol hash (`bcrypt` cost ≥ 12), JWT imzo tekshiruvi
- Access control (IDOR/role tekshiruvi)
- Secret'larни koda yozmaslik, token/parolни loglamaslik
- Fayl yuklash tekshiruvi (tип/o'lcham/kengaytma)

### 17.3 TOP-50 hujum va himoya (checklist)

**Injection & ma'lumot (1–14)**
1. SQL injection → parametrli query (§3.2), hech qachon string concat
2. NoSQL/ORM injection → validatsiya, tип-xavfsiz so'rov
3. OS command injection → user input'ni `exec`ga bermaslik
4. LDAP/SMTP injection → escaping/validatsiya
5. XSS (stored/reflected/DOM) → output escaping, Vue auto-escape, CSP
6. CSRF → `SameSite` cookie, OAuth `state` (§ HEMIS), token
7. SSRF → tashqi URL allowlist (faqat HEMIS), user URL'ni fetch qilmaslik
8. XXE → XML parserda external entity o'chirilgan
9. Open redirect → redirect URL'larni allowlist (deep link ham)
10. Path traversal → fayl nomi sanitize (media downloader)
11. HTTP header / CRLF injection → header'ga user input tozalab
12. Server-side template injection (SSTI)
13. Mass assignment / over-posting → DTO whitelisting (entity'ni to'g'ridan bind qilmaslik)
14. Insecure deserialization → ishonchsiz payload'ni deserialize qilmaslik

**Autentifikatsiya & sessiya (15–25)**
15. Brute force login → rate limit + (kerakda) lockout
16. Credential stuffing → rate limit, IP nazorati
17. Zaif parol → siyosat (min 8), `bcrypt` cost ≥ 12
18. JWT `alg=none` / alg confusion → qat'iy algoritm, imzo majburiy
19. JWT secret zaifligi → kuchli secret (.env), prod'da bo'sh bo'lmasligi (validate)
20. Token URL/log orqali sizishi → token URL'ga qo'yilmaydi (OAuth bir martalik code), loglanmaydi
21. Session fixation → loginда sessiya/JTI yangilanadi
22. Refresh token replay → rotatsiya + Redis revoke
23. OAuth CSRF → `state` tekshiruvi (qilingan)
24. OAuth `redirect_uri` manipulyatsiyasi → faqat ro'yxatdan o'tgan aniq URI
25. Sessiya muddati yo'qligi → access 15 min, refresh 7 kun TTL + revoke

**Ruxsat (access control) (26–30)**
26. IDOR / BOLA → har resursda egalik tekshiruvi (foydalanuvchi faqat o'zinikiga)
27. Privilege escalation → `RequireRole` middleware
28. Broken function-level auth (BFLA) → barcha admin endpoint himoyalangan
29. Forced browsing → barcha himoyalangan route'da `Auth`
30. CORS noto'g'ri sozlanishi → faqat allowlist origin (prod)

**Kripto & maxfiylik (31–35)**
31. Ochiq parol → `bcrypt`
32. Maxfiy ma'lumot oshkoraligi → TLS, hash, Redis'ga secret yozmaslik
33. Hardcoded secret → `.env`, `.gitignore`
34. Zaif TLS / HTTPS yo'qligi → prod'da TLS majburiy, HSTS
35. Xavfsiz cookie → `Secure`, `HttpOnly`, `SameSite`

**Input & API (36–44)**
36. Validatsiya yo'qligi → `go-playground/validator`
37. Excessive data exposure → DTO/read-model (`password json:"-"`)
38. Mass scraping → rate limit, paginatsiya cheklovi (max limit 100)
39. Resurs sarfi (DoS) → body o'lcham limiti, query timeout (`ctx`), paginatsiya
40. Rate limit bypass → IP+endpoint, Redis backend
41. Fayl yuklash suiiste'moli → tип/o'lcham/kengaytma tekshiruvi
42. MIME sniffing → `X-Content-Type-Options: nosniff`
43. Clickjacking → `X-Frame-Options` / CSP `frame-ancestors`
44. Xavfsizlik header'lari yo'qligi → HSTS, CSP, Referrer-Policy

**Infra & operatsiya (45–50)**
45. Batafsil xato/stack trace sizishi → ichki tafsilot mijozga chiqmaydi (§3.4)
46. Prod'da debug/swagger ochiq → `APP_ENV` gating (`ENABLE_SWAGGER=false`)
47. Prod'da seed/fake data → `SEED_FAKE_DATA=false` (config validate)
48. Bog'liqlik (dependency) zaifliklari → `govulncheck`, `npm audit` muntazam
49. Maxfiy ma'lumotni loglash → token/parol/PII loglanmaydi
50. Yetarli monitoring yo'qligi → muhim hodisalar (login, sync, xato) audit loglanadi

### 17.4 Tekshirish

- Har bir yangi endpoint qo'shilganda: auth, access control, validatsiya, rate limit ko'rib chiqilsin.
- `go test -race`, `govulncheck ./...`, `npm audit` — CI da.
- Yumshatish kodi doim `if cfg.App.Env == "local"` ichida; default — qattiq.
