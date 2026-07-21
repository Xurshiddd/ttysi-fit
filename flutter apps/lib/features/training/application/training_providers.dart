import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/training_models.dart';
import '../data/training_repository.dart';

/// trainingFilterProvider — tanlangan filtr (kategoriya + daraja).
final trainingFilterProvider =
    NotifierProvider<TrainingFilterController, TrainingFilter>(
        TrainingFilterController.new);

class TrainingFilterController extends Notifier<TrainingFilter> {
  @override
  TrainingFilter build() => const TrainingFilter();

  void setCategory(String c) => state = state.copyWith(category: c);
  void setLevel(String l) => state = state.copyWith(level: l);
}

/// trainingListProvider — filtrga mos mashg'ulotlar.
final trainingListProvider = FutureProvider.autoDispose<List<Training>>((ref) {
  final f = ref.watch(trainingFilterProvider);
  return ref.read(trainingRepositoryProvider).list(f);
});

/// trainingCategoriesProvider — filtr uchun kategoriyalar ro'yxati.
final trainingCategoriesProvider = FutureProvider<List<String>>(
  (ref) => ref.read(trainingRepositoryProvider).categories(),
);
