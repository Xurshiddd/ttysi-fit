import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'fitcoin_models.dart';

final fitCoinRepositoryProvider =
    Provider<FitCoinRepository>((ref) => FitCoinRepository(ref.read(dioProvider)));

/// FitCoinRepository — FIT Coin API chaqiruvlari.
class FitCoinRepository {
  FitCoinRepository(this._dio);
  final Dio _dio;

  Future<CoinBalance> balance() async {
    final res = await _dio.get('/fit-coins/balance');
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return CoinBalance.fromJson(data);
  }

  Future<List<CoinTx>> history({int limit = 50}) async {
    final res = await _dio.get('/fit-coins', queryParameters: {
      'page': 1,
      'limit': limit,
    });
    final data = (res.data['data'] as List?) ?? const [];
    return data
        .map((e) => CoinTx.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  /// claimReward — yakunlangan chellenj mukofotini oladi.
  ///
  /// Backend idempotent: takroriy so'rov 409 qaytaradi va ikkinchi coin
  /// berilmaydi. Shuning uchun mijozda takroriy bosishdan qo'rqmasa bo'ladi.
  Future<void> claimReward(String challengeId) async {
    await _dio.post('/challenges/$challengeId/claim-reward');
  }
}
