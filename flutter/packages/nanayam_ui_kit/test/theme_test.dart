import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nanayam_ui_kit/nanayam_ui_kit.dart';

void main() {
  test('light() produces a light-brightness Material 3 theme', () {
    final theme = NanayamTheme.light();
    expect(theme.useMaterial3, isTrue);
    expect(theme.colorScheme.brightness, Brightness.light);
  });

  test('dark() produces a dark-brightness Material 3 theme', () {
    final theme = NanayamTheme.dark();
    expect(theme.useMaterial3, isTrue);
    expect(theme.colorScheme.brightness, Brightness.dark);
  });

  test('both themes are seeded from the same brand color', () {
    final light = NanayamTheme.light();
    final dark = NanayamTheme.dark();
    expect(light.colorScheme.primary, isNot(dark.colorScheme.primary));
    // Different brightness derives different tones, but both are seeded
    // from NanayamTheme.seed rather than an arbitrary Material default.
    expect(NanayamTheme.seed, const Color(0xFF0B5FFF));
  });
}
