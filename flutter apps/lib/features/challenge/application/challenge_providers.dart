import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/challenge_models.dart';
import '../data/challenge_repository.dart';

/// challengeListProvider — aktiv chellenjlar ro'yxati.
final challengeListProvider = FutureProvider.autoDispose<List<Challenge>>(
  (ref) => ref.read(challengeRepositoryProvider).list(),
);

/// challengeJoinProvider — chellenjga qo'shilish amali.
final challengeJoinProvider =
    AsyncNotifierProvider.autoDispose<ChallengeJoinController, void>(
        ChallengeJoinController.new);

class ChallengeJoinController extends AutoDisposeAsyncNotifier<void> {
  @override
  Future<void> build() async {}

  /// join — qo'shiladi va ro'yxatni yangilaydi. true — muvaffaqiyatli.
  Future<bool> join(String id) async {
    state = const AsyncLoading();
    try {
      await ref.read(challengeRepositoryProvider).join(id);
      ref.invalidate(challengeListProvider);
      state = const AsyncData(null);
      return true;
    } catch (e, st) {
      state = AsyncError(e, st);
      return false;
    }
  }
}
