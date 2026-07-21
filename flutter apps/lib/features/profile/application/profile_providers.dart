import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/profile_models.dart';
import '../data/profile_repository.dart';

/// profileProvider — o'z profili. Tahrirlashdan keyin `ref.invalidate` bilan
/// yangilanadi.
final profileProvider = FutureProvider.autoDispose<UserProfile>(
  (ref) => ref.read(profileRepositoryProvider).get(),
);

/// profileEditProvider — profil tahrirlash amali (saqlash holati bilan).
final profileEditProvider =
    AsyncNotifierProvider.autoDispose<ProfileEditController, void>(
        ProfileEditController.new);

class ProfileEditController extends AutoDisposeAsyncNotifier<void> {
  @override
  Future<void> build() async {}

  /// save — profilni yangilaydi. true — muvaffaqiyatli.
  Future<bool> save(ProfileUpdate u) async {
    state = const AsyncLoading();
    try {
      await ref.read(profileRepositoryProvider).update(u);
      ref.invalidate(profileProvider);
      state = const AsyncData(null);
      return true;
    } catch (e, st) {
      state = AsyncError(e, st);
      return false;
    }
  }
}
