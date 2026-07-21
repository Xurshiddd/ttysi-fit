import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../prefs/app_prefs.dart';

/// themeControllerProvider — ilova mavzusi (tizim / yorug' / qorong'i).
///
/// Boshlang'ich qiymat `main()` da ombordan o'qib, override orqali beriladi:
/// shunda ilova ochilishida yorug' mavzu bir lahzaga chaqnab ketmaydi.
final themeControllerProvider =
    NotifierProvider<ThemeController, ThemeMode>(ThemeController.new);

class ThemeController extends Notifier<ThemeMode> {
  ThemeController([this._initial = ThemeMode.system]);
  final ThemeMode _initial;

  @override
  ThemeMode build() => _initial;

  void setMode(ThemeMode mode) {
    state = mode;
    ref.read(appPrefsProvider).setThemeMode(mode);
  }
}
