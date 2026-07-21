import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/auth/token_storage.dart';
import '../data/auth_models.dart';
import '../data/auth_repository.dart';

/// AuthState — autentifikatsiya holati (AsyncValue o'rniga oddiy holat).
class AuthState {
  final bool isAuthenticated;
  final bool loading;
  final UserInfo? user;
  final String? error;

  const AuthState({
    this.isAuthenticated = false,
    this.loading = false,
    this.user,
    this.error,
  });

  AuthState copyWith({
    bool? isAuthenticated,
    bool? loading,
    UserInfo? user,
    String? error,
  }) {
    return AuthState(
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      loading: loading ?? this.loading,
      user: user ?? this.user,
      error: error,
    );
  }
}

final authControllerProvider =
    NotifierProvider<AuthController, AuthState>(AuthController.new);

class AuthController extends Notifier<AuthState> {
  @override
  AuthState build() {
    _restore();
    return const AuthState();
  }

  /// Saqlangan token bo'lsa sessiyani tiklaydi va profilni yuklaydi.
  ///
  /// Token diskda saqlanadi, profil esa saqlanmaydi — shuning uchun uni
  /// serverdan qayta olamiz, aks holda ilova qayta ochilganda `user` null
  /// bo'lib qoladi (bosh sahifada ism ko'rinmaydi).
  Future<void> _restore() async {
    final token = await ref.read(tokenStorageProvider).getAccessToken();
    if (token == null || token.isEmpty) return;

    // Optimistik: token bor — login ekrani chaqnab ketmasin.
    state = state.copyWith(isAuthenticated: true);

    try {
      final user = await ref.read(authRepositoryProvider).me();
      state = state.copyWith(isAuthenticated: true, user: user);
    } on DioException catch (e) {
      // 401 — token eskirgan/bekor qilingan: tozalab login ekraniga qaytaramiz.
      if (e.response?.statusCode == 401) {
        await ref.read(tokenStorageProvider).clear();
        state = const AuthState();
        return;
      }
      // Tarmoq xatosi (backend o'chiq, Wi-Fi yo'q) — sessiyani buzmaymiz.
      // Foydalanuvchi ichkarida qoladi, ism keyingi urinishda yuklanadi.
    }
  }

  Future<bool> login(String email, String password) async {
    state = state.copyWith(loading: true, error: null);
    try {
      final tokens = await ref.read(authRepositoryProvider).login(email, password);
      await ref.read(tokenStorageProvider).saveTokens(
            access: tokens.accessToken,
            refresh: tokens.refreshToken,
          );
      state = AuthState(isAuthenticated: true, user: tokens.user);
      return true;
    } catch (e) {
      state = AuthState(isAuthenticated: false, error: _errorKey(e));
      return false;
    }
  }

  /// _errorKey — tarmoq xatosini parol xatosidan ajratadi (aniq diagnostika).
  static String _errorKey(Object e) {
    if (e is DioException) {
      // Server javob berdi: 401 = login/parol xato, boshqasi — server xatosi.
      if (e.response != null) {
        return e.response!.statusCode == 401 ? 'auth.error' : 'auth.serverError';
      }
      // Javob yo'q — ulanish muammosi (IP/firewall/Wi-Fi).
      return 'auth.network';
    }
    return 'auth.error';
  }

  /// loginWithHemis — HEMIS OAuth orqali kirish (provider: "student" | "employee").
  Future<bool> loginWithHemis(String provider) async {
    state = state.copyWith(loading: true, error: null);
    try {
      final tokens =
          await ref.read(authRepositoryProvider).loginWithHemis(provider);
      await ref.read(tokenStorageProvider).saveTokens(
            access: tokens.accessToken,
            refresh: tokens.refreshToken,
          );
      state = AuthState(isAuthenticated: true, user: tokens.user);
      return true;
    } catch (_) {
      state = const AuthState(isAuthenticated: false, error: 'auth.error');
      return false;
    }
  }

  Future<void> logout() async {
    await ref.read(tokenStorageProvider).clear();
    state = const AuthState();
  }
}
