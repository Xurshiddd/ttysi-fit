/// Achievement — `GET /achievements` javobidagi bitta qator
/// (yutuq shabloni + shu foydalanuvchi holati).
///
/// `criteria` ataylab `Map<String, dynamic>`: kalitlari TURGA bog'liq va
/// backend registrida aniqlanadi (§16). Yangi tur qo'shilganda bu model
/// o'zgarmasligi kerak.
class Achievement {
  final String id;
  final String type;
  final String title;
  final String description;
  final String awardMode;
  final int rewardCoins;
  final String iconUrl;
  final Map<String, dynamic> criteria;

  // Foydalanuvchi holati (backend qo'shib beradi)
  final bool earned;
  final DateTime? earnedAt;
  final double progress;
  final double target;
  final double progressPct;

  /// awardId — berilgan yutuq yozuvi. Sertifikat havolasi shundan yasaladi;
  /// qozonilmagan yutuqda null.
  final String? awardId;

  const Achievement({
    required this.id,
    required this.type,
    required this.title,
    required this.description,
    required this.awardMode,
    required this.rewardCoins,
    required this.iconUrl,
    required this.criteria,
    required this.earned,
    required this.earnedAt,
    required this.progress,
    required this.target,
    required this.progressPct,
    required this.awardId,
  });

  /// hasCertificate — sertifikat yuklab olish mumkinmi.
  bool get hasCertificate => earned && (awardId?.isNotEmpty ?? false);

  /// isManual — qo'lda beriladigan yutuq (progress ko'rsatilmaydi).
  bool get isManual => awardMode == 'manual';

  factory Achievement.fromJson(Map<String, dynamic> j) => Achievement(
        id: (j['id'] ?? '').toString(),
        type: (j['type'] ?? '').toString(),
        title: (j['title'] ?? '').toString(),
        description: (j['description'] ?? '').toString(),
        awardMode: (j['award_mode'] ?? '').toString(),
        rewardCoins: (j['reward_coins'] as num?)?.toInt() ?? 0,
        iconUrl: (j['icon_url'] ?? '').toString(),
        criteria: (j['criteria'] as Map?)?.cast<String, dynamic>() ?? const {},
        earned: j['earned'] == true,
        earnedAt: _date(j['earned_at']),
        progress: (j['progress'] as num?)?.toDouble() ?? 0,
        target: (j['target'] as num?)?.toDouble() ?? 0,
        progressPct: (j['progress_pct'] as num?)?.toDouble() ?? 0,
        awardId: (j['award_id'] as String?)?.isEmpty ?? true
            ? null
            : j['award_id'] as String,
      );

  static DateTime? _date(dynamic v) {
    if (v == null) return null;
    return DateTime.tryParse(v.toString());
  }

  /// progressLabel — progressni tur birligida ko'rsatadi.
  /// Backend `target` ni metrika birligida beradi (masofa — metrda).
  /// Qo'lda beriladigan turda maqsad yo'q — bo'sh qaytadi.
  String get progressLabel {
    if (target <= 0) return '';
    switch (type) {
      case 'distance_total':
        return '${(progress / 1000).toStringAsFixed(1)} / '
            '${(target / 1000).toStringAsFixed(1)} km';
      case 'active_days':
        return '${progress.toStringAsFixed(0)} / ${target.toStringAsFixed(0)} kun';
      case 'challenge_count':
        return '${progress.toStringAsFixed(0)} / ${target.toStringAsFixed(0)}';
      case 'steps_total':
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
