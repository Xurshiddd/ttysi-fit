import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/rating_providers.dart';
import '../data/rating_models.dart';

/// Joriy tanlangan filtr (kesim + davr).
final ratingFilterProvider =
    StateProvider.autoDispose<RatingFilter>((ref) => const RatingFilter());

/// RatingTab — reyting ro'yxati: kesim va davr chiplari bilan.
class RatingTab extends ConsumerWidget {
  const RatingTab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final filter = ref.watch(ratingFilterProvider);
    final list = ref.watch(ratingListProvider(filter));

    final types = [
      ('student', s.t('rating.students')),
      ('employee', s.t('rating.employees')),
      ('group', s.t('rating.groups')),
      ('faculty', s.t('rating.faculties')),
    ];
    final periods = [
      ('week', s.t('rating.week')),
      ('month', s.t('rating.month')),
      ('all', s.t('rating.all')),
    ];

    final isIndividual =
        filter.type == 'student' || filter.type == 'employee';

    return Column(
      children: [
        // Kesim chiplari
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
          child: Row(
            children: [
              for (final (value, label) in types) ...[
                ChoiceChip(
                  label: Text(label),
                  selected: filter.type == value,
                  selectedColor: AppColors.primary,
                  labelStyle: TextStyle(
                      color: filter.type == value
                          ? Colors.white
                          : null),
                  onSelected: (_) => ref
                      .read(ratingFilterProvider.notifier)
                      .state = filter.copyWith(type: value),
                ),
                const SizedBox(width: 8),
              ],
            ],
          ),
        ),
        // Davr chiplari
        SingleChildScrollView(
          scrollDirection: Axis.horizontal,
          padding: const EdgeInsets.fromLTRB(16, 4, 16, 8),
          child: Row(
            children: [
              for (final (value, label) in periods) ...[
                ChoiceChip(
                  label: Text(label),
                  selected: filter.period == value,
                  selectedColor: AppColors.accent,
                  labelStyle: TextStyle(
                      color: filter.period == value
                          ? Colors.white
                          : null),
                  onSelected: (_) => ref
                      .read(ratingFilterProvider.notifier)
                      .state = filter.copyWith(period: value),
                ),
                const SizedBox(width: 8),
              ],
            ],
          ),
        ),
        Expanded(
          child: list.when(
            loading: () =>
                const Center(child: CircularProgressIndicator()),
            error: (e, _) => Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(s.t('common.error'),
                      style: const TextStyle(color: AppColors.muted)),
                  TextButton(
                      onPressed: () =>
                          ref.invalidate(ratingListProvider(filter)),
                      child: Text(s.t('common.retry'))),
                ],
              ),
            ),
            data: (rows) => rows.isEmpty
                ? Center(
                    child: Text(s.t('common.empty'),
                        style:
                            const TextStyle(color: AppColors.muted)))
                : RefreshIndicator(
                    onRefresh: () async =>
                        ref.invalidate(ratingListProvider(filter)),
                    child: ListView.separated(
                      padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
                      itemCount: rows.length,
                      separatorBuilder: (_, __) =>
                          const SizedBox(height: 8),
                      itemBuilder: (ctx, i) => _RatingTile(
                          entry: rows[i],
                          isIndividual: isIndividual,
                          s: s),
                    ),
                  ),
          ),
        ),
      ],
    );
  }
}

class _RatingTile extends StatelessWidget {
  const _RatingTile(
      {required this.entry, required this.isIndividual, required this.s});
  final RatingEntry entry;
  final bool isIndividual;
  final S s;

  Color _rankColor(int rank) {
    switch (rank) {
      case 1:
        return const Color(0xFFF59E0B); // oltin
      case 2:
        return const Color(0xFF94A3B8); // kumush
      case 3:
        return const Color(0xFFB45309); // bronza
      default:
        return AppColors.muted;
    }
  }

  @override
  Widget build(BuildContext context) {
    final subtitle = isIndividual
        ? [entry.groupName, entry.facultyName]
            .where((e) => e.isNotEmpty)
            .join(' · ')
        : entry.facultyName;
    final value = isIndividual
        ? _fmt(entry.totalSteps)
        : _fmt(entry.avgSteps.round());
    final valueLabel = isIndividual
        ? s.t('activity.steps')
        : s.t('rating.avgSteps');

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(14),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withValues(alpha: 0.03),
              blurRadius: 8,
              offset: const Offset(0, 2)),
        ],
      ),
      child: Row(
        children: [
          SizedBox(
            width: 34,
            child: entry.rank <= 3
                ? Icon(Icons.emoji_events,
                    color: _rankColor(entry.rank), size: 26)
                : Text('${entry.rank}',
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                        fontWeight: FontWeight.w700,
                        color: AppColors.muted)),
          ),
          const SizedBox(width: 8),
          if (isIndividual)
            CircleAvatar(
              radius: 20,
              backgroundColor: AppColors.primary.withValues(alpha: 0.1),
              backgroundImage: entry.avatarUrl.isNotEmpty
                  ? NetworkImage(entry.avatarUrl)
                  : null,
              child: entry.avatarUrl.isEmpty
                  ? Text(
                      entry.name.isNotEmpty
                          ? entry.name[0].toUpperCase()
                          : '?',
                      style: const TextStyle(
                          color: AppColors.primary,
                          fontWeight: FontWeight.w700))
                  : null,
            )
          else
            CircleAvatar(
              radius: 20,
              backgroundColor: AppColors.accent.withValues(alpha: 0.12),
              child: const Icon(Icons.groups,
                  color: AppColors.accent, size: 22),
            ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(entry.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style:
                        const TextStyle(fontWeight: FontWeight.w600)),
                if (subtitle.isNotEmpty)
                  Text(subtitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                          color: AppColors.muted, fontSize: 12)),
              ],
            ),
          ),
          const SizedBox(width: 8),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(value,
                  style: const TextStyle(
                      fontWeight: FontWeight.w700, fontSize: 15)),
              Text(valueLabel,
                  style: const TextStyle(
                      color: AppColors.muted, fontSize: 11)),
            ],
          ),
        ],
      ),
    );
  }
}

String _fmt(int n) {
  final str = n.toString();
  final buf = StringBuffer();
  for (var i = 0; i < str.length; i++) {
    if (i > 0 && (str.length - i) % 3 == 0) buf.write(' ');
    buf.write(str[i]);
  }
  return buf.toString();
}
