/// Reward — do'kondagi sovg'a (`GET /rewards`).
class Reward {
  final String id;
  final String title;
  final String description;
  final String imageUrl;
  final String category;
  final int costCoins;

  /// stock — qolgan miqdor. null — cheksiz.
  final int? stock;

  /// perUserLimit — bitta odam necha marta ola oladi. null — cheklovsiz.
  final int? perUserLimit;

  const Reward({
    required this.id,
    required this.title,
    required this.description,
    required this.imageUrl,
    required this.category,
    required this.costCoins,
    required this.stock,
    required this.perUserLimit,
  });

  factory Reward.fromJson(Map<String, dynamic> j) => Reward(
        id: (j['id'] ?? '').toString(),
        title: (j['title'] ?? '').toString(),
        description: (j['description'] ?? '').toString(),
        imageUrl: (j['image_url'] ?? '').toString(),
        category: (j['category'] ?? 'other').toString(),
        costCoins: (j['cost_coins'] as num?)?.toInt() ?? 0,
        // null MUHIM: 0 (tugagan) bilan chalkashmasligi kerak.
        stock: (j['stock'] as num?)?.toInt(),
        perUserLimit: (j['per_user_limit'] as num?)?.toInt(),
      );

  /// lowStock — "oxirgi donalar" ogohlantirishi uchun.
  bool get lowStock => stock != null && stock! > 0 && stock! <= 3;

  bool get soldOut => stock != null && stock! <= 0;

  /// affordable — foydalanuvchi balansi yetadimi.
  bool affordable(int balance) => balance >= costCoins;
}

/// Redemption — almashtirish yozuvi (`GET /my-redemptions`).
class Redemption {
  final String id;
  final String code;
  final String status; // pending | issued | cancelled
  final int costCoins;
  final String rewardTitle;
  final String rewardImageUrl;
  final DateTime? createdAt;

  const Redemption({
    required this.id,
    required this.code,
    required this.status,
    required this.costCoins,
    required this.rewardTitle,
    required this.rewardImageUrl,
    required this.createdAt,
  });

  factory Redemption.fromJson(Map<String, dynamic> j) => Redemption(
        id: (j['id'] ?? '').toString(),
        code: (j['code'] ?? '').toString(),
        status: (j['status'] ?? 'pending').toString(),
        costCoins: (j['cost_coins'] as num?)?.toInt() ?? 0,
        rewardTitle: (j['reward_title'] ?? '').toString(),
        rewardImageUrl: (j['reward_image_url'] ?? '').toString(),
        createdAt: j['created_at'] == null
            ? null
            : DateTime.tryParse(j['created_at'].toString()),
      );

  bool get isPending => status == 'pending';
  bool get isCancelled => status == 'cancelled';
}
