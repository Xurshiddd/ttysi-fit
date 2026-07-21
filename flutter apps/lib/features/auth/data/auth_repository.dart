import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_web_auth_2/flutter_web_auth_2.dart';

import '../../../core/api/dio_client.dart';
import '../../../core/auth/device_identity.dart';
import '../../../core/config/app_config.dart';
import 'auth_models.dart';

/// DeviceConflict — hisob boshqa qurilmada ochiq (backend 409).
///
/// Login shu bilan to'xtaydi: foydalanuvchi rozilik bermaguncha kira
/// olmaydi (aks holda ikki kishi bitta hisobdan foydalanib reytingni
/// buzardi).
class DeviceConflict implements Exception {
  const DeviceConflict({required this.deviceName, required this.platform});
  final String deviceName;
  final String platform;
}

final authRepositoryProvider = Provider<AuthRepository>((ref) =>
    AuthRepository(ref.read(dioProvider), ref.read(deviceIdentityProvider)));

/// AuthRepository — auth API chaqiruvlari.
class AuthRepository {
  AuthRepository(this._dio, this._device);
  final Dio _dio;
  final DeviceIdentity _device;

  /// login — dev login (email/parol).
  ///
  /// [force] — foydalanuvchi "boshqa qurilmadan chiqarilsin" deb rozilik
  /// berdi. Faqat konflikt oynasidan keyin true bo'ladi.
  Future<AuthTokens> login(String email, String password,
      {bool force = false}) async {
    try {
      final res = await _dio.post('/auth/login', data: {
        'email': email,
        'password': password,
        'device': await _device.info(),
        if (force) 'force_device': true,
      });
      final data = (res.data['data'] as Map).cast<String, dynamic>();
      return AuthTokens.fromJson(data);
    } on DioException catch (e) {
      throw _mapConflict(e);
    }
  }

  /// _mapConflict — 409 ni tushunarli xatoga o'giradi.
  Object _mapConflict(DioException e) {
    if (e.response?.statusCode != 409) return e;
    final d = e.response?.data;
    if (d is Map && d['device'] is Map) {
      final dev = (d['device'] as Map).cast<String, dynamic>();
      return DeviceConflict(
        deviceName: (dev['name'] ?? '').toString(),
        platform: (dev['platform'] ?? '').toString(),
      );
    }
    return const DeviceConflict(deviceName: '', platform: '');
  }

  /// sessions — "Mening qurilmalarim".
  Future<List<UserSession>> sessions() async {
    final res = await _dio.get('/auth/sessions');
    final data = (res.data['data'] as List?) ?? [];
    return data
        .map((e) => UserSession.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  Future<void> revokeSession(String id) => _dio.delete('/auth/sessions/$id');

  /// currentDeviceId — ro'yxatda "shu qurilma" ni belgilash uchun.
  Future<String> currentDeviceId() => _device.deviceId();

  /// me — saqlangan token egasining profili.
  /// Sessiya tiklanganda ishlatiladi: token diskda qoladi, profil esa qolmaydi.
  Future<UserInfo> me() async {
    final res = await _dio.get('/users/me');
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return UserInfo.fromJson(data);
  }

  /// loginWithHemis — HEMIS OAuth (Custom Tab).
  /// provider: "student" | "employee".
  /// Oqim: authorize sahifa ochiladi → backend callback bir martalik `code`'ni
  /// `ttysifit://oauth/callback?code=...` deep link orqali qaytaradi → code token'ga almashtiriladi.
  Future<AuthTokens> loginWithHemis(String provider, {bool force = false}) async {
    final loginUrl = '${AppConfig.apiBaseUrl}/auth/hemis/$provider/login';

    final result = await FlutterWebAuth2.authenticate(
      url: loginUrl,
      callbackUrlScheme: 'ttysifit',
    );

    final code = Uri.parse(result).queryParameters['code'];
    if (code == null || code.isEmpty) {
      throw Exception('HEMIS: code qaytmadi');
    }

    return exchangeHemisCode(code, force: force);
  }

  /// exchangeHemisCode — kodni tokenga almashtiradi.
  ///
  /// Alohida metod: qurilma konfliktida foydalanuvchi rozilik bergach
  /// AYNAN SHU kod bilan qayta urinamiz — OAuth oynasini boshqatdan
  /// ochish shart emas. Kodning muddati qisqa, shuning uchun oyna uzoq
  /// ochiq qolsa qaytadan kirish kerak bo'ladi.
  Future<AuthTokens> exchangeHemisCode(String code, {bool force = false}) async {
    try {
      final res = await _dio.post('/auth/hemis/exchange', data: {
        'code': code,
        'device': await _device.info(),
        if (force) 'force_device': true,
      });
      final data = (res.data['data'] as Map).cast<String, dynamic>();
      return AuthTokens.fromJson(data);
    } on DioException catch (e) {
      throw _mapConflict(e);
    }
  }

  Future<AuthTokens> register({
    required String fullName,
    required String email,
    required String password,
    String role = 'student',
  }) async {
    final res = await _dio.post(
      '/auth/register',
      data: {
        'full_name': fullName,
        'email': email,
        'password': password,
        'role': role,
      },
    );
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return AuthTokens.fromJson(data);
  }
}
