// Qadam sinxroni testlari.
//
// NEGA BU TESTLAR BOR:
//
// 1) Sana. Avval mijoz `date` yubormas, backend esa `time.Now().UTC()` ni
//    olardi. O'zbekiston UTC+5 bo'lgani uchun mahalliy 00:00–05:00 dagi
//    qadamlar KECHAGI kunga yozilib, o'sha kunning to'liq yozuvi ustiga
//    kichik qiymat bilan yozilardi (12 000 → 300). Endi sanani telefon
//    aytadi — shu serializatsiya buzilmasligi kerak.
//
// 2) Throttle. Ilova har fokusga qaytganda sinxron qilinsa telefon va
//    backend bekorga yuklanadi. syncIfStale oralig'ni hurmat qilishi kerak.
//
// 3) Sovuq start. Pastdagi guruhga qarang — avtomatik sinxron ilova
//    ochilganda umuman ishlamay qolgan edi.

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:ttysi_fit/features/activity/application/health_sync_controller.dart';
import 'package:ttysi_fit/features/activity/data/activity_models.dart';

void main() {
  group('ActivityRecord sana', () {
    test('toJson sanani YYYY-MM-DD ko\'rinishida yuboradi', () {
      final r = ActivityRecord(
        date: DateTime(2026, 7, 21),
        steps: 12000,
        source: 'health_connect',
      );

      expect(r.toJson()['date'], '2026-07-21');
      expect(r.toJson()['steps'], 12000);
    });

    test('bir va ikki raqamli oy/kun nolga to\'ldiriladi', () {
      final r = ActivityRecord(date: DateTime(2026, 1, 5), steps: 10);
      expect(r.toJson()['date'], '2026-01-05');
    });

    test('sana mahalliy kun sifatida qoladi — UTC ga surilmaydi', () {
      // Mahalliy tunning boshi. UTC ga o'girilsa (masalan .toUtc()) sana
      // oldingi kunga tushib ketardi — aynan shu xato tuzatildi.
      final local = DateTime(2026, 7, 21, 0, 30);
      final r = ActivityRecord(date: local, steps: 300);

      expect(r.toJson()['date'], '2026-07-21');
    });

    test('sanasiz yozuvda date maydoni umuman yuborilmaydi', () {
      const r = ActivityRecord(steps: 500);
      expect(r.toJson().containsKey('date'), isFalse);
    });
  });

  group('Sinxron sozlamalari', () {
    test('backfill kamida bir haftani qamraydi', () {
      // Foydalanuvchi ilovani haftada bir marta ochsa ham kun yo'qolmasin.
      expect(kBackfillDays, greaterThanOrEqualTo(7));
    });

    test('avtomatik sinxron oralig\'i mavjud va juda tez-tez emas', () {
      expect(kAutoSyncInterval.inMinutes, greaterThanOrEqualTo(5));
    });

    test('skipped natijasi mavjud — throttle jim o\'tishi uchun', () {
      expect(HealthSyncResult.values, contains(HealthSyncResult.skipped));
    });
  });

  // ───────────────────────────────────────────────────────────────
  // REGRESSIYA: sovuq startda avtomatik sinxron ishlamay qolgan edi.
  //
  // Sabab: holat `AsyncNotifier` edi va `_run` "sinxron ketyaptimi" ni
  // `state.isLoading` orqali tekshirardi. `AsyncNotifier.build()` ASINXRON,
  // shuning uchun ilova endi ochilganda birinchi chaqiruvda state hali
  // `AsyncLoading` bo'lardi — sinxron "allaqachon ketyapti" deb o'tkazib
  // yuborilardi va HECH QACHON bajarilmasdi.
  //
  // Nuqson uzoq sezilmadi: fon'dan qaytish va "Sinxronlash" tugmasi
  // ishlardi. Buzilgani esa aynan eng ko'p uchraydigan holat edi —
  // foydalanuvchi ilovani ochishi.
  group('Sovuq start (regressiya)', () {
    test('boshlang\'ich holat DARROV tayyor — AsyncLoading bosqichi yo\'q', () {
      final c = ProviderContainer();
      addTearDown(c.dispose);

      // Sinxron o'qiladi: hech qanday await kerak emas.
      final st = c.read(healthSyncProvider);

      expect(st.running, isFalse,
          reason: 'sovuq startda "ketyapti" deb belgilangan — '
              'avtomatik sinxron o\'tkazib yuborilardi');
      expect(st.lastSyncAt, isNull);
      expect(st.synced, isFalse);
    });

    test('throttle faqat MUVAFFAQIYATLI sinxrondan keyin ishlaydi', () {
      // lastSyncAt null — hali sinxron bo'lmagan, to'siq yo'q.
      const fresh = HealthSyncState();
      expect(fresh.synced, isFalse);

      final done = HealthSyncState(lastSyncAt: DateTime.now());
      expect(done.synced, isTrue);
    });

    test('running va lastSyncAt alohida — biri ikkinchisini bildirmaydi', () {
      final s = HealthSyncState(lastSyncAt: DateTime.now(), running: true);
      expect(s.running, isTrue);
      expect(s.synced, isTrue);
    });

    test('copyWith holatni tasodifan tozalamaydi', () {
      final base = HealthSyncState(lastSyncAt: DateTime(2026, 7, 21));
      final busy = base.copyWith(running: true);

      expect(busy.lastSyncAt, base.lastSyncAt,
          reason: 'sinxron boshlanganda oxirgi vaqt yo\'qoldi — '
              'throttle qayta ishlamay qolardi');
    });
  });
}
