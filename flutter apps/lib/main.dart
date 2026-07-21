import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'core/i18n/app_localizations.dart';
import 'core/prefs/app_prefs.dart';
import 'core/router/app_router.dart';
import 'core/theme/app_theme.dart';
import 'core/theme/theme_controller.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Sozlamalar ilova chizilishidan OLDIN o'qiladi: aks holda qorong'i mavzu
  // tanlagan foydalanuvchida ekran bir lahzaga oq chaqnab ketardi, til esa
  // 'uz' dan tanlanganiga sakrardi.
  final prefs = AppPrefs();
  final themeMode = await prefs.themeMode();
  final locale = await prefs.locale();

  runApp(ProviderScope(
    overrides: [
      themeControllerProvider.overrideWith(() => ThemeController(themeMode)),
      if (locale != null)
        localeControllerProvider.overrideWith(() => LocaleController(locale)),
    ],
    child: const TtysiFitApp(),
  ));
}

class TtysiFitApp extends ConsumerWidget {
  const TtysiFitApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(appRouterProvider);
    final locale = ref.watch(localeControllerProvider);
    final themeMode = ref.watch(themeControllerProvider);

    return MaterialApp.router(
      title: 'TTYSI_FIT',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light(),
      darkTheme: AppTheme.dark(),
      themeMode: themeMode,
      locale: locale,
      supportedLocales: supportedLocales,
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      routerConfig: router,
    );
  }
}
