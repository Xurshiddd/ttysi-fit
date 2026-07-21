import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'competition_models.dart';

final competitionRepositoryProvider = Provider<CompetitionRepository>(
    (ref) => CompetitionRepository(ref.read(dioProvider)));

/// CompetitionRepository — musobaqa API chaqiruvlari.
class CompetitionRepository {
  CompetitionRepository(this._dio);
  final Dio _dio;

  /// list — musobaqalar. `status` bo'sh: backend draft'larni oddiy
  /// foydalanuvchidan yashiradi, shuning uchun bu yerda filtr shart emas.
  Future<List<Competition>> list({int limit = 50}) async {
    final res = await _dio.get('/competitions', queryParameters: {
      'page': 1,
      'limit': limit,
    });
    final data = (res.data['data'] as List?) ?? const [];
    return data
        .map((e) => Competition.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  Future<void> register(String id) async {
    await _dio.post('/competitions/$id/register');
  }

  Future<void> cancel(String id) async {
    await _dio.delete('/competitions/$id/register');
  }
}
