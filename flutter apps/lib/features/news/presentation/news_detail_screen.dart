import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/news_providers.dart';

/// NewsDetailScreen — yangilikning to'liq matni.
/// Ochilishi backend'da ko'rishlar sonini oshiradi.
class NewsDetailScreen extends ConsumerWidget {
  const NewsDetailScreen({super.key, required this.id});
  final String id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final detail = ref.watch(newsDetailProvider(id));

    return Scaffold(
      appBar: AppBar(title: Text(s.t('news.title'))),
      body: detail.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(s.t('common.error'),
                  style: const TextStyle(color: AppColors.muted)),
              TextButton(
                onPressed: () => ref.invalidate(newsDetailProvider(id)),
                child: Text(s.t('common.retry')),
              ),
            ],
          ),
        ),
        data: (n) => ListView(
          padding: EdgeInsets.zero,
          children: [
            if (n.coverUrl.isNotEmpty)
              // errorBuilder: rasm yuklanmasa (tarmoq/eski URL) ekran
              // buzilmasin — shunchaki rasmsiz ko'rsatiladi.
              Image.network(
                n.coverUrl,
                height: 200,
                width: double.infinity,
                fit: BoxFit.cover,
                errorBuilder: (_, __, ___) => const SizedBox.shrink(),
              ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(n.title,
                      style: Theme.of(context)
                          .textTheme
                          .titleLarge
                          ?.copyWith(fontWeight: FontWeight.w800)),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      if (n.publishedAt != null) ...[
                        const Icon(Icons.event,
                            size: 14, color: AppColors.muted),
                        const SizedBox(width: 4),
                        Text(_fmtDate(n.publishedAt!),
                            style: const TextStyle(
                                fontSize: 12, color: AppColors.muted)),
                        const SizedBox(width: 12),
                      ],
                      const Icon(Icons.visibility_outlined,
                          size: 14, color: AppColors.muted),
                      const SizedBox(width: 4),
                      Text('${n.views}',
                          style: const TextStyle(
                              fontSize: 12, color: AppColors.muted)),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Text(n.body,
                      style: const TextStyle(fontSize: 15, height: 1.55)),
                  const SizedBox(height: 24),
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
