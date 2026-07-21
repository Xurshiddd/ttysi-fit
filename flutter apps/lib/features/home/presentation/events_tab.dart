import 'package:flutter/material.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../challenge/presentation/challenge_tab.dart';
import '../../competition/presentation/competition_list.dart';

/// EventsTab — "Tadbirlar": Chellenjlar va Musobaqalar bitta tabda.
///
/// Nega birlashtirildi: Material pastki navigatsiyada 5 tadan ko'p element
/// tavsiya etilmaydi (yorliqlar siqilib, teginish nishoni kichrayadi).
/// Chellenj va musobaqa — mazmunan yaqin (ikkalasi ham tadbir, mukofotli),
/// shuning uchun segment bilan ajratish tabiiy.
class EventsTab extends StatefulWidget {
  const EventsTab({super.key});

  @override
  State<EventsTab> createState() => _EventsTabState();
}

class _EventsTabState extends State<EventsTab> {
  int _segment = 0;

  @override
  Widget build(BuildContext context) {
    final s = S.of(context);

    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
          child: SegmentedButton<int>(
            segments: [
              ButtonSegment(value: 0, label: Text(s.t('nav.challenge'))),
              ButtonSegment(value: 1, label: Text(s.t('nav.competition'))),
            ],
            selected: {_segment},
            onSelectionChanged: (v) => setState(() => _segment = v.first),
          ),
        ),
        // IndexedStack: segment almashganda ro'yxat holati (scroll) saqlanadi.
        Expanded(
          child: IndexedStack(
            index: _segment,
            children: const [ChallengeTab(), CompetitionList()],
          ),
        ),
      ],
    );
  }
}
