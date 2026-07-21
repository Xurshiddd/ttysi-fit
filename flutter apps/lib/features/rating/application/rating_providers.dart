import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/rating_models.dart';
import '../data/rating_repository.dart';

/// ratingListProvider — tanlangan filtr bo'yicha reyting ro'yxati.
final ratingListProvider = FutureProvider.autoDispose
    .family<List<RatingEntry>, RatingFilter>(
  (ref, f) => ref.read(ratingRepositoryProvider).list(f),
);

/// myRatingProvider — foydalanuvchining haftalik o'rni (bosh sahifa kartasi).
final myRatingProvider = FutureProvider.autoDispose<MyRating>(
  (ref) => ref.read(ratingRepositoryProvider).myRank(),
);
