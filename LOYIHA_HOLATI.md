# TTYSI_FIT — Loyiha holati hisoboti

**Sana:** 2026-07-21
**Asos:** `TTYSI_FIT.docx` (10 funksional yo'nalish + admin panel) bilan solishtirma.

TTYSI_FIT — institut talabalari va o'qituvchilarini sportga jalb qilish, jismoniy faollikni monitoring qilish va elektron reyting orqali rag'batlantirish uchun mobil platforma (Flutter ilova + Go backend + Nuxt admin panel).

---

## Umumiy holat

| Qatlam | Holat |
|--------|-------|
| Texnik poydevor (auth, HEMIS sync, infratuzilma) | Tayyor |
| Backend funksiyalari | ~95% (12 modul; analitika, eksport, do'kon tayyor) |
| Mobil ilova | ~85% (qadam sinxroni ishonchli, do'kon qo'shildi) |
| Admin panel | ~95% (15 sahifa) |

Hujjatdagi 10 yo'nalishning hammasi ishlaydigan holatda. Qolgan asosiy ish —
kichikroq bo'shliqlar (faollik turlari, turnir jadvallari, mashg'ulot
dasturlari) va institutdan muhr/imzo skanini olish.

---

## Hujjatdagi 10 yo'nalish bo'yicha holat

### 0. Sozlamalar (mobil) — 🟢 tayyor
- ✅ "Sozlamalar" bo'limi (5-tab): profil, til, mavzu, ilova haqida, chiqish
- ✅ **Profil shu bo'lim ichida** — karta bosilganda alohida ekran ochiladi;
  kartada FIT Coin va yutuq soni ko'rinib turadi
- ✅ Til va mavzu (tizim/yorug'/qorong'i) qurilmada saqlanadi va ilova
  ochilishida darrov qo'llanadi

### 1. Foydalanuvchi profili — 🟢 ~98%
- ✅ Talaba/o'qituvchi sifatida kirish (HEMIS OAuth — student + employee)
- ✅ Fakultet, kafedra, kurs, guruh ma'lumotlari (HEMIS sync orqali)
- ✅ Profil ekrani, sport faolligi tarixi
- ✅ **Yutuqlar va sertifikatlar**: profilda "Yutuqlarim" kartasi, to'liq ekran
  (qozonilgan + jarayondagi progress), sertifikat PDF yuklab olish
- ✅ **Sertifikat imzo bloki**: muhr + imzo skani + imzolovchi F.I.O./lavozim
  `.env` orqali (`CERT_*`); fayl yo'q bo'lsa hozirgidek bo'sh chiziq qoladi
- ✅ **Qurilmalarim va kirishlar** (17-migratsiya `user_sessions`): bir hisob —
  bir qurilma. Ikkinchi qurilmada kirilganda rozilik so'raladi; rozilik
  bermasa kira olmaydi. Eski qurilma `X-Device-Id` orqali DARROV chiqariladi
  (access token muddatini kutmasdan)

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
- ✅ **Ruxsat onboarding** — birinchi kirishda tushuntirish oynasi (nima uchun
  kerak + maxfiylik), ruxsat berilmasa bosh sahifada eslatma kartasi turadi
- ✅ **Reyting himoyasi** — reytingga faqat `health_connect` / `healthkit`
  manbali faollik kiradi; qo'lda kiritish mobil ilovadan olib tashlangan
- ✅ **Admin tuzatish** — `DELETE /admin/users/:id/activities?from=&to=`
  (hisobot sahifasida forma); audit loglanadi
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

### 5. Rag'batlantirish va bonus (FIT Coin) — 🟢 ~95%
- ✅ FIT Coin ledger (append-only), balans, tarix, admin grant/revoke
- ✅ Chellenj, musobaqa va **yutuq** mukofotlari ledger'ga idempotent yoziladi
- ✅ Mobil FIT Coin ko'rinishi, admin `fit-coins.vue`
- ✅ **Sovg'alar do'koni** — `rewards` + `reward_redemptions` (15-migratsiya):
  admin CRUD (`rewards.vue`), mobil do'kon, kod bilan topshirish
- ✅ **Almashtirish bitta tranzaksiyada**: sovg'a `FOR UPDATE` bilan bloklanadi,
  miqdor/shaxsiy limit/balans tekshiriladi, ledger'ga manfiy yozuv tushadi
- ✅ **Bekor qilishda coin qaytadi** (musbat ledger yozuvi) va miqdor tiklanadi
- ❌ Yetkazib berish/manzil (hozir faqat qo'lda topshirish)

### 6. Sport mashg'ulotlari bo'limi — 🟢 ~90%
- ✅ Video mashg'ulotlar CRUD, kategoriya (backenddan) + daraja filtri
- ✅ Mobil "Faollik → Mashg'ulot" segmenti, admin `trainings.vue`
- 🟡 Tayyor dasturlar (bir necha mashg'ulotdan iborat kurs) yo'q

### 7. Chellenjlar va marafonlar — 🟢 ~85%
- ✅ Chellenj CRUD (dinamik tur: qadam/masofa/faol daqiqa/custom)
- ✅ Qo'shilish, progress hisoblash, mukofot olish
- 🟡 Fakultet/guruhlararo bellashuv turi hali registrda yo'q

### 8. Axborot va yangiliklar — 🟢 ~95%
- ✅ Yangilik CRUD, draft/e'lon, pin, ko'rishlar hisobi
- ✅ Mobil bosh sahifada yangiliklar + batafsil ekran, admin `news.vue`
- ✅ **Bildirishnomalar (ilova ichida)** — 16-migratsiya: sovg'a topshirilgani/
  bekor qilingani va yangi yutuq avtomatik xabar qiladi; admin `announcements.vue`
  orqali e'lon yuboradi (fakultet/guruh/rol bo'yicha yoki hammaga)
- ✅ Mobil: AppBar da qo'ng'iroq + o'qilmaganlar nishoni, `/notifications` ekrani
- ❌ **Push (FCM)** — telefon ekraniga chiqadigan xabar hali yo'q
  (Firebase loyihasi + iOS uchun APNs kaliti kerak — tashkiliy qadam)
- ✅ Galaxy S23 Ultra da **qurilmada tasdiqlandi**: qo'ng'iroq nishoni,
  e'lon va sovg'a xabarlari, turlar bo'yicha ikonkalar

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
- ✅ 15 sahifa: users, faculties, departments, groups, hemis, ratings,
  challenges, competitions, fit-coins, news, trainings, achievements, reports,
  **rewards**, **redemptions**
- ✅ Dashboard analitikasi: grafiklar + filtrlar
- ✅ Light/Dark UI, responsive, Playwright e2e testlar (**42 ta**, shundan
  7 tasi 375/768/1280px kengliklarida gorizontal toshishni tekshiradi)
- ✅ Hisobot eksport (CSV)

---

## Texnik poydevor (bajarilgan) ✅

- **Autentifikatsiya:** JWT access + refresh (rotatsiya), avtomatik token yangilash
- **HEMIS OAuth:** talaba + xodim, bir martalik code orqali xavfsiz mobil oqim
- **HEMIS sync:** strukturalar, guruhlar, talabalar, xodimlar (rate-limit, dedup)
- **17 migratsiya:** users, faculties, structures, groups, activities, challenges,
  fit_coins, competitions, news, trainings, achievements, rewards, notifications, user_sessions
- **Dinamik kontent (§16):** chellenj, musobaqa va yutuq turlari registr orqali —
  yangi tur qo'shish migration ham, frontend o'zgarishi ham talab qilmaydi
- **Testlar:** Go (domain + service + handler + pkg), Flutter 69 test, Playwright 42 e2e
- **CI (`.github/workflows/ci.yml`):** 4 ta ish — Go (build/vet/`test -race`/
  gofmt/`govulncheck`), Nuxt (`npm audit --omit=dev`/build), Flutter
  (analyze/test), E2E (postgres+redis servis, migratsiya, backend, Playwright)
- **Bog'liqliklar zaifliksiz:** `govulncheck` — 0; `npm audit --omit=dev` — 0
  (yuqori va undan katta)

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
4. ~~**Sovg'alar do'koni**~~ — ✅ bajarildi (2026-07-21): admin CRUD, mobil
   do'kon, kod bilan topshirish, bekor qilishda coin qaytarish.

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
- **Do'kon almashtirish tranzaksiyasi** sovg'a qatorini `FOR UPDATE` bilan
  bloklaydi. Oxirgi donani ikki kishi bir vaqtda ololmaydi, lekin bu bitta
  sovg'aga kelgan xaridlar KETMA-KET bajarilishini bildiradi — juda ommabop
  sovg'ada kutish paydo bo'lishi mumkin.
- **Buyurtma kodi `crypto/rand` bilan** yasaladi (taxmin qilib bo'lmasin) va
  chalkashadigan belgilar (0/O, 1/I) olib tashlangan — kod og'zaki aytiladi.
- **Sovg'a narxi buyurtmada QOTIRILADI** — admin keyin narxni o'zgartirsa
  eski buyurtma va uning coin yozuvi mos qoladi.
- **`rating` COUNT** katta jadvalda sekinlashishi mumkin — §14.2 bo'yicha
  cache yoki taxminiy hisob kerak bo'ladi.
- **Faollik turlari** (yugurish/yurish/velosiped) ajratilmagan.
- **`APP_TIMEZONE` production'da ham qo'yilishi shart** — qo'yilmasa default
  `Asia/Tashkent` oladi, lekin noto'g'ri qiymat berilsa server umuman
  ishga tushmaydi (jim UTC'ga qaytmaydi — bu ataylab shunday).
- **Faollik upsert'i `GREATEST`** — qiymat kamaymaydi. Xato katta qiymat
  yozilsa qayta sinxron uni TUZATMAYDI. Endi buni admin panel orqali
  tuzatish mumkin: Hisobotlar → "Faollikni tuzatish" (oraliqni o'chiradi,
  telefon oxirgi 7 kunni qayta yuboradi). Bir martada maksimum 92 kun.
- **Reytingga faqat qurilma ma'lumoti kiradi** (`rating_repository.
  trustedSourceCond`). Qo'lda kiritilgan faollik shaxsiy statistikada
  qoladi, lekin musobaqaga ta'sir qilmaydi — aks holda `POST /activities`
  ga katta son yuborib birinchi o'ringa chiqish mumkin edi.
- **Avtomatik sinxron ruxsat so'ramaydi** (ataylab — ilova ochilishi bilan
  tizim oynasi chiqishi foydalanuvchini cho'chitadi). Buning o'rniga birinchi
  kirishda tushuntirish oynasi chiqadi, rad etilsa bosh sahifada eslatma
  kartasi qoladi. **Android'da ruxsat bir marta rad etilsa qayta so'rab
  bo'lmaydi** — shuning uchun oyna avval sababni tushuntiradi.
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
