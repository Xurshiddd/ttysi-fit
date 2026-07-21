import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';

const _coinColor = Color(0xFFF59E0B);

/// RewardShopCard — profildan do'konga kirish.
///
/// Alohida karta sifatida: FIT Coin balansi ko'ringanidan keyin darrov
/// "buni nimaga ishlatsam bo'ladi?" degan savol tug'iladi — javob shu
/// yerda turishi kerak.
class RewardShopCard extends StatelessWidget {
  const RewardShopCard({super.key});

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);

    return GestureDetector(
      onTap: () => context.push('/shop'),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppColors.accent.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: AppColors.accent.withValues(alpha: 0.25)),
        ),
        child: Row(
          children: [
            const Icon(Icons.card_giftcard, color: _coinColor, size: 30),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(s.t('shop.title'),
                      style: const TextStyle(fontWeight: FontWeight.w600)),
                  const SizedBox(height: 2),
                  Text(s.t('shop.cardHint'),
                      style: const TextStyle(
                          color: AppColors.muted, fontSize: 12)),
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
