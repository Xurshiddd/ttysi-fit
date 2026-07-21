import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'reward_models.dart';

final rewardRepositoryProvider =
    Provider<RewardRepository>((ref) => RewardRepository(ref.read(dioProvider)));

/// RewardRepository — FIT Coin do'koni API chaqiruvlari.
class RewardRepository {
  RewardRepository(this._dio);
  final Dio _dio;

  /// list — do'kon: faqat ayni damda olinishi mumkin bo'lgan sovg'alar
  /// (backend nofaol, tugagan va muddati o'tganlarni o'zi filtrlaydi).
  Future<List<Reward>> list({String? category}) async {
    final res = await _dio.get('/rewards', queryParameters: {
      'limit': 100,
      if (category != null && category.isNotEmpty) 'category': category,
    });
    final data = (res.data['data'] as List?) ?? [];
    return data
        .map((e) => Reward.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  /// categories — backenddan (kodda ro'yxat saqlanmaydi, §16.2).
  Future<List<String>> categories() async {
    final res = await _dio.get('/reward-categories');
    final data = (res.data['data'] as List?) ?? [];
    return data.map((e) => e.toString()).toList();
  }

  /// redeem — sovg'ani coinga almashtiradi.
  ///
  /// Balans yetmasa backend 409 qaytaradi — chaqiruvchi DioException ni
  /// tutib foydalanuvchiga tushunarli xabar ko'rsatadi.
  Future<Redemption> redeem(String rewardId) async {
    final res = await _dio.post('/rewards/$rewardId/redeem');
    return Redemption.fromJson(
        (res.data['data'] as Map).cast<String, dynamic>());
  }

  /// myRedemptions — mening buyurtmalarim (backend egalikni o'zi majburlaydi).
  Future<List<Redemption>> myRedemptions() async {
    final res = await _dio.get('/my-redemptions', queryParameters: {'limit': 50});
    final data = (res.data['data'] as List?) ?? [];
    return data
        .map((e) => Redemption.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }
}
