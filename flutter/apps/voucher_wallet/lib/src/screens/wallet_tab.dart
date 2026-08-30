import 'package:flutter/material.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';

import '../app_scope.dart';
import '../widgets/voucher_card.dart';
import 'voucher_detail_screen.dart';

/// Citizen view: "which vouchers has this holder ID been provisioned?".
///
/// The holder ID defaults to the signed-in username, but is editable so a
/// single demo login can browse any citizen's wallet — there is no separate
/// "citizen" account type in this sample, only ledger holder IDs.
class WalletTab extends StatefulWidget {
  const WalletTab({super.key});

  @override
  State<WalletTab> createState() => _WalletTabState();
}

class _WalletTabState extends State<WalletTab> {
  final TextEditingController _holderIdController = TextEditingController();
  Future<List<Voucher>>? _future;
  bool _initialized = false;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_initialized) return;
    _initialized = true;
    _holderIdController.text = AppScope.of(context).currentUser?.username ?? '';
    _load();
  }

  @override
  void dispose() {
    _holderIdController.dispose();
    super.dispose();
  }

  void _load() {
    final holderId = _holderIdController.text.trim();
    if (holderId.isEmpty) return;
    setState(() {
      _future = AppScope.of(context).vouchers.listVouchersForHolder(holderId);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _holderIdController,
                  decoration: const InputDecoration(labelText: 'Holder ID'),
                  onSubmitted: (_) => _load(),
                ),
              ),
              const SizedBox(width: 8),
              FilledButton(onPressed: _load, child: const Text('Load')),
            ],
          ),
        ),
        Expanded(
          child: FutureBuilder<List<Voucher>>(
            future: _future,
            builder: (context, snapshot) {
              if (_future == null) {
                return const EmptyView(
                  message: 'Enter a holder ID and tap Load.',
                  icon: Icons.wallet_outlined,
                );
              }
              if (snapshot.connectionState == ConnectionState.waiting) {
                return const LoadingView(message: 'Loading vouchers…');
              }
              if (snapshot.hasError) {
                return ErrorView(
                  message: '${snapshot.error}',
                  onRetry: _load,
                );
              }
              final vouchers = snapshot.data ?? const <Voucher>[];
              if (vouchers.isEmpty) {
                return const EmptyView(
                  message: 'No vouchers for this holder yet.',
                  icon: Icons.wallet_outlined,
                );
              }
              return RefreshIndicator(
                onRefresh: () async => _load(),
                child: ListView.separated(
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                  itemCount: vouchers.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 8),
                  itemBuilder: (context, index) {
                    final voucher = vouchers[index];
                    return VoucherCard(
                      voucher: voucher,
                      onTap: () => Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) =>
                              VoucherDetailScreen(voucherCode: voucher.code),
                        ),
                      ),
                    );
                  },
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}
