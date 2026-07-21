import 'dart:io' show Platform;

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:health/health.dart';
import 'package:permission_handler/permission_handler.dart';

import 'activity_models.dart';

final healthServiceProvider = Provider<HealthService>((ref) => HealthService());

/// HealthService — telefondagi sog'liq ma'lumotlarini o'qiydi:
/// Android → Health Connect (Samsung Health qadamlarini ham beradi),
/// iOS → Apple HealthKit.
///
/// Oqim: ruxsat so'rash → bugungi qadam/kaloriya/masofani o'qish →
/// chaqiruvchi (sync provider) backend'ga POST qiladi.
class HealthService {
  final Health _health = Health();
  bool _configured = false;

  Future<void> _ensureConfigured() async {
    if (!_configured) {
      await _health.configure();
      _configured = true;
    }
  }

  /// O'qiladigan turlar (platformaga qarab masofa turi farq qiladi).
  List<HealthDataType> get _types => [
        HealthDataType.STEPS,
        HealthDataType.ACTIVE_ENERGY_BURNED,
        Platform.isIOS
            ? HealthDataType.DISTANCE_WALKING_RUNNING
            : HealthDataType.DISTANCE_DELTA,
      ];

  /// requestPermissions — ACTIVITY_RECOGNITION (Android) + Health ruxsatlari.
  /// Kerak bo'lsa foydalanuvchiga ruxsat OYNASINI ko'rsatadi.
  /// true — barcha kerakli ruxsatlar berilgan.
  Future<bool> requestPermissions() async {
    await _ensureConfigured();

    if (Platform.isAndroid) {
      final status = await Permission.activityRecognition.request();
      if (!status.isGranted) return false;
    }

    final has = await _health.hasPermissions(_types);
    if (has == true) return true;
    return await _health.requestAuthorization(_types);
  }

  /// hasPermissions — ruxsat allaqachon berilganini SO'RAMASDAN tekshiradi.
  ///
  /// Avtomatik (jim) sinxron uchun: ilova ochilishi bilan ruxsat oynasi
  /// chiqib kelishi foydalanuvchini cho'chitadi va nima uchun so'ralayotgani
  /// tushunarsiz bo'ladi. Ruxsat faqat "Sinxronlash" tugmasi bosilganda,
  /// ya'ni foydalanuvchi o'zi xohlaganda so'raladi.
  Future<bool> hasPermissions() async {
    await _ensureConfigured();

    if (Platform.isAndroid &&
        !await Permission.activityRecognition.isGranted) {
      return false;
    }
    return await _health.hasPermissions(_types) ?? false;
  }

  /// Manba nomi (backend `source` ustuni uchun).
  String get _source => Platform.isIOS ? 'healthkit' : 'health_connect';

  /// readToday — bugungi yig'ma: qadam, kaloriya (kkal), masofa (m).
  /// Ma'lumot topilmasa null qaytadi.
  Future<ActivityRecord?> readToday() async {
    final now = DateTime.now();
    return _readDay(DateTime(now.year, now.month, now.day));
  }

  /// readRecentDays — oxirgi [days] kunni (bugun ham) kun-bakun o'qiydi.
  ///
  /// Nima uchun kerak: telefon qadamni ilova yopiq bo'lsa ham sanaydi va
  /// tarixni o'zida saqlaydi. Faqat bugunni yuborsak, foydalanuvchi 3 kun
  /// ilovani ochmasa o'sha 3 kun butunlay yo'qolardi — garchi ma'lumot
  /// telefonda turgan bo'lsa ham. Backfill shuni tiklaydi.
  ///
  /// Bo'sh kunlar (0 qadam) ro'yxatga qo'shilmaydi — bekorga so'rov
  /// kattalashtirmaslik uchun.
  ///
  /// Sikl ichida platforma chaqiruvi bor, lekin bu tarmoq/DB so'rovi emas
  /// (CLAUDE.md §3.1): Health Connect qadamni faqat berilgan oraliq uchun
  /// agregatlaydi, kunlarga bo'lingan yig'ma API'si yo'q. Natija baribir
  /// backend'ga BITTA batch so'rovda ketadi.
  Future<List<ActivityRecord>> readRecentDays({int days = 7}) async {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);

    final out = <ActivityRecord>[];
    for (var i = days - 1; i >= 0; i--) {
      final rec = await _readDay(today.subtract(Duration(days: i)));
      if (rec != null) out.add(rec);
    }
    return out;
  }

  /// _readDay — bitta mahalliy kunning [00:00, kun oxiri] yig'masi.
  ///
  /// Kun chegarasi MAHALLIY vaqtda olinadi (DateTime.now() mahalliy):
  /// backend ham shu sanani `date` maydonidan oladi, shuning uchun telefon
  /// va server bir xil kunni ko'radi.
  Future<ActivityRecord?> _readDay(DateTime dayStart) async {
    await _ensureConfigured();

    final now = DateTime.now();
    var end = dayStart.add(const Duration(days: 1));
    // Bugungi kun uchun kelajakni so'ramaymiz.
    if (end.isAfter(now)) end = now;
    if (!end.isAfter(dayStart)) return null;

    // Qadam — maxsus agregatsiya metodi (dublikatlarsiz).
    final steps = await _health.getTotalStepsInInterval(dayStart, end) ?? 0;

    // Kaloriya va masofa — nuqtalarni yig'ib chiqamiz.
    double calories = 0;
    double distanceM = 0;
    final points = await _health.getHealthDataFromTypes(
      types: _types.where((t) => t != HealthDataType.STEPS).toList(),
      startTime: dayStart,
      endTime: end,
    );
    for (final p in _health.removeDuplicates(points)) {
      final v = p.value;
      if (v is! NumericHealthValue) continue;
      final n = v.numericValue.toDouble();
      if (p.type == HealthDataType.ACTIVE_ENERGY_BURNED) {
        calories += n;
      } else {
        distanceM += n; // DISTANCE_DELTA / DISTANCE_WALKING_RUNNING — metr
      }
    }

    if (steps == 0 && calories == 0 && distanceM == 0) return null;

    return ActivityRecord(
      date: dayStart,
      steps: steps,
      calories: calories,
      distanceM: distanceM,
      activeMin: 0, // Health Connect'da to'g'ridan-to'g'ri yo'q — keyin EXERCISE_SESSION'dan
      source: _source,
    );
  }
}
