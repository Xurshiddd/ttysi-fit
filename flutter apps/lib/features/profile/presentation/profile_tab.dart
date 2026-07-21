import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../../core/theme/app_colors.dart';
import '../../achievement/application/achievement_providers.dart';
import '../../achievement/presentation/achievements_card.dart';
import '../../activity/application/activity_providers.dart';
import '../../fitcoin/presentation/coin_balance_card.dart';
import '../../rating/application/rating_providers.dart';
import '../application/profile_providers.dart';
import '../data/profile_models.dart';

/// ProfileTab — foydalanuvchi profili: shaxsiy ma'lumot, tashkiliy bog'lanish,
/// yig'ma statistika va tahrirlash.
class ProfileTab extends ConsumerWidget {
  const ProfileTab({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final profile = ref.watch(profileProvider);

    return RefreshIndicator(
      onRefresh: () async {
        ref.invalidate(profileProvider);
        ref.invalidate(activityStatsProvider);
        ref.invalidate(myRatingProvider);
        ref.invalidate(earnedAchievementsProvider);
      },
      child: profile.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => ListView(
          padding: const EdgeInsets.all(24),
          children: [
            const SizedBox(height: 80),
            Text(s.t('common.error'),
                textAlign: TextAlign.center,
                style: const TextStyle(color: AppColors.muted)),
            TextButton(
              onPressed: () => ref.invalidate(profileProvider),
              child: Text(s.t('common.retry')),
            ),
          ],
        ),
        data: (p) => ListView(
          padding: const EdgeInsets.all(16),
          children: [
            _Header(profile: p),
            const SizedBox(height: 20),
            const _StatsCard(),
            const SizedBox(height: 12),
            const CoinBalanceCard(),
            const SizedBox(height: 12),
            const AchievementsCard(),
            const SizedBox(height: 16),
            _InfoCard(profile: p),
            if (p.bio.isNotEmpty) ...[
              const SizedBox(height: 16),
              _BioCard(bio: p.bio),
            ],
            const SizedBox(height: 20),
            FilledButton.icon(
              style: FilledButton.styleFrom(
                backgroundColor: AppColors.primary,
                minimumSize: const Size.fromHeight(52),
              ),
              icon: const Icon(Icons.edit_outlined),
              label: Text(s.t('profile.edit')),
              onPressed: () => _showEditSheet(context, p),
            ),
            // Chiqish tugmasi Sozlamalar bo'limiga ko'chdi: u profil
            // ma'lumoti emas, hisob amali.
            const SizedBox(height: 12),
          ],
        ),
      ),
    );
  }

  /// Tahrirlash oynasi.
  void _showEditSheet(BuildContext context, UserProfile p) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(20))),
      builder: (_) => ProfileEditSheet(profile: p),
    );
  }
}

/// ProfileEditSheet — profil tahrirlash oynasi.
///
/// Alohida ConsumerStatefulWidget: `showModalBottomSheet` ichida tashqi `ref`
/// bilan `watch` qilinsa, oyna provider o'zgarganda QAYTA CHIZILMAYDI va
/// saqlash tugmasi birinchi o'qilgan holatda (AsyncLoading) muzlab qoladi.
/// O'z ref'i bo'lgan widget bu muammoni yechadi.
///
/// Faqat backend qabul qiladigan uch maydon bor — ism/fakultet/kurs HEMIS'dan
/// keladi va bu yerda o'zgartirilmaydi.
class ProfileEditSheet extends ConsumerStatefulWidget {
  const ProfileEditSheet({super.key, required this.profile});
  final UserProfile profile;

  @override
  ConsumerState<ProfileEditSheet> createState() => _ProfileEditSheetState();
}

class _ProfileEditSheetState extends ConsumerState<ProfileEditSheet> {
  late final TextEditingController _phone =
      TextEditingController(text: widget.profile.phone);
  late final TextEditingController _bio =
      TextEditingController(text: widget.profile.bio);
  late String _lang = widget.profile.language;

  @override
  void dispose() {
    _phone.dispose();
    _bio.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    final s = S.of(context);
    final ok = await ref.read(profileEditProvider.notifier).save(ProfileUpdate(
          phone: _phone.text.trim(),
          bio: _bio.text.trim(),
          language: _lang,
        ));
    if (!mounted) return;
    if (ok) {
      // Til o'zgargan bo'lsa ilova tilini ham moslaymiz.
      ref.read(localeControllerProvider.notifier).setLocale(Locale(_lang));
      Navigator.pop(context);
    } else {
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(s.t('profile.saveError'))));
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    final saving = ref.watch(profileEditProvider).isLoading;

    return Padding(
      padding: EdgeInsets.fromLTRB(
          20, 20, 20, MediaQuery.of(context).viewInsets.bottom + 20),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(s.t('profile.edit'),
              style: Theme.of(context).textTheme.titleLarge),
          const SizedBox(height: 4),
          Text(s.t('profile.hemisNote'),
              style: const TextStyle(color: AppColors.muted, fontSize: 12)),
          const SizedBox(height: 16),
          TextField(
            controller: _phone,
            keyboardType: TextInputType.phone,
            decoration: InputDecoration(
              labelText: s.t('profile.phone'),
              hintText: '+998901234567',
              prefixIcon: const Icon(Icons.phone_outlined),
            ),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _bio,
            maxLines: 3,
            maxLength: 500,
            decoration: InputDecoration(
              labelText: s.t('profile.bio'),
              hintText: s.t('profile.bioHint'),
              alignLabelWithHint: true,
            ),
          ),
          const SizedBox(height: 4),
          Text(s.t('profile.language'),
              style: const TextStyle(color: AppColors.muted, fontSize: 12)),
          const SizedBox(height: 8),
          // Qisqa yorliqlar: to'liq nomlar ("O'zbekcha") tor ekranda ikki
          // qatorga bo'linib ketadi.
          SegmentedButton<String>(
            segments: const [
              ButtonSegment(value: 'uz', label: Text('UZ')),
              ButtonSegment(value: 'ru', label: Text('RU')),
              ButtonSegment(value: 'en', label: Text('EN')),
            ],
            selected: {_lang},
            onSelectionChanged: (v) => setState(() => _lang = v.first),
          ),
          const SizedBox(height: 20),
          FilledButton(
            style: FilledButton.styleFrom(
                backgroundColor: AppColors.accent,
                minimumSize: const Size.fromHeight(52)),
            onPressed: saving ? null : _save,
            child: saving
                ? const SizedBox(
                    height: 20,
                    width: 20,
                    child: CircularProgressIndicator(
                        strokeWidth: 2, color: Colors.white))
                : Text(s.t('common.save')),
          ),
        ],
      ),
    );
  }
}

class _Header extends StatelessWidget {
  const _Header({required this.profile});
  final UserProfile profile;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    return Column(
      children: [
        CircleAvatar(
          radius: 44,
          backgroundColor: AppColors.primary,
          backgroundImage: profile.avatarUrl.isNotEmpty
              ? NetworkImage(profile.avatarUrl)
              : null,
          child: profile.avatarUrl.isEmpty
              ? Text(profile.initials,
                  style: const TextStyle(
                      fontSize: 30,
                      fontWeight: FontWeight.w700,
                      color: Colors.white))
              : null,
        ),
        const SizedBox(height: 12),
        Text(
          profile.fullName,
          textAlign: TextAlign.center,
          style: Theme.of(context)
              .textTheme
              .titleLarge
              ?.copyWith(fontWeight: FontWeight.w700),
        ),
        const SizedBox(height: 6),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
          decoration: BoxDecoration(
            color: AppColors.accent.withValues(alpha: 0.12),
            borderRadius: BorderRadius.circular(20),
          ),
          child: Text(
            s.t('role.${profile.role}'),
            style: const TextStyle(
                color: AppColors.accent,
                fontSize: 12,
                fontWeight: FontWeight.w600),
          ),
        ),
        if (profile.email.isNotEmpty) ...[
          const SizedBox(height: 6),
          Text(profile.email,
              style: const TextStyle(color: AppColors.muted, fontSize: 13)),
        ],
      ],
    );
  }
}

/// _StatsCard — jami qadam va reytingdagi o'rin (mavjud providerlardan).
class _StatsCard extends ConsumerWidget {
  const _StatsCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final s = S.of(context);
    final stats = ref.watch(activityStatsProvider);
    final rank = ref.watch(myRatingProvider);

    Widget cell(String label, String value) => Expanded(
          child: Column(children: [
            Text(value,
                style: const TextStyle(
                    fontWeight: FontWeight.w800, fontSize: 18)),
            const SizedBox(height: 2),
            Text(label,
                textAlign: TextAlign.center,
                style: const TextStyle(color: AppColors.muted, fontSize: 12)),
          ]),
        );

    return Container(
      padding: const EdgeInsets.symmetric(vertical: 16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(children: [
        cell(s.t('rating.all'),
            stats.valueOrNull == null ? '—' : _fmt(stats.value!.totalSteps)),
        cell(s.t('rating.myRank'),
            rank.valueOrNull == null ? '—' : '${rank.value!.globalRank}'),
        cell(s.t('rating.facultyRank'),
            rank.valueOrNull == null ? '—' : '${rank.value!.facultyRank}'),
      ]),
    );
  }
}

/// _InfoCard — tashkiliy va shaxsiy ma'lumotlar. Bo'sh maydonlar ko'rsatilmaydi
/// (admin va talaba profillari juda farq qiladi).
class _InfoCard extends StatelessWidget {
  const _InfoCard({required this.profile});
  final UserProfile profile;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);

    final rows = <({String label, String value})>[
      (label: s.t('profile.faculty'), value: profile.facultyName),
      (label: s.t('profile.department'), value: profile.departmentName),
      (label: s.t('profile.group'), value: profile.groupName),
      (label: s.t('profile.course'), value: profile.course?.toString() ?? ''),
      (label: s.t('profile.position'), value: profile.position),
      (label: s.t('profile.specialty'), value: profile.specialty),
      (label: s.t('profile.hemisLogin'), value: profile.hemisLogin),
      (label: s.t('profile.phone'), value: profile.phone),
    ].where((r) => r.value.isNotEmpty).toList();

    if (rows.isEmpty) return const SizedBox.shrink();

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        children: [
          for (var i = 0; i < rows.length; i++) ...[
            if (i > 0)
              Divider(height: 20, color: AppColors.muted.withValues(alpha: 0.15)),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                SizedBox(
                  width: 110,
                  child: Text(rows[i].label,
                      style: const TextStyle(
                          color: AppColors.muted, fontSize: 13)),
                ),
                Expanded(
                  child: Text(rows[i].value,
                      style: const TextStyle(
                          fontSize: 14, fontWeight: FontWeight.w500)),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }
}

class _BioCard extends StatelessWidget {
  const _BioCard({required this.bio});
  final String bio;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(s.t('profile.bio'),
              style: const TextStyle(color: AppColors.muted, fontSize: 12)),
          const SizedBox(height: 6),
          Text(bio, style: const TextStyle(fontSize: 14)),
        ],
      ),
    );
  }
}

String _fmt(int n) {
  final str = n.toString();
  final buf = StringBuffer();
  for (var i = 0; i < str.length; i++) {
    if (i > 0 && (str.length - i) % 3 == 0) buf.write(' ');
    buf.write(str[i]);
  }
  return buf.toString();
}
