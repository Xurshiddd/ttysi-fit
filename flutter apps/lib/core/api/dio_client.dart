import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../auth/device_identity.dart';
import '../auth/token_storage.dart';
import '../config/app_config.dart';
import '../i18n/app_localizations.dart';

/// dioProvider — sozlangan Dio klienti (token + til interceptor bilan).
final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(
    BaseOptions(
      baseUrl: AppConfig.apiBaseUrl,
      connectTimeout: AppConfig.connectTimeout,
      receiveTimeout: AppConfig.receiveTimeout,
      headers: {'Accept': 'application/json'},
    ),
  );

  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) async {
        // Accept-Language — joriy ilova tili (server javobi shu tilda keladi)
        options.headers['Accept-Language'] =
            ref.read(localeControllerProvider).languageCode;

        // Authorization — agar token bo'lsa
        final token = await ref.read(tokenStorageProvider).getAccessToken();
        if (token != null && token.isNotEmpty) {
          options.headers['Authorization'] = 'Bearer $token';
        }

        // X-Device-Id — server bekor qilingan qurilmani shu orqali taniydi
        // va darrov 401 qaytaradi (access token muddatini kutmasdan).
        options.headers['X-Device-Id'] =
            await ref.read(deviceIdentityProvider).deviceId();

        handler.next(options);
      },
    ),
  );

  return dio;
});
