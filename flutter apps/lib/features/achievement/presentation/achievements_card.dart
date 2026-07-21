import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/achievement_providers.dart';
import '../data/achievement_models.dart';

/// AchievementsCard — profildagi "Yutuqlarim" kartasi: qozonilgan medallarning
/// qisqa ko'rinishi, bosilganda to'liq ekran ochiladi.
///
/// Xato yoki bo'sh holatda ham karta ko'rinadi (bosib ochish mumkin) — bosh
/// sahifadagi yangiliklardan farqli, chunki bu profil bo'limining doimiy
/// elementi va yo'qolib qolsa foydalanuvchi yutuqlarni topa olmaydi.
class AchievementsCard extends ConsumerWidget {
  const AchievementsCard({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final earned = ref.watch(earnedAchievementsProvider);

    return InkWell(
      onTap: () => context.push('/achievements'),
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.emoji_events, color: AppColors.accent, size: 20),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(s.t('achievement.title'),
                      style: const TextStyle(
                          fontWeight: FontWeight.w700, fontSize: 15)),
                ),
                earned.maybeWhen(
                  data: (list) => Text('${list.length}',
                      style: const TextStyle(
                          color: AppColors.accent,
                          fontWeight: FontWeight.w700,
                          fontSize: 15)),
                  orElse: () => const SizedBox.shrink(),
                ),
                const Icon(Icons.chevron_right, color: AppColors.muted),
              ],
            ),
            const SizedBox(height: 12),
            earned.when(
              loading: () => const SizedBox(
                height: 44,
                child: Center(
                  child: SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                ),
              ),
              error: (_, __) => Text(s.t('common.error'),
                  style: const TextStyle(color: AppColors.muted, fontSize: 13)),
              data: (list) => list.isEmpty
                  ? Text(s.t('achievement.emptyShort'),
                      style:
                          const TextStyle(color: AppColors.muted, fontSize: 13))
                  : _MedalRow(items: list),
            ),
          ],
        ),
      ),
    );
  }
}

/// _MedalRow — oxirgi qozonilgan yutuqlar (ko'pi bilan 5 ta) gorizontal qatorda.
class _MedalRow extends StatelessWidget {
  const _MedalRow({required this.items});
  final List<Achievement> items;

  @override
  Widget build(BuildContext context) {
    const maxShown = 5;
    final shown = items.take(maxShown).toList();
    final extra = items.length - shown.length;

    return SizedBox(
      height: 44,
      child: Row(
        children: [
          for (final a in shown)
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: Tooltip(
                message: a.title,
                child: Container(
                  width: 40,
                  height: 40,
                  decoration: BoxDecoration(
                    color: AppColors.accent.withValues(alpha: 0.15),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.emoji_events,
                      color: AppColors.accent, size: 20),
                ),
              ),
            ),
          if (extra > 0)
            Container(
              width: 40,
              height: 40,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: AppColors.muted.withValues(alpha: 0.12),
                shape: BoxShape.circle,
              ),
              child: Text('+$extra',
                  style: const TextStyle(
                      color: AppColors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w600)),
            ),
        ],
      ),
    );
  }
}
