import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/activity_models.dart';
import '../data/activity_repository.dart';

/// activityStatsProvider — yig'ma statistika (AsyncValue bilan).
/// Yangi faollik yozilgach `ref.invalidate(activityStatsProvider)` chaqiriladi.
final activityStatsProvider = FutureProvider.autoDispose<ActivityStats>(
  (ref) => ref.read(activityRepositoryProvider).stats(),
);
