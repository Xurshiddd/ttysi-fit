/// Training — video mashg'ulot.
class Training {
  final String id;
  final String title;
  final String description;
  final String category;
  final String level;
  final String videoUrl;
  final String thumbnailUrl;
  final int? durationMin;
  final int views;

  const Training({
    required this.id,
    required this.title,
    required this.description,
    required this.category,
    required this.level,
    required this.videoUrl,
    required this.thumbnailUrl,
    required this.durationMin,
    required this.views,
  });

  factory Training.fromJson(Map<String, dynamic> j) => Training(
        id: (j['id'] ?? '').toString(),
        title: (j['title'] ?? '').toString(),
        description: (j['description'] ?? '').toString(),
        category: (j['category'] ?? '').toString(),
        level: (j['level'] ?? '').toString(),
        videoUrl: (j['video_url'] ?? '').toString(),
        thumbnailUrl: (j['thumbnail_url'] ?? '').toString(),
        durationMin: (j['duration_min'] as num?)?.toInt(),
        views: (j['views'] as num?)?.toInt() ?? 0,
      );

  /// durationLabel — "15 daq" yoki bo'sh (davomiylik ko'rsatilmagan).
  String get durationLabel => durationMin == null ? '' : '$durationMin daq';
}

/// TrainingFilter — ro'yxat filtri (provider family kaliti).
///
/// `category` erkin matn: ro'yxat backenddan keladi (§16 — kodda kategoriya
/// ro'yxati yo'q), shuning uchun bu yerda ham enum emas.
class TrainingFilter {
  final String category;
  final String level;

  const TrainingFilter({this.category = '', this.level = ''});

  TrainingFilter copyWith({String? category, String? level}) =>
      TrainingFilter(
        category: category ?? this.category,
        level: level ?? this.level,
      );

  @override
  bool operator ==(Object other) =>
      other is TrainingFilter &&
      other.category == category &&
      other.level == level;

  @override
  int get hashCode => Object.hash(category, level);
}
