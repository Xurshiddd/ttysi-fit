// i18n to'liqligi testlari.
//
// NEGA BU TESTLAR BOR: S.t() topilmagan kalitni xato tashlamay, kalitning
// o'zini qaytaradi (`lang[key] ?? uz[key] ?? key`). Shu sababli yetishmagan
// tarjima kompilyatsiyada ham, `flutter analyze` da ham, widget testlarida ham
// bilinmaydi — u faqat ekranda "common.all" degan xom matn bo'lib ko'rinadi.
// Aynan shunday xato mashg'ulotlar filtrida qurilmada topilgan edi.
//
// Shuning uchun bu yerda ikki narsa tekshiriladi:
//   1) uchala til bir xil kalitlar to'plamiga ega;
//   2) lib/ ichida ishlatilgan har bir kalit haqiqatan ta'riflangan.

import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:ttysi_fit/core/i18n/app_localizations.dart';

void main() {
  final messages = debugMessages;

  group('i18n kalitlari', () {
    test('uchala til bir xil kalitlarga ega', () {
      final uz = messages['uz']!.keys.toSet();

      for (final lang in ['ru', 'en']) {
        final keys = messages[lang]!.keys.toSet();
        expect(uz.difference(keys), isEmpty,
            reason: '$lang tilida yetishmayapti: ${uz.difference(keys)}');
        expect(keys.difference(uz), isEmpty,
            reason: 'uz tilida yetishmayapti: ${keys.difference(uz)}');
      }
    });

    test('hech bir tarjima bo‘sh emas', () {
      messages.forEach((lang, map) {
        map.forEach((key, value) {
          expect(value.trim(), isNotEmpty, reason: '$lang / $key bo‘sh');
        });
      });
    });

    test('lib/ da ishlatilgan har bir kalit ta’riflangan', () {
      // Dinamik kalitlar (`training.level.$l`) bu yerda tekshirilmaydi —
      // ularning qiymati ish vaqtida hosil bo'ladi.
      final pattern = RegExp(r"\.t\('([^'$]+)'\)");
      final defined = messages['uz']!.keys.toSet();
      final missing = <String, String>{};

      for (final file in Directory('lib')
          .listSync(recursive: true)
          .whereType<File>()
          .where((f) => f.path.endsWith('.dart'))) {
        for (final m in pattern.allMatches(file.readAsStringSync())) {
          final key = m.group(1)!;
          if (!defined.contains(key)) missing[key] = file.path;
        }
      }

      expect(missing, isEmpty,
          reason: 'Ta’riflanmagan kalitlar (ekranda xom matn bo‘lib chiqadi): '
              '$missing');
    });
  });
}
