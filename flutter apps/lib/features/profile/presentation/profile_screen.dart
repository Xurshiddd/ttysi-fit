import 'package:flutter/material.dart';

import '../../../core/i18n/app_localizations.dart';
import 'profile_tab.dart';

/// ProfileScreen — profil alohida ekran sifatida (Sozlamalar ichidan ochiladi).
///
/// ProfileTab o'zi Scaffold yasamaydi (u ilgari tab ichida turgan), shuning
/// uchun AppBar shu yerda qo'shiladi.
class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    return Scaffold(
      appBar: AppBar(title: Text(s.t('nav.profile'))),
      body: const ProfileTab(),
    );
  }
}
