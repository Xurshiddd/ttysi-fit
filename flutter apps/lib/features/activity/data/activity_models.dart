/// ActivityStats — backend GET /activities/stats javobi.
class ActivityStats {
  final int todaySteps;
  final double todayCalories;
  final double todayDistanceM;
  final int todayActiveMin;
  final int weekSteps;
  final int monthSteps;
  final int totalSteps;

  const ActivityStats({
    required this.todaySteps,
    required this.todayCalories,
    required this.todayDistanceM,
    required this.todayActiveMin,
    required this.weekSteps,
    required this.monthSteps,
    required this.totalSteps,
  });

  factory ActivityStats.fromJson(Map<String, dynamic> j) => ActivityStats(
        todaySteps: (j['today_steps'] as num?)?.toInt() ?? 0,
        todayCalories: (j['today_calories'] as num?)?.toDouble() ?? 0,
        todayDistanceM: (j['today_distance_m'] as num?)?.toDouble() ?? 0,
        todayActiveMin: (j['today_active_min'] as num?)?.toInt() ?? 0,
        weekSteps: (j['week_steps'] as num?)?.toInt() ?? 0,
        monthSteps: (j['month_steps'] as num?)?.toInt() ?? 0,
        totalSteps: (j['total_steps'] as num?)?.toInt() ?? 0,
      );
}

/// ActivityRecord — POST /activities so'rovi (qo'lda kiritish/pedometer).
class ActivityRecord {
  /// Qaysi kunga tegishli (telefon MAHALLIY sanasi).
  ///
  /// Sanani aynan mijoz aytadi: qadam tarixi telefonda mahalliy kunlar
  /// bo'yicha guruhlangan, server esa faqat so'rov KELGAN paytni biladi.
  /// null bo'lsa backend o'zining APP_TIMEZONE bugungi kunini oladi.
  final DateTime? date;
  final int steps;
  final double calories;
  final double distanceM;
  final int activeMin;
  final String source;

  const ActivityRecord({
    required this.steps,
    this.date,
    this.calories = 0,
    this.distanceM = 0,
    this.activeMin = 0,
    this.source = 'manual',
  });

  /// Backend kutgan format: YYYY-MM-DD (vaqt/mintaqasiz).
  static String formatDate(DateTime d) =>
      '${d.year.toString().padLeft(4, '0')}-'
      '${d.month.toString().padLeft(2, '0')}-'
      '${d.day.toString().padLeft(2, '0')}';

  Map<String, dynamic> toJson() => {
        if (date != null) 'date': formatDate(date!),
        'steps': steps,
        'calories': calories,
        'distance_m': distanceM,
        'active_min': activeMin,
        'source': source,
      };
}
