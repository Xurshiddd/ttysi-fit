# TTYSI_FIT — Mobil ilova (Flutter)

Riverpod v2 + Dio + go_router + flutter_secure_storage. Backend API'ga ulanadi.

## 1. Flutter SDK o'rnatish (Windows)

1. https://docs.flutter.dev/get-started/install/windows dan SDK'ni yuklab oling (yoki `git clone https://github.com/flutter/flutter.git -b stable`).
2. `flutter\bin` ni PATH ga qo'shing.
3. Tekshiring:
   ```bash
   flutter doctor
   ```
4. Android emulyator yoki qurilma uchun **Android Studio** o'rnating (SDK + emulator).

## 2. Platforma papkalarini generatsiya qilish

Bu papkada `lib/`, `pubspec.yaml` allaqachon bor (skeletim). Faqat `android/ios/web` papkalarini qo'shish kerak — mavjud kodni buzmasdan, **vaqtinchalik** loyiha orqali:

```bash
cd "/c/Users/user/Documents/ttysi fit"

# vaqtinchalik loyiha (faqat platforma papkalari uchun)
flutter create --org uz.ttyesi --project-name ttysi_fit _tmp_flutter

# kerakli papkalarni ko'chiramiz
mv _tmp_flutter/android "flutter apps/"
mv _tmp_flutter/ios "flutter apps/"
mv _tmp_flutter/web "flutter apps/" 2>/dev/null || true
mv _tmp_flutter/.metadata "flutter apps/" 2>/dev/null || true
rm -rf _tmp_flutter
```

> Eslatma: `flutter create .` ni to'g'ridan-to'g'ri shu papkada ishlatmang — u mening `pubspec.yaml` va `lib/main.dart` larimni qayta yozadi.

## 3. Ishga tushirish

```bash
cd "flutter apps"
flutter pub get
flutter run --dart-define=API_BASE_URL=http://10.0.2.2:8090/api/v1
```

- **Android emulyator**: `10.0.2.2` = kompyuter localhost (default).
- **iOS simulyator / web**: `--dart-define=API_BASE_URL=http://localhost:8090/api/v1`.
- **Haqiqiy telefon (Wi-Fi)**: kompyuter IP manzili, masalan `http://192.168.1.10:8090/api/v1`.
  Telefon va kompyuter bir tarmoqda bo'lishi shart.

### Haqiqiy telefon, USB kabel orqali (tavsiya etiladi)

`10.0.2.2` faqat emulyatorda ishlaydi, IP manzil esa tarmoq o'zgarsa buziladi.
Kabel bilan `adb reverse` eng ishonchlisi — telefondagi `localhost:8090`
kompyuterga yo'naltiriladi, Wi-Fi umuman kerak emas:

```bash
adb reverse tcp:8090 tcp:8090
flutter run -d <device-id> --dart-define=API_BASE_URL=http://localhost:8090/api/v1
```

`adb devices` — qurilma id sini ko'rsatadi. `adb reverse` telefon uzilib
ulanganda qaytadan bajarilishi kerak.

Bu usulda avatar/media rasmlari ham ishlaydi: backend `.env.local` dagi
`MEDIA_PUBLIC_BASE_URL=http://localhost:8090` telefonda ham to'g'ri hal bo'ladi.

> HTTP (TLS'siz) faqat **debug** build'da ochiq — qarang
> `android/app/src/debug/AndroidManifest.xml`. Release build'da Android cleartext'ni
> bloklaydi, ya'ni prod ilova majburan HTTPS ishlatadi (CLAUDE.md §17.2).

Backend ishlab turishi va admin/foydalanuvchi mavjud bo'lishi kerak.

## Arxitektura (CLAUDE.md 6)

```
lib/
├── core/
│   ├── api/        # Dio client + interceptors (token, Accept-Language)
│   ├── auth/       # TokenStorage (flutter_secure_storage)
│   ├── config/     # AppConfig (API manzil)
│   ├── i18n/       # Lokalizatsiya (uz/ru/en) + locale provider
│   ├── router/     # go_router (auth redirect)
│   └── theme/      # Ranglar (#1E3A5F + #00C896), Inter shrift
└── features/
    ├── auth/       # data / application (Riverpod) / presentation
    └── home/       # presentation
```

## Holat (state)

- **Riverpod v2**: `authControllerProvider` (login/logout/restore), `localeControllerProvider` (til).
- Token `flutter_secure_storage` da, har bir so'rovga `Authorization` va `Accept-Language` avtomatik qo'shiladi.

## Keyingi feature'lar

`features/` ostiga qo'shing: `profile`, `activity` (qadam), `rating`, `competition`, `challenge`, `training`, `news` — har biri `data / application / presentation` bilan.

> Agar `phosphor_flutter` ikonkasi nomida xato chiqsa, Material `Icons.*` ga almashtiring.
