/// Competition — `GET /competitions` javobidagi bitta qator (foydalanuvchi holati bilan).
///
/// `config` turga bog'liq (§16) — typed maydonlar yo'q, aks holda backendda
/// yangi tur qo'shilganda bu model ham o'zgarishi kerak bo'lardi.
class Competition {
  final String id;
  final String type;
  final String title;
  final String description;
  final String status;
  final String location;
  final int rewardCoins;
  final int? maxParticipants;
  final DateTime? startsAt;
  final DateTime? endsAt;
  final DateTime? regEndsAt;
  final Map<String, dynamic> config;

  // Foydalanuvchi holati (backend qo'shib beradi)
  final bool registered;
  final int? place;
  final bool rewardGranted;
  final int participantCount;
  final bool regOpen;

  const Competition({
    required this.id,
    required this.type,
    required this.title,
    required this.description,
    required this.status,
    required this.location,
    required this.rewardCoins,
    required this.maxParticipants,
    required this.startsAt,
    required this.endsAt,
    required this.regEndsAt,
    required this.config,
    required this.registered,
    required this.place,
    required this.rewardGranted,
    required this.participantCount,
    required this.regOpen,
  });

  factory Competition.fromJson(Map<String, dynamic> j) => Competition(
        id: (j['id'] ?? '').toString(),
        type: (j['type'] ?? '').toString(),
        title: (j['title'] ?? '').toString(),
        description: (j['description'] ?? '').toString(),
        status: (j['status'] ?? '').toString(),
        location: (j['location'] ?? '').toString(),
        rewardCoins: (j['reward_coins'] as num?)?.toInt() ?? 0,
        maxParticipants: (j['max_participants'] as num?)?.toInt(),
        startsAt: _date(j['starts_at']),
        endsAt: _date(j['ends_at']),
        regEndsAt: _date(j['reg_ends_at']),
        config: (j['config'] as Map?)?.cast<String, dynamic>() ?? const {},
        registered: j['registered'] == true,
        place: (j['place'] as num?)?.toInt(),
        rewardGranted: j['reward_granted'] == true,
        participantCount: (j['participant_count'] as num?)?.toInt() ?? 0,
        regOpen: j['reg_open'] == true,
      );

  static DateTime? _date(dynamic v) =>
      v == null ? null : DateTime.tryParse(v.toString());

  /// sport — config'dagi sport turi (barcha turlarda bor, custom'dan tashqari).
  String get sport => (config['sport'] ?? '').toString();

  /// slotsLabel — "12 / 30" yoki cheklovsiz bo'lsa faqat ishtirokchilar soni.
  /// max_participants 0 ham cheklovsiz degani (backend shunday talqin qiladi).
  String get slotsLabel {
    final max = maxParticipants;
    if (max == null || max == 0) return '$participantCount';
    return '$participantCount / $max';
  }

  bool get isFull {
    final max = maxParticipants;
    return max != null && max > 0 && participantCount >= max;
  }
}
