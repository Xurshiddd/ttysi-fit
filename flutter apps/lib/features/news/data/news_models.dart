/// NewsItem — `GET /news` ro'yxatidagi bitta yangilik.
///
/// `body` YO'Q: backend ro'yxatda uni yubormaydi (uzun matn, trafik).
/// To'liq matn detal ekranida `GET /news/:id` orqali olinadi.
class NewsItem {
  final String id;
  final String title;
  final String excerpt;
  final String coverUrl;
  final bool pinned;
  final int views;
  final DateTime? publishedAt;

  const NewsItem({
    required this.id,
    required this.title,
    required this.excerpt,
    required this.coverUrl,
    required this.pinned,
    required this.views,
    required this.publishedAt,
  });

  factory NewsItem.fromJson(Map<String, dynamic> j) => NewsItem(
        id: (j['id'] ?? '').toString(),
        title: (j['title'] ?? '').toString(),
        excerpt: (j['excerpt'] ?? '').toString(),
        coverUrl: (j['cover_url'] ?? '').toString(),
        pinned: j['pinned'] == true,
        views: (j['views'] as num?)?.toInt() ?? 0,
        publishedAt: j['published_at'] == null
            ? null
            : DateTime.tryParse(j['published_at'].toString()),
      );
}

/// NewsDetail — `GET /news/:id` javobi (to'liq matn bilan).
class NewsDetail {
  final String id;
  final String title;
  final String body;
  final String coverUrl;
  final int views;
  final DateTime? publishedAt;

  const NewsDetail({
    required this.id,
    required this.title,
    required this.body,
    required this.coverUrl,
    required this.views,
    required this.publishedAt,
  });

  factory NewsDetail.fromJson(Map<String, dynamic> j) => NewsDetail(
        id: (j['id'] ?? '').toString(),
        title: (j['title'] ?? '').toString(),
        body: (j['body'] ?? '').toString(),
        coverUrl: (j['cover_url'] ?? '').toString(),
        views: (j['views'] as num?)?.toInt() ?? 0,
        publishedAt: j['published_at'] == null
            ? null
            : DateTime.tryParse(j['published_at'].toString()),
      );
}
