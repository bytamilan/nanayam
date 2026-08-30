import 'package:flutter/material.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';

import '../app_scope.dart';

/// Business/merchant view: redeem an amount against a voucher code, and see
/// this merchant's recent redemptions.
class RedeemTab extends StatefulWidget {
  const RedeemTab({super.key});

  @override
  State<RedeemTab> createState() => _RedeemTabState();
}

class _RedeemTabState extends State<RedeemTab> {
  final _formKey = GlobalKey<FormState>();
  final _merchantIdController = TextEditingController();
  final _voucherCodeController = TextEditingController();
  final _amountController = TextEditingController();
  bool _submitting = false;
  String? _error;
  Future<List<VoucherRedemption>>? _historyFuture;
  bool _initialized = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_initialized) return;
    _initialized = true;
    _merchantIdController.text = AppScope.of(context).currentUser?.username ?? '';
    _refreshHistory();
  }

  @override
  void dispose() {
    _merchantIdController.dispose();
    _voucherCodeController.dispose();
    _amountController.dispose();
    super.dispose();
  }

  void _refreshHistory() {
    final merchantId = _merchantIdController.text.trim();
    if (merchantId.isEmpty) return;
    setState(() {
      _historyFuture =
          AppScope.of(context).vouchers.listMerchantRedemptions(merchantId);
    });
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _submitting = true;
      _error = null;
    });

    try {
      final amountCents =
          (double.parse(_amountController.text.trim()) * 100).round();
      final redemption = await AppScope.of(context).vouchers.redeemVoucher(
            code: _voucherCodeController.text.trim(),
            merchantId: _merchantIdController.text.trim(),
            amountCents: amountCents,
          );
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Redeemed ${formatCents(redemption.amountCents)}. '
            'Remaining: ${formatCents(redemption.remainingAfterCents)}',
          ),
        ),
      );
      _voucherCodeController.clear();
      _amountController.clear();
      _refreshHistory();
    } on VoucherException catch (e) {
      setState(() => _error = e.message);
    } catch (e) {
      setState(() => _error = '$e');
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextFormField(
                controller: _merchantIdController,
                decoration: const InputDecoration(labelText: 'Merchant ID'),
                validator: (v) => (v == null || v.isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _voucherCodeController,
                decoration: const InputDecoration(labelText: 'Voucher code'),
                textCapitalization: TextCapitalization.characters,
                validator: (v) => (v == null || v.isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _amountController,
                decoration: const InputDecoration(
                  labelText: 'Amount to redeem',
                  prefixText: r'$',
                ),
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                validator: (v) {
                  if (v == null || v.isEmpty) return 'Required';
                  final parsed = double.tryParse(v);
                  if (parsed == null || parsed <= 0) return 'Enter an amount';
                  return null;
                },
              ),
              if (_error != null) ...[
                const SizedBox(height: 12),
                Text(
                  _error!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ],
              const SizedBox(height: 16),
              FilledButton(
                onPressed: _submitting ? null : _submit,
                child: Text(_submitting ? 'Redeeming…' : 'Redeem'),
              ),
            ],
          ),
        ),
        const SizedBox(height: 24),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'Recent redemptions',
              style: Theme.of(context).textTheme.titleMedium,
            ),
            IconButton(
              onPressed: _refreshHistory,
              icon: const Icon(Icons.refresh),
            ),
          ],
        ),
        FutureBuilder<List<VoucherRedemption>>(
          future: _historyFuture,
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: LoadingView(),
              );
            }
            if (snapshot.hasError) {
              return ErrorView(
                message: '${snapshot.error}',
                onRetry: _refreshHistory,
              );
            }
            final history = snapshot.data ?? const <VoucherRedemption>[];
            if (history.isEmpty) {
              return const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Text('No redemptions recorded yet.'),
              );
            }
            return Column(
              children: history
                  .map(
                    (r) => Card(
                      child: ListTile(
                        title: Text(
                          '${r.voucherCode} · ${formatCents(r.amountCents)}',
                        ),
                        subtitle: Text(r.redeemedAt.toLocal().toString()),
                      ),
                    ),
                  )
                  .toList(),
            );
          },
        ),
      ],
    );
  }
}
