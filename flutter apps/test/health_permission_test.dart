// Qadam sanagich ruxsati testlari.
//
// NEGA BU TESTLAR BOR:
//
// Avtomatik sinxron ataylab ruxsat SO'RAMAYDI (ilova ochilishi bilan tizim
// oynasi chiqib kelishi foydalanuvchini cho'chitadi). Natijada "Faollik →
// Sinxronlash" tugmasini hech qachon bosmagan foydalanuvchining qadamlari
// umuman yuklanmasdi va u reytingda doim 0 bo'lib turardi — butun reyting
// tizimi shu ma'lumotga bog'liq bo'lsa ham.
//
// Tushuntirish oynasi va eslatma kartasi aynan shu teshikni yopadi.
// Ular jimgina yo'qolib qolmasligi kerak.

import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:ttysi_fit/core/i18n/app_localizations.dart';
import 'package:ttysi_fit/core/theme/app_theme.dart';
import 'package:ttysi_fit/features/activity/application/health_permission_controller.dart';
import 'package:ttysi_fit/features/activity/presentation/health_permission_prompt.dart';

/// _wrap — kartani mustaqil ravishda chizish uchun minimal ilova.
Widget _wrap(Widget child, {required AsyncValue<bool> permission}) {
  return ProviderScope(
    overrides: [
      healthPermissionProvider.overrideWith(() => _FakePermission(permission)),
    ],
    child: MaterialApp(
      theme: AppTheme.light(),
      darkTheme: AppTheme.dark(),
      locale: const Locale('uz'),
      supportedLocales: supportedLocales,
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      home: Scaffold(body: SingleChildScrollView(child: child)),
    ),
  );
}

/// _FakePermission — tizimga tegmasdan berilgan holatni qaytaradi.
class _FakePermission extends HealthPermissionController {
  _FakePermission(this._value);
  final AsyncValue<bool> _value;

  @override
  Future<bool> build() async {
    return _value.when(
      data: (v) => v,
      loading: () => throw UnimplementedError(),
      error: (e, _) => throw e,
    );
  }
}

void main() {
  group('HealthPermissionCard', () {
    testWidgets('ruxsat YO‘Q bo‘lsa eslatma ko‘rinadi', (tester) async {
      await tester.pumpWidget(
        _wrap(const HealthPermissionCard(), permission: const AsyncData(false)),
      );
      await tester.pumpAndSettle();

      expect(find.text('Qadamlar yuklanmayapti'), findsOneWidget);
      expect(find.text('Yoqish'), findsOneWidget);
      // Nima uchun muhimligi aytilsin — shunchaki "ruxsat bering" emas.
      expect(find.textContaining('reytingda 0 qadam'), findsOneWidget);
    });

    testWidgets('ruxsat BOR bo‘lsa karta butunlay yashirinadi', (tester) async {
      await tester.pumpWidget(
        _wrap(const HealthPermissionCard(), permission: const AsyncData(true)),
      );
      await tester.pumpAndSettle();

      expect(find.text('Qadamlar yuklanmayapti'), findsNothing);
      expect(find.text('Yoqish'), findsNothing);
    });

    testWidgets('layout xatosisiz chiziladi (tor ekranda ham)', (tester) async {
      tester.view.physicalSize = const Size(375, 812);
      tester.view.devicePixelRatio = 1.0;
      addTearDown(tester.view.resetPhysicalSize);

      await tester.pumpWidget(
        _wrap(const HealthPermissionCard(), permission: const AsyncData(false)),
      );
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
    });

    testWidgets('tugma bosilganda tushuntirish oynasi ochiladi', (tester) async {
      await tester.pumpWidget(
        _wrap(const HealthPermissionCard(), permission: const AsyncData(false)),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Yoqish'));
      await tester.pumpAndSettle();

      expect(find.text('Qadamlaringiz hisobga olinsin'), findsOneWidget);
      expect(find.text('Ruxsat berish'), findsOneWidget);
      expect(find.text('Keyinroq'), findsOneWidget);
    });

    testWidgets('tushuntirishda maxfiylik izohi bo‘lsin', (tester) async {
      await tester.pumpWidget(
        _wrap(const HealthPermissionCard(), permission: const AsyncData(false)),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Yoqish'));
      await tester.pumpAndSettle();

      // Ruxsat so'raganda eng ko'p beriladigan savol — javobi ko'rinib tursin.
      expect(find.textContaining('Faqat qadam'), findsOneWidget);
    });

    testWidgets('"Keyinroq" oynani yopadi, karta joyida qoladi',
        (tester) async {
      await tester.pumpWidget(
        _wrap(const HealthPermissionCard(), permission: const AsyncData(false)),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Yoqish'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Keyinroq'));
      await tester.pumpAndSettle();

      expect(find.text('Qadamlaringiz hisobga olinsin'), findsNothing);
      // Karta yo'qolmasin — bu yagona tiklash yo'li.
      expect(find.text('Yoqish'), findsOneWidget);
    });
  });

  group('Tarjimalar', () {
    test('health.* kalitlari uchala tilda ham bor', () {
      const keys = [
        'health.promptTitle',
        'health.promptBody',
        'health.promptPrivacy',
        'health.allow',
        'health.later',
        'health.granted',
        'health.deniedHint',
        'health.cardTitle',
        'health.cardBody',
        'health.enable',
      ];

      for (final lang in ['uz', 'ru', 'en']) {
        final s = S(Locale(lang));
        for (final k in keys) {
          // t() topilmagan kalit uchun kalitning o'zini qaytaradi —
          // foydalanuvchi ekranda "health.allow" ko'rib qolmasin.
          expect(s.t(k), isNot(k), reason: '$lang tilida $k tarjimasi yo‘q');
        }
      }
    });
  });
}
