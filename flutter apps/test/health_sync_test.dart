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
      // Foydalanuvchi ilovani haftada bir marta ochsa ham kun yo\'qolmasin.
      expect(kBackfillDays, greaterThanOrEqualTo(7));
    });

    test('avtomatik sinxron oralig\'i mavjud va juda tez-tez emas', () {
      expect(kAutoSyncInterval.inMinutes, greaterThanOrEqualTo(5));
    });

    test('skipped natijasi mavjud — throttle jim o\'tishi uchun', () {
      expect(HealthSyncResult.values, contains(HealthSyncResult.skipped));
    });
  });
}
