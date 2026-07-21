import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/competition_models.dart';
import '../data/competition_repository.dart';

/// competitionListProvider — musobaqalar ro'yxati.
final competitionListProvider = FutureProvider.autoDispose<List<Competition>>(
  (ref) => ref.read(competitionRepositoryProvider).list(),
);

/// Ro'yxatdan o'tish natijasi.
///
/// `unavailable` — joy tugagan YOKI muddat o'tgan YOKI holat o'zgargan.
/// Ularni ajratmaymiz: backend sababni faqat tarjima qilinmagan xom matnda
/// beradi va unga tayanish mo'rt bo'lardi. Buning o'rniga ro'yxat yangilanadi
/// va karta sababni o'zi ko'rsatadi (joy soni / muddat / holat).
enum RegResult { success, alreadyRegistered, unavailable, error, cancelled }

/// competitionRegProvider — yozilish/bekor qilish amali.
final competitionRegProvider =
    AsyncNotifierProvider.autoDispose<CompetitionRegController, void>(
        CompetitionRegController.new);

class CompetitionRegController extends AutoDisposeAsyncNotifier<void> {
  @override
  Future<void> build() async {}

  Future<RegResult> register(String id) async {
    state = const AsyncLoading();
    try {
      await ref.read(competitionRepositoryProvider).register(id);
      ref.invalidate(competitionListProvider);
      state = const AsyncData(null);
      return RegResult.success;
    } on DioException catch (e, st) {
      state = AsyncError(e, st);
      ref.invalidate(competitionListProvider);

      // 409 — allaqachon yozilgan. Foydalanuvchi uchun xato emas: ro'yxat
      // yangilanadi va tinch xabar ko'rsatiladi.
      if (e.response?.statusCode == 409) {
        state = const AsyncData(null);
        return RegResult.alreadyRegistered;
      }
      // 400 — yozilish mumkin emas (joy tugagan / muddat o'tgan / yopilgan).
      // Tugma faqat reg_open bo'lganda faol, ya'ni bu yerga tushish holat
      // so'rov orasida o'zgarganini bildiradi — ro'yxat yangilanib, karta
      // sababni ko'rsatadi.
      if (e.response?.statusCode == 400) return RegResult.unavailable;
      return RegResult.error;
    } catch (e, st) {
      state = AsyncError(e, st);
      return RegResult.error;
    }
  }

  Future<RegResult> cancel(String id) async {
    state = const AsyncLoading();
    try {
      await ref.read(competitionRepositoryProvider).cancel(id);
      ref.invalidate(competitionListProvider);
      state = const AsyncData(null);
      return RegResult.cancelled;
    } catch (e, st) {
      state = AsyncError(e, st);
      return RegResult.error;
    }
  }
}
