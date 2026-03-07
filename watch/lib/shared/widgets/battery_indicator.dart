import 'package:flutter/material.dart';
import '../../core/theme/app_theme.dart';

class BatteryIndicator extends StatelessWidget {
  final int level;
  final bool isCharging;

  const BatteryIndicator({
    super.key,
    required this.level,
    this.isCharging = false,
  });

  Color get _color {
    if (level > 50) return AppTheme.cta;
    if (level > 20) return AppTheme.warning;
    return AppTheme.error;
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (isCharging) ...[
          Icon(
            Icons.electric_bolt,
            size: 14,
            color: AppTheme.cta,
          ),
          const SizedBox(width: AppTheme.spaceXs),
        ],
        Icon(
          _getBatteryIcon(),
          size: 18,
          color: _color,
        ),
        const SizedBox(width: AppTheme.spaceXs),
        Text(
          '$level%',
          style: AppTheme.bodyStyle(
            fontSize: 11,
            fontWeight: FontWeight.w500,
            color: _color,
          ),
        ),
      ],
    );
  }

  IconData _getBatteryIcon() {
    if (level > 80) return Icons.battery_full;
    if (level > 60) return Icons.battery_5_bar;
    if (level > 40) return Icons.battery_4_bar;
    if (level > 20) return Icons.battery_2_bar;
    return Icons.battery_alert;
  }
}
