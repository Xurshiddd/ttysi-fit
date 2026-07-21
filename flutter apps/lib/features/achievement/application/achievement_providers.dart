import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/achievement_models.dart';
import '../data/achievement_repository.dart';

/// achievementListProvider — barcha aktiv yutuqlar (qozonilgan + jarayondagi).
final achievementListProvider = FutureProvider.autoDispose<List<Achievement>>(
  (ref) => ref.read(achievementRepositoryProvider).list(),
);

/// earnedAchievementsProvider — faqat qozonilganlari (profil kartasi uchun).
///
/// Alohida provider: profil kartasi qisqa ro'yxat ko'rsatadi va to'liq
/// ekranni ochmasdan turib ham yangilanishi kerak.
final earnedAchievementsProvider = FutureProvider.autoDispose<List<Achievement>>(
  (ref) => ref.read(achievementRepositoryProvider).earned(),
);
