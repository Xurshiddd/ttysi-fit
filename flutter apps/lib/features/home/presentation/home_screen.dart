import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/prefs/app_prefs.dart';
import '../../../core/theme/app_colors.dart';
import '../../activity/application/activity_providers.dart';
import '../../activity/application/health_permission_controller.dart';
import '../../activity/application/health_sync_controller.dart';
import '../../activity/presentation/health_permission_prompt.dart';
import '../../auth/application/auth_controller.dart';
import '../../news/application/news_providers.dart';
import '../../news/presentation/news_section.dart';
import '../../settings/presentation/settings_tab.dart';
import 'activity_hub_tab.dart';
import 'events_tab.dart';
import '../../rating/application/rating_providers.dart';
import '../../rating/presentation/rating_tab.dart';

/// HomeScreen — asosiy ekran: 3 tab (Bosh sahifa / Reyting / Faollik).
class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen>
    with WidgetsBindingObserver {
  int _index = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      // Ilova ochilishi bilan jim sinxron.
      await _autoSync();
      // Keyin — birinchi kirishda ruxsat tushuntirishi.
      await _askHealthOnce();
    });
  }

  /// _askHealthOnce — qadam sanagich ruxsatini BIR MARTA so'raydi.
  ///
  /// Bu bo'lmasa ilova jim buziladi: avtomatik sinxron ataylab ruxsat
  /// so'ramaydi, ya'ni "Faollik → Sinxronlash" tugmasini hech qachon
  /// bosmagan foydalanuvchining qadamlari umuman yuklanmasdi va u
  /// reytingda doim 0 bo'lib turardi.
  Future<void> _askHealthOnce() async {
    if (!mounted) return;

    final prefs = ref.read(appPrefsProvider);
    if (await prefs.healthAsked()) return;

    // Health Connect o'rnatilmagan yoki qurilma qo'llab-quvvatlamasa
    // hasPermissions xato beradi. Bu ilova ochilishini buzmasligi kerak:
    // qadam sanagich — foydali qo'shimcha, majburiy shart emas.
    bool granted;
    try {
      granted = await ref.read(healthPermissionProvider.future);
    } catch (_) {
      return; // keyingi ochilishda yana urinamiz (bayroq qo'yilmaydi)
    }

    await prefs.setHealthAsked();
    if (granted || !mounted) return;

    await showHealthPermissionSheet(context, ref);
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state != AppLifecycleState.resumed) return;
    // Foydalanuvchi ilovaga qaytdi — telefon shu orada sanagan qadamlarni
    // olib kelamiz. Throttle (kAutoSyncInterval) syncIfStale ichida.
    _autoSync();
    // Tizim sozlamalaridan ruxsat berib qaytgan bo'lishi mumkin — holatni
    // qayta o'qiymiz, aks holda eslatma kartasi noto'g'ri turaverardi.
    ref.read(healthPermissionProvider.notifier).refresh();
  }

  /// _autoSync — jim sinxron: xato bo'lsa ham foydalanuvchini bezovta
  /// qilmaydi (SnackBar yo'q). Ruxsat berilmagan bo'lsa ham jim o'tadi —
  /// ruxsat so'rovi faqat "Sinxronlash" tugmasi orqali ko'rinadi.
  Future<void> _autoSync() async {
    if (!mounted) return;
    await ref.read(healthSyncProvider.notifier).syncIfStale();
  }

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);

    final titles = [
      s.t('home.title'),
      s.t('nav.rating'),
      s.t('nav.activity'),
      s.t('nav.events'),
      s.t('nav.settings'),
    ];

    return Scaffold(
      // Chiqish tugmasi Profil tabiga ko'chdi — har bir tabda takrorlanmaydi.
      appBar: AppBar(title: Text(titles[_index])),
      body: IndexedStack(
        index: _index,
        children: [
          _HomeTab(onGoTo: (i) => setState(() => _index = i)),
          const RatingTab(),
          const ActivityHubTab(),
          const EventsTab(),
          const SettingsTab(),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: [
          NavigationDestination(
              icon: const Icon(Icons.home_outlined),
              selectedIcon: const Icon(Icons.home),
              label: s.t('home.title')),
          NavigationDestination(
              icon: const Icon(Icons.emoji_events_outlined),
              selectedIcon: const Icon(Icons.emoji_events),
              label: s.t('nav.rating')),
          NavigationDestination(
              icon: const Icon(Icons.directions_walk_outlined),
              selectedIcon: const Icon(Icons.directions_walk),
              label: s.t('nav.activity')),
          NavigationDestination(
              icon: const Icon(Icons.flag_outlined),
              selectedIcon: const Icon(Icons.flag),
              label: s.t('nav.events')),
          NavigationDestination(
              icon: const Icon(Icons.settings_outlined),
              selectedIcon: const Icon(Icons.settings),
              label: s.t('nav.settings')),
        ],
      ),
    );
  }
}

/// _HomeTab — salomlashish + bugungi qadam + mening o'rnim kartalari.
class _HomeTab extends ConsumerWidget {
  const _HomeTab({required this.onGoTo});
  final void Function(int tabIndex) onGoTo;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final user = ref.watch(authControllerProvider).user;
    final stats = ref.watch(activityStatsProvider);
    final myRank = ref.watch(myRatingProvider);

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(activityStatsProvider);
        ref.invalidate(myRatingProvider);
        ref.invalidate(newsListProvider);
      },
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text('${s.t('home.hello')}, ${user?.fullName ?? ''} 👋',
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 16),

          // Ruxsat berilmagan bo'lsa eslatma. Berilgan bo'lsa o'zi yashirinadi.
          const HealthPermissionCard(),

          // Bugungi qadam — hero karta
          GestureDetector(
            onTap: () => onGoTo(2),
            child: Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                    colors: [AppColors.primary, Color(0xFF2D5486)],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight),
                borderRadius: BorderRadius.circular(18),
                boxShadow: [
                  BoxShadow(
                      color: AppColors.primary.withValues(alpha: 0.3),
                      blurRadius: 14,
                      offset: const Offset(0, 6)),
                ],
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(s.t('activity.todaySteps'),
                            style: TextStyle(
                                color:
                                    Colors.white.withValues(alpha: 0.75),
                                fontSize: 13)),
                        const SizedBox(height: 6),
                        stats.when(
                          loading: () => const SizedBox(
                              height: 34,
                              width: 34,
                              child: CircularProgressIndicator(
                                  strokeWidth: 2, color: Colors.white)),
                          error: (_, __) => const Text('—',
                              style: TextStyle(
                                  color: Colors.white, fontSize: 30)),
                          data: (st) => Text(_fmt(st.todaySteps),
                              style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 32,
                                  fontWeight: FontWeight.w800)),
                        ),
                      ],
                    ),
                  ),
                  const Icon(Icons.directions_walk,
                      color: AppColors.accent, size: 48),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),

          // Mening o'rnim
          GestureDetector(
            onTap: () => onGoTo(1),
            child: Container(
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
              child: myRank.when(
                loading: () => const Center(
                    child: Padding(
                        padding: EdgeInsets.all(8),
                        child: CircularProgressIndicator(strokeWidth: 2))),
                error: (_, __) => Row(children: [
                  const Icon(Icons.emoji_events_outlined,
                      color: AppColors.muted),
                  const SizedBox(width: 12),
                  Text(s.t('rating.noRank'),
                      style: const TextStyle(color: AppColors.muted)),
                ]),
                data: (m) => Row(
                  children: [
                    const Icon(Icons.emoji_events,
                        color: Color(0xFFF59E0B), size: 34),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(s.t('rating.myRank'),
                              style: const TextStyle(
                                  color: AppColors.muted, fontSize: 12)),
                          Text(
                              '${m.globalRank} / ${m.totalUsers}',
                              style: const TextStyle(
                                  fontWeight: FontWeight.w800,
                                  fontSize: 20)),
                        ],
                      ),
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(s.t('rating.facultyRank'),
                            style: const TextStyle(
                                color: AppColors.muted, fontSize: 12)),
                        Text('${m.facultyRank}',
                            style: const TextStyle(
                                fontWeight: FontWeight.w700,
                                fontSize: 18,
                                color: AppColors.accent)),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),

          // Yangiliklar — ro'yxat bo'sh yoki yiqilsa o'zi yashirinadi.
          const NewsSection(),
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
