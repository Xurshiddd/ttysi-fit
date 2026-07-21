/// AppNotification — `GET /notifications` javobidagi bitta xabar.
///
/// Nomi `Notification` EMAS: Flutter'da shu nomli o'z klassi bor
/// (`dart:ui`), ustma-ust tushsa import chalkashligi chiqadi.
class AppNotification {
  final String id;
  final String type;
  final String title;
  final String body;
  final String refType;
  final String refId;
  final DateTime? createdAt;
  final DateTime? readAt;

  const AppNotification({
    required this.id,
    required this.type,
    required this.title,
    required this.body,
    required this.refType,
    required this.refId,
    required this.createdAt,
    required this.readAt,
  });

  factory AppNotification.fromJson(Map<String, dynamic> j) => AppNotification(
        id: (j['id'] ?? '').toString(),
        type: (j['type'] ?? '').toString(),
        title: (j['title'] ?? '').toString(),
        body: (j['body'] ?? '').toString(),
        refType: (j['ref_type'] ?? '').toString(),
        refId: (j['ref_id'] ?? '').toString(),
        createdAt: j['created_at'] == null
            ? null
            : DateTime.tryParse(j['created_at'].toString()),
        readAt: j['read_at'] == null
            ? null
            : DateTime.tryParse(j['read_at'].toString()),
      );

  bool get isUnread => readAt == null;
}
