import 'package:characters/characters.dart';

/// UserProfile — `GET /users/me` javobi (fakultet/kafedra/guruh nomi bilan).
class UserProfile {
  final String id;
  final String fullName;
  final String email;
  final String phone;
  final String role;
  final String avatarUrl;
  final String bio;
  final String language;

  final String gender;
  final int? course;
  final String position;
  final String specialty;
  final String hemisLogin;

  final String facultyName;
  final String departmentName;
  final String groupName;

  const UserProfile({
    required this.id,
    required this.fullName,
    required this.email,
    required this.phone,
    required this.role,
    required this.avatarUrl,
    required this.bio,
    required this.language,
    required this.gender,
    required this.course,
    required this.position,
    required this.specialty,
    required this.hemisLogin,
    required this.facultyName,
    required this.departmentName,
    required this.groupName,
  });

  factory UserProfile.fromJson(Map<String, dynamic> j) => UserProfile(
        id: (j['id'] ?? '').toString(),
        fullName: (j['full_name'] ?? '').toString(),
        email: (j['email'] ?? '').toString(),
        phone: (j['phone'] ?? '').toString(),
        role: (j['role'] ?? '').toString(),
        avatarUrl: (j['avatar_url'] ?? '').toString(),
        bio: (j['bio'] ?? '').toString(),
        language: (j['language'] ?? 'uz').toString(),
        gender: (j['gender'] ?? '').toString(),
        course: j['course'] is num ? (j['course'] as num).toInt() : null,
        position: (j['position'] ?? '').toString(),
        specialty: (j['specialty'] ?? '').toString(),
        hemisLogin: (j['hemis_login'] ?? '').toString(),
        facultyName: (j['faculty_name'] ?? '').toString(),
        departmentName: (j['department_name'] ?? '').toString(),
        groupName: (j['group_name'] ?? '').toString(),
      );

  /// initials — avatar rasmi bo'lmaganda ko'rsatiladigan bosh harflar.
  String get initials {
    final parts = fullName.trim().split(RegExp(r'\s+'))
      ..removeWhere((p) => p.isEmpty);
    if (parts.isEmpty) return '?';
    if (parts.length == 1) return parts.first.characters.first.toUpperCase();
    return (parts[0].characters.first + parts[1].characters.first).toUpperCase();
  }
}

/// ProfileUpdate — `PUT /users/me` so'rovi.
///
/// null — "tegilmasin", '' — "tozalansin". Backend faqat shu uch maydonni
/// qabul qiladi: qolganini (ism, fakultet, kurs...) HEMIS sync boshqaradi va
/// bu yerdan o'zgartirib bo'lmaydi.
class ProfileUpdate {
  final String? phone;
  final String? bio;
  final String? language;

  const ProfileUpdate({this.phone, this.bio, this.language});

  Map<String, dynamic> toJson() => {
        if (phone != null) 'phone': phone,
        if (bio != null) 'bio': bio,
        if (language != null) 'language': language,
      };
}
