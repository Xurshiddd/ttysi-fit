// Tab layout testlari.
//
// NEGA BU TESTLAR BOR: AppTheme dagi `minimumSize: Size.fromHeight(52)` minimal
// kenglikni Infinity qilib qo'ygan edi. Row ichidagi FilledButton (Faollik
// tabidagi "Sinxronlash") layout qilinmay, "BoxConstraints forces an infinite
// width" assertion tashlagan va BUTUN tab bo'sh chizilgan edi.
//
// `flutter analyze` bu bug'ni ko'rmaydi — u faqat chizishda chiqadi. Shu sababli
// har bir tab kamida bir marta chizilib, exception yo'qligi tekshiriladi.

import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:ttysi_fit/core/i18n/app_localizations.dart';
import 'package:ttysi_fit/core/theme/app_theme.dart';
import 'package:ttysi_fit/features/achievement/application/achievement_providers.dart';
import 'package:ttysi_fit/features/achievement/data/achievement_models.dart';
import 'package:ttysi_fit/features/achievement/presentation/achievements_card.dart';
import 'package:ttysi_fit/features/achievement/presentation/achievements_screen.dart';
import 'package:ttysi_fit/features/activity/application/activity_providers.dart';
import 'package:ttysi_fit/features/activity/data/activity_models.dart';
import 'package:ttysi_fit/features/activity/presentation/activity_tab.dart';
import 'package:ttysi_fit/features/challenge/application/challenge_providers.dart';
import 'package:ttysi_fit/features/challenge/data/challenge_models.dart';
import 'package:ttysi_fit/features/challenge/presentation/challenge_tab.dart';
import 'package:ttysi_fit/features/competition/application/competition_providers.dart';
import 'package:ttysi_fit/features/competition/data/competition_models.dart';
import 'package:ttysi_fit/features/competition/presentation/competition_list.dart';
import 'package:ttysi_fit/features/fitcoin/application/fitcoin_providers.dart';
import 'package:ttysi_fit/features/fitcoin/data/fitcoin_models.dart';
import 'package:ttysi_fit/features/news/application/news_providers.dart';
import 'package:ttysi_fit/features/news/data/news_models.dart';
import 'package:ttysi_fit/features/news/presentation/news_section.dart';
import 'package:ttysi_fit/features/training/application/training_providers.dart';
import 'package:ttysi_fit/features/training/data/training_models.dart';
import 'package:ttysi_fit/features/training/presentation/training_list.dart';
import 'package:ttysi_fit/features/profile/application/profile_providers.dart';
import 'package:ttysi_fit/features/profile/data/profile_models.dart';
import 'package:ttysi_fit/features/profile/presentation/profile_tab.dart';
import 'package:ttysi_fit/features/rating/application/rating_providers.dart';
import 'package:ttysi_fit/features/rating/data/rating_models.dart';
import 'package:ttysi_fit/features/settings/presentation/settings_tab.dart';

const _stats = ActivityStats(
  todaySteps: 1726,
  todayCalories: 42,
  todayDistanceM: 1200,
  todayActiveMin: 15,
  weekSteps: 1726,
  monthSteps: 1726,
  totalSteps: 1726,
);

const _myRating = MyRating(
  globalRank: 1,
  facultyRank: 1,
  totalUsers: 1,
  totalSteps: 1726,
);

const _balance = CoinBalance(balance: 125, earned: 125, spent: 0);

const _profile = UserProfile(
  id: '64deb502-f3fa-4203-882f-3b42e07dbdd2',
  fullName: 'TO‘RAYEV ZAFARJON O‘KTAMOVICH',
  email: 'test@ttyesi.uz',
  phone: '+998901234567',
  role: 'student',
  avatarUrl: '', // bo'sh — tarmoqqa chiqmasin, initials ko'rsatiladi
  bio: 'Sport ishqibozi',
  language: 'uz',
  gender: 'male',
  course: 2,
  position: '',
  specialty: 'Iqtisodiyot',
  hemisLogin: '15703',
  facultyName: 'Iqtisodiyot fakulteti',
  departmentName: '',
  groupName: '13-25-M',
);

/// _phoneViewport — testni real telefon o'lchamiga qo'yadi (360×800 dp).
///
/// Default test oynasi 800×600 (landshaft planshet). ListView ekrandan tashqari
/// bolalarni qurmaydi, shuning uchun uzun ekranda pastdagi elementlar testda
/// "yo'q" bo'lib ko'rinadi va tekshiruv yolg'on yiqiladi.
void _phoneViewport(WidgetTester tester) {
  tester.view.physicalSize = const Size(1080, 2400);
  tester.view.devicePixelRatio = 3.0;
  addTearDown(tester.view.reset);
}

/// _harness — real AppTheme bilan o'raydi. Tema MAJBURIY: bug aynan temada edi,
/// default ThemeData bilan sinasak uni o'tkazib yuborardik.
Widget _harness(Widget child, {required List<Override> overrides}) {
  return ProviderScope(
    overrides: overrides,
    child: MaterialApp(
      theme: AppTheme.light(),
      darkTheme: AppTheme.dark(),
      locale: const Locale('uz'),
      supportedLocales: supportedLocales,
      localizationsDelegates: const [
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      home: Scaffold(body: child),
    ),
  );
}

void main() {
  group('ActivityTab', () {
    testWidgets('layout xatosisiz chiziladi va Sinxronlash tugmasi ko‘rinadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ActivityTab(),
        overrides: [
          activityStatsProvider.overrideWith((ref) async => _stats),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull,
          reason: 'Row ichidagi tugma layout assertion tashlamasligi kerak');
      expect(find.text('Sinxronlash'), findsOneWidget);
      expect(find.text('Qadam sanagich'), findsOneWidget);
      expect(find.text('1 726'), findsWidgets);
    });

    testWidgets('stats xatosida ham karta va tugma qoladi', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ActivityTab(),
        overrides: [
          activityStatsProvider
              .overrideWith((ref) async => throw Exception('500')),
        ],
      ));
      await tester.pumpAndSettle();

      // Statistika yiqilsa ham sinxronlash kartasi ko'rinishi kerak —
      // aks holda foydalanuvchi qadamlarini yuklay olmaydi.
      expect(find.text('Sinxronlash'), findsOneWidget);
    });
  });

  group('ProfileTab', () {
    testWidgets('layout xatosisiz chiziladi va profil ma‘lumoti ko‘rinadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ProfileTab(),
        overrides: [
          profileProvider.overrideWith((ref) async => _profile),
          activityStatsProvider.overrideWith((ref) async => _stats),
          myRatingProvider.overrideWith((ref) async => _myRating),
          coinBalanceProvider.overrideWith((ref) async => _balance),
          earnedAchievementsProvider.overrideWith((ref) async => [_achEarned]),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('TO‘RAYEV ZAFARJON O‘KTAMOVICH'), findsOneWidget);
      expect(find.text('Talaba'), findsOneWidget); // role.student tarjimasi
      expect(find.text('Iqtisodiyot fakulteti'), findsOneWidget);
      expect(find.text('13-25-M'), findsOneWidget);

      // Bio kartasi ro'yxatning pastida — ListView ekrandan tashqaridagi
      // bolalarni qurmaydi, shuning uchun avval unga surish kerak.
      await tester.scrollUntilVisible(find.text('Sport ishqibozi'), 200,
          scrollable: find.byType(Scrollable).first);
      expect(find.text('Sport ishqibozi'), findsOneWidget);
    });

    testWidgets('bo‘sh maydonlar ko‘rsatilmaydi', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ProfileTab(),
        overrides: [
          profileProvider.overrideWith((ref) async => _profile),
          activityStatsProvider.overrideWith((ref) async => _stats),
          myRatingProvider.overrideWith((ref) async => _myRating),
          coinBalanceProvider.overrideWith((ref) async => _balance),
          earnedAchievementsProvider.overrideWith((ref) async => [_achEarned]),
        ],
      ));
      await tester.pumpAndSettle();

      // departmentName va position bo'sh — yorliqlari chiqmasligi kerak.
      expect(find.text('Kafedra'), findsNothing);
      expect(find.text('Lavozim'), findsNothing);
      // Guruh esa to'ldirilgan — chiqishi kerak.
      expect(find.text('Guruh'), findsOneWidget);
    });
  });

  group('ProfileEditSheet', () {
    // REGRESSIYA: oyna avval showModalBottomSheet ichida tashqi `ref` bilan
    // watch qilardi. Oyna provider o'zgarganda qayta chizilmagani uchun
    // saqlash tugmasi AsyncLoading holatida muzlab qolardi — ya'ni profilni
    // umuman saqlab bo'lmasdi. `flutter analyze` buni ko'rmaydi, chunki bu
    // kompilyatsiya emas, hayot-sikli xatosi.
    testWidgets('saqlash tugmasi faol bo‘ladi (spinner’da qotmaydi)',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ProfileEditSheet(profile: _profile),
        overrides: [
          profileProvider.overrideWith((ref) async => _profile),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Saqlash'), findsOneWidget);

      // Tugma bosiladigan holatda bo'lishi kerak (onPressed != null).
      final btn = tester.widget<FilledButton>(
        find.ancestor(
          of: find.text('Saqlash'),
          matching: find.byType(FilledButton),
        ),
      );
      expect(btn.onPressed, isNotNull,
          reason: 'saqlash tugmasi o‘chirilgan holatda qolmasligi kerak');
    });

    testWidgets('mavjud qiymatlar formaga to‘ldiriladi', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ProfileEditSheet(profile: _profile),
        overrides: [profileProvider.overrideWith((ref) async => _profile)],
      ));
      await tester.pumpAndSettle();

      // Maydon qiymatini controller orqali tekshiramiz: find.text bo'yicha
      // qidirish hint matni ("+998901234567") bilan to'qnashadi.
      final fields = tester.widgetList<TextField>(find.byType(TextField));
      expect(fields.map((f) => f.controller?.text),
          containsAll(['+998901234567', 'Sport ishqibozi']));

      // HEMIS boshqaradigan maydonlar haqida ogohlantirish ko'rinishi kerak.
      expect(find.textContaining('HEMIS'), findsOneWidget);
    });
  });

  group('ChallengeTab', () {
    // Backenddagi haqiqiy holat: 1726 qadam / 10000 maqsad = 17.3%.
    final joined = Challenge(
      id: '1', type: 'steps', title: '10 000 qadam',
      description: 'Kuniga 10 ming qadam', status: 'active', rewardCoins: 50,
      startsAt: DateTime(2026, 7, 1), endsAt: DateTime(2026, 8, 1),
      config: const {'target_steps': 10000},
      joined: true, progress: 1726, completed: false, rewardGranted: false,
      target: 10000, progressPct: 17.26,
    );
    // Yakunlangan, mukofoti hali olinmagan — "Olish" tugmasi chiqishi kerak.
    const claimable = Challenge(
      id: '3', type: 'steps', title: 'Test: 100 qadam',
      description: '', status: 'active', rewardCoins: 25,
      startsAt: null, endsAt: null, config: {'target_steps': 100},
      joined: true, progress: 1726, completed: true, rewardGranted: false,
      target: 100, progressPct: 100,
    );
    // Mukofot olingan — tugma o'rniga holat matni.
    const claimed = Challenge(
      id: '4', type: 'steps', title: 'Olingan chellenj',
      description: '', status: 'active', rewardCoins: 25,
      startsAt: null, endsAt: null, config: {'target_steps': 100},
      joined: true, progress: 1726, completed: true, rewardGranted: true,
      target: 100, progressPct: 100,
    );
    const notJoined = Challenge(
      id: '2', type: 'distance', title: 'Marafon',
      description: '', status: 'active', rewardCoins: 100,
      startsAt: null, endsAt: null, config: {'target_km': 42.2},
      joined: false, progress: 0, completed: false, rewardGranted: false,
      target: 42200, progressPct: 0,
    );

    testWidgets('layout xatosisiz chiziladi, progress va tugma ko‘rinadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ChallengeTab(),
        overrides: [
          challengeListProvider.overrideWith((ref) async => [joined, notJoined]),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('10 000 qadam'), findsOneWidget);
      expect(find.text('Marafon'), findsOneWidget);
      // Qo'shilgan chellenjda progress, qo'shilmaganida "Qo'shilish" tugmasi.
      expect(find.text('1 726 / 10 000'), findsOneWidget);
      expect(find.text('17%'), findsOneWidget);
      expect(find.text('Qo‘shilish'), findsOneWidget);
      expect(find.text('Davom etmoqda'), findsOneWidget);
    });

    testWidgets('bo‘sh ro‘yxatda tushuntirish ko‘rsatiladi', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ChallengeTab(),
        overrides: [challengeListProvider.overrideWith((ref) async => [])],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Hozircha aktiv chellenj yo‘q'), findsOneWidget);
    });

    testWidgets('yakunlangan chellenjda mukofot olish tugmasi chiqadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const ChallengeTab(),
        overrides: [
          challengeListProvider.overrideWith((ref) async => [claimable, claimed]),
          coinBalanceProvider.overrideWith(
              (ref) async => const CoinBalance(balance: 100, earned: 100, spent: 0)),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      // Olinmagan mukofot -> tugma; olingan -> holat matni.
      expect(find.text('Olish +25'), findsOneWidget);
      expect(find.text('Mukofot olingan ✓'), findsOneWidget);
    });

    test('progressLabel turga qarab birlikni o‘giradi', () {
      // Backend masofani metrda beradi — UI km ko'rsatishi kerak.
      expect(notJoined.progressLabel, '0.0 / 42.2 km');
      expect(joined.progressLabel, '1 726 / 10 000');
    });

    test('canClaim faqat yakunlangan va olinmagan mukofotda true', () {
      expect(claimable.canClaim, isTrue);
      expect(claimed.canClaim, isFalse, reason: 'mukofot allaqachon olingan');
      expect(joined.canClaim, isFalse, reason: 'hali yakunlanmagan');
      expect(notJoined.canClaim, isFalse, reason: 'mukofoti ham yo‘q');
    });
  });

  group('CompetitionList', () {
    // Ro'yxat ochiq, joy bor -> "Yozilish" tugmasi.
    final open = Competition(
      id: 'c1', type: 'individual', title: 'Universitet krossi',
      description: '5 km yugurish', status: 'registration', location: 'Stadion',
      rewardCoins: 40, maxParticipants: 30,
      startsAt: DateTime(2026, 8, 1), endsAt: null, regEndsAt: DateTime(2026, 7, 30),
      config: const {'sport': 'yugurish'},
      registered: false, place: null, rewardGranted: false,
      participantCount: 12, regOpen: true,
    );
    // Yozilgan -> "Bekor qilish".
    const mine = Competition(
      id: 'c2', type: 'team', title: 'Futbol turniri',
      description: '', status: 'registration', location: '',
      rewardCoins: 0, maxParticipants: 0,
      startsAt: null, endsAt: null, regEndsAt: null,
      config: {'sport': 'futbol', 'team_size': 5},
      registered: true, place: null, rewardGranted: false,
      participantCount: 8, regOpen: true,
    );
    // Joy to'lgan -> tugma emas, sabab.
    const full = Competition(
      id: 'c3', type: 'individual', title: 'Shaxmat',
      description: '', status: 'registration', location: '',
      rewardCoins: 0, maxParticipants: 4,
      startsAt: null, endsAt: null, regEndsAt: null,
      config: {'sport': 'shaxmat'},
      registered: false, place: null, rewardGranted: false,
      participantCount: 4, regOpen: false,
    );
    // Yakunlangan, o'rin bor.
    const done = Competition(
      id: 'c4', type: 'individual', title: 'Bahorgi kross',
      description: '', status: 'finished', location: '',
      rewardCoins: 20, maxParticipants: null,
      startsAt: null, endsAt: null, regEndsAt: null,
      config: {'sport': 'yugurish'},
      registered: true, place: 2, rewardGranted: true,
      participantCount: 50, regOpen: false,
    );

    testWidgets('layout xatosisiz chiziladi, holatlar to‘g‘ri', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const CompetitionList(),
        overrides: [
          competitionListProvider
              .overrideWith((ref) async => [open, mine, full, done]),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Universitet krossi'), findsOneWidget);
      expect(find.text('Yozilish'), findsOneWidget); // faqat `open` da
      expect(find.text('Bekor qilish'), findsWidgets); // yozilganlarda
      expect(find.text('Joylar to‘lgan'), findsOneWidget); // `full`
      expect(find.text('2-o‘rin'), findsOneWidget); // `done`
      expect(find.text('Ro‘yxat ochiq'), findsWidgets); // status chip
    });

    testWidgets('bo‘sh ro‘yxatda tushuntirish', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const CompetitionList(),
        overrides: [competitionListProvider.overrideWith((ref) async => [])],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Hozircha musobaqa yo‘q'), findsOneWidget);
    });

    test('slotsLabel: max=0 va null cheklovsiz degani', () {
      // Backend 0 ni "cheklovsiz" deb talqin qiladi — UI ham shunday
      // ko'rsatishi kerak, aks holda "8 / 0" chiqib qolardi.
      expect(mine.slotsLabel, '8');
      expect(done.slotsLabel, '50');
      expect(open.slotsLabel, '12 / 30');
      expect(full.slotsLabel, '4 / 4');
    });

    test('isFull faqat haqiqiy cheklovda true', () {
      expect(full.isFull, isTrue);
      expect(open.isFull, isFalse);
      expect(mine.isFull, isFalse, reason: 'max=0 — cheklovsiz');
      expect(done.isFull, isFalse, reason: 'max=null — cheklovsiz');
    });
  });

  group('NewsSection', () {
    final pinned = NewsItem(
      id: 'n1', title: 'Universitet krossi boshlanadi',
      excerpt: '15-avgust kuni markaziy stadionda', coverUrl: '',
      pinned: true, views: 42, publishedAt: DateTime(2026, 7, 17),
    );
    const plain = NewsItem(
      id: 'n2', title: 'Yangi sport zali ochildi',
      excerpt: 'Zamonaviy jihozlar bilan', coverUrl: '',
      pinned: false, views: 7, publishedAt: null,
    );

    testWidgets('layout xatosisiz chiziladi va yangiliklar ko‘rinadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const SingleChildScrollView(child: NewsSection()),
        overrides: [
          newsListProvider.overrideWith((ref) async => [pinned, plain]),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Yangiliklar'), findsOneWidget); // bo'lim sarlavhasi
      expect(find.text('Universitet krossi boshlanadi'), findsOneWidget);
      expect(find.text('Yangi sport zali ochildi'), findsOneWidget);
      expect(find.text('42'), findsOneWidget); // ko'rishlar
      // Pinned yangilikda qadash belgisi bo'lishi kerak.
      expect(find.byIcon(Icons.push_pin), findsOneWidget);
    });

    testWidgets('bo‘sh ro‘yxatda butunlay yashiriladi', (tester) async {
      // Bosh sahifada bo'sh "Yangiliklar" sarlavhasi osilib qolmasin.
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const SingleChildScrollView(child: NewsSection()),
        overrides: [newsListProvider.overrideWith((ref) async => [])],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Yangiliklar'), findsNothing);
    });

    testWidgets('xato bo‘lsa ham bosh sahifani buzmaydi', (tester) async {
      // Yangilik — yordamchi kontent: yiqilsa jimgina yashirinadi, chunki
      // bosh sahifaning asosiy qismi (qadam, reyting) muhimroq.
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const SingleChildScrollView(child: NewsSection()),
        overrides: [
          newsListProvider.overrideWith((ref) async => throw Exception('500')),
        ],
      ));
      await tester.pumpAndSettle();

      expect(find.text('Yangiliklar'), findsNothing);
    });
  });

  group('TrainingList', () {
    const items = [
      Training(
        id: 't1', title: 'Ertalabki mashqlar',
        description: 'Kuningizni energiya bilan boshlang',
        category: 'Kardio', level: 'beginner',
        videoUrl: 'https://www.youtube.com/watch?v=x', thumbnailUrl: '',
        durationMin: 15, views: 12,
      ),
      Training(
        id: 't2', title: 'Qorin mushaklari', description: '',
        category: 'Kuch', level: 'intermediate',
        videoUrl: 'https://www.youtube.com/watch?v=y', thumbnailUrl: '',
        durationMin: null, views: 0,
      ),
    ];

    testWidgets('layout xatosisiz chiziladi, kartalar va filtrlar ko‘rinadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const TrainingList(),
        overrides: [
          trainingListProvider.overrideWith((ref) async => items),
          // Kategoriyalar backenddan keladi — kodda ro'yxat yo'q (§16).
          trainingCategoriesProvider
              .overrideWith((ref) async => ['Kardio', 'Kuch', 'Yoga']),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Ertalabki mashqlar'), findsOneWidget);
      expect(find.text('Qorin mushaklari'), findsOneWidget);
      // Davomiylik faqat ko'rsatilganda.
      expect(find.text('15 daq'), findsOneWidget);

      // Filtr chiplari backenddan kelgan kategoriyalardan yasaladi (§16 —
      // kodda ro'yxat yo'q). Chip matn bo'yicha emas, kalit bo'yicha topiladi:
      // "Kardio" ayni paytda kartaning kategoriya yozuvida ham chiqadi.
      expect(find.byKey(const ValueKey('training-cat-Kardio')), findsOneWidget);
      // Chiplar qatori gorizontal suriladi, shuning uchun uzoqdagi "Yoga" ni
      // ko'rish uchun surish kerak — ListView ekrandan tashqaridagi bolalarni
      // qurmaydi.
      await tester.scrollUntilVisible(
          find.byKey(const ValueKey('training-cat-Yoga')), 100,
          scrollable: find.byType(Scrollable).first);
      expect(find.byKey(const ValueKey('training-cat-Yoga')), findsOneWidget);
    });

    testWidgets('bo‘sh ro‘yxatda tushuntirish, filtrlar qoladi', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const TrainingList(),
        overrides: [
          trainingListProvider.overrideWith((ref) async => []),
          trainingCategoriesProvider.overrideWith((ref) async => ['Kardio']),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Hozircha mashg‘ulot yo‘q'), findsOneWidget);
      // Filtrlar yo'qolmasligi kerak — foydalanuvchi boshqasini tanlay olsin.
      expect(find.byKey(const ValueKey('training-cat-Kardio')), findsOneWidget);
    });

    test('durationLabel: null bo‘lsa bo‘sh', () {
      expect(items[0].durationLabel, '15 daq');
      expect(items[1].durationLabel, '');
    });
  });

  group('UserProfile.initials', () {
    test('ko‘p so‘zli ismdan dastlabki ikki so‘zning bosh harfi', () {
      // TO‘RAYEV ZAFARJON O‘KTAMOVICH -> T + Z
      expect(_profile.initials, 'TZ');
    });

    test('bir so‘zli ismdan bitta harf', () {
      const p = UserProfile(
        id: '1', fullName: 'Aziz', email: '', phone: '', role: 'student',
        avatarUrl: '', bio: '', language: 'uz', gender: '', course: null,
        position: '', specialty: '', hemisLogin: '', facultyName: '',
        departmentName: '', groupName: '',
      );
      expect(p.initials, 'A');
    });

    test('bo‘sh ism qulamaydi', () {
      const p = UserProfile(
        id: '1', fullName: '   ', email: '', phone: '', role: 'student',
        avatarUrl: '', bio: '', language: 'uz', gender: '', course: null,
        position: '', specialty: '', hemisLogin: '', facultyName: '',
        departmentName: '', groupName: '',
      );
      expect(p.initials, '?');
    });
  });

  _achievementTests();
  _settingsTests();
}

// ─────────────────────────── Yutuqlar ───────────────────────────

const _achEarned = Achievement(
  id: 'a1', type: 'steps_total', title: 'Birinchi ming qadam',
  description: '1000 qadam bosdingiz', awardMode: 'auto', rewardCoins: 10,
  iconUrl: '', criteria: {'threshold': 1000},
  earned: true, earnedAt: null, progress: 4238, target: 1000, progressPct: 100,
  awardId: 'award-1',
);

const _achInProgress = Achievement(
  id: 'a2', type: 'distance_total', title: 'Marafon masofasi',
  description: '', awardMode: 'auto', rewardCoins: 100,
  iconUrl: '', criteria: {'threshold': 42},
  earned: false, earnedAt: null, progress: 12000, target: 42000,
  progressPct: 28.6, awardId: null,
);

void _achievementTests() {
  group('AchievementsScreen', () {
    testWidgets('layout xatosisiz chiziladi, ikkala bo‘lim ko‘rinadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const AchievementsScreen(),
        overrides: [
          achievementListProvider
              .overrideWith((ref) async => [_achEarned, _achInProgress]),
          earnedAchievementsProvider.overrideWith((ref) async => [_achEarned]),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Birinchi ming qadam'), findsOneWidget);
      expect(find.text('Marafon masofasi'), findsOneWidget);
      // Qozonilgan bo'limi sonni ko'rsatadi.
      expect(find.text('Qozonilgan (1)'), findsOneWidget);
      expect(find.text('Jarayonda'), findsOneWidget);
      // Sertifikat tugmasi FAQAT qozonilgan yutuqda.
      expect(find.text('Sertifikatni yuklab olish'), findsOneWidget);
    });

    testWidgets('bo‘sh ro‘yxatda tushuntirish chiqadi', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const AchievementsScreen(),
        overrides: [
          achievementListProvider.overrideWith((ref) async => []),
          earnedAchievementsProvider.overrideWith((ref) async => []),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.textContaining('Hozircha yutuq yo‘q'), findsOneWidget);
    });

    testWidgets('AchievementsCard profilda chiziladi', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const AchievementsCard(),
        overrides: [
          earnedAchievementsProvider.overrideWith((ref) async => [_achEarned]),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Yutuqlarim'), findsOneWidget);
      expect(find.text('1'), findsOneWidget);
    });
  });

  group('Achievement modeli', () {
    test('progressLabel masofani km ga o‘giradi', () {
      // Backend target'ni metrda beradi — km ga o'girilmasa "12000 / 42000"
      // chiqib, foydalanuvchi nimani bildirishini tushunmasdi.
      expect(_achInProgress.progressLabel, '12.0 / 42.0 km');
    });

    test('progressLabel qadamni guruhlaydi', () {
      expect(_achEarned.progressLabel, '4 238 / 1 000');
    });

    test('maqsadsiz (manual) yutuqda progressLabel bo‘sh', () {
      const manual = Achievement(
        id: 'm', type: 'manual', title: 'G‘olib', description: '',
        awardMode: 'manual', rewardCoins: 0, iconUrl: '', criteria: {},
        earned: true, earnedAt: null, progress: 0, target: 0, progressPct: 0,
        awardId: 'x',
      );
      expect(manual.progressLabel, '');
      expect(manual.isManual, isTrue);
    });

    test('hasCertificate faqat qozonilgan va award_id bor bo‘lsa', () {
      expect(_achEarned.hasCertificate, isTrue);
      expect(_achInProgress.hasCertificate, isFalse);
    });
  });
}

// ─────────────────────────── Sozlamalar ───────────────────────────

void _settingsTests() {
  group('SettingsTab', () {
    testWidgets('layout xatosisiz chiziladi, profil va sozlamalar ko‘rinadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const SettingsTab(),
        overrides: [
          profileProvider.overrideWith((ref) async => _profile),
          coinBalanceProvider.overrideWith((ref) async => _balance),
          earnedAchievementsProvider.overrideWith((ref) async => [_achEarned]),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      // Profil kartasi — sozlamalar ICHIDA.
      expect(find.text('TO‘RAYEV ZAFARJON O‘KTAMOVICH'), findsOneWidget);
      expect(find.text('Iqtisodiyot fakulteti'), findsOneWidget);
      // Asosiy raqamlar profilga kirmasdan ko'rinib turishi kerak.
      expect(find.text('125'), findsOneWidget); // coin balansi
      expect(find.text('1'), findsOneWidget); // yutuqlar soni

      expect(find.text('Til'), findsOneWidget);
      expect(find.text('Mavzu'), findsOneWidget);
      expect(find.text('Chiqish'), findsOneWidget);
    });

    testWidgets('profil yuklanmasa ham sozlamalar ishlaydi', (tester) async {
      // Profil so'rovi yiqilsa sozlamalar (til, mavzu, chiqish) baribir
      // ochiq qolishi kerak — aks holda foydalanuvchi tilni ham, chiqishni
      // ham topa olmasdi.
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const SettingsTab(),
        overrides: [
          profileProvider.overrideWith((ref) async => throw Exception('500')),
          coinBalanceProvider.overrideWith((ref) async => throw Exception('500')),
          earnedAchievementsProvider
              .overrideWith((ref) async => throw Exception('500')),
        ],
      ));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);
      expect(find.text('Til'), findsOneWidget);
      expect(find.text('Mavzu'), findsOneWidget);
      expect(find.text('Chiqish'), findsOneWidget);
    });

    testWidgets('til tanlash oynasi ochiladi', (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const SettingsTab(),
        overrides: [
          profileProvider.overrideWith((ref) async => _profile),
          coinBalanceProvider.overrideWith((ref) async => _balance),
          earnedAchievementsProvider.overrideWith((ref) async => [_achEarned]),
        ],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Til'));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);

      // Finder oyna ICHIGA cheklanadi: joriy til qatorning o'ng tomonida ham
      // yozilgan, shuning uchun find.text('O‘zbekcha') ikkitani topadi.
      final sheet = find.byType(RadioGroup<String>);
      expect(sheet, findsOneWidget);
      expect(find.descendant(of: sheet, matching: find.text('O‘zbekcha')),
          findsOneWidget);
      expect(find.descendant(of: sheet, matching: find.text('Русский')),
          findsOneWidget);
      expect(find.descendant(of: sheet, matching: find.text('English')),
          findsOneWidget);
    });

    testWidgets('mavzu tanlash oynasi uch variantni ko‘rsatadi',
        (tester) async {
      _phoneViewport(tester);
      await tester.pumpWidget(_harness(
        const SettingsTab(),
        overrides: [
          profileProvider.overrideWith((ref) async => _profile),
          coinBalanceProvider.overrideWith((ref) async => _balance),
          earnedAchievementsProvider.overrideWith((ref) async => [_achEarned]),
        ],
      ));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Mavzu'));
      await tester.pumpAndSettle();

      expect(tester.takeException(), isNull);

      final sheet = find.byType(RadioGroup<ThemeMode>);
      expect(sheet, findsOneWidget);
      expect(find.descendant(of: sheet, matching: find.text('Tizim')),
          findsOneWidget);
      expect(find.descendant(of: sheet, matching: find.text('Yorug‘')),
          findsOneWidget);
      expect(find.descendant(of: sheet, matching: find.text('Qorong‘i')),
          findsOneWidget);
    });
  });
}
