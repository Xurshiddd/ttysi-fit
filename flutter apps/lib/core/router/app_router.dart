import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/achievement/presentation/achievements_screen.dart';
import '../../features/auth/application/auth_controller.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/home/presentation/home_screen.dart';
import '../../features/news/presentation/news_detail_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/reward/presentation/reward_shop_screen.dart';

/// appRouterProvider — auth holatiga qarab yo'naltiruvchi router.
final appRouterProvider = Provider<GoRouter>((ref) {
  // Auth o'zgarganda router'ni yangilash uchun Listenable.
  final refresh = ValueNotifier<bool>(ref.read(authControllerProvider).isAuthenticated);
  ref.listen<AuthState>(authControllerProvider, (_, next) {
    refresh.value = next.isAuthenticated;
  });
  ref.onDispose(refresh.dispose);

  return GoRouter(
    initialLocation: '/',
    refreshListenable: refresh,
    redirect: (context, state) {
      final loggedIn = ref.read(authControllerProvider).isAuthenticated;
      final loggingIn = state.matchedLocation == '/login';
      if (!loggedIn && !loggingIn) return '/login';
      if (loggedIn && loggingIn) return '/';
      return null;
    },
    routes: [
      GoRoute(path: '/login', builder: (context, state) => const LoginScreen()),
      GoRoute(path: '/', builder: (context, state) => const HomeScreen()),
      // Yangilik detali — bosh sahifadagi kartadan ochiladi (push).
      GoRoute(
        path: '/news/:id',
        builder: (context, state) =>
            NewsDetailScreen(id: state.pathParameters['id']!),
      ),
      // Yutuqlar — profil kartasidan ochiladi (push).
      GoRoute(
        path: '/achievements',
        builder: (context, state) => const AchievementsScreen(),
      ),
      // Do'kon — profildagi kartadan ochiladi (push).
      GoRoute(
        path: '/shop',
        builder: (context, state) => const RewardShopScreen(),
      ),
      // Profil — Sozlamalar bo'limi ichidan ochiladi (push).
      GoRoute(
        path: '/profile',
        builder: (context, state) => const ProfileScreen(),
      ),
    ],
  );
});
