import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'notification_models.dart';

final notificationRepositoryProvider = Provider<NotificationRepository>(
    (ref) => NotificationRepository(ref.read(dioProvider)));

/// NotificationRepository — bildirishnoma API chaqiruvlari.
class NotificationRepository {
  NotificationRepository(this._dio);
  final Dio _dio;

  Future<List<AppNotification>> list() async {
    final res = await _dio.get('/notifications', queryParameters: {'limit': 50});
    final data = (res.data['data'] as List?) ?? [];
    return data
        .map((e) => AppNotification.fromJson((e as Map).cast<String, dynamic>()))
        .toList();
  }

  /// unreadCount — qo'ng'iroq nishoni uchun yengil so'rov.
  Future<int> unreadCount() async {
    final res = await _dio.get('/notifications/unread-count');
    return ((res.data['data'] as Map)['unread'] as num?)?.toInt() ?? 0;
  }

  Future<void> markRead(String id) => _dio.post('/notifications/$id/read');

  Future<void> markAllRead() => _dio.post('/notifications/read-all');
}
