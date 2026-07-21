# TTYSI_FIT — Xavfsizlik auditi

**Sana:** 2026-07-02
**Qamrov:** `go backend/` (Gin API, PostgreSQL, Redis, JWT, HEMIS OAuth, media) + `nuxt-admin-panel/`
**Metodika:** Anthropic-Cybersecurity-Skills (`security-skills/`) yo'riqnomalari + CLAUDE.md §17 TOP-50 checklist.

## Umumiy xulosa

Backend xavfsizlik yadrosi **kuchli**: SQL to'liq parametrli, JWT alg-confusion himoyalangan, bcrypt cost 12, refresh rotatsiya + Redis revoke, OAuth `state` CSRF, access control (IDOR yo'q). Kritik (injection/authn) darajada zaiflik topilmadi.

Asosiy kamchiliklar — **perimetr himoyasi**: inbound rate limiting, xavfsizlik header'lari, SSRF va DoS himoyasi hozircha yo'q. Bular CLAUDE.md §17 talab qilgan, lekin hali yozilmagan qismlar.

| Daraja | Soni |
|--------|------|
| 🔴 High | 1 |
| 🟠 Medium | 5 |
| 🟡 Low / Info | 3 |
| ✅ Tasdiqlangan (yaxshi) | 13+ |

---

## 🔴 HIGH

### H1 — Inbound rate limiting yo'q (brute-force / credential stuffing)
**Fayl:** `cmd/api/main.go` (middleware zanjirida yo'q)
`/auth/login`, `/auth/register`, `/auth/refresh` cheklovsiz. Hujumchi parolni cheksiz sinab ko'radi. CLAUDE.md §17.3 #15/#16/#40 buni majburiy qiladi; Redis allaqachon ulangan, lekin faqat HEMIS **outbound** limiter bor (`pkg/hemis/client.go`), inbound yo'q.
**Yechim:** `ratelimit:{ip}:{endpoint}` kaliti bilan Redis-backed middleware (masalan `INCR`+`EXPIRE` 1 min). `security-skills/skills/implementing-api-rate-limiting-and-throttling/`. `local` da yumshoq, prod'da qattiq (§17.1).

---

## 🟠 MEDIUM

### M1 — Xavfsizlik header'lari yo'q
**Fayl:** `cmd/api/main.go`
`HSTS`, `Content-Security-Policy`, `X-Frame-Options`, `X-Content-Type-Options: nosniff`, `Referrer-Policy` — hech biri o'rnatilmagan (§17.3 #42 clickjacking, #43 MIME sniffing, #44). Ayniqsa `r.Static` orqali xizmat qilinadigan avatarlar uchun `nosniff` muhim.
**Yechim:** `SecurityHeaders()` middleware qo'shish; prod'da HSTS + CSP.

### M2 — SSRF: media downloader URL'ni cheklamaydi
**Fayl:** `pkg/media/avatar.go` → `Save()`
`srcURL` (HEMIS profil rasmi) to'g'ridan-to'g'ri `http.Get` qilinadi. Domen allowlist yoki private-IP/localhost/metadata (`169.254.169.254`) bloklash yo'q. HEMIS ma'lumoti buzilsa yoki ichki URL qaytsa — server ichki tarmoqqa so'rov yuboradi (§17.3 #7).
**Yechim:** faqat HEMIS domenlariga allowlist; DNS resolve qilib private/loopback IP'larni rad etish. `security-skills/skills/exploiting-oauth-misconfiguration/` va SSRF pattern'lari.

### M3 — Request body hajm limiti yo'q (DoS)
**Fayl:** barcha handler'lar — `c.ShouldBindJSON`
Kiruvchi body cheklanmagan; katta payload xotira/CPU yeydi (§17.3 #39).
**Yechim:** global `http.MaxBytesReader` yoki `r.MaxMultipartMemory` + JSON uchun body limit middleware (masalan 1 MB).

### M4 — Config/muhit ajratish nozikliklari
**Fayl:** `cmd/api/main.go:29`, `config/config.go` → `validate()`
`config.Load(".env.local")` **doim** hardcode qilingan — prod'da ham `.env.local` yuklashga urinadi. `validate()` faqat JWT bo'shligi va seed'ni tekshiradi; **tekshirilmaydi:** JWT secret min-uzunligi (#19 — hozir 32 belgi, yaxshi, lekin majburlanmagan), prod'da `DB_SSLMODE=disable` (#34), TLS majburiyligi, `ENABLE_SWAGGER=false`.
**Yechim:** env fayl nomini `APP_ENV` orqali tanlash; `validate()` da prod uchun `sslmode != disable`, JWT secret uzunligi ≥ 32, swagger off tekshiruvi.

### M5 — Nuxt admin: token himoyasiz cookie'da
**Fayl:** `nuxt-admin-panel/app/composables/useAuth.ts`
`tf_token`/`tf_refresh` JS-o'qiladigan cookie'da (`useCookie`, `sameSite:'lax'`), **`secure` flag yo'q** → prod'da HTTP orqali oqishi mumkin; `httpOnly` emasligi sabab XSS token o'g'irlashi mumkin (§17.3 #20/#35).
**Yechim:** `secure: true` (prod), imkoni bo'lsa refresh tokenni `httpOnly` cookie sifatida backend o'rnatishi. `v-html` ishlatilmagani (XSS yuzasi past) — yaxshi.

---

## 🟡 LOW / INFO

### L1 — Login user-enumeration timing
**Fayl:** `internal/service/auth_service.go` → `Login()`
Foydalanuvchi topilmasa bcrypt taqqoslashsiz darhol qaytadi; noto'g'ri parolda bcrypt ishlaydi → javob vaqti farqi mavjudlikni oshkor qiladi. Xato xabari bir xil (yaxshi), faqat timing.
**Yechim:** topilmagan holatda ham dummy bcrypt taqqoslash.

### L2 — OAuth'da PKCE yo'q
**Fayl:** `pkg/hemis/oauth.go`
Confidential client (client_secret bor), shuning uchun past ustuvorlik, lekin `code_challenge`/`verifier` qo'shilsa code-interception himoyasi kuchayadi (§17.3 #24 qisman).

### L3 — Ochiq `register` da `role=employee` o'zi tanlanadi
**Fayl:** `internal/dto/auth.go`, `auth_service.go` → `Register()`
Har kim parol bilan `employee` bo'lib ro'yxatdan o'tishi mumkin. Admin bo'lib bo'lmaydi (yaxshi), lekin xodim odatda faqat HEMIS orqali kelishi kerak bo'lsa — biznes qoidasini tekshiring.

---

## ✅ Tasdiqlangan kuchli tomonlar

- **SQL injection yo'q** — barcha so'rov parametrli (`?`), `ILIKE` ham placeholder bilan (`user_repository.go`).
- **JWT** — HS256, `SigningMethodHMAC` tekshiruvi bilan alg-confusion bloklangan; access 15 min, refresh 7 kun; `typ` tekshiruvi (`pkg/security/jwt.go`).
- **Parol** — bcrypt cost 12 (`pkg/security/password.go`).
- **Refresh** — rotatsiya + Redis JTI revoke; logout Redis'dan o'chiradi (`auth_service.go`).
- **OAuth** — `state` CSRF (`crypto/rand` 20 bayt, bir martalik, Redis TTL); `redirect_uri` config'dan (manipulyatsiya yo'q); token URL'ga tushmaydi (bir martalik exchange code).
- **Access control** — activity/stats `uid` token'dan olinadi (IDOR yo'q); barcha `/admin/*` route `RequireRole(admin)` bilan (`user/roster/structure` handler).
- **Path traversal** — avatar fayl nomi `sanitize()` bilan tozalanadi, atomik `tmp`→`rename`.
- **Media** — `io.LimitReader` hajm cheklovi, content-type tekshiruvi.
- **Xato ishlash** — ichki tafsilot mijozga chiqmaydi, `zap` bilan loglanadi (`handleError`).
- **Secret** — koda yo'q; `.env`, `.env.local`, `.env.production` `.gitignore`'da.
- **Mass assignment** — DTO whitelisting + `validator` teglari (`min`, `oneof`, `email`, `e164`).
- **Graceful shutdown** + `context` timeout'lar.
- **Frontend** — `v-html` yo'q, Vue avtomatik escape.
- **DoS/pagination** — admin ro'yxatida `limit > 100` cheklangan.

---

## Bajarildi ✅
- Anthropic-Cybersecurity-Skills'dan stack'ga mos 18 skill `security-skills/` ga qo'shildi (indeks: `security-skills/README.md`).
- Backend + admin panel to'liq audit qilindi, TOP-50 checklist bo'yicha.

## Keyin qilinishi kerak 🔜
- [x] H1: inbound rate limit middleware (Redis) — `internal/middleware/ratelimit.go` (2026-07-02)
- [x] M1: xavfsizlik header'lari middleware — `internal/middleware/security_headers.go` (2026-07-02)
- [x] M2: media downloader SSRF allowlist + private-IP bloklash — `pkg/media/avatar.go` (2026-07-02)
- [x] M3: request body hajm limiti — `internal/middleware/body_limit.go` (2026-07-02)
- [x] M4: `APP_ENV` bo'yicha env tanlash (`config.EnvFile`) + `validate()` kengaytirildi (2026-07-02)
- [x] M5: Nuxt cookie `secure` (prod) — `useAuth.ts` (2026-07-02)
- [ ] L1: login timing (dummy bcrypt), L2: PKCE, L3: register'da role cheklovi — past ustuvorlik.
- [ ] CI: `govulncheck ./...`, `npm audit`, gitleaks, semgrep (skill'lar bor).

## E'tibor berish kerak ⚠️
- Yumshatishlar **faqat** `APP_ENV=local` ostida bo'lsin (§17.1); default — qattiq.
- SSRF (M2) HEMIS ma'lumotiga ishonganda ham xatarli — ichki tarmoqqa kirish yo'lini yopish muhim.
