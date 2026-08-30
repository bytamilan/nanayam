import 'package:flutter_test/flutter_test.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';

void main() {
  test('formatCents renders dollars and cents', () {
    expect(formatCents(1050), r'$10.50');
    expect(formatCents(5), r'$0.05');
    expect(formatCents(0), r'$0.00');
  });

  test('formatCents handles negative amounts', () {
    expect(formatCents(-250), r'-$2.50');
  });
}
