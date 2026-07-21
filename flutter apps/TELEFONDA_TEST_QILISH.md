# Flutter ilovani telefonda (Android) sinab ko'rish — to'liq qo'llanma

Bu qo'llanma Windows kompyuter + Android telefon uchun. (iOS uchun Mag kompyuter shart — pastda eslatma bor.)

---

## 0. Umumiy tushuncha (nima nimaga kerak)

| Narsa | Nima uchun kerak |
|-------|------------------|
| **Flutter SDK** | Ilovani kompilyatsiya qilish |
| **Android Studio** | Android SDK, platform-tools (`adb`), drayverlar, emulyator |
| **USB kabel** yoki **Wi-Fi** | Telefonni kompyuterga ulash |
| **Telefonda Developer mode + USB debugging** | Kompyuter telefonga ilova o'rnatishi uchun |
| **Backend ishlab turishi** | Ilova login va ma'lumot uchun API'ga ulanadi |
| **Bir xil Wi-Fi tarmoq** | Telefon kompyuterdagi backend'ni ko'rishi uchun |

---

## 1. Flutter SDK o'rnatish

1. https://docs.flutter.dev/get-started/install/windows/mobile saytiga kiring.
2. Flutter SDK (zip) ni yuklab oling va, masalan, `C:\src\flutter` ga oching.
   - **Muhim:** `C:\Program Files` ga qo'ymang (ruxsat muammosi bo'ladi).
3. PATH ga qo'shing:
   - Windows qidiruv → "environment variables" → **Edit the system environment variables**
   - **Environment Variables** → **Path** → **Edit** → **New** → `C:\src\flutter\bin`
   - OK bosing, yangi terminal oching.
4. Tekshiring:
   ```bash
   flutter --version
   flutter doctor
   ```
   `flutter doctor` qaysi narsa yetishmayotganini ✗ bilan ko'rsatadi.

---

## 2. Android Studio o'rnatish (Android SDK uchun)

1. https://developer.android.com/studio dan Android Studio'ni o'rnating.
2. Birinchi ochilishda **Standard** o'rnatishni tanlang — u quyidagilarni yuklaydi:
   - Android SDK
   - Android SDK Platform-Tools (`adb` shu yerda)
   - Emulyator
3. Android litsenziyalarini qabul qiling (terminal):
   ```bash
   flutter doctor --android-licenses
   ```
   Hammasiga `y` deb javob bering.
4. Yana tekshiring:
   ```bash
   flutter doctor
   ```
   `[✓] Android toolchain` bo'lishi kerak.

> `flutter doctor` da faqat **Android** va **Flutter** qatorlari ✓ bo'lsa yetarli. Chrome/VS Code/Visual Studio shart emas.

---

## 3. Telefonni tayyorlash (Developer mode + USB debugging)

1. Telefonda **Sozlamalar (Settings) → Telefon haqida (About phone)**.
2. **Build number** (yoki "Версия сборки") ni **7 marta** ketma-ket bosing — "Siz endi dasturchisiz" chiqadi.
3. Orqaga qaytib **Sozlamalar → Tizim → Developer options** (yoki qidiruvda "Developer").
4. Quyidagilarni yoqing:
   - **USB debugging** (USB orqali nosozliklarni tuzatish) — ✅ majburiy
   - **Install via USB** (agar bor bo'lsa) — ✅
5. Telefonni USB kabel bilan kompyuterga ulang.
6. Telefonda **"USB debugging'ga ruxsat berasizmi?"** so'rovi chiqadi → **Allow / Ruxsat** (va "always allow" belgilang).

### Ulanishni tekshirish

```bash
flutter devices
```

Telefoningiz ro'yxatda ko'rinishi kerak (masalan `SM-A515F (mobile)`). Ko'rinmasa:

```bash
adb devices
```

- `unauthorized` chiqsa — telefondagi ruxsat so'rovini tasdiqlang.
- Hech narsa chiqmasa — boshqa USB kabel/port sinab ko'ring (ba'zi kabellar faqat quvvat uzatadi), yoki telefon USB rejimini "File transfer (MTP)" ga o'tkazing.

---

## 4. Eng muhim qadam — telefon backend'ga ula olishi

Telefon **alohida qurilma**, shuning uchun `localhost` (telefon o'zini bildiradi) backend'ga mos kelmaydi. Telefon kompyuteringizning **Wi-Fi IP manzili** orqali ulanadi.

### 4.1. Telefon va kompyuter bir xil Wi-Fi'da bo'lsin
Ikkalasi ham bitta Wi-Fi routerga ulangan bo'lishi shart.

### 4.2. Kompyuter IP manzilini toping
```bash
ipconfig
```
**Wireless LAN adapter Wi-Fi** bo'limidagi **IPv4 Address** ni oling, masalan:
```
IPv4 Address. . . . . . . . . . . : 192.168.1.105
```
Demak API manzili: `http://192.168.1.105:8090/api/v1`

### 4.3. Windows Firewall — 8090 portga ruxsat
Birinchi marta telefon ulanganda Windows so'rashi mumkin — **Allow** bosing. Yoki qo'lda (PowerShell **administrator** sifatida):
```powershell
New-NetFirewallRule -DisplayName "TTYSI_FIT API" -Direction Inbound -LocalPort 8090 -Protocol TCP -Action Allow
```

### 4.4. Backend barcha interfeyslarda tinglashini tekshiring
Backend `:8090` da ishlaydi — bu barcha tarmoq interfeyslarini tinglaydi (to'g'ri). Telefon brauzeridan tekshiring:
```
http://192.168.1.105:8090/api/v1/health
```
`{"data":{"status":"ok",...}}` chiqsa — telefon backend'ni ko'ryapti. ✅

---

## 5. Ilovani ishga tushirish

Backend ishlab tursin (alohida terminalda):
```bash
cd "/c/Users/user/Documents/ttysi fit/go backend"
./run.sh dev
```

Flutter ilovani telefonda ishga tushiring (IP'ni o'zingiznikiga almashtiring):
```bash
cd "/c/Users/user/Documents/ttysi fit/flutter apps"
flutter pub get
flutter run --dart-define=API_BASE_URL=http://192.168.1.105:8090/api/v1
```

Bir nechta qurilma bo'lsa, `-d` bilan tanlang:
```bash
flutter devices
flutter run -d <device_id> --dart-define=API_BASE_URL=http://192.168.1.105:8090/api/v1
```

Birinchi build 5–15 daqiqa olishi mumkin (keyingilari tez). Ilova telefonda ochiladi.

### Foydali tugmalar (flutter run ishlab turganda)
- `r` — hot reload (o'zgarishni darrov ko'rsatadi)
- `R` — hot restart
- `q` — to'xtatish

---

## 6. Test login

Ilova ochilгach login ekrani chiqadi. Sinash uchun foydalanuvchi kerak:
```bash
cd "/c/Users/user/Documents/ttysi fit/go backend"
./run.sh create-admin test@ttyesi.uz "parol123" "Test User"
```
Keyin ilovada `test@ttyesi.uz` / `parol123` bilan kiring.

---

## 7. Simsiz (Wi-Fi) ulanish — kabelsiz

**Android 11+ da kabel UMUMAN shart emas.** Telefon va kompyuter bitta Wi-Fi'da bo'lsa yetarli.

### 7.1. Android 11 va undan yuqori — kabelsiz (tavsiya etiladi)

1. Telefon: **Developer options → Wireless debugging** → yoqing.
2. Shu ekranda **"Pair device with pairing code"** ni bosing. Chiqadi:
   - `192.168.1.50:41234` (juftlash manzili — **har safar o'zgaradi**)
   - 6 xonali kod, masalan `482913`
3. Kompyuterda (kod ekranda turganda, u ~2 daqiqada eskiradi):
   ```bash
   adb pair 192.168.1.50:41234
   # "Enter pairing code:" so'raganda kodni kiriting
   ```
4. Endi **Wireless debugging** asosiy ekraniga qaytib, u yerdagi
   **boshqa** portni oling (`192.168.1.50:37000` kabi — juftlash porti emas!) va ulang:
   ```bash
   adb connect 192.168.1.50:37000
   adb devices        # ro'yxatda ko'rinsin
   flutter devices
   ```
5. Keyin odatdagidek:
   ```bash
   flutter run --dart-define=API_BASE_URL=http://192.168.1.105:8090/api/v1
   ```

> **Juftlash bir marta** qilinadi — keyingi safar faqat `adb connect` yetadi.
> Lekin **port telefon qayta ulanganda o'zgaradi**, shuning uchun har seansda
> Wireless debugging ekranidagi yangi portni ko'rib oling.

### 7.2. Android 10 va past — bir martalik kabel kerak

```bash
# Kabel ulangan holda:
adb tcpip 5555
# Endi kabelni uzing:
adb connect 192.168.1.50:5555
```
Telefon o'chirib-yoqilsa bu rejim tushadi — kabelni qayta ulash kerak bo'ladi.

### 7.3. Telefon IP manzilini topish

Sozlamalar → Wi-Fi → ulangan tarmoq → **IP address**. Yoki kabel ulangan bo'lsa:
```bash
adb shell ip route
```

### 7.4. Simsiz ulanish muammolari

| Muammo | Sabab / yechim |
|--------|----------------|
| `failed to connect` | Telefon va kompyuter turli tarmoqda (biri mobil internetda) |
| Ulanadi, lekin darrov uziladi | Router'da **AP isolation / Client isolation** yoqilgan — router sozlamasidan o'chiring yoki telefon hotspot'i orqali ishlang |
| `adb pair` kod qabul qilmaydi | Kod eskirgan — telefonda oynani yopib qayta oching |
| Har safar qayta ulash kerak | Normal: Wireless debugging porti o'zgaradi. `adb connect` ni qayta bajaring |
| Build/o'rnatish sekin | Wi-Fi USB'dan sekinroq — birinchi o'rnatish uzoqroq, hot reload (`r`) esa tez |

> **Eslatma:** simsiz debug faqat ilovani o'rnatish/loglar uchun. Backend'ga
> ulanish alohida masala — 4-bo'limdagi IP va firewall sozlamalari baribir kerak.

---

## 8. Tez-tez uchraydigan muammolar

| Muammo | Yechim |
|--------|--------|
| `flutter devices` da telefon yo'q | Boshqa kabel/port; USB rejim "File transfer"; USB debugging ruxsatini tasdiqlang |
| `adb: unauthorized` | Telefondagi "Allow USB debugging" so'rovini bosing |
| Ilova ochiladi, lekin login "Connection refused/timeout" | IP noto'g'ri yoki firewall; telefon brauzeridan `/health` ni tekshiring; bir xil Wi-Fi'da bo'ling |
| `/health` telefonda ochilmaydi | Windows Firewall 8090 ni bloklayapti — 4.3 qadamni bajaring |
| `flutter doctor` Android ✗ | `flutter doctor --android-licenses` ni bajaring, Android Studio'da SDK o'rnatilganini tekshiring |
| Build juda sekin | Birinchi marta normal; antivirus Flutter papkasini skanlashini cheklang |
| phosphor ikonka xatosi | `Icons.*` (Material) ga almashtiring |

---

## iOS (iPhone) haqida

iPhone'da sinash uchun **macOS kompyuter + Xcode** shart (Windows'da iOS build qilib bo'lmaydi). Apple Developer hisobi (bepul) bilan o'z qurilmangizga o'rnatishingiz mumkin. Hozircha Android orqali sinang.

---

## Qadam sanagich (Health Connect) — Samsung Galaxy S23 Ultra

Ilova qadamlarni **Health Connect** orqali o'qiydi. Samsung Health barcha qadamlarni Health Connect'ga yozadi — alohida sensor kerak emas.

### Telefonni tayyorlash (bir marta)

1. **Health Connect** o'rnatilganini tekshiring: Sozlamalar → qidiruv → "Health Connect". S23 Ultra'da (One UI 6+/Android 14+) tizimga o'rnatilgan. Bo'lmasa Play Store'dan o'rnating.
2. **Samsung Health → Health Connect ulanishi:** Samsung Health → profil → Sozlamalar → **Health Connect** → ulang va "Qadamlar" (Steps) ga ruxsat bering. Shu bo'lmasa ilova 0 qadam ko'radi!
3. Samsung Health'da bugun qadam borligini tekshiring (telefon bilan biroz yuring 🙂).

### Qadamni telefon o'zi sanaydi

Qadamni **ilova emas, telefonning o'zi** sanaydi — ilova yopiq bo'lsa ham
(Health Connect / HealthKit apparat sensori). "Sinxronlash" degani o'sha tayyor
raqamni backend'ga **ko'chirish**, sanashni boshlash emas.

Ikki xil sinxron bor:

| Turi | Qachon | Ruxsat so'raydimi |
|------|--------|-------------------|
| **Avtomatik (jim)** | Ilova ochilganda va fon'dan qaytganda, 15 daqiqada bir martadan ko'p emas | ❌ Yo'q — faqat ruxsat allaqachon berilgan bo'lsa ishlaydi |
| **Qo'lda** | "Sinxronlash" tugmasi | ✅ Ha, kerak bo'lsa oyna ochadi |

Har sinxronda **oxirgi 7 kun** qayta yuboriladi ("backfill"): foydalanuvchi
bir hafta ilovani ochmasa ham hech qanday kun yo'qolmaydi.

### Ilovada sinash

1. Login → **Faollik** tabi → yashil **"Qadam sanagich"** kartasida **Sinxronlash** bosing.
2. Birinchi safar: "Jismoniy faollik" ruxsati (Android dialog) → keyin **Health Connect ruxsat oynasi** ochiladi → Steps / Active calories / Distance ga **Allow** bering.
3. "Qadamlar yuklandi" chiqadi → statistika yangilanadi → **Reyting** tabida o'z o'rningiz o'zgaradi.
4. **Avtomatikni sinash:** ilovani fon'ga chiqaring (Home tugmasi), biroz yuring,
   15 daqiqadan keyin qaytib kiring — tugmani bosmasdan raqam yangilanadi.

### Muammolar

| Muammo | Yechim |
|--------|--------|
| "Ruxsat berilmadi" | Health Connect → App permissions → ttysi_fit → hammasini yoqing |
| "Ma'lumot topilmadi" | Samsung Health ↔ Health Connect ulanishini tekshiring (yuqorida 2-qadam) |
| Ruxsat oynasi ochilmaydi | Ilovani to'liq o'chirib qayta o'rnating (`flutter run`) — manifest o'zgargan |
| Qadam kam ko'rinadi | Health Connect faqat ruxsat berilgandan keyingi 30 kunni beradi — bugun uchun muammo emas |
| Avtomatik sinxron ishlamayapti | Ruxsat berilmagan (jim rejim ruxsat so'ramaydi) — bir marta "Sinxronlash" bosing |
| Tunda (00:00–05:00) sana noto'g'ri | Backend `.env` da `APP_TIMEZONE=Asia/Tashkent` borligini tekshiring |

---

## Qisqacha tartib (checklist)

1. ☐ Flutter SDK + PATH
2. ☐ Android Studio + `flutter doctor --android-licenses`
3. ☐ Telefon: Developer mode + USB debugging
4. ☐ USB ulang → `flutter devices` da ko'rinsin
5. ☐ `ipconfig` → kompyuter IP
6. ☐ Firewall 8090 ruxsat
7. ☐ Telefon brauzerida `/health` ishlasin
8. ☐ Backend: `./run.sh dev`
9. ☐ `flutter run --dart-define=API_BASE_URL=http://<IP>:8090/api/v1`
10. ☐ `create-admin` bilan test foydalanuvchi → login
11. ☐ Samsung Health ↔ Health Connect ulangan
12. ☐ Faollik tabi → Sinxronlash → ruxsatlar → qadamlar yuklandi
