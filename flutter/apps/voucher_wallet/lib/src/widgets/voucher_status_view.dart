import 'package:flutter/material.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';

/// Renders a [VoucherStatus] as a [StatusBadge] with this app's chosen
/// label/color mapping — `nanayam_ui_kit`'s `StatusBadge` itself has no
/// opinion on what any status means.
class VoucherStatusView extends StatelessWidget {
  const VoucherStatusView({required this.status, super.key});

  final VoucherStatus status;

  @override
  Widget build(BuildContext context) {
    final (label, color) = switch (status) {
      VoucherStatus.issued => ('Issued', Colors.blue),
      VoucherStatus.partiallyRedeemed => ('Partially redeemed', Colors.orange),
      VoucherStatus.fullyRedeemed => ('Fully redeemed', Colors.grey),
      VoucherStatus.expired => ('Expired', Colors.red),
    };
    return StatusBadge(label: label, color: color);
  }
}
