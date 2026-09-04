# TTYSI_FIT — Backend (Go)

Clean Architecture asosidagi REST API. PostgreSQL (lokal) + Redis (Docker).

## Talablar

- Go 1.24+
- Docker (PostgreSQL + Redis uchun)
- goose (migratsiya): `go install github.com/pressly/goose/v3/cmd/goose@latest`

## Ishga tushirish (lokal)

```bash
# 1. .env.local yaratish (.env.example dan nusxa) va DB_PASSWORD ni to'ldirish
#    MUHIM: config APP_ENV bo'yicha .env.local ni o'qiydi — oddiy .env emas.
cp .env.example .env.local

# 2. Postgres + Redis ni Docker'da ko'tarish (baza avtomatik yaratiladi)
make db-up

# 3. Migratsiya
make migrate-up

# 4. Serverni ishga tushirish
make dev
```

Portlar: Postgres **5433**, Redis **6380** (host'dagi 5432/6379 band bo'lgani uchun).
Compose o'zgaruvchilari `.env.local` dan olinadi (`--env-file .env.local`).

Server: http://localhost:8090

## Tekshirish

```bash
curl http://localhost:8090/api/v1/health
# {"data":{"status":"ok","postgres":"up","redis":"up"}}
```

## Auth endpointlari

```
POST /api/v1/auth/register   # ro'yxatdan o'tish
POST /api/v1/auth/login      # kirish
POST /api/v1/auth/refresh    # tokenni yangilash
POST /api/v1/auth/logout     # chiqish (Bearer token kerak)
```

Sinov (Git Bash):

```bash
# Ro'yxatdan o'tish
curl -X POST http://localhost:8090/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Xurshid Marifov","email":"x@test.uz","password":"parol1234","role":"student"}'

# Kirish
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"x@test.uz","password":"parol1234"}'

# Logout (access_token ni qo'ying)
curl -X POST http://localhost:8090/api/v1/auth/logout \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

JWT: HS256, access 15 daqiqa, refresh 7 kun (Redis da `refresh:{user_id}` JTI bilan saqlanadi, rotatsiya + revoke qo'llab-quvvatlanadi).

## Ko'p tillilik (i18n)

3 til: **uz** (default), **ru**, **en**. Javob xabarlari foydalanuvchi tiliga moslashadi — kod ichida matn qattiq yozilmaydi, faqat xabar kodi (`internal/i18n/messages.go`).

Til aniqlash ustuvorligi: `?lang=ru` query > `Accept-Language` header > default `uz`. Foydalanuvchining DB dagi `language` ustuni server-initiated xabarlar (email/SMS/push) uchun ishlatiladi.

Validatsiya xatolari ham 3 tilga tarjima qilinadi — maydon bo'yicha:

```bash
curl -X POST "http://localhost:8090/api/v1/auth/register?lang=ru" \
  -H "Content-Type: application/json" -d '{"email":"notanemail","password":"123"}'
# {
#   "error": "Переданные данные некорректны",
#   "fields": {
#     "full_name": "Поле «ФИО» обязательно",
#     "email": "Email должен быть корректным email",
#     "password": "Пароль должен содержать не менее 8 символов"
#   }
# }
```

```bash
# Rus tilida xato javobi
curl -X POST "http://localhost:8090/api/v1/auth/login?lang=ru" \
  -H "Content-Type: application/json" \
  -d '{"email":"x@test.uz","password":"xato"}'
# {"error":"Неверный логин или пароль"}

# Yoki header orqali
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Accept-Language: en" -H "Content-Type: application/json" \
  -d '{"email":"x@test.uz","password":"xato"}'
# {"error":"Invalid login or password"}
```

## HEMIS integratsiyasi

Tashkiliy birliklar (fakultet, kafedra, bo'lim) HEMIS'dan sinxronlanadi va yagona `structures` jadvalida saqlanadi (daraxt, `parent_id` self-reference). `.env` da faqat asosiy URL + token:

```bash
HEMIS_BASE_URL=https://student.ttyesi.uz/rest/v1
HEMIS_TOKEN=<sizning_tokeningiz>
```

Qolgan sozlamalar (`HEMIS_STRUCTURE_PATH`, `HEMIS_FACULTY_TYPE_CODE`, `HEMIS_PAGE_LIMIT`, `HEMIS_RATE_LIMIT`) kod default'lariga ega — kerak bo'lsagina `.env` da override qilinadi.

**Rate limiting:** HEMIS sekundiga 10 ta so'rovdan oshganda bloklaydi. Shuning uchun client har bir so'rovni `rate.Limiter` (default 10 req/sek, teng oraliqda) orqali o'tkazadi — limitdan oshmaydi.

Endpointlar:

```
GET  /api/v1/faculties                       # fakultetlar (type 11)
GET  /api/v1/departments?faculty_id=<uuid>   # kafedralar (type 12)
GET  /api/v1/groups?faculty_id=<uuid>        # guruhlar
POST /api/v1/admin/hemis/sync/structures     # admin only
POST /api/v1/admin/hemis/sync/groups         # admin only
POST /api/v1/admin/hemis/sync/students       # admin only
POST /api/v1/admin/hemis/sync/employees      # admin only
```

structureType kodlari (TTYESI): `11`=Fakultet, `12`=Kafedra, `13`=Bo'lim, `15`=Markaz, `16`=Rektorat.

CLI sync (admin shart emas, bootstrap uchun):

```bash
./run.sh sync              # hammasi: structures → groups → students → employees
./run.sh sync structures   # alohida bosqich
./run.sh sync students
```

Talaba/o'qituvchi `hemis_id` bo'yicha upsert qilinadi. Talaba: `faculty_id` + `group_id` + `course`; o'qituvchi: `department_id` (kafedra) + `position`. Sinxron foydalanuvchilarda parol yo'q (`hemis_login` = student/employee id), login usuli (HEMIS hisobi/OTP) keyin qo'shiladi.

Sync (admin token bilan):

```bash
curl -X POST http://localhost:8090/api/v1/admin/hemis/sync/structures \
  -H "Authorization: Bearer <ADMIN_ACCESS_TOKEN>"
# {"data":{"total":120,"created":120,"updated":0},"message":"Sinxronizatsiya muvaffaqiyatli yakunlandi"}
```

Sync `hemis_id` bo'yicha **upsert** qiladi (qayta ishga tushirsa dublikat bo'lmaydi), so'ng `parent_hemis_id` → `parent_id` ni bitta SQL bilan bog'laydi. To'liq HEMIS javobi `raw JSONB` da saqlanadi (forward-compat). Kelajakda admin panel shu endpointni chaqiradi.

> Eslatma: real HEMIS endpoint yo'li va `FacultyTypeCode` tashkilotingizga qarab farq qilishi mumkin — `.env` da moslang.

## Struktura

```
go backend/
├── cmd/api/            # main.go — kirish nuqtasi
├── config/             # .env loader
├── internal/
│   ├── domain/         # Entity + repository interfeyslari
│   ├── repository/     # gorm implementatsiyalar
│   ├── handler/        # HTTP handlerlar
│   └── dto/            # Request/Response strukturalar
├── migrations/         # goose SQL migratsiyalar
├── pkg/
│   ├── database/       # Postgres + Redis ulanish
│   └── logger/         # zap logger
├── docker-compose.yml  # Redis
└── Makefile
```
