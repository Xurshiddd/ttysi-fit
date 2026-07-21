import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_web_auth_2/flutter_web_auth_2.dart';

import '../../../core/api/dio_client.dart';
import '../../../core/config/app_config.dart';
import 'auth_models.dart';

final authRepositoryProvider =
    Provider<AuthRepository>((ref) => AuthRepository(ref.read(dioProvider)));

/// AuthRepository — auth API chaqiruvlari.
class AuthRepository {
  AuthRepository(this._dio);
  final Dio _dio;

  Future<AuthTokens> login(String email, String password) async {
    final res = await _dio.post(
      '/auth/login',
      data: {'email': email, 'password': password},
    );
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return AuthTokens.fromJson(data);
  }

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
  Future<AuthTokens> loginWithHemis(String provider) async {
    final loginUrl = '${AppConfig.apiBaseUrl}/auth/hemis/$provider/login';

    final result = await FlutterWebAuth2.authenticate(
      url: loginUrl,
      callbackUrlScheme: 'ttysifit',
    );

    final code = Uri.parse(result).queryParameters['code'];
    if (code == null || code.isEmpty) {
      throw Exception('HEMIS: code qaytmadi');
    }

    final res = await _dio.post('/auth/hemis/exchange', data: {'code': code});
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return AuthTokens.fromJson(data);
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
