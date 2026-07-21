import 'package:flutter/foundation.dart';

/// AppConfig — ilova muhiti sozlamalari.
///
/// API manzilini build paytida o'zgartirish mumkin:
///   flutter run --dart-define=API_BASE_URL=http://192.168.1.10:8090/api/v1
///
/// Standart qiymat Android emulyator uchun (10.0.2.2 = host localhost).
/// iOS simulyator yoki web uchun http://localhost:8090/api/v1 ishlating.
class AppConfig {
  static const String apiBaseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8090/api/v1',
  );

  /// devLogin — dev/test uchun oddiy email+parol login ko'rsatiladimi.
  /// Default: debug build'da yoqilgan, release'da o'chiq.
  /// Majburan boshqarish: --dart-define=DEV_LOGIN=true|false
  static const bool devLogin = bool.fromEnvironment(
    'DEV_LOGIN',
    defaultValue: kDebugMode,
  );

  static const Duration connectTimeout = Duration(seconds: 15);
  static const Duration receiveTimeout = Duration(seconds: 20);
}
