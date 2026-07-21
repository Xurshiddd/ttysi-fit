/// CoinBalance — `GET /fit-coins/balance` javobi.
class CoinBalance {
  final int balance;
  final int earned;
  final int spent;

  const CoinBalance({
    required this.balance,
    required this.earned,
    required this.spent,
  });

  factory CoinBalance.fromJson(Map<String, dynamic> j) => CoinBalance(
        balance: (j['balance'] as num?)?.toInt() ?? 0,
        earned: (j['earned'] as num?)?.toInt() ?? 0,
        spent: (j['spent'] as num?)?.toInt() ?? 0,
      );
}

/// CoinTx — ledger yozuvi (`GET /fit-coins`).
///
/// Ledger o'zgarmas: har bir qator bir marta yoziladi va tahrirlanmaydi.
/// Shuning uchun mijozda ham faqat o'qish uchun model.
class CoinTx {
  final String id;
  final int amount;
  final String reason;
  final String refType;
  final String note;
  final DateTime? createdAt;

  const CoinTx({
    required this.id,
    required this.amount,
    required this.reason,
    required this.refType,
    required this.note,
    required this.createdAt,
  });

  factory CoinTx.fromJson(Map<String, dynamic> j) => CoinTx(
        id: (j['id'] ?? '').toString(),
        amount: (j['amount'] as num?)?.toInt() ?? 0,
        reason: (j['reason'] ?? '').toString(),
        refType: (j['ref_type'] ?? '').toString(),
        note: (j['note'] ?? '').toString(),
        createdAt: j['created_at'] == null
            ? null
            : DateTime.tryParse(j['created_at'].toString()),
      );

  bool get isEarned => amount > 0;

  /// signedAmount — ro'yxatda ko'rsatiladigan qiymat ("+25" / "-10").
  String get signedAmount => amount > 0 ? '+$amount' : '$amount';
}
