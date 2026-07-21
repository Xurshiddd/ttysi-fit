import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/health_permission_controller.dart';

/// showHealthPermissionSheet — ruxsat so'rashdan OLDINGI tushuntirish.
///
/// NEGA TIZIM OYNASINI DARROV CHIQARMAYMIZ ("priming"): Health Connect
/// ruxsati bir marta rad etilsa, Android uni qayta so'rashga yo'l qo'ymaydi —
/// foydalanuvchi sozlamalardan qo'lda yoqishi kerak bo'ladi. Sababi
/// tushunarsiz oyna esa ko'pincha rad etiladi. Shuning uchun avval NIMA
/// UCHUN kerakligini aytamiz, keyin tizim oynasini ochamiz.
Future<void> showHealthPermissionSheet(BuildContext context, WidgetRef ref) {
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    backgroundColor: Theme.of(context).colorScheme.surface,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    builder: (_) => const _HealthPermissionSheet(),
  );
}

class _HealthPermissionSheet extends ConsumerStatefulWidget {
  const _HealthPermissionSheet();

  @override
  ConsumerState<_HealthPermissionSheet> createState() =>
      _HealthPermissionSheetState();
}

class _HealthPermissionSheetState
    extends ConsumerState<_HealthPermissionSheet> {
  bool _busy = false;

  Future<void> _allow() async {
    setState(() => _busy = true);
    final result = await ref.read(healthPermissionProvider.notifier).request();
    if (!mounted) return;

    setState(() => _busy = false);
    Navigator.of(context).pop();

    final s = S.of(context);
    final msg = switch (result) {
      HealthPermissionResult.granted => s.t('health.granted'),
      HealthPermissionResult.denied => s.t('activity.syncNoPerm'),
      HealthPermissionResult.permanentlyDenied => s.t('health.deniedHint'),
    };
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(msg)));
  }

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);

    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(24, 8, 24, 24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(
              child: Container(
                height: 64,
                width: 64,
                decoration: BoxDecoration(
                  color: AppColors.accent.withValues(alpha: 0.12),
                  shape: BoxShape.circle,
                ),
                child: const Icon(Icons.directions_walk,
                    color: AppColors.accent, size: 34),
              ),
            ),
            const SizedBox(height: 16),
            Text(
              s.t('health.promptTitle'),
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    fontWeight: FontWeight.w700,
                  ),
            ),
            const SizedBox(height: 10),
            Text(
              s.t('health.promptBody'),
              style: const TextStyle(color: AppColors.muted, height: 1.45),
            ),
            const SizedBox(height: 14),

            // Maxfiylik — ruxsat so'raganda eng ko'p beriladigan savol.
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.lock_outline,
                    size: 18, color: AppColors.muted),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    s.t('health.promptPrivacy'),
                    style: const TextStyle(
                        color: AppColors.muted, fontSize: 12, height: 1.4),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 22),

            SizedBox(
              width: double.infinity,
              child: FilledButton(
                style: FilledButton.styleFrom(
                  backgroundColor: AppColors.accent,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                ),
                onPressed: _busy ? null : _allow,
                child: _busy
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white))
                    : Text(s.t('health.allow')),
              ),
            ),
            const SizedBox(height: 6),
            SizedBox(
              width: double.infinity,
              child: TextButton(
                onPressed:
                    _busy ? null : () => Navigator.of(context).pop(),
                child: Text(s.t('health.later')),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// HealthPermissionCard — bosh sahifadagi eslatma.
///
/// Tushuntirish oynasida "Keyinroq" bosgan yoki ruxsatni keyin o'chirgan
/// foydalanuvchi uchun TIKLASH yo'li. Ruxsat berilgan bo'lsa butunlay
/// yashirinadi — kerak bo'lmagan joyni egallamaydi.
class HealthPermissionCard extends ConsumerWidget {
  const HealthPermissionCard({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final granted = ref.watch(healthPermissionProvider);

    // Yuklanayotganda yoki ruxsat bor bo'lsa — hech narsa ko'rsatmaymiz.
    // Xato bo'lsa ham jim: karta ilovaning asosiy oqimi emas.
    if (granted.valueOrNull != false) return const SizedBox.shrink();

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.accent.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.accent.withValues(alpha: 0.3)),
      ),
      child: Row(
        children: [
          const Icon(Icons.directions_walk, color: AppColors.accent, size: 30),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(s.t('health.cardTitle'),
                    style: const TextStyle(fontWeight: FontWeight.w600)),
                const SizedBox(height: 2),
                Text(s.t('health.cardBody'),
                    style: const TextStyle(
                        color: AppColors.muted, fontSize: 12, height: 1.35)),
              ],
            ),
          ),
          const SizedBox(width: 8),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: AppColors.accent,
              // Barmoq uchun qulay nishon (min 48dp — CLAUDE.md §6.3).
              minimumSize: const Size(0, 44),
              padding: const EdgeInsets.symmetric(horizontal: 16),
            ),
            onPressed: () => showHealthPermissionSheet(context, ref),
            child: Text(s.t('health.enable')),
          ),
        ],
      ),
    );
  }
}
