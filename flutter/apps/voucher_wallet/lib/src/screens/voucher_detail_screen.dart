import 'package:flutter/material.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';

import '../app_scope.dart';
import '../widgets/voucher_status_view.dart';

/// Full detail for one voucher: balance, expiry, and every redemption
/// recorded against it on the ledger.
class VoucherDetailScreen extends StatefulWidget {
  const VoucherDetailScreen({required this.voucherCode, super.key});

  final String voucherCode;

  @override
  State<VoucherDetailScreen> createState() => _VoucherDetailScreenState();
}

class _VoucherDetailScreenState extends State<VoucherDetailScreen> {
  Future<Voucher>? _future;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _future ??= _load();
  }

  Future<Voucher> _load() =>
      AppScope.of(context).vouchers.getVoucher(widget.voucherCode);

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(widget.voucherCode)),
      body: FutureBuilder<Voucher>(
        future: _future,
        builder: (context, snapshot) {
          if (snapshot.connectionState == ConnectionState.waiting) {
            return const LoadingView();
          }
          if (snapshot.hasError) {
            return ErrorView(
              message: '${snapshot.error}',
              onRetry: () => setState(() => _future = _load()),
            );
          }
          final voucher = snapshot.data!;
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    voucher.program,
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                  VoucherStatusView(status: voucher.status),
                ],
              ),
              const SizedBox(height: 4),
              Text(
                '${voucher.category} · holder ${voucher.holderId}',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const SizedBox(height: 16),
              _StatRow(
                label: 'Remaining',
                value: formatCents(voucher.remainingCents),
              ),
              _StatRow(
                label: 'Face value',
                value: formatCents(voucher.faceValueCents),
              ),
              _StatRow(
                label: 'Redeemed',
                value: formatCents(voucher.redeemedCents),
              ),
              _StatRow(
                label: 'Expires',
                value: voucher.expiresAt.toLocal().toString().split(' ').first,
              ),
              const SizedBox(height: 24),
              Text(
                'Redemption history',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              const SizedBox(height: 8),
              if (voucher.redemptions.isEmpty)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 16),
                  child: Text('No redemptions yet.'),
                )
              else
                ...voucher.redemptions.reversed.map(
                  (redemption) => Card(
                    child: ListTile(
                      title: Text(formatCents(redemption.amountCents)),
                      subtitle: Text(
                        'Merchant ${redemption.merchantId}\n'
                        '${redemption.redeemedAt.toLocal()}\n'
                        'Balance after: ${formatCents(redemption.remainingAfterCents)}',
                      ),
                      isThreeLine: true,
                    ),
                  ),
                ),
            ],
          );
        },
      ),
    );
  }
}

class _StatRow extends StatelessWidget {
  const _StatRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: Theme.of(context).textTheme.bodyMedium),
          Text(
            value,
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(fontWeight: FontWeight.bold),
          ),
        ],
      ),
    );
  }
}
