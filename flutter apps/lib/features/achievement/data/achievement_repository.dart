import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';

import '../../../core/api/dio_client.dart';
import 'achievement_models.dart';

final achievementRepositoryProvider = Provider<AchievementRepository>(
    (ref) => AchievementRepository(ref.read(dioProvider)));

/// AchievementRepository — yutuq API chaqiruvlari.
class AchievementRepository {
  AchievementRepository(this._dio);
  final Dio _dio;

  /// list — aktiv yutuqlar (foydalanuvchi progressi bilan).
  Future<List<Achievement>> list({int limit = 50}) async {
    final res = await _dio.get('/achievements', queryParameters: {
      'page': 1,
      'limit': limit,
    });
    return _parse(res.data);
  }

  /// earned — faqat qozonilgan yutuqlar (yangi -> eski).
  Future<List<Achievement>> earned({int limit = 50}) async {
    final res = await _dio.get('/achievements/me', queryParameters: {
      'page': 1,
      'limit': limit,
    });
    return _parse(res.data);
  }

  /// downloadCertificate — sertifikat PDF sini yuklab olib, lokal fayl
  /// yo'lini qaytaradi.
  ///
  /// URL'ni brauzerda ochib bo'lmaydi: endpoint Authorization sarlavhasini
  /// talab qiladi (sertifikat shaxsiy ma'lumot). Shuning uchun Dio orqali
  /// yuklab olinadi — token interceptor avtomatik qo'shadi.
  Future<File> downloadCertificate(String awardId) async {
    final res = await _dio.get<List<int>>(
      '/achievements/awards/$awardId/certificate',
      options: Options(responseType: ResponseType.bytes),
    );

    final dir = await getTemporaryDirectory();
    final file = File('${dir.path}/sertifikat-$awardId.pdf');
    await file.writeAsBytes(res.data ?? const []);
    return file;
  }

  List<Achievement> _parse(dynamic data) {
    final list = (data['data'] as List?) ?? const [];
    return list
        .map((e) => Achievement.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }
}
