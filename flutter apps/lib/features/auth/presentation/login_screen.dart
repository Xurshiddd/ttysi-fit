import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/config/app_config.dart';
import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/auth_controller.dart';
import '../data/auth_repository.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  bool _obscure = true;
  bool _showDev = false;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    super.dispose();
  }

  void _onResult(bool ok) {
    if (!mounted) return;
    if (ok) {
      context.go('/');
    } else {
      // Aniq xato kaliti controller'dan (tarmoq/server/parol farqlanadi).
      final key = ref.read(authControllerProvider).error ?? 'auth.error';
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(S.of(context).t(key)),
          backgroundColor: AppColors.danger,
        ),
      );
    }
  }

  Future<void> _loginHemis(String provider) async {
    await _withDeviceConflict(
      (force) => ref
          .read(authControllerProvider.notifier)
          .loginWithHemis(provider, force: force),
    );
  }

  Future<void> _loginDev() async {
    FocusScope.of(context).unfocus();
    await _withDeviceConflict(
      (force) => ref
          .read(authControllerProvider.notifier)
          .login(_email.text.trim(), _password.text, force: force),
    );
  }

  /// _withDeviceConflict — login urinishini bajaradi va hisob boshqa
  /// qurilmada ochiq bo'lsa foydalanuvchidan rozilik so'raydi.
  ///
  /// Rozilik bermasa — kirmaydi. Bu ATAYLAB: bitta hisobdan ikki kishi
  /// foydalansa reyting buziladi (qadamlar bir joyga tushadi).
  Future<void> _withDeviceConflict(Future<bool> Function(bool force) attempt) async {
    try {
      _onResult(await attempt(false));
    } on DeviceConflict catch (c) {
      if (!mounted) return;
      final s = S.of(context);

      final agreed = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: Text(s.t('device.conflictTitle')),
          content: Text(
            s.t('device.conflictBody').replaceAll(
                  '{device}',
                  c.deviceName.isEmpty ? s.t('device.unknown') : c.deviceName,
                ),
          ),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(ctx, false),
                child: Text(s.t('common.cancel'))),
            FilledButton(
                onPressed: () => Navigator.pop(ctx, true),
                child: Text(s.t('device.continueHere'))),
          ],
        ),
      );
      if (agreed != true || !mounted) return;

      try {
        _onResult(await attempt(true));
      } on DeviceConflict {
        if (!mounted) return;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(S.of(context).t('common.error'))),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    final state = ref.watch(authControllerProvider);
    final locale = ref.watch(localeControllerProvider);

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 420),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Container(
                    width: 64,
                    height: 64,
                    decoration: BoxDecoration(
                      color: AppColors.primary,
                      borderRadius: BorderRadius.circular(18),
                    ),
                    alignment: Alignment.center,
                    child: const Text('TF',
                        style: TextStyle(
                            color: Colors.white,
                            fontSize: 22,
                            fontWeight: FontWeight.bold)),
                  ),
                  const SizedBox(height: 20),
                  Text(s.t('auth.welcome'),
                      textAlign: TextAlign.center,
                      style: Theme.of(context)
                          .textTheme
                          .headlineSmall
                          ?.copyWith(fontWeight: FontWeight.bold)),
                  const SizedBox(height: 4),
                  Text(s.t('auth.subtitle'),
                      textAlign: TextAlign.center,
                      style: TextStyle(color: AppColors.muted)),
                  const SizedBox(height: 28),

                  if (state.loading)
                    const Padding(
                      padding: EdgeInsets.symmetric(vertical: 12),
                      child: Center(
                          child: CircularProgressIndicator(strokeWidth: 2.6)),
                    )
                  else ...[
                    FilledButton.icon(
                      onPressed: () => _loginHemis('student'),
                      icon: const Icon(Icons.school_outlined),
                      label: Text(s.t('auth.loginStudent')),
                    ),
                    const SizedBox(height: 12),
                    OutlinedButton.icon(
                      onPressed: () => _loginHemis('employee'),
                      icon: const Icon(Icons.badge_outlined),
                      label: Text(s.t('auth.loginEmployee')),
                    ),
                  ],

                  const SizedBox(height: 16),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.lock_outline, size: 14, color: AppColors.muted),
                      const SizedBox(width: 6),
                      Text(s.t('auth.hemisHint'),
                          style:
                              TextStyle(fontSize: 12, color: AppColors.muted)),
                    ],
                  ),

                  // Dev/test login — faqat debug yoki DEV_LOGIN=true bo'lganda.
                  if (AppConfig.devLogin) _devLoginSection(s, state),

                  const SizedBox(height: 12),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: supportedLocales.map((l) {
                      final selected = l.languageCode == locale.languageCode;
                      return TextButton(
                        onPressed: () => ref
                            .read(localeControllerProvider.notifier)
                            .setLocale(l),
                        child: Text(
                          l.languageCode.toUpperCase(),
                          style: TextStyle(
                            fontWeight:
                                selected ? FontWeight.bold : FontWeight.normal,
                            color:
                                selected ? AppColors.accent : AppColors.muted,
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _devLoginSection(S s, AuthState state) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const SizedBox(height: 12),
        TextButton.icon(
          onPressed: () => setState(() => _showDev = !_showDev),
          icon: Icon(_showDev ? Icons.expand_less : Icons.expand_more, size: 18),
          label: Text(s.t('auth.devLogin')),
        ),
        if (_showDev) ...[
          TextField(
            controller: _email,
            keyboardType: TextInputType.emailAddress,
            decoration: InputDecoration(
              labelText: s.t('auth.email'),
              prefixIcon: const Icon(Icons.email_outlined),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _password,
            obscureText: _obscure,
            decoration: InputDecoration(
              labelText: s.t('auth.password'),
              prefixIcon: const Icon(Icons.lock_outline),
              suffixIcon: IconButton(
                icon: Icon(_obscure
                    ? Icons.visibility_outlined
                    : Icons.visibility_off_outlined),
                onPressed: () => setState(() => _obscure = !_obscure),
              ),
            ),
          ),
          const SizedBox(height: 14),
          FilledButton(
            onPressed: state.loading ? null : _loginDev,
            child: Text(s.t('auth.signIn')),
          ),
        ],
      ],
    );
  }
}
