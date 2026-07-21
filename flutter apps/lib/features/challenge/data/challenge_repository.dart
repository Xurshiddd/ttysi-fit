import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'challenge_models.dart';

final challengeRepositoryProvider = Provider<ChallengeRepository>(
    (ref) => ChallengeRepository(ref.read(dioProvider)));

/// ChallengeRepository — chellenj API chaqiruvlari.
class ChallengeRepository {
  ChallengeRepository(this._dio);
  final Dio _dio;

  /// list — aktiv chellenjlar (foydalanuvchi progressi bilan).
  Future<List<Challenge>> list({int limit = 50}) async {
    final res = await _dio.get('/challenges', queryParameters: {
      'status': 'active',
      'page': 1,
      'limit': limit,
    });
    final data = (res.data['data'] as List?) ?? const [];
    return data
        .map((e) => Challenge.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  /// join — chellenjga qo'shilish.
  Future<void> join(String id) async {
    await _dio.post('/challenges/$id/join');
  }
}
