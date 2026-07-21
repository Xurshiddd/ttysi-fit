import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/notification_models.dart';
import '../data/notification_repository.dart';

/// notificationsProvider — xabarlar ro'yxati.
final notificationsProvider =
    FutureProvider.autoDispose<List<AppNotification>>(
        (ref) => ref.read(notificationRepositoryProvider).list());

/// unreadCountProvider — qo'ng'iroq nishonidagi son.
///
/// autoDispose EMAS: nishon barcha tablarda ko'rinadi, tab almashganda
/// qayta so'rov ketmasligi kerak. Yangilanish `refresh()` orqali.
final unreadCountProvider =
    AsyncNotifierProvider<UnreadCountController, int>(UnreadCountController.new);

class UnreadCountController extends AsyncNotifier<int> {
  @override
  Future<int> build() => ref.read(notificationRepositoryProvider).unreadCount();

  Future<void> refresh() async {
    // Xato bo'lsa eski qiymat qoladi: nishon ilovaning asosiy oqimi emas,
    // tarmoq uzilishi uchun ekranda xato ko'rsatishning ma'nosi yo'q.
    try {
      state = AsyncData(
          await ref.read(notificationRepositoryProvider).unreadCount());
    } catch (_) {}
  }

  /// markRead — bitta xabarni o'qilgan qiladi va nishonni kamaytiradi.
  Future<void> markRead(String id) async {
    await ref.read(notificationRepositoryProvider).markRead(id);
    ref.invalidate(notificationsProvider);
    await refresh();
  }

  Future<void> markAllRead() async {
    await ref.read(notificationRepositoryProvider).markAllRead();
    ref.invalidate(notificationsProvider);
    await refresh();
  }
}
