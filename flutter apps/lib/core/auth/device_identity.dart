import 'dart:io' show Platform;
import 'dart:math';

import 'package:device_info_plus/device_info_plus.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

final deviceIdentityProvider =
    Provider<DeviceIdentity>((ref) => DeviceIdentity());

/// DeviceIdentity — qurilmani tanish uchun barqaror ID va o'qiladigan nom.
///
/// ID NIMA UCHUN O'ZIMIZ YASAYMIZ: `device_info_plus` beradigan apparat
/// identifikatorlari (androidId, identifierForVendor) ishonchsiz —
/// ular fabrika sozlamalarida, ba'zan oddiy yangilanishda ham o'zgaradi,
/// emulyatorlarda esa turli qurilmalarda BIR XIL bo'lishi mumkin. Bir xil
/// ID ikki kishida chiqsa ular bir-birini hisobdan chiqarib yuborardi.
///
/// O'rniga: birinchi ishga tushishda tasodifiy ID yasab, xavfsiz omborda
/// saqlaymiz. U ilova o'chirilmaguncha o'zgarmaydi — aynan kerakli xatti-harakat
/// ("bu o'sha telefonmi?").
class DeviceIdentity {
  static const _idKey = 'device_id';

  final FlutterSecureStorage _storage = const FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
  );

  String? _cachedId;
  String? _cachedName;

  /// deviceId — barqaror identifikator (yo'q bo'lsa yasaladi).
  Future<String> deviceId() async {
    if (_cachedId != null) return _cachedId!;

    try {
      final saved = await _storage.read(key: _idKey);
      if (saved != null && saved.isNotEmpty) {
        _cachedId = saved;
        return saved;
      }
    } catch (_) {
      // Ombor o'qilmasa ham davom etamiz — pastda yangisi yasaladi.
    }

    final fresh = _randomId();
    try {
      await _storage.write(key: _idKey, value: fresh);
    } catch (_) {
      // Saqlanmasa ham joriy sessiya uchun ishlaydi.
    }
    _cachedId = fresh;
    return fresh;
  }

  /// deviceName — foydalanuvchi tanishi uchun nom ("Samsung SM-S918B").
  Future<String> deviceName() async {
    if (_cachedName != null) return _cachedName!;

    var name = 'Qurilma';
    try {
      final info = DeviceInfoPlugin();
      if (!kIsWeb && Platform.isAndroid) {
        final a = await info.androidInfo;
        name = '${a.manufacturer} ${a.model}'.trim();
      } else if (!kIsWeb && Platform.isIOS) {
        final i = await info.iosInfo;
        name = i.name.isNotEmpty ? i.name : i.utsname.machine;
      }
    } catch (_) {
      // Nom — faqat qulaylik: aniqlanmasa umumiy nom qoladi.
    }
    _cachedName = name;
    return name;
  }

  /// platform — backend kutgan qiymat: android | ios | web.
  String get platform {
    if (kIsWeb) return 'web';
    if (Platform.isAndroid) return 'android';
    if (Platform.isIOS) return 'ios';
    return 'web';
  }

  /// info — login so'roviga qo'shiladigan to'plam.
  Future<Map<String, dynamic>> info() async => {
        'device_id': await deviceId(),
        'device_name': await deviceName(),
        'platform': platform,
      };

  /// _randomId — 32 belgili tasodifiy identifikator.
  /// Random.secure(): taxmin qilinadigan ID boshqaning sessiyasini
  /// nishonga olishga imkon berardi.
  String _randomId() {
    const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
    final rnd = Random.secure();
    return List.generate(32, (_) => chars[rnd.nextInt(chars.length)]).join();
  }
}
