import 'package:flutter/material.dart';

/// A small pill badge for showing a domain status string (e.g. a voucher's
/// `issued` / `redeemed` / `expired` state) with a caller-chosen color.
///
/// Kept generic on purpose — this package doesn't know what "issued" or
/// "redeemed" mean, only how to render `(label, color)` as a badge. Domain
/// packages own mapping their own enums to a color.
class StatusBadge extends StatelessWidget {
  const StatusBadge({
    required this.label,
    required this.color,
    super.key,
  });

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.12),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: color.withOpacity(0.4)),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: color,
          fontWeight: FontWeight.w600,
          fontSize: 12,
        ),
      ),
    );
  }
}
