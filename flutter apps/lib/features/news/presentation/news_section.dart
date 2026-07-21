import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/news_providers.dart';
import '../data/news_models.dart';

/// NewsSection — bosh sahifadagi yangiliklar bo'limi.
///
/// Alohida tab EMAS: pastki panelda allaqachon 5 ta element (Material
/// maksimumi). Yangilik bosh sahifada tabiiy joyda turadi — foydalanuvchi
/// ilovani ochishi bilan ko'radi.
class NewsSection extends ConsumerWidget {
  const NewsSection({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final news = ref.watch(newsListProvider);

    return news.when(
      // Yangilik yordamchi kontent — yuklanayotgani yoki yiqilgani bosh
      // sahifaning asosiy qismini (qadam, reyting) bezovta qilmasin.
      loading: () => const SizedBox.shrink(),
      error: (_, __) => const SizedBox.shrink(),
      data: (items) {
        if (items.isEmpty) return const SizedBox.shrink();
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 20),
            Text(s.t('news.title'),
                style: Theme.of(context).textTheme.titleMedium),
            const SizedBox(height: 10),
            for (final n in items) ...[
              _NewsCard(item: n),
              const SizedBox(height: 10),
            ],
          ],
        );
      },
    );
  }
}

class _NewsCard extends StatelessWidget {
  const _NewsCard({required this.item});
  final NewsItem item;

  @override
  Widget build(BuildContext context) {
    final n = item;

    return GestureDetector(
      onTap: () => context.push('/news/${n.id}'),
      child: Container(
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          borderRadius: BorderRadius.circular(16),
        ),
        clipBehavior: Clip.antiAlias,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (n.coverUrl.isNotEmpty)
              Image.network(
                n.coverUrl,
                height: 140,
                width: double.infinity,
                fit: BoxFit.cover,
                errorBuilder: (_, __, ___) => const SizedBox.shrink(),
              ),
            Padding(
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (n.pinned) ...[
                        const Icon(Icons.push_pin,
                            size: 14, color: AppColors.accent),
                        const SizedBox(width: 5),
                      ],
                      Expanded(
                        child: Text(n.title,
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                                fontWeight: FontWeight.w700, fontSize: 15)),
                      ),
                    ],
                  ),
                  if (n.excerpt.isNotEmpty) ...[
                    const SizedBox(height: 5),
                    Text(n.excerpt,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                            color: AppColors.muted, fontSize: 13)),
                  ],
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      if (n.publishedAt != null) ...[
                        Text(_fmtDate(n.publishedAt!),
                            style: const TextStyle(
                                fontSize: 11, color: AppColors.muted)),
                        const SizedBox(width: 10),
                      ],
                      const Icon(Icons.visibility_outlined,
                          size: 12, color: AppColors.muted),
                      const SizedBox(width: 3),
                      Text('${n.views}',
                          style: const TextStyle(
                              fontSize: 11, color: AppColors.muted)),
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
}

String _fmtDate(DateTime d) =>
    '${d.day.toString().padLeft(2, '0')}.${d.month.toString().padLeft(2, '0')}.${d.year}';
