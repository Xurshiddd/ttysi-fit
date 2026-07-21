import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'activity_models.dart';

final activityRepositoryProvider =
    Provider<ActivityRepository>((ref) => ActivityRepository(ref.read(dioProvider)));

/// ActivityRepository — faollik API chaqiruvlari.
class ActivityRepository {
  ActivityRepository(this._dio);
  final Dio _dio;

  /// stats — bugun/hafta/oy/jami yig'ma.
  Future<ActivityStats> stats() async {
    final res = await _dio.get('/activities/stats');
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return ActivityStats.fromJson(data);
  }

  /// record — bir kunlik faollikni yozadi/yangilaydi (bir kun — bir yozuv).
  Future<void> record(ActivityRecord r) async {
    await _dio.post('/activities', data: r.toJson());
  }

  /// recordBatch — bir necha kunni BITTA so'rovda yuboradi (backfill).
  ///
  /// Har kunni alohida POST qilish sikl ichida tarmoq so'rovi bo'lardi:
  /// sekin, qisman muvaffaqiyat (yarmi yozilib yarmi yozilmaslik) xavfi bor.
  Future<void> recordBatch(List<ActivityRecord> records) async {
    if (records.isEmpty) return;
    await _dio.post('/activities/batch', data: {
      'items': records.map((r) => r.toJson()).toList(),
    });
  }
}
