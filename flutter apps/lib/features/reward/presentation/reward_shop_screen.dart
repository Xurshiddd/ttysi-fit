import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../../fitcoin/application/fitcoin_providers.dart';
import '../application/reward_providers.dart';
import '../data/reward_models.dart';
import 'my_redemptions_sheet.dart';

const _coinColor = Color(0xFFF59E0B);

/// RewardShopScreen — FIT Coin do'koni.
///
/// Balans doim tepada turadi: foydalanuvchi nima sotib ola olishini
/// ro'yxatni aylantirmasdan ko'rishi kerak.
class RewardShopScreen extends ConsumerWidget {
  const RewardShopScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final rewards = ref.watch(rewardsProvider);
    final balance = ref.watch(coinBalanceProvider).valueOrNull?.balance ?? 0;

    return Scaffold(
      appBar: AppBar(
        title: Text(s.t('shop.title')),
        actions: [
          IconButton(
            tooltip: s.t('shop.myOrders'),
            icon: const Icon(Icons.receipt_long_outlined),
            onPressed: () => showMyRedemptionsSheet(context),
          ),
        ],
      ),
      body: Column(
        children: [
          _BalanceBar(balance: balance),
          const _CategoryChips(),
          Expanded(
            child: rewards.when(
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (_, __) => _Empty(text: s.t('common.error')),
              data: (list) => list.isEmpty
                  ? _Empty(text: s.t('shop.empty'))
                  : RefreshIndicator(
                      onRefresh: () async {
                        ref.invalidate(rewardsProvider);
                        ref.invalidate(coinBalanceProvider);
                      },
                      child: ListView.separated(
                        padding: const EdgeInsets.all(16),
                        itemCount: list.length,
                        separatorBuilder: (_, __) => const SizedBox(height: 12),
                        itemBuilder: (_, i) =>
                            _RewardCard(reward: list[i], balance: balance),
                      ),
                    ),
            ),
          ),
        ],
      ),
    );
  }
}

class _BalanceBar extends StatelessWidget {
  const _BalanceBar({required this.balance});
  final int balance;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      color: _coinColor.withValues(alpha: 0.10),
      child: Row(
        children: [
          const Icon(Icons.monetization_on, color: _coinColor, size: 26),
          const SizedBox(width: 10),
          Text(s.t('coin.balance'),
              style: const TextStyle(color: AppColors.muted, fontSize: 13)),
          const Spacer(),
          Text('$balance',
              style: const TextStyle(
                  fontSize: 20, fontWeight: FontWeight.w800, color: _coinColor)),
        ],
      ),
    );
  }
}

class _CategoryChips extends ConsumerWidget {
  const _CategoryChips();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final cats = ref.watch(rewardCategoriesProvider).valueOrNull ?? const [];
    if (cats.isEmpty) return const SizedBox.shrink();

    final selected = ref.watch(rewardCategoryProvider);

    return SizedBox(
      height: 48,
      child: ListView(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        children: [
          _chip(context, ref, '', s.t('common.all'), selected.isEmpty),
          for (final c in cats)
            _chip(context, ref, c, s.t('shop.cat.$c'), selected == c),
        ],
      ),
    );
  }

  Widget _chip(BuildContext context, WidgetRef ref, String value, String label,
      bool active) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 6),
      child: ChoiceChip(
        label: Text(label),
        selected: active,
        onSelected: (_) =>
            ref.read(rewardCategoryProvider.notifier).state = value,
      ),
    );
  }
}

class _RewardCard extends ConsumerStatefulWidget {
  const _RewardCard({required this.reward, required this.balance});
  final Reward reward;
  final int balance;

  @override
  ConsumerState<_RewardCard> createState() => _RewardCardState();
}

class _RewardCardState extends ConsumerState<_RewardCard> {
  bool _busy = false;

  Future<void> _redeem() async {
    final s = S.of(context);
    final r = widget.reward;

    // Tasdiqlash MAJBURIY: coin qaytarib bo'lmaydigan tarzda yechiladi
    // (faqat admin bekor qila oladi).
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(r.title),
        content: Text(s
            .t('shop.confirmBody')
            .replaceAll('{cost}', '${r.costCoins}')
            .replaceAll('{left}', '${widget.balance - r.costCoins}')),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: Text(s.t('common.cancel'))),
          FilledButton(
              style: FilledButton.styleFrom(backgroundColor: AppColors.accent),
              onPressed: () => Navigator.pop(ctx, true),
              child: Text(s.t('shop.redeem'))),
        ],
      ),
    );
    if (ok != true || !mounted) return;

    setState(() => _busy = true);
    try {
      final red = await ref.read(redeemRewardProvider)(r.id);
      if (!mounted) return;
      // Kodni ko'rsatamiz — sovg'ani olishda shu kod so'raladi.
      await showDialog<void>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: Text(s.t('shop.doneTitle')),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(s.t('shop.doneBody'),
                  textAlign: TextAlign.center,
                  style: const TextStyle(color: AppColors.muted)),
              const SizedBox(height: 14),
              SelectableText(
                red.code,
                style: const TextStyle(
                    fontSize: 26,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 3),
              ),
            ],
          ),
          actions: [
            TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: Text(s.t('common.ok'))),
          ],
        ),
      );
    } on DioException catch (e) {
      if (!mounted) return;
      // 409 — balans yetmadi yoki limit tugadi; backend sababni aytadi.
      final msg = e.response?.statusCode == 409
          ? s.t('shop.notEnough')
          : (e.response?.data is Map
              ? (e.response!.data['error']?.toString() ?? s.t('common.error'))
              : s.t('common.error'));
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(msg)));
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(s.t('common.error'))));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    final r = widget.reward;
    final canAfford = r.affordable(widget.balance);

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
            color: Theme.of(context).dividerColor.withValues(alpha: 0.4)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _Thumb(url: r.imageUrl),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(r.title,
                    style: const TextStyle(
                        fontWeight: FontWeight.w700, fontSize: 15)),
                if (r.description.isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(r.description,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                          color: AppColors.muted, fontSize: 12, height: 1.3)),
                ],
                const SizedBox(height: 8),
                Row(
                  children: [
                    const Icon(Icons.monetization_on,
                        color: _coinColor, size: 18),
                    const SizedBox(width: 4),
                    Text('${r.costCoins}',
                        style: const TextStyle(
                            fontWeight: FontWeight.w800, color: _coinColor)),
                    if (r.lowStock) ...[
                      const SizedBox(width: 10),
                      Text(
                        s.t('shop.left').replaceAll('{n}', '${r.stock}'),
                        style: const TextStyle(
                            color: Colors.redAccent, fontSize: 11),
                      ),
                    ],
                  ],
                ),
                const SizedBox(height: 8),
                SizedBox(
                  width: double.infinity,
                  child: FilledButton(
                    style: FilledButton.styleFrom(
                      backgroundColor:
                          canAfford ? AppColors.accent : Colors.grey.shade400,
                      // Barmoq uchun qulay nishon (§6.3).
                      minimumSize: const Size(0, 44),
                    ),
                    onPressed: (!canAfford || _busy) ? null : _redeem,
                    child: _busy
                        ? const SizedBox(
                            height: 18,
                            width: 18,
                            child: CircularProgressIndicator(
                                strokeWidth: 2, color: Colors.white))
                        : Text(canAfford
                            ? s.t('shop.redeem')
                            : s.t('shop.notEnoughShort')),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _Thumb extends StatelessWidget {
  const _Thumb({required this.url});
  final String url;

  @override
  Widget build(BuildContext context) {
    const size = 72.0;
    if (url.isEmpty) {
      return Container(
        height: size,
        width: size,
        decoration: BoxDecoration(
          color: _coinColor.withValues(alpha: 0.12),
          borderRadius: BorderRadius.circular(12),
        ),
        child: const Icon(Icons.card_giftcard, color: _coinColor, size: 30),
      );
    }
    return ClipRRect(
      borderRadius: BorderRadius.circular(12),
      child: Image.network(
        url,
        height: size,
        width: size,
        fit: BoxFit.cover,
        // Rasm yuklanmasa karta buzilmasin.
        errorBuilder: (_, __, ___) => Container(
          height: size,
          width: size,
          color: _coinColor.withValues(alpha: 0.12),
          child: const Icon(Icons.card_giftcard, color: _coinColor, size: 30),
        ),
      ),
    );
  }
}

class _Empty extends StatelessWidget {
  const _Empty({required this.text});
  final String text;

  @override
  Widget build(BuildContext context) => Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Text(text,
              textAlign: TextAlign.center,
              style: const TextStyle(color: AppColors.muted)),
        ),
      );
}
