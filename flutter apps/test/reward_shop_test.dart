// FIT Coin do'koni testlari.
//
// NEGA BU TESTLAR BOR: do'kon PUL harakati bilan ishlaydi. Modeldagi jimgina
// xato — masalan `stock: null` (cheksiz) ni 0 (tugagan) deb o'qish — barcha
// cheksiz sovg'alarni do'kondan yo'qotib yuborardi. `affordable` xatosi esa
// balansi yetmagan foydalanuvchiga tugmani ochib, backend 409 xatosiga
// olib borardi.

import 'package:flutter_test/flutter_test.dart';

import 'package:ttysi_fit/features/reward/data/reward_models.dart';

Reward reward({int cost = 50, int? stock, int? perUser}) => Reward(
      id: 'r1',
      title: 'Futbolka',
      description: '',
      imageUrl: '',
      category: 'merch',
      costCoins: cost,
      stock: stock,
      perUserLimit: perUser,
    );

void main() {
  group('Reward modeli', () {
    test('stock null — CHEKSIZ, tugagan emas', () {
      final r = Reward.fromJson({
        'id': 'x', 'title': 'T', 'cost_coins': 10, 'stock': null,
      });
      expect(r.stock, isNull);
      expect(r.soldOut, isFalse, reason: 'cheksiz sovg‘a tugagan deb belgilandi');
      expect(r.lowStock, isFalse);
    });

    test('stock 0 — TUGAGAN', () {
      final r = Reward.fromJson({'id': 'x', 'title': 'T', 'cost_coins': 10, 'stock': 0});
      expect(r.soldOut, isTrue);
    });

    test('oz qolganda ogohlantirish', () {
      expect(reward(stock: 3).lowStock, isTrue);
      expect(reward(stock: 1).lowStock, isTrue);
      expect(reward(stock: 4).lowStock, isFalse);
      // Tugagan sovg'a "oxirgi donalar" emas.
      expect(reward(stock: 0).lowStock, isFalse);
    });

    test('affordable — balans aynan yetganda ham TRUE', () {
      final r = reward(cost: 50);
      expect(r.affordable(50), isTrue, reason: 'aynan yetarli balans rad etildi');
      expect(r.affordable(51), isTrue);
      expect(r.affordable(49), isFalse);
      expect(r.affordable(0), isFalse);
    });

    test('bo‘sh/buzuq JSON qulatmaydi', () {
      final r = Reward.fromJson({});
      expect(r.id, '');
      expect(r.costCoins, 0);
      expect(r.stock, isNull);
      expect(r.category, 'other');
    });
  });

  group('Redemption modeli', () {
    test('holatlar to‘g‘ri o‘qiladi', () {
      final p = Redemption.fromJson({'status': 'pending', 'code': 'ABC12345'});
      expect(p.isPending, isTrue);
      expect(p.isCancelled, isFalse);
      expect(p.code, 'ABC12345');

      final c = Redemption.fromJson({'status': 'cancelled'});
      expect(c.isCancelled, isTrue);
      expect(c.isPending, isFalse);

      final i = Redemption.fromJson({'status': 'issued'});
      expect(i.isPending, isFalse);
      expect(i.isCancelled, isFalse);
    });

    test('status berilmasa pending deb qaraladi', () {
      expect(Redemption.fromJson({}).isPending, isTrue);
    });

    test('sana buzuq bo‘lsa null, qulamaydi', () {
      expect(Redemption.fromJson({'created_at': 'axlat'}).createdAt, isNull);
      expect(
        Redemption.fromJson({'created_at': '2026-07-21T10:00:00Z'}).createdAt,
        isNotNull,
      );
    });
  });
}
