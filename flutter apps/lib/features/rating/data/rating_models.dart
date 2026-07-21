/// RatingEntry — backend GET /ratings javobidagi bitta qator.
class RatingEntry {
  final int rank;
  final String id;
  final String name;
  final String avatarUrl;
  final String facultyName;
  final String groupName;
  final int memberCount;
  final double avgSteps;
  final int totalSteps;
  final double totalDistanceM;
  final int activeDays;

  const RatingEntry({
    required this.rank,
    required this.id,
    required this.name,
    required this.avatarUrl,
    required this.facultyName,
    required this.groupName,
    required this.memberCount,
    required this.avgSteps,
    required this.totalSteps,
    required this.totalDistanceM,
    required this.activeDays,
  });

  factory RatingEntry.fromJson(Map<String, dynamic> j) => RatingEntry(
        rank: (j['rank'] as num?)?.toInt() ?? 0,
        id: (j['id'] ?? '').toString(),
        name: (j['name'] ?? '').toString(),
        avatarUrl: (j['avatar_url'] ?? '').toString(),
        facultyName: (j['faculty_name'] ?? '').toString(),
        groupName: (j['group_name'] ?? '').toString(),
        memberCount: (j['member_count'] as num?)?.toInt() ?? 0,
        avgSteps: (j['avg_steps'] as num?)?.toDouble() ?? 0,
        totalSteps: (j['total_steps'] as num?)?.toInt() ?? 0,
        totalDistanceM: (j['total_distance_m'] as num?)?.toDouble() ?? 0,
        activeDays: (j['active_days'] as num?)?.toInt() ?? 0,
      );
}

/// MyRating — GET /ratings/me javobi.
class MyRating {
  final int globalRank;
  final int facultyRank;
  final int totalUsers;
  final int totalSteps;

  const MyRating({
    required this.globalRank,
    required this.facultyRank,
    required this.totalUsers,
    required this.totalSteps,
  });

  factory MyRating.fromJson(Map<String, dynamic> j) => MyRating(
        globalRank: (j['global_rank'] as num?)?.toInt() ?? 0,
        facultyRank: (j['faculty_rank'] as num?)?.toInt() ?? 0,
        totalUsers: (j['total_users'] as num?)?.toInt() ?? 0,
        totalSteps: (j['total_steps'] as num?)?.toInt() ?? 0,
      );
}

/// RatingFilter — ro'yxat so'rovi parametrlari (provider family kaliti).
class RatingFilter {
  final String type; // student | employee | group | faculty
  final String period; // week | month | all

  const RatingFilter({this.type = 'student', this.period = 'week'});

  RatingFilter copyWith({String? type, String? period}) =>
      RatingFilter(type: type ?? this.type, period: period ?? this.period);

  @override
  bool operator ==(Object other) =>
      other is RatingFilter && other.type == type && other.period == period;

  @override
  int get hashCode => Object.hash(type, period);
}
