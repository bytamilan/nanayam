import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:voucher_wallet/src/app.dart';

void main() {
  testWidgets('shows the login form on launch', (tester) async {
    await tester.pumpWidget(const VoucherWalletApp());

    expect(find.text('Nanayam Voucher Wallet'), findsOneWidget);
    expect(find.widgetWithText(TextFormField, 'Gateway URL'), findsOneWidget);
    expect(find.text('Sign in'), findsOneWidget);
  });
}
