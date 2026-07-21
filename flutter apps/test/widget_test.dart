// Ilova darajasidagi smoke test.
//
// Eslatma: bu fayl avval Flutter'ning standart "counter" shabloni edi va mavjud
// bo'lmagan `MyApp` klassiga murojaat qilardi — shu sababli `flutter test`
// umuman ishlamasdi. Ekran testlari `tab_layout_test.dart` da.

import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:ttysi_fit/core/i18n/app_localizations.dart';
import 'package:ttysi_fit/core/theme/app_theme.dart';

void main() {
  group('AppTheme', () {
    // Regressiya: minimumSize kengligi Infinity bo'lsa, Row ichidagi tugma
    // "BoxConstraints forces an infinite width" bilan yiqiladi va uni o'z
    // ichiga olgan butun ekran bo'sh chiziladi. To'liq kenglik kerak bo'lsa
    // layoutdan olinadi (Column stretch / lokal minimumSize), temadan emas.
    test('filledButtonTheme minimal kengligi cheksiz emas', () {
      for (final theme in [AppTheme.light(), AppTheme.dark()]) {
        final size = theme.filledButtonTheme.style?.minimumSize
            ?.resolve(<WidgetState>{});
        expect(size, isNotNull);
        expect(size!.width.isFinite, isTrue,
            reason:
                'Size.fromHeight() ishlatilgan bo‘lsa kenglik Infinity bo‘ladi');
        expect(size.height, 52);
      }
    });

    test('yorug‘ va qorong‘i mavzular yasaladi', () {
      expect(AppTheme.light().brightness, Brightness.light);
      expect(AppTheme.dark().brightness, Brightness.dark);
    });
  });

  group('S (i18n)', () {
    test('uch tilda ham nav kalitlari bor', () {
      for (final loc in supportedLocales) {
        final s = S(loc);
        for (final key in ['nav.rating', 'nav.activity', 'nav.profile']) {
          expect(s.t(key), isNot(key),
              reason: '$key kaliti ${loc.languageCode} tilida yo‘q');
        }
      }
    });

    test('noma‘lum kalit kalitning o‘zini qaytaradi', () {
      expect(S(const Locale('uz')).t('yoq.kalit'), 'yoq.kalit');
    });
  });

  testWidgets('MaterialApp AppTheme bilan qulamay ochiladi', (tester) async {
    await tester.pumpWidget(MaterialApp(
      theme: AppTheme.light(),
      locale: const Locale('uz'),
      supportedLocales: supportedLocales,
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      home: const Scaffold(body: Center(child: Text('ok'))),
    ));
    await tester.pumpAndSettle();

    expect(tester.takeException(), isNull);
    expect(find.text('ok'), findsOneWidget);
  });
}
