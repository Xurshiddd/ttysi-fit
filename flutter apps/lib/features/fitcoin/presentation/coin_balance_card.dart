import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/fitcoin_providers.dart';
import '../data/fitcoin_models.dart';

const _coinColor = Color(0xFFF59E0B);

/// CoinBalanceCard — FIT Coin balansi. Bosilganda tarix oynasi ochiladi.
class CoinBalanceCard extends ConsumerWidget {
  const CoinBalanceCard({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final bal = ref.watch(coinBalanceProvider);

    return GestureDetector(
      onTap: () => showModalBottomSheet(
        context: context,
        isScrollControlled: true,
        shape: const RoundedRectangleBorder(
            borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
        builder: (_) => const CoinHistorySheet(),
      ),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: _coinColor.withValues(alpha: 0.10),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: _coinColor.withValues(alpha: 0.25)),
        ),
        child: Row(
          children: [
            const Icon(Icons.monetization_on, color: _coinColor, size: 32),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(s.t('coin.balance'),
                      style: const TextStyle(
                          color: AppColors.muted, fontSize: 12)),
                  const SizedBox(height: 2),
                  bal.when(
                    loading: () => const SizedBox(
                        height: 22,
                        width: 22,
                        child:
                            CircularProgressIndicator(strokeWidth: 2)),
                    error: (_, __) => const Text('—',
                        style: TextStyle(
                            fontSize: 22, fontWeight: FontWeight.w800)),
                    data: (b) => Text('${b.balance}',
                        style: const TextStyle(
                            fontSize: 22,
                            fontWeight: FontWeight.w800,
                            color: _coinColor)),
                  ),
                ],
              ),
            ),
            const Icon(Icons.chevron_right, color: AppColors.muted),
          ],
        ),
      ),
    );
  }
}

/// CoinHistorySheet — tranzaksiyalar tarixi (ledger).
class CoinHistorySheet extends ConsumerWidget {
  const CoinHistorySheet({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final history = ref.watch(coinHistoryProvider);
    final bal = ref.watch(coinBalanceProvider);

    return DraggableScrollableSheet(
      initialChildSize: 0.7,
      minChildSize: 0.4,
      maxChildSize: 0.95,
      expand: false,
      builder: (_, scrollCtrl) => Column(
        children: [
          const SizedBox(height: 12),
          Container(
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: AppColors.muted.withValues(alpha: 0.3),
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: Text(s.t('coin.history'),
                      style: Theme.of(context).textTheme.titleLarge),
                ),
                bal.maybeWhen(
                  data: (b) => _Summary(balance: b),
                  orElse: () => const SizedBox.shrink(),
                ),
              ],
            ),
          ),
          Expanded(
            child: history.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (_, __) => Center(
                child: TextButton(
                  onPressed: () => ref.invalidate(coinHistoryProvider),
                  child: Text(s.t('common.retry')),
                ),
              ),
              data: (rows) {
                if (rows.isEmpty) {
                  return Center(
                    child: Text(s.t('coin.empty'),
                        style: const TextStyle(color: AppColors.muted)),
                  );
                }
                return ListView.separated(
                  controller: scrollCtrl,
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 24),
                  itemCount: rows.length,
                  separatorBuilder: (_, __) => Divider(
                      height: 16,
                      color: AppColors.muted.withValues(alpha: 0.12)),
                  itemBuilder: (_, i) => _TxRow(tx: rows[i]),
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _Summary extends StatelessWidget {
  const _Summary({required this.balance});
  final CoinBalance balance;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Icon(Icons.monetization_on, color: _coinColor, size: 20),
        const SizedBox(width: 6),
        Text('${balance.balance}',
            style: const TextStyle(
                fontWeight: FontWeight.w800,
                fontSize: 18,
                color: _coinColor)),
      ],
    );
  }
}

class _TxRow extends StatelessWidget {
  const _TxRow({required this.tx});
  final CoinTx tx;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    // Sabab yorlig'i i18n'dan; noma'lum sabab bo'lsa kalitning o'zi qaytadi va
    // shunda note (yoki xom sabab) ko'rsatiladi — bo'sh qator chiqmaydi.
    final reasonKey = 'coin.reason.${tx.reason}';
    var label = s.t(reasonKey);
    if (label == reasonKey) label = tx.reason;

    return Row(
      children: [
        Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            color: (tx.isEarned ? AppColors.accent : AppColors.danger)
                .withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Icon(
            tx.isEarned ? Icons.arrow_downward : Icons.arrow_upward,
            size: 18,
            color: tx.isEarned ? AppColors.accent : AppColors.danger,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label,
                  style: const TextStyle(
                      fontWeight: FontWeight.w600, fontSize: 14)),
              if (tx.note.isNotEmpty)
                Text(tx.note,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                        color: AppColors.muted, fontSize: 12)),
            ],
          ),
        ),
        const SizedBox(width: 8),
        Text(tx.signedAmount,
            style: TextStyle(
                fontWeight: FontWeight.w800,
                fontSize: 15,
                color: tx.isEarned ? AppColors.accent : AppColors.danger)),
      ],
    );
  }
}
