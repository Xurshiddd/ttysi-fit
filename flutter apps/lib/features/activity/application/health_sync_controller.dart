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

/// HealthSyncState — sinxron holati.
///
/// `lastSyncAt` va `running` ATAYLAB ajratilgan. Avval `AsyncNotifier` va
/// uning `AsyncLoading` holati ikkalasini ham bildirar edi va shu bilan
/// jimgina nuqson tug'dirardi: `AsyncNotifier.build()` ASINXRON, ya'ni ilova
/// sovuq ishga tushganda birinchi chaqiruvda state hali `AsyncLoading`
/// bo'ladi. "Sinxron allaqachon ketyapti" deb hisoblanib, ilova ochilishidagi
/// avtomatik sinxron HECH QACHON bajarilmasdi (fon'dan qaytish va tugma
/// ishlagani uchun sezilmagan).
class HealthSyncState {
  const HealthSyncState({this.lastSyncAt, this.running = false, this.failed = false});

  /// lastSyncAt — oxirgi MUVAFFAQIYATLI sinxron vaqti (throttle hisobi).
  final DateTime? lastSyncAt;

  /// running — ayni damda sinxron ketyaptimi (takroriy ishga tushmasin).
  final bool running;

  /// failed — oxirgi urinish xato bilan tugadimi (UI ko'rsatishi uchun).
  final bool failed;

  HealthSyncState copyWith({DateTime? lastSyncAt, bool? running, bool? failed}) =>
      HealthSyncState(
        lastSyncAt: lastSyncAt ?? this.lastSyncAt,
        running: running ?? this.running,
        failed: failed ?? this.failed,
      );

  bool get synced => lastSyncAt != null;
}

/// HealthSyncController — telefon → backend qadam sinxroni.
///
/// `Notifier` (AsyncNotifier emas): `build()` sinxron bo'lgani uchun holat
/// birinchi o'qishdayoq tayyor va yuqoridagi poyga umuman yuzaga kelmaydi.
class HealthSyncController extends Notifier<HealthSyncState> {
  @override
  HealthSyncState build() => const HealthSyncState();

  /// sync — qo'lda (tugma) sinxron: throttle'siz, doim bajariladi.
  Future<HealthSyncResult> sync() => _run(force: true);

  /// syncIfStale — avtomatik (jim) sinxron: oxirgisidan
  /// [kAutoSyncInterval] o'tmagan bo'lsa hech narsa qilmaydi.
  Future<HealthSyncResult> syncIfStale() => _run(force: false);

  Future<HealthSyncResult> _run({required bool force}) async {
    // Takroriy ishga tushishdan himoya — qo'lda ham, avtomatik ham.
    if (state.running) return HealthSyncResult.skipped;

    if (!force) {
      final last = state.lastSyncAt;
      if (last != null && DateTime.now().difference(last) < kAutoSyncInterval) {
        return HealthSyncResult.skipped;
      }
    }

    state = state.copyWith(running: true, failed: false);
    try {
      final health = ref.read(healthServiceProvider);

      if (force) {
        // Qo'lda: kerak bo'lsa ruxsat oynasini ko'rsatamiz.
        if (!await health.requestPermissions()) {
          state = state.copyWith(running: false);
          return HealthSyncResult.noPermission;
        }
      } else {
        // Avtomatik: ruxsat SO'RALMAYDI (oyna ilova ochilishi bilan
        // o'zicha chiqib kelmasin).
        //
        // Lekin `hasPermissions()` bilan ham GATE qilmaymiz: Health Connect
        // o'qish ruxsatini ishonchli tekshirib bo'lmaydi va u ko'p holatda
        // `null` (noma'lum) qaytaradi. Avval null "rad etilgan" deb
        // hisoblanib, ilova ochilishidagi sinxron JIMGINA to'xtardi.
        //
        // O'rniga shunchaki O'QIB KO'RAMIZ: ruxsat yo'q bo'lsa o'qish bo'sh
        // qaytadi yoki xato beradi — ikkalasi ham quyida jim ushlanadi.
        // O'qish hech qanday oyna ko'rsatmaydi.
        if (await health.hasPermissions() == false) {
          state = state.copyWith(running: false);
          return HealthSyncResult.noPermission;
        }
      }

      // Bugun emas, oxirgi kunlar: ilova ochilmagan kunlar ham tiklanadi.
      final records = await health.readRecentDays(days: kBackfillDays);
      if (records.isEmpty) {
        state = state.copyWith(running: false);
        return HealthSyncResult.noData;
      }

      // Barcha kun BITTA so'rovda (CLAUDE.md §3.1).
      await ref.read(activityRepositoryProvider).recordBatch(records);

      // Statistika qayta o'qilsin (backend upsert qildi).
      ref.invalidate(activityStatsProvider);

      state = HealthSyncState(lastSyncAt: DateTime.now());
      return HealthSyncResult.success;
    } catch (_) {
      state = state.copyWith(running: false, failed: true);
      return HealthSyncResult.error;
    }
  }
}

/// autoDispose EMAS: oxirgi sinxron vaqti tab almashganda ham saqlanishi
/// kerak, aks holda throttle ishlamay har tab ochilganda so'rov ketardi.
final healthSyncProvider =
    NotifierProvider<HealthSyncController, HealthSyncState>(
        HealthSyncController.new);
