import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../challenge/application/challenge_providers.dart';
import '../data/fitcoin_models.dart';
import '../data/fitcoin_repository.dart';

/// coinBalanceProvider — joriy balans.
final coinBalanceProvider = FutureProvider.autoDispose<CoinBalance>(
  (ref) => ref.read(fitCoinRepositoryProvider).balance(),
);

/// coinHistoryProvider — tranzaksiyalar tarixi.
final coinHistoryProvider = FutureProvider.autoDispose<List<CoinTx>>(
  (ref) => ref.read(fitCoinRepositoryProvider).history(),
);

/// Mukofot olish natijasi.
enum ClaimResult { success, alreadyClaimed, notCompleted, error }

/// claimRewardProvider — chellenj mukofotini olish amali.
final claimRewardProvider =
    AsyncNotifierProvider.autoDispose<ClaimRewardController, void>(
        ClaimRewardController.new);

class ClaimRewardController extends AutoDisposeAsyncNotifier<void> {
  @override
  Future<void> build() async {}

  Future<ClaimResult> claim(String challengeId) async {
    state = const AsyncLoading();
    try {
      await ref.read(fitCoinRepositoryProvider).claimReward(challengeId);
      ref.invalidate(coinBalanceProvider);
      ref.invalidate(coinHistoryProvider);
      ref.invalidate(challengeListProvider);
      state = const AsyncData(null);
      return ClaimResult.success;
    } on DioException catch (e, st) {
      state = AsyncError(e, st);
      // 409 — allaqachon olingan (backend idempotent). Bu foydalanuvchi uchun
      // xato emas: balansni yangilab, tinch xabar ko'rsatamiz.
      if (e.response?.statusCode == 409) {
        ref.invalidate(coinBalanceProvider);
        ref.invalidate(challengeListProvider);
        state = const AsyncData(null);
        return ClaimResult.alreadyClaimed;
      }
      if (e.response?.statusCode == 400) return ClaimResult.notCompleted;
      return ClaimResult.error;
    } catch (e, st) {
      state = AsyncError(e, st);
      return ClaimResult.error;
    }
  }
}
