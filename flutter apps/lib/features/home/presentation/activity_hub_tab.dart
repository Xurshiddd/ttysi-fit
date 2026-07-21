import 'package:flutter/material.dart';

import '../../../core/i18n/app_localizations.dart';
import '../../activity/presentation/activity_tab.dart';
import '../../training/presentation/training_list.dart';

/// ActivityHubTab — "Faollik": shaxsiy statistika va video mashg'ulotlar.
///
/// EventsTab bilan bir xil andoza: pastki panelda 5 ta element (Material
/// maksimumi), shuning uchun yaqin mazmunli ekranlar segment bilan birlashadi.
/// Mashg'ulot va faollik bog'liq: biri "men nima qildim", ikkinchisi "nima
/// qilishim mumkin".
class ActivityHubTab extends StatefulWidget {
  const ActivityHubTab({super.key});

  @override
  State<ActivityHubTab> createState() => _ActivityHubTabState();
}

class _ActivityHubTabState extends State<ActivityHubTab> {
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
              ButtonSegment(value: 0, label: Text(s.t('nav.activity'))),
              ButtonSegment(value: 1, label: Text(s.t('nav.training'))),
            ],
            selected: {_segment},
            onSelectionChanged: (v) => setState(() => _segment = v.first),
          ),
        ),
        // IndexedStack: segment almashganda scroll holati saqlanadi.
        Expanded(
          child: IndexedStack(
            index: _segment,
            children: const [ActivityTab(), TrainingList()],
          ),
        ),
      ],
    );
  }
}
