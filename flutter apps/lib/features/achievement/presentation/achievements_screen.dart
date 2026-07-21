import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:open_filex/open_filex.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/achievement_providers.dart';
import '../data/achievement_models.dart';
import '../data/achievement_repository.dart';

/// AchievementsScreen — barcha yutuqlar: qozonilganlar tepada, qolganlari
/// progress bilan pastda.
class AchievementsScreen extends ConsumerWidget {
  const AchievementsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final list = ref.watch(achievementListProvider);

    return Scaffold(
      appBar: AppBar(title: Text(s.t('achievement.title'))),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(achievementListProvider);
          ref.invalidate(earnedAchievementsProvider);
        },
        child: list.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (_, __) => ListView(
            padding: const EdgeInsets.all(24),
            children: [
              const SizedBox(height: 60),
              Text(s.t('common.error'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(color: AppColors.muted)),
              TextButton(
                onPressed: () => ref.invalidate(achievementListProvider),
                child: Text(s.t('common.retry')),
              ),
            ],
          ),
          data: (items) {
            if (items.isEmpty) {
              return ListView(
                padding: const EdgeInsets.all(24),
                children: [
                  const SizedBox(height: 60),
                  const Icon(Icons.emoji_events_outlined,
                      size: 56, color: AppColors.muted),
                  const SizedBox(height: 12),
                  Text(s.t('achievement.empty'),
                      textAlign: TextAlign.center,
                      style: const TextStyle(color: AppColors.muted)),
                ],
              );
            }

            // Qozonilganlar tepada: foydalanuvchi avval o'z yutug'ini ko'rsin.
            final earned = items.where((a) => a.earned).toList();
            final rest = items.where((a) => !a.earned).toList();

            return ListView(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
              children: [
                if (earned.isNotEmpty) ...[
                  _SectionTitle(
                      text: '${s.t('achievement.earned')} (${earned.length})'),
                  ...earned.map((a) => _AchievementCard(achievement: a)),
                  const SizedBox(height: 8),
                ],
                if (rest.isNotEmpty) ...[
                  _SectionTitle(text: s.t('achievement.inProgress')),
                  ...rest.map((a) => _AchievementCard(achievement: a)),
                ],
              ],
            );
          },
        ),
      ),
    );
  }
}

class _SectionTitle extends StatelessWidget {
  const _SectionTitle({required this.text});
  final String text;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.fromLTRB(2, 12, 2, 8),
        child: Text(text,
            style: const TextStyle(fontWeight: FontWeight.w700, fontSize: 15)),
      );
}

class _AchievementCard extends ConsumerStatefulWidget {
  const _AchievementCard({required this.achievement});
  final Achievement achievement;

  @override
  ConsumerState<_AchievementCard> createState() => _AchievementCardState();
}

class _AchievementCardState extends ConsumerState<_AchievementCard> {
  bool _downloading = false;

  /// _openCertificate — PDF ni yuklab olib tizim ko'ruvchisida ochadi.
  ///
  /// URL'ni to'g'ridan-to'g'ri brauzerga bera olmaymiz: endpoint Authorization
  /// talab qiladi. Shuning uchun avval Dio bilan yuklab, keyin faylni ochamiz.
  Future<void> _openCertificate() async {
    final s = S.of(context);
    final id = widget.achievement.awardId;
    if (id == null || _downloading) return;

    setState(() => _downloading = true);
    try {
      final file = await ref
          .read(achievementRepositoryProvider)
          .downloadCertificate(id);
      final res = await OpenFilex.open(file.path);
      if (res.type != ResultType.done && mounted) {
        // PDF ko'ruvchi o'rnatilmagan bo'lishi mumkin — fayl saqlanganini
        // aytamiz, foydalanuvchi uni o'zi topa oladi.
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(s.t('achievement.certificateNoViewer'))),
        );
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(s.t('achievement.certificateError'))),
        );
      }
    } finally {
      if (mounted) setState(() => _downloading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    final a = widget.achievement;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: a.earned
            ? Border.all(color: AppColors.accent.withValues(alpha: 0.5))
            : null,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _Medal(earned: a.earned),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(a.title,
                        style: const TextStyle(
                            fontWeight: FontWeight.w700, fontSize: 15)),
                    if (a.description.isNotEmpty) ...[
                      const SizedBox(height: 2),
                      Text(a.description,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                              color: AppColors.muted, fontSize: 13)),
                    ],
                  ],
                ),
              ),
              if (a.rewardCoins > 0)
                Padding(
                  padding: const EdgeInsets.only(left: 8),
                  child: Text('+${a.rewardCoins}',
                      style: const TextStyle(
                          color: AppColors.accent,
                          fontWeight: FontWeight.w700,
                          fontSize: 13)),
                ),
            ],
          ),

          // Progress faqat maqsadi bor va hali qozonilmagan yutuqda.
          if (!a.earned && a.target > 0) ...[
            const SizedBox(height: 12),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: (a.progressPct / 100).clamp(0.0, 1.0),
                minHeight: 6,
                backgroundColor: AppColors.muted.withValues(alpha: 0.2),
                valueColor:
                    const AlwaysStoppedAnimation<Color>(AppColors.accent),
              ),
            ),
            const SizedBox(height: 6),
            Text(a.progressLabel,
                style: const TextStyle(color: AppColors.muted, fontSize: 12)),
          ],

          if (a.earned) ...[
            const SizedBox(height: 10),
            Row(
              children: [
                const Icon(Icons.check_circle,
                    size: 16, color: AppColors.accent),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    a.earnedAt != null
                        ? '${s.t('achievement.earnedAt')} ${_fmtDate(a.earnedAt!)}'
                        : s.t('achievement.earned'),
                    style: const TextStyle(
                        color: AppColors.accent,
                        fontSize: 12,
                        fontWeight: FontWeight.w600),
                  ),
                ),
              ],
            ),
            if (a.hasCertificate) ...[
              const SizedBox(height: 8),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: _downloading ? null : _openCertificate,
                  icon: _downloading
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.picture_as_pdf_outlined, size: 18),
                  label: Text(_downloading
                      ? s.t('common.loading')
                      : s.t('achievement.certificate')),
                ),
              ),
            ],
          ],
        ],
      ),
    );
  }

  String _fmtDate(DateTime d) =>
      '${d.day.toString().padLeft(2, '0')}.${d.month.toString().padLeft(2, '0')}.${d.year}';
}

/// _Medal — qozonilgan yutuqda rangli, qolganida so'nik medal.
class _Medal extends StatelessWidget {
  const _Medal({required this.earned});
  final bool earned;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 44,
      height: 44,
      decoration: BoxDecoration(
        color: earned
            ? AppColors.accent.withValues(alpha: 0.15)
            : AppColors.muted.withValues(alpha: 0.12),
        shape: BoxShape.circle,
      ),
      child: Icon(
        earned ? Icons.emoji_events : Icons.emoji_events_outlined,
        color: earned ? AppColors.accent : AppColors.muted,
        size: 24,
      ),
    );
  }
}
