import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/news_models.dart';
import '../data/news_repository.dart';

/// newsListProvider — bosh sahifadagi yangiliklar ro'yxati.
final newsListProvider = FutureProvider.autoDispose<List<NewsItem>>(
  (ref) => ref.read(newsRepositoryProvider).list(),
);

/// newsDetailProvider — bitta yangilikning to'liq matni.
///
/// autoDispose EMAS `keepAlive` bilan: detal ekranidan chiqib qayta kirilsa
/// ko'rishlar soni yana oshib ketmasin desak keshlash kerak bo'lardi — lekin
/// har ochilish haqiqiy ko'rish hisoblanadi, shuning uchun autoDispose to'g'ri.
final newsDetailProvider =
    FutureProvider.autoDispose.family<NewsDetail, String>(
  (ref, id) => ref.read(newsRepositoryProvider).get(id),
);
