import 'dart:math';

import 'package:flutter/material.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';

import '../app_scope.dart';

const List<String> _categories = ['groceries', 'hawker', 'general'];

/// Government/program-issuer view: provisions a new voucher to a citizen.
///
/// There is no separate "admin" login flow in this demo — any signed-in
/// user can provision, the same way any of them can browse a wallet or
/// redeem, because the point of the example is the ledger interaction, not
/// building out full role-based access control.
class ProvisionTab extends StatefulWidget {
  const ProvisionTab({super.key});

  @override
  State<ProvisionTab> createState() => _ProvisionTabState();
}

class _ProvisionTabState extends State<ProvisionTab> {
  final _formKey = GlobalKey<FormState>();
  final _codeController =
      TextEditingController(text: _generateCode());
  final _holderIdController = TextEditingController();
  final _programController =
      TextEditingController(text: 'CDC Vouchers 2026');
  final _valueController = TextEditingController(text: '300');
  String _category = _categories.first;
  DateTime _expiresAt = DateTime.now().add(const Duration(days: 365));
  bool _submitting = false;
  String? _error;
  Voucher? _lastProvisioned;

  static String _generateCode() {
    const chars = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
    final random = Random();
    final suffix = List.generate(
      6,
      (_) => chars[random.nextInt(chars.length)],
    ).join();
    return 'CDC-$suffix';
  }

  @override
  void dispose() {
    _codeController.dispose();
    _holderIdController.dispose();
    _programController.dispose();
    _valueController.dispose();
    super.dispose();
  }

  Future<void> _pickExpiry() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _expiresAt,
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365 * 3)),
    );
    if (picked != null) setState(() => _expiresAt = picked);
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _submitting = true;
      _error = null;
    });

    try {
      final faceValueCents =
          (double.parse(_valueController.text.trim()) * 100).round();
      final voucher = await AppScope.of(context).vouchers.provisionVoucher(
            code: _codeController.text.trim(),
            holderId: _holderIdController.text.trim(),
            category: _category,
            program: _programController.text.trim(),
            faceValueCents: faceValueCents,
            expiresAt: _expiresAt,
          );
      if (!mounted) return;
      setState(() {
        _lastProvisioned = voucher;
        _codeController.text = _generateCode();
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Provisioned ${voucher.code}')),
      );
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
                controller: _codeController,
                decoration: const InputDecoration(labelText: 'Voucher code'),
                textCapitalization: TextCapitalization.characters,
                validator: (v) => (v == null || v.isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _holderIdController,
                decoration: const InputDecoration(
                  labelText: 'Holder ID (citizen)',
                ),
                validator: (v) => (v == null || v.isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<String>(
                value: _category,
                decoration: const InputDecoration(labelText: 'Category'),
                items: _categories
                    .map((c) => DropdownMenuItem(value: c, child: Text(c)))
                    .toList(),
                onChanged: (value) {
                  if (value != null) setState(() => _category = value);
                },
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _programController,
                decoration: const InputDecoration(labelText: 'Program name'),
                validator: (v) => (v == null || v.isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _valueController,
                decoration: const InputDecoration(
                  labelText: 'Face value',
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
              const SizedBox(height: 12),
              ListTile(
                contentPadding: EdgeInsets.zero,
                title: const Text('Expires'),
                subtitle: Text(_expiresAt.toString().split(' ').first),
                trailing: const Icon(Icons.calendar_today),
                onTap: _pickExpiry,
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
                child: Text(_submitting ? 'Provisioning…' : 'Provision voucher'),
              ),
            ],
          ),
        ),
        if (_lastProvisioned != null) ...[
          const SizedBox(height: 24),
          Text(
            'Last provisioned',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          Card(
            child: ListTile(
              title: Text(_lastProvisioned!.code),
              subtitle: Text(
                '${_lastProvisioned!.holderId} · '
                '${formatCents(_lastProvisioned!.faceValueCents)}',
              ),
            ),
          ),
        ],
      ],
    );
  }
}
