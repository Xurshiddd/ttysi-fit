import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

final appPrefsProvider = Provider<AppPrefs>((ref) => AppPrefs());

/// AppPrefs — ilova sozlamalarini qurilmada saqlaydi (til, mavzu).
///
/// TokenStorage bilan bir xil omborni ishlatadi: maxfiy ma'lumot emas, lekin
/// shu tufayli qo'shimcha paket (shared_preferences) kerak bo'lmaydi.
class AppPrefs {
  static const _themeKey = 'app_theme_mode';
  static const _localeKey = 'app_locale';
  static const _healthAskedKey = 'health_permission_asked';

  final FlutterSecureStorage _storage = const FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
  );

  /// themeMode — saqlangan mavzu. Saqlanmagan bo'lsa tizim mavzusi.
  Future<ThemeMode> themeMode() async {
    switch (await _read(_themeKey)) {
      case 'light':
        return ThemeMode.light;
      case 'dark':
        return ThemeMode.dark;
      default:
        return ThemeMode.system;
    }
  }

  Future<void> setThemeMode(ThemeMode mode) =>
      _write(_themeKey, mode.name); // system | light | dark

  /// locale — saqlangan til. Saqlanmagan bo'lsa null (chaqiruvchi standartni tanlaydi).
  Future<Locale?> locale() async {
    final code = await _read(_localeKey);
    if (code == null || code.isEmpty) return null;
    // Faqat qo'llab-quvvatlanadigan til qabul qilinadi: omborda buzuq qiymat
    // qolsa ilova noma'lum tilda ochilib qolmasin.
    for (final l in supportedLanguageCodes) {
      if (l == code) return Locale(code);
    }
    return null;
  }

  Future<void> setLocale(Locale locale) => _write(_localeKey, locale.languageCode);

  /// healthAsked — qadam sanagich ruxsati haqida tushuntirish KO'RSATILGANMI.
  ///
  /// Ruxsat berilgan-berilmaganini bildirmaydi (uni tizimdan so'raymiz) —
  /// faqat "biz bir marta so'radik" degani. Shusiz ilova har ochilganda
  /// tushuntirish chiqib, foydalanuvchini bezovta qilardi.
  Future<bool> healthAsked() async => await _read(_healthAskedKey) == '1';

  Future<void> setHealthAsked() => _write(_healthAskedKey, '1');

  /// _read — ombor xatosi (masalan qurilma kaliti almashgan) sozlamani
  /// yo'qotadi, lekin ilovani yiqitmasligi kerak.
  Future<String?> _read(String key) async {
    try {
      return await _storage.read(key: key);
    } catch (_) {
      return null;
    }
  }

  Future<void> _write(String key, String value) async {
    try {
      await _storage.write(key: key, value: value);
    } catch (_) {
      // Sozlama saqlanmadi — joriy sessiyada baribir qo'llanadi.
    }
  }
}

/// supportedLanguageCodes — app_localizations dagi ro'yxat bilan bir xil.
const supportedLanguageCodes = ['uz', 'ru', 'en'];
