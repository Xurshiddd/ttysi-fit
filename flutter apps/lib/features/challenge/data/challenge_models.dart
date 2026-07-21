/// Challenge — `GET /challenges` javobidagi bitta qator (foydalanuvchi holati bilan).
///
/// `config` ataylab `Map<String, dynamic>`: uning kalitlari TURGA bog'liq va
/// backend registrida aniqlanadi (§16). Yangi tur qo'shilganda bu model
/// o'zgarmasligi kerak — shuning uchun typed maydonlar yo'q.
class Challenge {
  final String id;
  final String type;
  final String title;
  final String description;
  final String status;
  final int rewardCoins;
  final DateTime? startsAt;
  final DateTime? endsAt;
  final Map<String, dynamic> config;

  // Foydalanuvchi holati (backend qo'shib beradi)
  final bool joined;
  final double progress;
  final bool completed;
  final bool rewardGranted;
  final double target;
  final double progressPct;

  const Challenge({
    required this.id,
    required this.type,
    required this.title,
    required this.description,
    required this.status,
    required this.rewardCoins,
    required this.startsAt,
    required this.endsAt,
    required this.config,
    required this.joined,
    required this.progress,
    required this.completed,
    required this.rewardGranted,
    required this.target,
    required this.progressPct,
  });

  /// canClaim — mukofot olish mumkinmi: yakunlangan, mukofoti bor va hali olinmagan.
  bool get canClaim => completed && rewardCoins > 0 && !rewardGranted;

  factory Challenge.fromJson(Map<String, dynamic> j) => Challenge(
        id: (j['id'] ?? '').toString(),
        type: (j['type'] ?? '').toString(),
        title: (j['title'] ?? '').toString(),
        description: (j['description'] ?? '').toString(),
        status: (j['status'] ?? '').toString(),
        rewardCoins: (j['reward_coins'] as num?)?.toInt() ?? 0,
        startsAt: _date(j['starts_at']),
        endsAt: _date(j['ends_at']),
        config: (j['config'] as Map?)?.cast<String, dynamic>() ?? const {},
        joined: j['joined'] == true,
        progress: (j['progress'] as num?)?.toDouble() ?? 0,
        completed: j['completed'] == true,
        rewardGranted: j['reward_granted'] == true,
        target: (j['target'] as num?)?.toDouble() ?? 0,
        progressPct: (j['progress_pct'] as num?)?.toDouble() ?? 0,
      );

  static DateTime? _date(dynamic v) {
    if (v == null) return null;
    return DateTime.tryParse(v.toString());
  }

  /// daysLeft — tugashiga necha kun qoldi (null — muddatsiz).
  int? get daysLeft {
    if (endsAt == null) return null;
    final d = endsAt!.difference(DateTime.now()).inDays;
    return d < 0 ? 0 : d;
  }

  /// progressLabel — progressni tur birligida ko'rsatadi.
  /// Backend `target` ni metrika birligida beradi (masofa — metrda).
  String get progressLabel {
    switch (type) {
      case 'distance':
        return '${(progress / 1000).toStringAsFixed(1)} / ${(target / 1000).toStringAsFixed(1)} km';
      case 'active_min':
        return '${progress.toStringAsFixed(0)} / ${target.toStringAsFixed(0)} daq';
      case 'steps':
        return '${_fmt(progress.toInt())} / ${_fmt(target.toInt())}';
      default:
        return '';
    }
  }
}

String _fmt(int n) {
  final str = n.toString();
  final buf = StringBuffer();
  for (var i = 0; i < str.length; i++) {
    if (i > 0 && (str.length - i) % 3 == 0) buf.write(' ');
    buf.write(str[i]);
  }
  return buf.toString();
}
