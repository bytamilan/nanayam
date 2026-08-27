/// Formats an integer amount of minor currency units (cents) as a display
/// string, e.g. `1050` -> `$10.50`.
///
/// Every ledger amount in these sample apps is stored as an integer number
/// of cents (matching the chaincode's `int32` fields, which have no native
/// decimal type) — this is the single place that turns that back into
/// something human-readable.
String formatCents(int cents, {String symbol = '\$'}) {
  final isNegative = cents < 0;
  final abs = cents.abs();
  final dollars = abs ~/ 100;
  final remainder = (abs % 100).toString().padLeft(2, '0');
  final sign = isNegative ? '-' : '';
  return '$sign$symbol$dollars.$remainder';
}
