import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:permission_handler/permission_handler.dart';

import '../data/health_service.dart';
import 'health_sync_controller.dart';

/// Ruxsat so'rash natijasi.
enum HealthPermissionResult {
  /// Berildi — qadamlar endi yuklanadi.
  granted,

  /// Rad etildi, lekin qayta so'rash mumkin.
  denied,

  /// Butunlay rad etilgan: tizim endi so'rov oynasini ko'rsatmaydi,
  /// foydalanuvchi sozlamalardan qo'lda yoqishi kerak.
  permanentlyDenied,
}

/// HealthPermissionController — qadam sanagich ruxsatining holati.
///
/// NEGA ALOHIDA KONTROLLER: avtomatik sinxron ruxsat SO'RAMAYDI (ilova
/// ochilishi bilan tizim oynasi chiqib kelishi foydalanuvchini cho'chitadi).
/// Natijada "Sinxronlash" tugmasini hech qachon bosmagan foydalanuvchining
/// qadamlari umuman yuklanmasdi — reyting esa butunlay shu ma'lumotga
/// bog'liq. Bu kontroller ruxsatni ko'rinadigan holatga aylantiradi:
/// bosh sahifa uni so'ramagan foydalanuvchiga eslatib turadi.
class HealthPermissionController extends AsyncNotifier<bool> {
  @override
  Future<bool> build() async {
    // null — HOLAT NOMA'LUM (Health Connect ishonchli javob bermaydi).
    // Bunday holatda ruxsat BOR deb hisoblaymiz: noma'lumlik uchun
    // foydalanuvchini "ruxsat bering" kartasi bilan bezovta qilish
    // noto'g'ri bo'lardi. Ruxsat aniq yo'q bo'lsa (ACTIVITY_RECOGNITION
    // rad etilgan) `hasPermissions()` aniq `false` qaytaradi.
    return await ref.read(healthServiceProvider).hasPermissions() ?? true;
  }

  /// request — tizim ruxsat oynasini ko'rsatadi.
  ///
  /// Ruxsat berilsa darrov sinxron qilinadi: foydalanuvchi natijani
  /// (qadamlari ko'rinishini) shu zahoti ko'rsin.
  Future<HealthPermissionResult> request() async {
    final health = ref.read(healthServiceProvider);

    final granted = await health.requestPermissions();
    state = AsyncData(granted);

    if (granted) {
      await ref.read(healthSyncProvider.notifier).sync();
      return HealthPermissionResult.granted;
    }

    // Android: "boshqa so'ramang" tanlangan bo'lsa qayta so'rash foyda
    // bermaydi — foydalanuvchini sozlamalarga yo'naltirish kerak.
    if (await Permission.activityRecognition.isPermanentlyDenied) {
      return HealthPermissionResult.permanentlyDenied;
    }
    return HealthPermissionResult.denied;
  }

  /// openSettings — tizim sozlamalarini ochadi (butunlay rad etilgan holat).
  Future<void> openSettings() => openAppSettings();

  /// refresh — sozlamalardan qaytgach holatni qayta o'qiydi.
  Future<void> refresh() async {
    // build() bilan bir xil qoida: noma'lum (null) — bezovta qilmaymiz.
    state = AsyncData(
        await ref.read(healthServiceProvider).hasPermissions() ?? true);
  }
}

/// autoDispose EMAS: holat tab almashganda ham saqlanib qolsin.
final healthPermissionProvider =
    AsyncNotifierProvider<HealthPermissionController, bool>(
        HealthPermissionController.new);
