import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';

void main() {
  testWidgets('StatusBadge renders its label', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: Scaffold(
          body: StatusBadge(label: 'Issued', color: Colors.blue),
        ),
      ),
    );

    expect(find.text('Issued'), findsOneWidget);
  });
}
