import 'package:flutter/material.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';

import 'voucher_status_view.dart';

class VoucherCard extends StatelessWidget {
  const VoucherCard({required this.voucher, this.onTap, super.key});

  final Voucher voucher;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Expanded(
                    child: Text(
                      voucher.code,
                      style: Theme.of(context)
                          .textTheme
                          .titleMedium
                          ?.copyWith(fontWeight: FontWeight.bold),
                    ),
                  ),
                  VoucherStatusView(status: voucher.status),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                '${voucher.program} · ${voucher.category}',
                style: Theme.of(context).textTheme.bodySmall,
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    '${formatCents(voucher.remainingCents)} left',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  Text(
                    'of ${formatCents(voucher.faceValueCents)}',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                'Expires ${voucher.expiresAt.toLocal().toString().split(' ').first}',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
