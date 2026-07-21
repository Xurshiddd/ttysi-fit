import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../fitcoin/application/fitcoin_providers.dart';
import '../data/reward_models.dart';
import '../data/reward_repository.dart';

/// rewardCategoryProvider — tanlangan filtr ('' — barchasi).
final rewardCategoryProvider = StateProvider<String>((ref) => '');

/// rewardsProvider — do'kon ro'yxati (kategoriya filtri bilan).
final rewardsProvider = FutureProvider.autoDispose<List<Reward>>((ref) {
  final cat = ref.watch(rewardCategoryProvider);
  return ref.read(rewardRepositoryProvider).list(category: cat);
});

/// rewardCategoriesProvider — filtr chiplari uchun.
final rewardCategoriesProvider = FutureProvider<List<String>>(
    (ref) => ref.read(rewardRepositoryProvider).categories());

/// myRedemptionsProvider — mening buyurtmalarim.
final myRedemptionsProvider = FutureProvider.autoDispose<List<Redemption>>(
    (ref) => ref.read(rewardRepositoryProvider).myRedemptions());

/// redeemRewardProvider — almashtirish amali.
///
/// Muvaffaqiyatli bo'lsa BALANS, DO'KON va BUYURTMALAR ro'yxati qayta
/// o'qiladi: coin yechildi, miqdor kamaydi, yangi buyurtma paydo bo'ldi —
/// uchalasi ham ekranda darrov yangilanishi kerak.
final redeemRewardProvider =
    Provider<Future<Redemption> Function(String)>((ref) {
  return (String rewardId) async {
    final red = await ref.read(rewardRepositoryProvider).redeem(rewardId);
    ref.invalidate(coinBalanceProvider);
    ref.invalidate(rewardsProvider);
    ref.invalidate(myRedemptionsProvider);
    return red;
  };
});
