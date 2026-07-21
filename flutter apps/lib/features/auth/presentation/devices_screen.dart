import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../data/auth_models.dart';
import '../data/auth_repository.dart';

/// devicesProvider — foydalanuvchining faol qurilmalari.
final devicesProvider = FutureProvider.autoDispose<List<UserSession>>(
    (ref) => ref.read(authRepositoryProvider).sessions());

/// currentDeviceIdProvider — ro'yxatda "shu qurilma" ni belgilash uchun.
final currentDeviceIdProvider = FutureProvider<String>(
    (ref) => ref.read(authRepositoryProvider).currentDeviceId());

/// DevicesScreen — "Qurilmalarim va kirishlar".
///
/// Nima uchun kerak: hisob bir vaqtda bitta qurilmada ochiladi, lekin
/// foydalanuvchi qayerda ochiq ekanini va oxirgi marta qachon kirilganini
/// ko'ra olishi kerak — begona kirishni shu yerda payqaydi.
class DevicesScreen extends ConsumerWidget {
  const DevicesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final list = ref.watch(devicesProvider);
    final currentId = ref.watch(currentDeviceIdProvider).valueOrNull ?? '';

    return Scaffold(
      appBar: AppBar(title: Text(s.t('device.title'))),
      body: list.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => _Msg(text: s.t('common.error')),
        data: (rows) => RefreshIndicator(
          onRefresh: () async => ref.invalidate(devicesProvider),
          child: ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Text(s.t('device.hint'),
                  style: const TextStyle(
                      color: AppColors.muted, fontSize: 12, height: 1.4)),
              const SizedBox(height: 16),
              if (rows.isEmpty)
                _Msg(text: s.t('device.empty'))
              else
                ...rows.map((d) => _DeviceCard(
                      d: d,
                      isCurrent: d.deviceId == currentId,
                    )),
            ],
          ),
        ),
      ),
    );
  }
}

class _DeviceCard extends ConsumerStatefulWidget {
  const _DeviceCard({required this.d, required this.isCurrent});
  final UserSession d;
  final bool isCurrent;

  @override
  ConsumerState<_DeviceCard> createState() => _DeviceCardState();
}

class _DeviceCardState extends ConsumerState<_DeviceCard> {
  bool _busy = false;

  Future<void> _revoke() async {
    final s = S.of(context);

    // Joriy qurilmani o'chirish = o'zini chiqarish. Ogohlantiramiz.
    final msg = widget.isCurrent
        ? s.t('device.revokeCurrentConfirm')
        : s.t('device.revokeConfirm');

    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(s.t('device.revoke')),
        content: Text(msg),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: Text(s.t('common.cancel'))),
          FilledButton(
              style: FilledButton.styleFrom(backgroundColor: Colors.redAccent),
              onPressed: () => Navigator.pop(ctx, true),
              child: Text(s.t('device.revoke'))),
        ],
      ),
    );
    if (ok != true || !mounted) return;

    setState(() => _busy = true);
    try {
      await ref.read(authRepositoryProvider).revokeSession(widget.d.id);
      ref.invalidate(devicesProvider);
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(S.of(context).t('common.error'))),
        );
      }
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  IconData get _icon => switch (widget.d.platform) {
        'ios' => Icons.phone_iphone,
        'android' => Icons.smartphone,
        _ => Icons.computer,
      };

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    final d = widget.d;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: widget.isCurrent
              ? AppColors.accent.withValues(alpha: 0.5)
              : Theme.of(context).dividerColor.withValues(alpha: 0.4),
        ),
      ),
      child: Row(
        children: [
          Icon(_icon, color: AppColors.muted, size: 28),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Flexible(
                      child: Text(
                        d.deviceName.isEmpty ? s.t('device.unknown') : d.deviceName,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(fontWeight: FontWeight.w600),
                      ),
                    ),
                    if (widget.isCurrent) ...[
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 8, vertical: 2),
                        decoration: BoxDecoration(
                          color: AppColors.accent.withValues(alpha: 0.15),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Text(s.t('device.current'),
                            style: const TextStyle(
                                color: AppColors.accent,
                                fontSize: 10,
                                fontWeight: FontWeight.w700)),
                      ),
                    ],
                  ],
                ),
                const SizedBox(height: 3),
                Text(
                  '${s.t('device.lastSeen')}: ${_fmt(d.lastSeenAt)}',
                  style: const TextStyle(color: AppColors.muted, fontSize: 12),
                ),
                if (d.ip.isNotEmpty)
                  Text('IP: ${d.ip}',
                      style: const TextStyle(
                          color: AppColors.muted, fontSize: 11)),
              ],
            ),
          ),
          IconButton(
            onPressed: _busy ? null : _revoke,
            icon: _busy
                ? const SizedBox(
                    height: 18,
                    width: 18,
                    child: CircularProgressIndicator(strokeWidth: 2))
                : const Icon(Icons.logout, color: Colors.redAccent),
            tooltip: s.t('device.revoke'),
          ),
        ],
      ),
    );
  }

  String _fmt(DateTime? t) {
    if (t == null) return '—';
    final l = t.toLocal();
    String two(int n) => n.toString().padLeft(2, '0');
    return '${two(l.day)}.${two(l.month)}.${l.year} ${two(l.hour)}:${two(l.minute)}';
  }
}

class _Msg extends StatelessWidget {
  const _Msg({required this.text});
  final String text;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.all(24),
        child: Text(text,
            textAlign: TextAlign.center,
            style: const TextStyle(color: AppColors.muted)),
      );
}
