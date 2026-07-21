import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/training_providers.dart';
import '../data/training_models.dart';

/// TrainingList — video mashg'ulotlar (Faollik tabining "Mashg'ulot" segmenti).
///
/// Kontent to'liq admin panel boshqaruvida (§16): kategoriyalar ham backenddan
/// keladi, bu yerda ro'yxat qattiq yozilmagan.
class TrainingList extends ConsumerWidget {
  const TrainingList({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final list = ref.watch(trainingListProvider);

    return Column(
      children: [
        const _Filters(),
        Expanded(
          child: RefreshIndicator(
            onRefresh: () async {
              ref.invalidate(trainingListProvider);
              ref.invalidate(trainingCategoriesProvider);
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
                    onPressed: () => ref.invalidate(trainingListProvider),
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
                      const Icon(Icons.ondemand_video_outlined,
                          size: 56, color: AppColors.muted),
                      const SizedBox(height: 12),
                      Text(s.t('training.empty'),
                          textAlign: TextAlign.center,
                          style: const TextStyle(color: AppColors.muted)),
                    ],
                  );
                }
                return ListView.separated(
                  padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
                  itemCount: items.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 12),
                  itemBuilder: (_, i) => _TrainingCard(training: items[i]),
                );
              },
            ),
          ),
        ),
      ],
    );
  }
}

/// _Filters — kategoriya (backenddan) + daraja (enum).
class _Filters extends ConsumerWidget {
  const _Filters();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final filter = ref.watch(trainingFilterProvider);
    final cats = ref.watch(trainingCategoriesProvider);

    return SizedBox(
      height: 44,
      child: ListView(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
        children: [
          // Kategoriya: "Barchasi" + backenddan kelganlar.
          // Kalitlar (ValueKey) — chip matni karta ichidagi kategoriya yozuvi
          // bilan bir xil bo'lishi mumkin, shuning uchun testlar chipni matn
          // bo'yicha emas, kalit bo'yicha topadi.
          _Chip(
            key: const ValueKey('training-cat-'),
            label: s.t('common.all'),
            selected: filter.category.isEmpty,
            onTap: () =>
                ref.read(trainingFilterProvider.notifier).setCategory(''),
          ),
          ...cats.maybeWhen(
            data: (list) => list.map((c) => _Chip(
                  key: ValueKey('training-cat-$c'),
                  label: c,
                  selected: filter.category == c,
                  onTap: () =>
                      ref.read(trainingFilterProvider.notifier).setCategory(c),
                )),
            orElse: () => const <Widget>[],
          ),
          const SizedBox(width: 8),
          Container(width: 1, margin: const EdgeInsets.symmetric(vertical: 8),
              color: AppColors.muted.withValues(alpha: 0.2)),
          const SizedBox(width: 8),
          // Daraja: enum (chegaralangan shkala).
          for (final l in const ['beginner', 'intermediate', 'advanced'])
            _Chip(
              key: ValueKey('training-level-$l'),
              label: s.t('training.level.$l'),
              selected: filter.level == l,
              onTap: () => ref
                  .read(trainingFilterProvider.notifier)
                  .setLevel(filter.level == l ? '' : l),
            ),
        ],
      ),
    );
  }
}

class _Chip extends StatelessWidget {
  const _Chip(
      {super.key,
      required this.label,
      required this.selected,
      required this.onTap});
  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: GestureDetector(
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 14),
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: selected
                ? AppColors.accent.withValues(alpha: 0.15)
                : Colors.transparent,
            border: Border.all(
                color: selected
                    ? AppColors.accent
                    : AppColors.muted.withValues(alpha: 0.3)),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Text(label,
              style: TextStyle(
                  fontSize: 13,
                  fontWeight: selected ? FontWeight.w600 : FontWeight.w400,
                  color: selected ? AppColors.accent : AppColors.muted)),
        ),
      ),
    );
  }
}

class _TrainingCard extends StatelessWidget {
  const _TrainingCard({required this.training});
  final Training training;

  /// _open — videoni tashqi ilovada ochadi (YouTube/brauzer).
  /// Ilovada video pleyer yo'q: darsliklar tashqi xostingda turadi.
  Future<void> _open(BuildContext context) async {
    final s = S.of(context);
    final uri = Uri.tryParse(training.videoUrl);
    final ok = uri != null &&
        await launchUrl(uri, mode: LaunchMode.externalApplication);
    if (!ok && context.mounted) {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(s.t('training.openError'))));
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    final t = training;

    return GestureDetector(
      onTap: () => _open(context),
      child: Container(
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: BorderRadius.circular(16),
        ),
        clipBehavior: Clip.antiAlias,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Thumbnail + play tugmasi. Rasm bo'lmasa ham play ko'rinadi.
            Stack(
              alignment: Alignment.center,
              children: [
                if (t.thumbnailUrl.isNotEmpty)
                  Image.network(
                    t.thumbnailUrl,
                    height: 150,
                    width: double.infinity,
                    fit: BoxFit.cover,
                    errorBuilder: (_, __, ___) => _placeholder(context),
                  )
                else
                  _placeholder(context),
                Container(
                  width: 52,
                  height: 52,
                  decoration: BoxDecoration(
                    color: Colors.black.withValues(alpha: 0.55),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.play_arrow,
                      color: Colors.white, size: 32),
                ),
                if (t.durationLabel.isNotEmpty)
                  Positioned(
                    right: 8,
                    bottom: 8,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.black.withValues(alpha: 0.7),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Text(t.durationLabel,
                          style: const TextStyle(
                              color: Colors.white, fontSize: 11)),
                    ),
                  ),
              ],
            ),
            Padding(
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(t.title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                          fontWeight: FontWeight.w700, fontSize: 15)),
                  if (t.description.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(t.description,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                            color: AppColors.muted, fontSize: 13)),
                  ],
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 12,
                    runSpacing: 4,
                    children: [
                      if (t.category.isNotEmpty)
                        _Meta(icon: Icons.category_outlined, text: t.category),
                      _Meta(
                          icon: Icons.signal_cellular_alt,
                          text: s.t('training.level.${t.level}')),
                      _Meta(
                          icon: Icons.visibility_outlined,
                          text: '${t.views}'),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _placeholder(BuildContext context) => Container(
        height: 150,
        width: double.infinity,
        color: AppColors.primary.withValues(alpha: 0.25),
      );
}

class _Meta extends StatelessWidget {
  const _Meta({required this.icon, required this.text});
  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 13, color: AppColors.muted),
        const SizedBox(width: 4),
        Text(text,
            style: const TextStyle(fontSize: 12, color: AppColors.muted)),
      ],
    );
  }
}
