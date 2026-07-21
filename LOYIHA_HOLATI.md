# TTYSI_FIT — Loyiha holati hisoboti

**Sana:** 2026-07-18
**Asos:** `TTYSI_FIT.docx` (10 funksional yo'nalish + admin panel) bilan solishtirma.

TTYSI_FIT — institut talabalari va o'qituvchilarini sportga jalb qilish, jismoniy faollikni monitoring qilish va elektron reyting orqali rag'batlantirish uchun mobil platforma (Flutter ilova + Go backend + Nuxt admin panel).

---

## Umumiy holat

| Qatlam | Holat |
|--------|-------|
| Texnik poydevor (auth, HEMIS sync, infratuzilma) | Tayyor |
| Backend funksiyalari | ~85% (10 modul; qolgani analitika/eksport, sertifikat PDF) |
| Mobil ilova | ~70% (asosiy ekranlar ishlaydi, qadam sanagich ulangan) |
| Admin panel | ~75% (11 sahifa; qolgani yutuqlar sahifasi, hisobot eksport) |

Hujjatdagi 10 yo'nalishning 8 tasi ishlaydigan holatda. Qolgan asosiy ish —
**analitika/hisobot eksport** va **sertifikat generatsiyasi (PDF)**.

---

## Hujjatdagi 10 yo'nalish bo'yicha holat

### 0. Sozlamalar (mobil) — 🟢 tayyor
- ✅ "Sozlamalar" bo'limi (5-tab): profil, til, mavzu, ilova haqida, chiqish
- ✅ **Profil shu bo'lim ichida** — karta bosilganda alohida ekran ochiladi;
  kartada FIT Coin va yutuq soni ko'rinib turadi
- ✅ Til va mavzu (tizim/yorug'/qorong'i) qurilmada saqlanadi va ilova
  ochilishida darrov qo'llanadi

### 1. Foydalanuvchi profili — 🟢 ~95%
- ✅ Talaba/o'qituvchi sifatida kirish (HEMIS OAuth — student + employee)
- ✅ Fakultet, kafedra, kurs, guruh ma'lumotlari (HEMIS sync orqali)
- ✅ Profil ekrani, sport faolligi tarixi
- ✅ **Yutuqlar va sertifikatlar**: profilda "Yutuqlarim" kartasi, to'liq ekran
  (qozonilgan + jarayondagi progress), sertifikat PDF yuklab olish
- ✅ **Sertifikat imzo bloki**: muhr + imzo skani + imzolovchi F.I.O./lavozim
  `.env` orqali (`CERT_*`); fayl yo'q bo'lsa hozirgidek bo'sh chiziq qoladi

### 2. Kunlik jismoniy faollik monitoringi — 🟢 ~92%
- ✅ Qadam/kaloriya/masofa/faol daqiqa yozish (`POST /activities`)
- ✅ Kunlik/haftalik/oylik/jami statistika
- ✅ **Qadam sanagich**: `health` paketi, Health Connect (Android) + HealthKit (iOS),
  ruxsatlar manifestda, "Sinxronlash" tugmasi ishlaydi
- ✅ **Avtomatik jim sinxron** — ilova ochilganda/fon'dan qaytganda, 15 daq throttle;
  ruxsat oynasini o'zicha ochmaydi (faqat tugma so'raydi)
- ✅ **Backfill** — har sinxronda oxirgi 7 kun `POST /activities/batch` orqali
  bitta so'rovda; foydalanuvchi bir hafta ilovani ochmasa ham kun yo'qolmaydi
- ✅ **Vaqt mintaqasi** — kun chegarasi `APP_TIMEZONE` (Asia/Tashkent) bo'yicha
- ❌ Yugurish/yurish/velosiped turlarini ajratish (hozir umumiy)
- ❌ Faol daqiqa (`active_min`) hali doim 0 — Health Connect EXERCISE_SESSION'dan olinmagan

### 3. Elektron sport reytingi — 🟢 ~90%
- ✅ Talaba/xodim/guruh/fakultet reytingi (`RANK()` bilan bitta so'rov)
- ✅ Davr bo'yicha filtr (hafta/oy/butun davr), Redis cache
- ✅ Mobil "Reyting" tabi, admin `ratings.vue`
- 🟡 Guruh/fakultet reytingi jon boshiga o'rtacha bo'yicha (katta guruh ustunlik olmasin)

### 4. Sport tadbirlari va musobaqalar — 🟢 ~85%
- ✅ Musobaqa CRUD (dinamik tur + `config JSONB`), ro'yxatdan o'tish, ishtirokchilar
- ✅ Mobil "Tadbirlar" tabi, admin `competitions.vue`
- 🟡 Turnir jadvallari / bosqichli natijalar hali yo'q

### 5. Rag'batlantirish va bonus (FIT Coin) — 🟢 ~85%
- ✅ FIT Coin ledger (append-only), balans, tarix, admin grant/revoke
- ✅ Chellenj, musobaqa va **yutuq** mukofotlari ledger'ga idempotent yoziladi
- ✅ Mobil FIT Coin ko'rinishi, admin `fit-coins.vue`
- ❌ Sovg'alarga almashtirish (do'kon) yo'q

### 6. Sport mashg'ulotlari bo'limi — 🟢 ~90%
- ✅ Video mashg'ulotlar CRUD, kategoriya (backenddan) + daraja filtri
- ✅ Mobil "Faollik → Mashg'ulot" segmenti, admin `trainings.vue`
- 🟡 Tayyor dasturlar (bir necha mashg'ulotdan iborat kurs) yo'q

### 7. Chellenjlar va marafonlar — 🟢 ~85%
- ✅ Chellenj CRUD (dinamik tur: qadam/masofa/faol daqiqa/custom)
- ✅ Qo'shilish, progress hisoblash, mukofot olish
- 🟡 Fakultet/guruhlararo bellashuv turi hali registrda yo'q

### 8. Axborot va yangiliklar — 🟢 ~90%
- ✅ Yangilik CRUD, draft/e'lon, pin, ko'rishlar hisobi
- ✅ Mobil bosh sahifada yangiliklar + batafsil ekran, admin `news.vue`

### 9. Analitika va hisobotlar — 🟢 ~85%
- ✅ Admin dashboard: foydalanuvchi/fakultet/kafedra/guruh sonlari
- ✅ **Faollik dinamikasi grafigi** — kunlik chiziq (`ChartLine.vue`), bo'sh
  kunlar `generate_series` bilan 0 qilib to'ldiriladi
- ✅ **Fakultetlar kesimi** — jon boshiga o'rtacha + qatnashuv ulushi
  (`ChartBars.vue`); jami bo'yicha emas, aks holda katta fakultet doim yutardi
- ✅ **Umumiy raqamlar** — jami qadam, masofa, faol/jami foydalanuvchi, qamrov %
- ✅ **Davr va fakultet filtri** (hafta/oy/butun davr)
- ✅ **CSV eksport** — `GET /admin/reports/users.csv`, stream (xotirada
  yig'ilmaydi), UTF-8 BOM + `;` ajratgich (Excel uchun), formula injection
  himoyasi (`pkg/report`)
- ❌ XLSX (formatlangan Excel) — CSV yetarli bo'lmasa keyin
- ❌ Rejali (avtomatik) hisobot yuborish — hozir faqat qo'lda yuklab olish

### 10. Admin-panel — 🟢 ~95%
- ✅ 13 sahifa: users, faculties, departments, groups, hemis, ratings,
  challenges, competitions, fit-coins, news, trainings, achievements, **reports**
- ✅ Dashboard analitikasi: grafiklar + filtrlar
- ✅ Light/Dark UI, responsive, Playwright e2e testlar (**32 ta**, shundan
  3 tasi 375/768/1280px kengliklarida gorizontal toshishni tekshiradi)
- ✅ Hisobot eksport (CSV)

---

## Texnik poydevor (bajarilgan) ✅

- **Autentifikatsiya:** JWT access + refresh (rotatsiya), avtomatik token yangilash
- **HEMIS OAuth:** talaba + xodim, bir martalik code orqali xavfsiz mobil oqim
- **HEMIS sync:** strukturalar, guruhlar, talabalar, xodimlar (rate-limit, dedup)
- **14 migratsiya:** users, faculties, structures, groups, activities, challenges,
  fit_coins, competitions, news, trainings, achievements
- **Dinamik kontent (§16):** chellenj, musobaqa va yutuq turlari registr orqali —
  yangi tur qo'shish migration ham, frontend o'zgarishi ham talab qilmaydi
- **Testlar:** Go (domain + handler), Flutter 32 test, Playwright e2e

---

## Tavsiya etilgan keyingi tartib

0. ~~**Qadam sinxron ishonchliligi**~~ — ✅ bajarildi (2026-07-21): vaqt mintaqasi,
   backfill, batch endpoint, `GREATEST` upsert, avtomatik jim sinxron.
1. ~~**Analitika/eksport (9-yo'nalish)**~~ — ✅ bajarildi (2026-07-21): dashboard
   grafiklari, fakultetlar kesimi, CSV eksport.
2. ~~**Sertifikatga muhr va imzo rasmi**~~ — ✅ infratuzilma tayyor (2026-07-21):
   `CERT_STAMP_PATH` / `CERT_SIGNATURE_PATH` / `CERT_SIGNER_NAME` /
   `CERT_SIGNER_TITLE`. **Qoldi:** institutdan skanerlangan muhr va imzo PNG
   olib, serverga qo'yish — kod o'zgarishi kerak emas.
3. Kichikroq bo'shliqlar: faollik turlari (yugurish/velosiped), turnir
   jadvallari, mashg'ulot dasturlari (bir necha mashg'ulotdan iborat kurs).
4. **Sovg'alar do'koni** — FIT Coin'ni nimagadir almashtirish (hozir faqat
   yig'iladi).

---

## E'tibor berish kerak ⚠️

- **Sertifikat diskda saqlanmaydi** — har so'ralganda qaytadan chiziladi
  (`pkg/certificate`). Shu sababli shablon o'zgarsa avval berilgan sertifikatlar
  ham darrov yangilanadi; fayl tozalash ham kerak emas.
- **Sertifikat shrifti** — DejaVu binarga singdirilgan (~1.4 MB).
  Go'ning o'z shrifti o'zbek kirillchasidagi Қ, Ғ, Ҳ ni qoplamaydi; shriftni
  almashtirmoqchi bo'lsangiz `pkg/certificate/certificate_test.go` dagi
  qamrov testi ogohlantiradi.
- **Muhr/imzo repozitoriyda saqlanmaydi** — ataylab: git tarixiga tushgan
  muhr tasvirini olib tashlab bo'lmaydi va u soxta hujjat yasashga yaraydi.
  Fayllar serverda, `.env` orqali ko'rsatiladi.
- **`CERT_SAMPLE_STAMP` — sinov belgisi** ("NAMUNA — haqiqiy muhr emas").
  Production'da yoqilsa server ishga tushmaydi. Haqiqiy muhr berilsa
  avtomatik o'chadi.
- **Muhr PNG alfa kanalisiz bo'lishi SHART** — startupda tekshiriladi,
  yaroqsiz bo'lsa server ko'tarilmaydi (buzuq sertifikat bergandan yaxshi).
- **Sertifikat emblemasi** — ttysi.uz dagi rasmiy logotipdan
  (`pkg/certificate/assets/`, manba va qayta yaratish yo'riqnomasi shu yerdagi
  README da). PNG **alfa kanalisiz** bo'lishi shart: fpdf shaffof piksellarni
  qora qilib chizadi. Buni ham test tekshiradi.
- **Avtomatik yutuq baholash** faollik yozilganda sinxron ishlaydi
  (`ActivityService.Record`). Foydalanuvchi ko'payganda bu qadam sekinlashsa,
  navbatga (queue) ko'chirish kerak bo'ladi.
- **`rating` COUNT** katta jadvalda sekinlashishi mumkin — §14.2 bo'yicha
  cache yoki taxminiy hisob kerak bo'ladi.
- **Faollik turlari** (yugurish/yurish/velosiped) ajratilmagan.
- **`APP_TIMEZONE` production'da ham qo'yilishi shart** — qo'yilmasa default
  `Asia/Tashkent` oladi, lekin noto'g'ri qiymat berilsa server umuman
  ishga tushmaydi (jim UTC'ga qaytmaydi — bu ataylab shunday).
- **Faollik upsert'i `GREATEST`** — qiymat kamaymaydi. Ya'ni xato katta qiymat
  (masalan test paytida) yozilsa, uni oddiy qayta sinxron bilan tuzatib
  bo'lmaydi; DB dan qo'lda o'chirish kerak.
- **Avtomatik sinxron ruxsat so'ramaydi** — foydalanuvchi hech qachon
  "Sinxronlash" tugmasini bosmasa, qadamlari umuman yuklanmaydi. Onboarding'da
  bir marta so'rash kerak bo'ladi.
- **Fakultet = `structures` yozuvi**, `faculties` jadvali 00004-migratsiyada
  olib tashlangan. Yangi so'rov yozganda `users.faculty_id → structures.id`
  bog'lanishidan foydalaning.
- **Analitika "butun davr" 366 kun bilan cheklangan** (`maxRangeDays`) —
  cheksiz oraliq `generate_series` da millionlab qator yasab serverni
  cho'ktirardi. Ko'proq kerak bo'lsa shu constant o'zgartiriladi.
- **CSV eksport paginatsiyasiz** — 9 600 foydalanuvchida ~1 MB va bir necha
  soniya. Foydalanuvchi soni bir necha barobar oshsa `LIMIT` yoki fon
  navbatiga (queue) o'tkazish kerak bo'ladi.
- **Eksportda formula injection himoyasi bor** (`pkg/report.Sanitize`).
  Yangi eksport qo'shilsa `report.Writer` dan foydalaning — to'g'ridan-to'g'ri
  `encoding/csv` ishlatilsa himoya chetlab o'tiladi.
- Dev DB'da sinov yutuqlari qoldi ("Birinchi ming qadam", "Universitet krossi
  g'olibi") — admin sahifasini yasashda namuna sifatida ishlatiladi.
