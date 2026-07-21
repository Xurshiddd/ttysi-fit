import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/notification_providers.dart';
import '../data/notification_models.dart';

/// NotificationScreen — bildirishnomalar ro'yxati.
class NotificationScreen extends ConsumerWidget {
  const NotificationScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final list = ref.watch(notificationsProvider);
    final unread = ref.watch(unreadCountProvider).valueOrNull ?? 0;

    return Scaffold(
      appBar: AppBar(
        title: Text(s.t('notif.title')),
        actions: [
          // Matnli tugma emas, IKONKA: "Bildirishnomalar" sarlavhasi
          // uzun va matnli tugma uni kesib qo'yardi ("Bildirishnomal…").
          if (unread > 0)
            IconButton(
              icon: const Icon(Icons.done_all),
              tooltip: s.t('notif.readAll'),
              onPressed: () =>
                  ref.read(unreadCountProvider.notifier).markAllRead(),
            ),
        ],
      ),
      body: list.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => _Message(text: s.t('common.error')),
        data: (rows) => rows.isEmpty
            ? _Message(text: s.t('notif.empty'))
            : RefreshIndicator(
                onRefresh: () async {
                  ref.invalidate(notificationsProvider);
                  await ref.read(unreadCountProvider.notifier).refresh();
                },
                child: ListView.separated(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  itemCount: rows.length,
                  separatorBuilder: (_, __) => const Divider(height: 1),
                  itemBuilder: (_, i) => _Tile(n: rows[i]),
                ),
              ),
      ),
    );
  }
}

class _Tile extends ConsumerWidget {
  const _Tile({required this.n});
  final AppNotification n;

  /// Tur bo'yicha ikonka va rang. Ro'yxatni bir qarashda ajratish uchun.
  (IconData, Color) get _look => switch (n.type) {
        'achievement' => (Icons.emoji_events, const Color(0xFFF59E0B)),
        'challenge' => (Icons.flag, AppColors.accent),
        'competition' => (Icons.sports_score, AppColors.primary),
        'reward_issued' => (Icons.card_giftcard, AppColors.accent),
        'reward_cancel' => (Icons.undo, Colors.redAccent),
        'coins' => (Icons.monetization_on, const Color(0xFFF59E0B)),
        _ => (Icons.campaign, AppColors.primary),
      };

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final (icon, color) = _look;

    return InkWell(
      onTap: n.isUnread
          ? () => ref.read(unreadCountProvider.notifier).markRead(n.id)
          : null,
      child: Container(
        // O'qilmagan xabar ajralib tursin.
        color: n.isUnread ? color.withValues(alpha: 0.06) : null,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              height: 40,
              width: 40,
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.12),
                shape: BoxShape.circle,
              ),
              child: Icon(icon, color: color, size: 20),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    n.title,
                    style: TextStyle(
                      fontWeight:
                          n.isUnread ? FontWeight.w700 : FontWeight.w500,
                    ),
                  ),
                  if (n.body.isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(n.body,
                        style: const TextStyle(
                            color: AppColors.muted, fontSize: 13, height: 1.35)),
                  ],
                  const SizedBox(height: 4),
                  Text(_ago(context, n.createdAt),
                      style: const TextStyle(
                          color: AppColors.muted, fontSize: 11)),
                ],
              ),
            ),
            if (n.isUnread)
              Container(
                margin: const EdgeInsets.only(top: 6, left: 8),
                height: 8,
                width: 8,
                decoration: BoxDecoration(color: color, shape: BoxShape.circle),
              ),
          ],
        ),
      ),
    );
  }

  /// _ago — "5 daqiqa oldin" ko'rinishidagi nisbiy vaqt.
  String _ago(BuildContext context, DateTime? t) {
    final s = S.of(context);
    if (t == null) return '';
    final d = DateTime.now().difference(t);
    if (d.inMinutes < 1) return s.t('notif.justNow');
    if (d.inHours < 1) return '${d.inMinutes} ${s.t('notif.minAgo')}';
    if (d.inDays < 1) return '${d.inHours} ${s.t('notif.hourAgo')}';
    if (d.inDays < 7) return '${d.inDays} ${s.t('notif.dayAgo')}';
    return '${t.day.toString().padLeft(2, '0')}.${t.month.toString().padLeft(2, '0')}.${t.year}';
  }
}

class _Message extends StatelessWidget {
  const _Message({required this.text});
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

/// NotificationBell — AppBar dagi qo'ng'iroq + o'qilmaganlar nishoni.
class NotificationBell extends ConsumerWidget {
  const NotificationBell({super.key, required this.onTap});
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final unread = ref.watch(unreadCountProvider).valueOrNull ?? 0;

    return Stack(
      alignment: Alignment.center,
      children: [
        IconButton(
          icon: const Icon(Icons.notifications_outlined),
          onPressed: onTap,
          tooltip: S.of(context).t('notif.title'),
        ),
        if (unread > 0)
          Positioned(
            top: 8,
            right: 8,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
              constraints: const BoxConstraints(minWidth: 16),
              decoration: BoxDecoration(
                color: Colors.redAccent,
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                // 99 dan oshsa "99+" — nishon kengayib ketmasin.
                unread > 99 ? '99+' : '$unread',
                textAlign: TextAlign.center,
                style: const TextStyle(
                    color: Colors.white,
                    fontSize: 10,
                    fontWeight: FontWeight.w700),
              ),
            ),
          ),
      ],
    );
  }
}
