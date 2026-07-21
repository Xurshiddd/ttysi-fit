import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'rating_models.dart';

final ratingRepositoryProvider =
    Provider<RatingRepository>((ref) => RatingRepository(ref.read(dioProvider)));

/// RatingRepository — reyting API chaqiruvlari.
class RatingRepository {
  RatingRepository(this._dio);
  final Dio _dio;

  /// list — reyting jadvali (kesim + davr, birinchi 50 qator).
  Future<List<RatingEntry>> list(RatingFilter f, {int limit = 50}) async {
    final res = await _dio.get('/ratings', queryParameters: {
      'type': f.type,
      'period': f.period,
      'page': 1,
      'limit': limit,
    });
    final data = (res.data['data'] as List?) ?? const [];
    return data
        .map((e) => RatingEntry.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  /// myRank — foydalanuvchining o'z o'rni (umumiy + fakultet ichida).
  Future<MyRating> myRank({String period = 'week'}) async {
    final res =
        await _dio.get('/ratings/me', queryParameters: {'period': period});
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return MyRating.fromJson(data);
  }
}
