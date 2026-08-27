import 'package:flutter/material.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';

import 'app_scope.dart';
import 'screens/login_screen.dart';
import 'session_controller.dart';

class VoucherWalletApp extends StatefulWidget {
  /// [session] lets a test (or a future entry point) supply a
  /// pre-configured [SessionController] — e.g. one built with a fake
  /// [LedgerClientFactory] — instead of the real one `main.dart` uses.
  const VoucherWalletApp({this.session, super.key});

  final SessionController? session;

  @override
  State<VoucherWalletApp> createState() => _VoucherWalletAppState();
}

class _VoucherWalletAppState extends State<VoucherWalletApp> {
  late final SessionController _session = widget.session ?? SessionController();

  @override
  void dispose() {
    _session.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AppScope(
      session: _session,
      child: MaterialApp(
        title: 'Nanayam Voucher Wallet',
        debugShowCheckedModeBanner: false,
        theme: NanayamTheme.light(),
        darkTheme: NanayamTheme.dark(),
        home: const LoginScreen(),
      ),
    );
  }
}
