import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/activity_repository.dart';
import '../data/health_service.dart';
import 'activity_providers.dart';

/// Sinxron natijasi (UI xabari uchun).
enum HealthSyncResult { success, noPermission, noData, error, skipped }

/// Backfill chuqurligi: har sinxronda oxirgi shuncha kun qayta yuboriladi.
///
/// 7 kun — foydalanuvchi bir hafta ilovani ochmasa ham hech qanday kun
/// yo'qolmaydi. Ko'proq olish ham mumkin, lekin Health Connect ruxsat
/// berilgunga qadar bo'lgan tarixni bermaydi, foyda kamayadi.
const kBackfillDays = 7;

/// Avtomatik sinxron oralig'i: shu vaqt ichida qayta so'ralsa o'tkazib
/// yuboriladi. Ilova har ochilganda/fokusga qaytganda tekin so'rov
/// yubormaslik uchun (batareya + trafik + backend yuki).
const kAutoSyncInterval = Duration(minutes: 15);

/// HealthSyncController — telefon → backend qadam sinxroni.
///
/// State — oxirgi MUVAFFAQIYATLI sinxron vaqti (UI "oxirgi yangilanish"
/// ni ko'rsatishi va throttle hisobi uchun).
class HealthSyncController extends AsyncNotifier<DateTime?> {
  @override
  Future<DateTime?> build() async => null;

  /// sync — qo'lda (tugma) sinxron: throttle'siz, doim bajariladi.
  Future<HealthSyncResult> sync() => _run(force: true);

  /// syncIfStale — avtomatik (jim) sinxron: oxirgisidan
  /// [kAutoSyncInterval] o'tmagan bo'lsa hech narsa qilmaydi.
  Future<HealthSyncResult> syncIfStale() => _run(force: false);

  Future<HealthSyncResult> _run({required bool force}) async {
    if (!force) {
      final last = state.valueOrNull;
      if (last != null &&
          DateTime.now().difference(last) < kAutoSyncInterval) {
        return HealthSyncResult.skipped;
      }
      // Allaqachon ishlab turgan bo'lsa ikkinchisini boshlamaymiz.
      if (state.isLoading) return HealthSyncResult.skipped;
    }

    final previous = state.valueOrNull;
    state = const AsyncLoading();
    try {
      final health = ref.read(healthServiceProvider);

      // Qo'lda — kerak bo'lsa ruxsat so'raymiz; avtomatik — faqat
      // allaqachon berilgan bo'lsa davom etamiz (ruxsat oynasi ilova
      // ochilishi bilan o'zicha chiqib kelmasin).
      final granted = force
          ? await health.requestPermissions()
          : await health.hasPermissions();
      if (!granted) {
        state = AsyncData(previous);
        return HealthSyncResult.noPermission;
      }

      // Bugun emas, oxirgi kunlar: ilova ochilmagan kunlar ham tiklanadi.
      final records = await health.readRecentDays(days: kBackfillDays);
      if (records.isEmpty) {
        state = AsyncData(previous);
        return HealthSyncResult.noData;
      }

      // Barcha kun BITTA so'rovda (CLAUDE.md §3.1).
      await ref.read(activityRepositoryProvider).recordBatch(records);

      // Statistika qayta o'qilsin (backend upsert qildi).
      ref.invalidate(activityStatsProvider);

      state = AsyncData(DateTime.now());
      return HealthSyncResult.success;
    } catch (e, st) {
      state = AsyncError(e, st);
      return HealthSyncResult.error;
    }
  }
}

/// autoDispose EMAS: oxirgi sinxron vaqti tab almashganda ham saqlanishi
/// kerak, aks holda throttle ishlamay har tab ochilganda so'rov ketardi.
final healthSyncProvider =
    AsyncNotifierProvider<HealthSyncController, DateTime?>(
        HealthSyncController.new);
