import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';
import 'package:voucher_wallet/src/app.dart';
import 'package:voucher_wallet/src/session_controller.dart';

import 'support/fake_gateway.dart';

void main() {
  testWidgets('empty required fields show validation errors, not a login attempt',
      (tester) async {
    await tester.pumpWidget(const VoucherWalletApp());

    await tester.enterText(
      find.widgetWithText(TextFormField, 'Gateway URL'),
      '',
    );
    await tester.tap(find.text('Sign in'));
    await tester.pump();

    expect(find.text('Required'), findsWidgets);
  });

  testWidgets('valid credentials sign in and land on the wallet home screen',
      (tester) async {
    final gateway = FakeGateway();
    final session = SessionController(
      clientFactory: (baseUrl) => NanayamLedgerClient(
        baseUrl: baseUrl,
        httpClient: gateway.asClient(),
      ),
    );

    await tester.pumpWidget(VoucherWalletApp(session: session));

    // The form is pre-filled with the demo admin/admin credentials.
    await tester.tap(find.text('Sign in'));
    await tester.pumpAndSettle();

    expect(find.text('Wallet'), findsWidgets);
    expect(find.byIcon(Icons.logout), findsOneWidget);
  });

  testWidgets('bad credentials show an inline error and stay on the login screen',
      (tester) async {
    final gateway = FakeGateway();
    final session = SessionController(
      clientFactory: (baseUrl) => NanayamLedgerClient(
        baseUrl: baseUrl,
        httpClient: gateway.asClient(),
      ),
    );

    await tester.pumpWidget(VoucherWalletApp(session: session));

    await tester.enterText(
      find.widgetWithText(TextFormField, 'Password'),
      'wrong-password',
    );
    await tester.tap(find.text('Sign in'));
    await tester.pumpAndSettle();

    expect(find.text('invalid username or password'), findsOneWidget);
    expect(find.text('Nanayam Voucher Wallet'), findsOneWidget);
  });
}
