import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/theme_controller.dart';
import '../../achievement/application/achievement_providers.dart';
import '../../auth/application/auth_controller.dart';
import '../../fitcoin/application/fitcoin_providers.dart';
import '../../profile/application/profile_providers.dart';

/// SettingsTab — sozlamalar bo'limi. Profil shu yerning ichida: tepadagi
/// karta bosilganda to'liq profil ekrani ochiladi.
///
/// Profil kartasida asosiy raqamlar (coin, yutuq) ko'rinib turadi — profil
/// bir tegish uzoqlashgani bilan foydalanuvchi ularni baribir darrov ko'radi.
class SettingsTab extends ConsumerWidget {
  const SettingsTab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(profileProvider);
        ref.invalidate(coinBalanceProvider);
        ref.invalidate(earnedAchievementsProvider);
      },
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const _ProfileCard(),
          const SizedBox(height: 20),

          _SectionTitle(text: s.t('settings.appearance')),
          const _LanguageTile(),
          const SizedBox(height: 8),
          const _ThemeTile(),

          const SizedBox(height: 20),
          _SectionTitle(text: s.t('settings.about')),
          const _DevicesTile(),
          const _AboutTile(),

          const SizedBox(height: 24),
          OutlinedButton.icon(
            style: OutlinedButton.styleFrom(
              minimumSize: const Size.fromHeight(52),
              foregroundColor: AppColors.danger,
              side: const BorderSide(color: AppColors.danger),
            ),
            icon: const Icon(Icons.logout),
            label: Text(s.t('common.logout')),
            onPressed: () async {
              await ref.read(authControllerProvider.notifier).logout();
              if (context.mounted) context.go('/login');
            },
          ),
          const SizedBox(height: 12),
        ],
      ),
    );
  }
}

class _SectionTitle extends StatelessWidget {
  const _SectionTitle({required this.text});
  final String text;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(left: 4, bottom: 8),
        child: Text(text,
            style: const TextStyle(
                color: AppColors.muted,
                fontSize: 12,
                fontWeight: FontWeight.w600,
                letterSpacing: 0.5)),
      );
}

/// _ProfileCard — profil bo'limiga kirish. Avatar, ism, rol + coin/yutuq soni.
class _ProfileCard extends ConsumerWidget {
  const _ProfileCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final profile = ref.watch(profileProvider);
    final coins = ref.watch(coinBalanceProvider);
    final earned = ref.watch(earnedAchievementsProvider);

    return InkWell(
      onTap: () => context.push('/profile'),
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Column(
          children: [
            Row(
              children: [
                CircleAvatar(
                  radius: 26,
                  backgroundColor: AppColors.primary,
                  child: profile.maybeWhen(
                    data: (p) => Text(p.initials,
                        style: const TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.w700,
                            fontSize: 16)),
                    orElse: () => const Icon(Icons.person, color: Colors.white),
                  ),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: profile.when(
                    loading: () => const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                    error: (_, __) => Text(s.t('common.error'),
                        style: const TextStyle(color: AppColors.muted)),
                    data: (p) => Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(p.fullName,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                                fontWeight: FontWeight.w700, fontSize: 15)),
                        const SizedBox(height: 2),
                        Text(
                          p.facultyName.isNotEmpty
                              ? p.facultyName
                              : s.t('role.${p.role}'),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                              color: AppColors.muted, fontSize: 13),
                        ),
                      ],
                    ),
                  ),
                ),
                const Icon(Icons.chevron_right, color: AppColors.muted),
              ],
            ),
            const SizedBox(height: 14),
            const Divider(height: 1),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: _MiniStat(
                    icon: Icons.monetization_on_outlined,
                    color: const Color(0xFFF59E0B),
                    label: s.t('coin.balance'),
                    value: coins.maybeWhen(
                        data: (c) => '${c.balance}', orElse: () => '—'),
                  ),
                ),
                Container(width: 1, height: 32, color: AppColors.muted.withValues(alpha: 0.2)),
                Expanded(
                  child: _MiniStat(
                    icon: Icons.emoji_events_outlined,
                    color: AppColors.accent,
                    label: s.t('achievement.title'),
                    value: earned.maybeWhen(
                        data: (list) => '${list.length}', orElse: () => '—'),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _MiniStat extends StatelessWidget {
  const _MiniStat({
    required this.icon,
    required this.color,
    required this.label,
    required this.value,
  });
  final IconData icon;
  final Color color;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Icon(icon, color: color, size: 18),
        const SizedBox(height: 4),
        Text(value,
            style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 16)),
        Text(label,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(color: AppColors.muted, fontSize: 11)),
      ],
    );
  }
}

/// _LanguageTile — til tanlash. Tanlov qurilmada saqlanadi va Dio
/// Accept-Language header'iga ham tushadi (server javobi ham shu tilda keladi).
class _LanguageTile extends ConsumerWidget {
  const _LanguageTile();

  static const _names = {'uz': 'O‘zbekcha', 'ru': 'Русский', 'en': 'English'};

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final current = ref.watch(localeControllerProvider);

    return _SettingsTile(
      icon: Icons.language,
      title: s.t('settings.language'),
      trailing: Text(_names[current.languageCode] ?? current.languageCode,
          style: const TextStyle(color: AppColors.muted, fontSize: 13)),
      onTap: () => _showChoiceSheet<String>(
        context: context,
        options: _names.entries
            .map((e) => _Choice(value: e.key, label: e.value))
            .toList(),
        selected: current.languageCode,
        onSelected: (v) =>
            ref.read(localeControllerProvider.notifier).setLocale(Locale(v)),
      ),
    );
  }
}

/// _Choice — tanlov oynasidagi bitta variant.
class _Choice<T> {
  const _Choice({required this.value, required this.label});
  final T value;
  final String label;
}

/// _showChoiceSheet — til va mavzu uchun umumiy tanlov oynasi.
///
/// `RadioGroup` ishlatiladi: `RadioListTile.groupValue`/`onChanged` Flutter
/// 3.32 dan keyin eskirgan.
void _showChoiceSheet<T>({
  required BuildContext context,
  required List<_Choice<T>> options,
  required T selected,
  required void Function(T value) onSelected,
}) {
  showModalBottomSheet<void>(
    context: context,
    shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
    builder: (sheetContext) => SafeArea(
      child: RadioGroup<T>(
        groupValue: selected,
        onChanged: (v) {
          if (v != null) onSelected(v);
          Navigator.of(sheetContext).pop();
        },
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(height: 8),
            for (final o in options)
              RadioListTile<T>(value: o.value, title: Text(o.label)),
            const SizedBox(height: 8),
          ],
        ),
      ),
    ),
  );
}

/// _ThemeTile — mavzu tanlash (tizim / yorug' / qorong'i).
class _ThemeTile extends ConsumerWidget {
  const _ThemeTile();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final mode = ref.watch(themeControllerProvider);

    String label(ThemeMode m) => switch (m) {
          ThemeMode.light => s.t('settings.themeLight'),
          ThemeMode.dark => s.t('settings.themeDark'),
          ThemeMode.system => s.t('settings.themeSystem'),
        };

    return _SettingsTile(
      icon: mode == ThemeMode.dark
          ? Icons.dark_mode_outlined
          : Icons.light_mode_outlined,
      title: s.t('settings.theme'),
      trailing: Text(label(mode),
          style: const TextStyle(color: AppColors.muted, fontSize: 13)),
      onTap: () => _showChoiceSheet<ThemeMode>(
        context: context,
        options: ThemeMode.values
            .map((m) => _Choice(value: m, label: label(m)))
            .toList(),
        selected: mode,
        onSelected: (v) => ref.read(themeControllerProvider.notifier).setMode(v),
      ),
    );
  }
}

/// _DevicesTile — "Qurilmalarim va kirishlar".
///
/// Sozlamalar ichida: hisob xavfsizligi bilan bog'liq, kundalik amal emas.
class _DevicesTile extends StatelessWidget {
  const _DevicesTile();

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    return _SettingsTile(
      icon: Icons.devices_outlined,
      title: s.t('device.title'),
      onTap: () => context.push('/devices'),
    );
  }
}

/// _AboutTile — ilova haqida.
class _AboutTile extends StatelessWidget {
  const _AboutTile();

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    return _SettingsTile(
      icon: Icons.info_outline,
      title: s.t('settings.aboutApp'),
      trailing: const Text(appVersion,
          style: TextStyle(color: AppColors.muted, fontSize: 13)),
      onTap: () => showAboutDialog(
        context: context,
        applicationName: 'TTYSI_FIT',
        applicationVersion: appVersion,
        applicationLegalese: s.t('settings.legalese'),
      ),
    );
  }
}

/// appVersion — pubspec dagi versiya bilan qo'lda mos yuritiladi.
/// (package_info_plus qo'shilsa avtomatik olinadi.)
const appVersion = '1.0.0';

/// _SettingsTile — bir xil ko'rinishdagi sozlama qatori.
class _SettingsTile extends StatelessWidget {
  const _SettingsTile({
    required this.icon,
    required this.title,
    required this.onTap,
    this.trailing,
  });
  final IconData icon;
  final String title;
  final VoidCallback onTap;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Theme.of(context).colorScheme.surface,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          // 56dp — barmoq uchun qulay tegish maydoni (CLAUDE.md §6.3).
          constraints: const BoxConstraints(minHeight: 56),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          child: Row(
            children: [
              Icon(icon, size: 20, color: AppColors.muted),
              const SizedBox(width: 14),
              Expanded(child: Text(title, style: const TextStyle(fontSize: 15))),
              if (trailing != null) trailing!,
              const SizedBox(width: 6),
              const Icon(Icons.chevron_right, size: 18, color: AppColors.muted),
            ],
          ),
        ),
      ),
    );
  }
}
