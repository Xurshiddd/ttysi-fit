import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/competition_providers.dart';
import '../data/competition_models.dart';

/// CompetitionList — musobaqalar ro'yxati (Tadbirlar tabining "Musobaqa" segmenti).
///
/// Kontent to'liq admin panel boshqaruvida (§16): bu yerda hech qanday musobaqa
/// nomi yoki sport turi qattiq yozilmagan.
class CompetitionList extends ConsumerWidget {
  const CompetitionList({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final list = ref.watch(competitionListProvider);

    return RefreshIndicator(
      onRefresh: () async => ref.invalidate(competitionListProvider),
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
              onPressed: () => ref.invalidate(competitionListProvider),
              child: Text(s.t('common.retry')),
            ),
          ],
        ),
        data: (items) {
          if (items.isEmpty) {
            return ListView(
              padding: const EdgeInsets.all(24),
              children: [
                const SizedBox(height: 80),
                const Icon(Icons.emoji_events_outlined,
                    size: 56, color: AppColors.muted),
                const SizedBox(height: 12),
                Text(s.t('comp.empty'),
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: AppColors.muted)),
              ],
            );
          }
          return ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: items.length,
            separatorBuilder: (_, __) => const SizedBox(height: 12),
            itemBuilder: (_, i) => _CompetitionCard(competition: items[i]),
          );
        },
      ),
    );
  }
}

class _CompetitionCard extends ConsumerWidget {
  const _CompetitionCard({required this.competition});
  final Competition competition;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final c = competition;
    final busy = ref.watch(competitionRegProvider).isLoading;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: c.registered
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
              const SizedBox(width: 8),
              _StatusChip(status: c.status),
            ],
          ),
          if (c.description.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(c.description,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(color: AppColors.muted, fontSize: 13)),
          ],
          const SizedBox(height: 10),

          // Meta qatorlari — bo'sh maydonlar ko'rsatilmaydi.
          Wrap(
            spacing: 14,
            runSpacing: 6,
            children: [
              if (c.sport.isNotEmpty)
                _Meta(icon: Icons.sports_soccer, text: c.sport),
              if (c.startsAt != null)
                _Meta(icon: Icons.event, text: _fmtDate(c.startsAt!)),
              if (c.location.isNotEmpty)
                _Meta(icon: Icons.place_outlined, text: c.location),
              _Meta(
                icon: Icons.group_outlined,
                text: c.slotsLabel,
                color: c.isFull ? AppColors.danger : null,
              ),
              if (c.rewardCoins > 0)
                _Meta(
                    icon: Icons.monetization_on_outlined,
                    text: '+${c.rewardCoins}',
                    color: const Color(0xFFF59E0B)),
            ],
          ),
          const SizedBox(height: 12),

          Row(
            children: [
              Expanded(
                child: c.place == null
                    ? const SizedBox.shrink()
                    : Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.emoji_events,
                              size: 16, color: Color(0xFFF59E0B)),
                          const SizedBox(width: 4),
                          Flexible(
                            child: Text('${c.place}-${s.t('comp.place')}',
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                    fontSize: 12,
                                    fontWeight: FontWeight.w700,
                                    color: Color(0xFFF59E0B))),
                          ),
                        ],
                      ),
              ),
              const SizedBox(width: 8),
              _ActionButton(competition: c, busy: busy),
            ],
          ),
        ],
      ),
    );
  }
}

/// _ActionButton — holатga qarab: yozilish / bekor qilish / yopiq.
class _ActionButton extends ConsumerWidget {
  const _ActionButton({required this.competition, required this.busy});
  final Competition competition;
  final bool busy;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final c = competition;

    Future<void> act(Future<RegResult> Function() run) async {
      final r = await run();
      if (!context.mounted) return;
      final msg = switch (r) {
        RegResult.success => s.t('comp.registered'),
        RegResult.cancelled => s.t('comp.cancelled'),
        RegResult.alreadyRegistered => s.t('comp.already'),
        RegResult.unavailable => s.t('comp.unavailable'),
        RegResult.error => s.t('common.error'),
      };
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    }

    if (c.registered) {
      return OutlinedButton(
        style: OutlinedButton.styleFrom(
          foregroundColor: AppColors.danger,
          side: const BorderSide(color: AppColors.danger),
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
          minimumSize: const Size(0, 36),
        ),
        onPressed: busy
            ? null
            : () => act(() =>
                ref.read(competitionRegProvider.notifier).cancel(c.id)),
        child: Text(s.t('comp.cancel')),
      );
    }

    // Yozilish yopiq bo'lsa tugma o'rniga sabab (joy tugagan / muddat / holat).
    if (!c.regOpen) {
      return Text(
        c.isFull ? s.t('comp.full') : s.t('comp.regClosed'),
        style: const TextStyle(color: AppColors.muted, fontSize: 13),
      );
    }

    return FilledButton(
      style: FilledButton.styleFrom(
        backgroundColor: AppColors.accent,
        padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 8),
      ),
      onPressed: busy
          ? null
          : () => act(
              () => ref.read(competitionRegProvider.notifier).register(c.id)),
      child: Text(s.t('comp.register')),
    );
  }
}

class _StatusChip extends StatelessWidget {
  const _StatusChip({required this.status});
  final String status;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    final (color, key) = switch (status) {
      'registration' => (AppColors.accent, 'comp.status.registration'),
      'ongoing' => (AppColors.primary, 'comp.status.ongoing'),
      'finished' => (AppColors.muted, 'comp.status.finished'),
      _ => (AppColors.muted, 'comp.status.draft'),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(s.t(key),
          style: TextStyle(
              color: color, fontSize: 11, fontWeight: FontWeight.w600)),
    );
  }
}

class _Meta extends StatelessWidget {
  const _Meta({required this.icon, required this.text, this.color});
  final IconData icon;
  final String text;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final c = color ?? AppColors.muted;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: c),
        const SizedBox(width: 4),
        Text(text, style: TextStyle(fontSize: 12, color: c)),
      ],
    );
  }
}

String _fmtDate(DateTime d) =>
    '${d.day.toString().padLeft(2, '0')}.${d.month.toString().padLeft(2, '0')}.${d.year}';
