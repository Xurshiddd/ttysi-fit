import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../application/activity_providers.dart';
import '../application/health_sync_controller.dart';
import '../data/activity_models.dart';
import '../data/activity_repository.dart';

/// ActivityTab — faollik statistikasi + qo'lda kiritish (test/manual source).
/// Keyinchalik pedometer (HealthKit/Google Fit) shu repository'ga ulanadi.
class ActivityTab extends ConsumerWidget {
  const ActivityTab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final stats = ref.watch(activityStatsProvider);

    return RefreshIndicator(
      onRefresh: () async => ref.invalidate(activityStatsProvider),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const _HealthSyncCard(),
          const SizedBox(height: 16),
          Text(s.t('activity.today'),
              style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 12),
          stats.when(
            loading: () => const Center(
                child: Padding(
                    padding: EdgeInsets.all(24),
                    child: CircularProgressIndicator())),
            error: (e, _) => _ErrorCard(onRetry: () => ref.invalidate(activityStatsProvider)),
            data: (st) => Column(
              children: [
                Row(children: [
                  Expanded(
                      child: _StatCard(
                          icon: Icons.directions_walk,
                          color: AppColors.accent,
                          value: _fmt(st.todaySteps),
                          label: s.t('activity.steps'))),
                  const SizedBox(width: 12),
                  Expanded(
                      child: _StatCard(
                          icon: Icons.local_fire_department,
                          color: AppColors.warning,
                          value: st.todayCalories.toStringAsFixed(0),
                          label: s.t('activity.calories'))),
                ]),
                const SizedBox(height: 12),
                Row(children: [
                  Expanded(
                      child: _StatCard(
                          icon: Icons.route,
                          color: AppColors.primary,
                          value:
                              '${(st.todayDistanceM / 1000).toStringAsFixed(1)} km',
                          label: s.t('activity.distance'))),
                  const SizedBox(width: 12),
                  Expanded(
                      child: _StatCard(
                          icon: Icons.timer_outlined,
                          color: AppColors.danger,
                          value: '${st.todayActiveMin}',
                          label: s.t('activity.activeMin'))),
                ]),
                const SizedBox(height: 20),
                _PeriodRow(stats: st, s: s),
              ],
            ),
          ),
          const SizedBox(height: 24),
          FilledButton.icon(
            style: FilledButton.styleFrom(
              backgroundColor: AppColors.accent,
              minimumSize: const Size.fromHeight(52),
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(14)),
            ),
            icon: const Icon(Icons.add),
            label: Text(s.t('activity.add')),
            onPressed: () => _showEntrySheet(context, ref, s),
          ),
        ],
      ),
    );
  }

  /// Qo'lda kiritish bottom sheet (test uchun; keyin pedometer avtomatlashadi).
  void _showEntrySheet(BuildContext context, WidgetRef ref, S s) {
    final stepsCtrl = TextEditingController();
    final calCtrl = TextEditingController();
    final distCtrl = TextEditingController();
    final minCtrl = TextEditingController();
    var saving = false;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setState) => Padding(
          padding: EdgeInsets.fromLTRB(
              20, 20, 20, MediaQuery.of(ctx).viewInsets.bottom + 20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text(s.t('activity.add'),
                  style: Theme.of(ctx).textTheme.titleLarge),
              const SizedBox(height: 16),
              TextField(
                controller: stepsCtrl,
                keyboardType: TextInputType.number,
                decoration: InputDecoration(
                    labelText: s.t('activity.steps'),
                    prefixIcon: const Icon(Icons.directions_walk)),
              ),
              const SizedBox(height: 12),
              Row(children: [
                Expanded(
                    child: TextField(
                        controller: calCtrl,
                        keyboardType: TextInputType.number,
                        decoration: InputDecoration(
                            labelText: s.t('activity.calories')))),
                const SizedBox(width: 12),
                Expanded(
                    child: TextField(
                        controller: distCtrl,
                        keyboardType: TextInputType.number,
                        decoration: InputDecoration(
                            labelText: '${s.t('activity.distance')} (m)'))),
              ]),
              const SizedBox(height: 12),
              TextField(
                controller: minCtrl,
                keyboardType: TextInputType.number,
                decoration: InputDecoration(
                    labelText: s.t('activity.activeMin'),
                    prefixIcon: const Icon(Icons.timer_outlined)),
              ),
              const SizedBox(height: 20),
              FilledButton(
                style: FilledButton.styleFrom(
                    backgroundColor: AppColors.primary,
                    minimumSize: const Size.fromHeight(50)),
                onPressed: saving
                    ? null
                    : () async {
                        final steps = int.tryParse(stepsCtrl.text) ?? 0;
                        if (steps <= 0) return;
                        setState(() => saving = true);
                        try {
                          await ref
                              .read(activityRepositoryProvider)
                              .record(ActivityRecord(
                                steps: steps,
                                calories:
                                    double.tryParse(calCtrl.text) ?? 0,
                                distanceM:
                                    double.tryParse(distCtrl.text) ?? 0,
                                activeMin: int.tryParse(minCtrl.text) ?? 0,
                              ));
                          ref.invalidate(activityStatsProvider);
                          if (ctx.mounted) Navigator.pop(ctx);
                        } catch (_) {
                          setState(() => saving = false);
                          if (ctx.mounted) {
                            ScaffoldMessenger.of(ctx).showSnackBar(SnackBar(
                                content: Text(s.t('common.error'))));
                          }
                        }
                      },
                child: saving
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white))
                    : Text(s.t('common.save')),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// _HealthSyncCard — telefon (Health Connect / HealthKit) dan bugungi
/// qadamlarni o'qib backend'ga yuborish kartasi.
class _HealthSyncCard extends ConsumerWidget {
  const _HealthSyncCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final syncState = ref.watch(healthSyncProvider);
    final syncing = syncState.isLoading;

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.accent.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.accent.withValues(alpha: 0.25)),
      ),
      child: Row(
        children: [
          const Icon(Icons.favorite, color: AppColors.accent, size: 30),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(s.t('activity.healthTitle'),
                    style: const TextStyle(fontWeight: FontWeight.w600)),
                Text(
                  syncState.valueOrNull != null
                      ? s.t('activity.syncOk')
                      : s.t('activity.healthHint'),
                  style: const TextStyle(
                      color: AppColors.muted, fontSize: 12),
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          FilledButton(
            style: FilledButton.styleFrom(
                backgroundColor: AppColors.accent,
                padding: const EdgeInsets.symmetric(
                    horizontal: 16, vertical: 10)),
            onPressed: syncing
                ? null
                : () async {
                    final result = await ref
                        .read(healthSyncProvider.notifier)
                        .sync();
                    if (!context.mounted) return;
                    final msg = switch (result) {
                      HealthSyncResult.success => s.t('activity.syncOk'),
                      HealthSyncResult.noPermission =>
                        s.t('activity.syncNoPerm'),
                      HealthSyncResult.noData => s.t('activity.syncNoData'),
                      // skipped — tugma bosilganda bo'lmaydi (force: true),
                      // lekin switch to'liq bo'lishi uchun.
                      HealthSyncResult.skipped => s.t('activity.syncOk'),
                      HealthSyncResult.error => s.t('common.error'),
                    };
                    ScaffoldMessenger.of(context)
                        .showSnackBar(SnackBar(content: Text(msg)));
                  },
            child: syncing
                ? const SizedBox(
                    height: 18,
                    width: 18,
                    child: CircularProgressIndicator(
                        strokeWidth: 2, color: Colors.white))
                : Text(s.t('activity.sync')),
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

class _StatCard extends StatelessWidget {
  const _StatCard(
      {required this.icon,
      required this.color,
      required this.value,
      required this.label});
  final IconData icon;
  final Color color;
  final String value;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withValues(alpha: 0.04),
              blurRadius: 10,
              offset: const Offset(0, 2)),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: color, size: 26),
          const SizedBox(height: 10),
          Text(value,
              style: Theme.of(context)
                  .textTheme
                  .titleLarge
                  ?.copyWith(fontWeight: FontWeight.w700)),
          Text(label,
              style: const TextStyle(color: AppColors.muted, fontSize: 12)),
        ],
      ),
    );
  }
}

class _PeriodRow extends StatelessWidget {
  const _PeriodRow({required this.stats, required this.s});
  final ActivityStats stats;
  final S s;

  @override
  Widget build(BuildContext context) {
    Widget cell(String label, int v) => Expanded(
          child: Column(children: [
            Text(_fmt(v),
                style: const TextStyle(
                    fontWeight: FontWeight.w700, fontSize: 16)),
            const SizedBox(height: 2),
            Text(label,
                style:
                    const TextStyle(color: AppColors.muted, fontSize: 12)),
          ]),
        );

    return Container(
      padding: const EdgeInsets.symmetric(vertical: 14),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(children: [
        cell(s.t('rating.week'), stats.weekSteps),
        cell(s.t('rating.month'), stats.monthSteps),
        cell(s.t('rating.all'), stats.totalSteps),
      ]),
    );
  }
}

class _ErrorCard extends StatelessWidget {
  const _ErrorCard({required this.onRetry});
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    return Column(children: [
      Text(s.t('common.error'),
          style: const TextStyle(color: AppColors.muted)),
      TextButton(onPressed: onRetry, child: Text(s.t('common.retry'))),
    ]);
  }
}
