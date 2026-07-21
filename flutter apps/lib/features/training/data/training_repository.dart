import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'training_models.dart';

final trainingRepositoryProvider = Provider<TrainingRepository>(
    (ref) => TrainingRepository(ref.read(dioProvider)));

/// TrainingRepository — mashg'ulotlar API chaqiruvlari.
class TrainingRepository {
  TrainingRepository(this._dio);
  final Dio _dio;

  Future<List<Training>> list(TrainingFilter f, {int limit = 50}) async {
    final res = await _dio.get('/trainings', queryParameters: {
      if (f.category.isNotEmpty) 'category': f.category,
      if (f.level.isNotEmpty) 'level': f.level,
      'page': 1,
      'limit': limit,
    });
    final data = (res.data['data'] as List?) ?? const [];
    return data
        .map((e) => Training.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  /// categories — mavjud kategoriyalar. Kodda ro'yxat yo'q: admin yangisini
  /// qo'shsa filtr avtomatik yangilanadi (§16).
  Future<List<String>> categories() async {
    final res = await _dio.get('/training-categories');
    final data = (res.data['data'] as List?) ?? const [];
    return data.map((e) => e.toString()).toList();
  }

  /// get — to'liq mashg'ulot (ko'rishlar hisobini oshiradi).
  Future<Training> get(String id) async {
    final res = await _dio.get('/trainings/$id');
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return Training.fromJson(data);
  }
}
