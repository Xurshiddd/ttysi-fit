import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'news_models.dart';

final newsRepositoryProvider =
    Provider<NewsRepository>((ref) => NewsRepository(ref.read(dioProvider)));

/// NewsRepository — yangiliklar API chaqiruvlari.
class NewsRepository {
  NewsRepository(this._dio);
  final Dio _dio;

  /// list — e'lon qilingan yangiliklar (backend draft'ni yashiradi).
  Future<List<NewsItem>> list({int limit = 20}) async {
    final res = await _dio.get('/news', queryParameters: {
      'page': 1,
      'limit': limit,
    });
    final data = (res.data['data'] as List?) ?? const [];
    return data
        .map((e) => NewsItem.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  /// get — to'liq matn. Har chaqiruv ko'rishlar sonini oshiradi (backend).
  Future<NewsDetail> get(String id) async {
    final res = await _dio.get('/news/$id');
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return NewsDetail.fromJson(data);
  }
}
