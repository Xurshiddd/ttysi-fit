import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api/dio_client.dart';
import 'profile_models.dart';

final profileRepositoryProvider =
    Provider<ProfileRepository>((ref) => ProfileRepository(ref.read(dioProvider)));

/// ProfileRepository — profil API chaqiruvlari.
class ProfileRepository {
  ProfileRepository(this._dio);
  final Dio _dio;

  /// get — o'z profili (fakultet/guruh nomi bilan).
  Future<UserProfile> get() async {
    final res = await _dio.get('/users/me');
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return UserProfile.fromJson(data);
  }

  /// update — profilni yangilaydi va yangilangan holatini qaytaradi.
  Future<UserProfile> update(ProfileUpdate u) async {
    final res = await _dio.put('/users/me', data: u.toJson());
    final data = (res.data['data'] as Map).cast<String, dynamic>();
    return UserProfile.fromJson(data);
  }
}
