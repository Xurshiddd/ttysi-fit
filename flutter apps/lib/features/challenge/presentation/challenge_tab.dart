import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../../fitcoin/application/fitcoin_providers.dart';
import '../application/challenge_providers.dart';
import '../data/challenge_models.dart';

/// ChallengeTab — aktiv chellenjlar ro'yxati.
///
/// Kontent to'liq admin panel boshqaruvida (§16): bu yerda hech qanday chellenj
/// nomi yoki maqsadi qattiq yozilmagan — hammasi backenddan keladi.
class ChallengeTab extends ConsumerWidget {
  const ChallengeTab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final list = ref.watch(challengeListProvider);

    return RefreshIndicator(
      onRefresh: () async => ref.invalidate(challengeListProvider),
      child: list.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => ListView(
          padding: const EdgeInsets.all(24),
          children: [
            const SizedBox(height: 80),
            Text(s.t('common.error'),
                textAlign: TextAlign.center,
                style: const TextStyle(color: AppColors.muted)),
            TextButton(
              onPressed: () => ref.invalidate(challengeListProvider),
              child: Text(s.t('common.retry')),
            ),
          ],
        ),
        data: (items) {
          if (items.isEmpty) {
            return ListView(
              padding: const EdgeInsets.all(24),
              children: [
                const SizedBox(height: 100),
                const Icon(Icons.flag_outlined,
                    size: 56, color: AppColors.muted),
                const SizedBox(height: 12),
                Text(s.t('challenge.empty'),
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: AppColors.muted)),
              ],
            );
          }
          return ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: items.length,
            separatorBuilder: (_, __) => const SizedBox(height: 12),
            itemBuilder: (_, i) => _ChallengeCard(challenge: items[i]),
          );
        },
      ),
    );
  }
}

/// _ClaimButton — mukofot olish. Backend idempotent (409 — allaqachon olingan),
/// shuning uchun takroriy bosish xavfsiz: ikkinchi coin berilmaydi.
class _ClaimButton extends ConsumerWidget {
  const _ClaimButton({required this.challengeId, required this.coins});
  final String challengeId;
  final int coins;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final claiming = ref.watch(claimRewardProvider).isLoading;

    return FilledButton.icon(
      style: FilledButton.styleFrom(
        backgroundColor: const Color(0xFFF59E0B),
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      ),
      icon: claiming
          ? const SizedBox(
              height: 14,
              width: 14,
              child: CircularProgressIndicator(
                  strokeWidth: 2, color: Colors.white))
          : const Icon(Icons.card_giftcard, size: 16),
      label: Text('${s.t('challenge.claim')} +$coins'),
      onPressed: claiming
          ? null
          : () async {
              final r = await ref
                  .read(claimRewardProvider.notifier)
                  .claim(challengeId);
              if (!context.mounted) return;
              final msg = switch (r) {
                ClaimResult.success => s.t('challenge.rewardOk'),
                ClaimResult.alreadyClaimed => s.t('challenge.rewardClaimed'),
                ClaimResult.notCompleted => s.t('challenge.notCompleted'),
                ClaimResult.error => s.t('common.error'),
              };
              ScaffoldMessenger.of(context)
                  .showSnackBar(SnackBar(content: Text(msg)));
            },
    );
  }
}

class _ChallengeCard extends ConsumerWidget {
  const _ChallengeCard({required this.challenge});
  final Challenge challenge;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final c = challenge;
    final joining = ref.watch(challengeJoinProvider).isLoading;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: c.completed
            ? Border.all(color: AppColors.accent.withValues(alpha: 0.5))
            : null,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Text(c.title,
                    style: const TextStyle(
                        fontWeight: FontWeight.w700, fontSize: 16)),
              ),
              if (c.completed)
                const Icon(Icons.check_circle,
                    color: AppColors.accent, size: 22)
              else if (c.rewardCoins > 0)
                Container(
                  padding: const EdgeInsets.symmetric(
                      horizontal: 8, vertical: 3),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF59E0B).withValues(alpha: 0.15),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text('+${c.rewardCoins}',
                      style: const TextStyle(
                          color: Color(0xFFF59E0B),
                          fontSize: 12,
                          fontWeight: FontWeight.w700)),
                ),
            ],
          ),
          if (c.description.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(c.description,
                style: const TextStyle(color: AppColors.muted, fontSize: 13)),
          ],
          const SizedBox(height: 12),

          // Progress — faqat maqsadli turlarda (custom da target=0).
          if (c.joined && c.target > 0) ...[
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(c.progressLabel,
                    style: const TextStyle(
                        fontSize: 12, fontWeight: FontWeight.w600)),
                Text('${c.progressPct.toStringAsFixed(0)}%',
                    style: const TextStyle(
                        fontSize: 12, color: AppColors.muted)),
              ],
            ),
            const SizedBox(height: 6),
            ClipRRect(
              borderRadius: BorderRadius.circular(6),
              child: LinearProgressIndicator(
                value: (c.progressPct / 100).clamp(0.0, 1.0),
                minHeight: 8,
                backgroundColor: AppColors.muted.withValues(alpha: 0.15),
                valueColor: AlwaysStoppedAnimation(
                    c.completed ? AppColors.accent : AppColors.primary),
              ),
            ),
            const SizedBox(height: 12),
          ],

          Row(
            children: [
              // Expanded + ellipsis: tor ekranda (360dp) "N kun qoldi" va
              // "Olish +25" tugmasi birga sig'masdi va Row toshib ketardi.
              Expanded(
                child: c.daysLeft == null
                    ? const SizedBox.shrink()
                    : Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.schedule,
                              size: 14, color: AppColors.muted),
                          const SizedBox(width: 4),
                          Flexible(
                            child: Text(
                                '${c.daysLeft} ${s.t('challenge.daysLeft')}',
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                    fontSize: 12, color: AppColors.muted)),
                          ),
                        ],
                      ),
              ),
              const SizedBox(width: 8),
              if (!c.joined)
                FilledButton(
                  style: FilledButton.styleFrom(
                    backgroundColor: AppColors.accent,
                    padding: const EdgeInsets.symmetric(
                        horizontal: 18, vertical: 8),
                  ),
                  onPressed: joining
                      ? null
                      : () async {
                          final ok = await ref
                              .read(challengeJoinProvider.notifier)
                              .join(c.id);
                          if (!context.mounted) return;
                          ScaffoldMessenger.of(context).showSnackBar(SnackBar(
                              content: Text(ok
                                  ? s.t('challenge.joined')
                                  : s.t('common.error'))));
                        },
                  child: Text(s.t('challenge.join')),
                )
              // Yakunlangan va mukofoti hali olinmagan — olish tugmasi.
              else if (c.canClaim)
                _ClaimButton(challengeId: c.id, coins: c.rewardCoins)
              else if (c.completed)
                Text(
                    c.rewardGranted
                        ? s.t('challenge.rewardClaimed')
                        : s.t('challenge.done'),
                    style: const TextStyle(
                        color: AppColors.accent,
                        fontWeight: FontWeight.w700,
                        fontSize: 13))
              else
                Text(s.t('challenge.inProgress'),
                    style: const TextStyle(
                        color: AppColors.muted, fontSize: 13)),
            ],
          ),
        ],
      ),
    );
  }
}
