/// UserInfo — backend qaytaradigan xavfsiz foydalanuvchi ma'lumotlari.
class UserInfo {
  final String id;
  final String fullName;
  final String email;
  final String role;
  final String language;

  const UserInfo({
    required this.id,
    required this.fullName,
    required this.email,
    required this.role,
    required this.language,
  });

  factory UserInfo.fromJson(Map<String, dynamic> j) => UserInfo(
        id: (j['id'] ?? '').toString(),
        fullName: (j['full_name'] ?? '').toString(),
        email: (j['email'] ?? '').toString(),
        role: (j['role'] ?? '').toString(),
        language: (j['language'] ?? 'uz').toString(),
      );
}

/// AuthTokens — login/register javobi: { access_token, refresh_token, user }.
class AuthTokens {
  final String accessToken;
  final String refreshToken;
  final UserInfo user;

  const AuthTokens({
    required this.accessToken,
    required this.refreshToken,
    required this.user,
  });

  factory AuthTokens.fromJson(Map<String, dynamic> j) => AuthTokens(
        accessToken: (j['access_token'] ?? '').toString(),
        refreshToken: (j['refresh_token'] ?? '').toString(),
        user: UserInfo.fromJson(
          (j['user'] as Map?)?.cast<String, dynamic>() ?? const {},
        ),
      );
}

/// UserSession — "Mening qurilmalarim" ro'yxatidagi qator.
class UserSession {
  final String id;
  final String deviceId;
  final String deviceName;
  final String platform;
  final String ip;
  final DateTime? lastSeenAt;

  const UserSession({
    required this.id,
    required this.deviceId,
    required this.deviceName,
    required this.platform,
    required this.ip,
    required this.lastSeenAt,
  });

  factory UserSession.fromJson(Map<String, dynamic> j) => UserSession(
        id: (j['id'] ?? '').toString(),
        deviceId: (j['device_id'] ?? '').toString(),
        deviceName: (j['device_name'] ?? '').toString(),
        platform: (j['platform'] ?? '').toString(),
        ip: (j['ip'] ?? '').toString(),
        lastSeenAt: j['last_seen_at'] == null
            ? null
            : DateTime.tryParse(j['last_seen_at'].toString()),
      );
}
