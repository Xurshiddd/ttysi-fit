import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/reward_providers.dart';
import '../data/reward_models.dart';

/// showMyRedemptionsSheet — "Buyurtmalarim" oynasi.
Future<void> showMyRedemptionsSheet(BuildContext context) {
  return showModalBottomSheet<void>(
    context: context,
    isScrollControlled: true,
    showDragHandle: true,
    backgroundColor: Theme.of(context).colorScheme.surface,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    builder: (_) => const _MyRedemptionsSheet(),
  );
}

class _MyRedemptionsSheet extends ConsumerWidget {
  const _MyRedemptionsSheet();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final list = ref.watch(myRedemptionsProvider);

    return DraggableScrollableSheet(
      expand: false,
      initialChildSize: 0.6,
      maxChildSize: 0.9,
      builder: (_, scroll) => Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 4, 20, 12),
            child: Row(
              children: [
                Text(s.t('shop.myOrders'),
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.w700,
                        )),
              ],
            ),
          ),
          Expanded(
            child: list.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (_, __) => Center(
                child: Text(s.t('common.error'),
                    style: const TextStyle(color: AppColors.muted)),
              ),
              data: (rows) => rows.isEmpty
                  ? Center(
                      child: Padding(
                        padding: const EdgeInsets.all(24),
                        child: Text(s.t('shop.noOrders'),
                            textAlign: TextAlign.center,
                            style: const TextStyle(color: AppColors.muted)),
                      ),
                    )
                  : ListView.separated(
                      controller: scroll,
                      padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                      itemCount: rows.length,
                      separatorBuilder: (_, __) => const Divider(height: 20),
                      itemBuilder: (_, i) => _RedemptionRow(r: rows[i]),
                    ),
            ),
          ),
        ],
      ),
    );
  }
}

class _RedemptionRow extends StatelessWidget {
  const _RedemptionRow({required this.r});
  final Redemption r;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);

    final (color, label) = switch (r.status) {
      'issued' => (Colors.green, s.t('shop.statusIssued')),
      'cancelled' => (Colors.redAccent, s.t('shop.statusCancelled')),
      _ => (Colors.orange, s.t('shop.statusPending')),
    };

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(r.rewardTitle,
                  style: const TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: color.withValues(alpha: 0.12),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(label,
                        style: TextStyle(
                            color: color,
                            fontSize: 11,
                            fontWeight: FontWeight.w600)),
                  ),
                  const SizedBox(width: 8),
                  Text('${r.costCoins} coin',
                      style: const TextStyle(
                          color: AppColors.muted, fontSize: 12)),
                ],
              ),
              // Kod faqat kutayotgan buyurtmada kerak: topshirilgan yoki
              // bekor qilingandan keyin uning ma'nosi yo'q.
              if (r.isPending) ...[
                const SizedBox(height: 6),
                SelectableText(
                  r.code,
                  style: const TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w800,
                      letterSpacing: 2),
                ),
                Text(s.t('shop.showCode'),
                    style: const TextStyle(
                        color: AppColors.muted, fontSize: 11)),
              ],
            ],
          ),
        ),
      ],
    );
  }
}
