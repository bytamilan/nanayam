import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';

void main() {
  testWidgets('LoadingView shows a spinner and optional message', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(body: LoadingView(message: 'Loading vouchers…')),
      ),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    expect(find.text('Loading vouchers…'), findsOneWidget);
  });

  testWidgets('LoadingView with no message shows just the spinner', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: Scaffold(body: LoadingView())),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    expect(find.byType(Text), findsNothing);
  });

  testWidgets('ErrorView shows the message and invokes onRetry', (tester) async {
    var retried = false;

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ErrorView(
            message: 'Something went wrong',
            onRetry: () => retried = true,
          ),
        ),
      ),
    );

    expect(find.text('Something went wrong'), findsOneWidget);
    expect(find.text('Retry'), findsOneWidget);

    await tester.tap(find.text('Retry'));
    await tester.pump();

    expect(retried, isTrue);
  });

  testWidgets('ErrorView with no onRetry does not show a Retry button', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(body: ErrorView(message: 'Oops')),
      ),
    );

    expect(find.text('Retry'), findsNothing);
  });

  testWidgets('EmptyView shows its message', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(body: EmptyView(message: 'No vouchers yet.')),
      ),
    );

    expect(find.text('No vouchers yet.'), findsOneWidget);
  });
}
