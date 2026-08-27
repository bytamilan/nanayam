import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';
import 'package:test/test.dart';

void main() {
  test('VoucherNotFoundException mentions the code', () {
    const err = VoucherNotFoundException('CDC-001');
    expect(err.message, contains('CDC-001'));
    expect(err.toString(), err.message);
  });

  test('VoucherAlreadyExistsException mentions the code', () {
    const err = VoucherAlreadyExistsException('CDC-001');
    expect(err.message, contains('already exists'));
    expect(err.message, contains('CDC-001'));
  });

  test('VoucherExpiredException mentions the code and expiry', () {
    final expiredAt = DateTime.utc(2024, 1, 1);
    final err = VoucherExpiredException('CDC-001', expiredAt);
    expect(err.message, contains('CDC-001'));
    expect(err.message, contains('2024'));
  });

  test('InsufficientVoucherBalanceException exposes the numbers', () {
    const err = InsufficientVoucherBalanceException(
      code: 'CDC-001',
      requestedCents: 5000,
      remainingCents: 2000,
    );
    expect(err.code, 'CDC-001');
    expect(err.requestedCents, 5000);
    expect(err.remainingCents, 2000);
    expect(err.message, contains('2000'));
    expect(err.message, contains('5000'));
  });

  test('VoucherLedgerRejectedException carries the gateway error', () {
    const err = VoucherLedgerRejectedException('chaincode says no');
    expect(err.message, contains('chaincode says no'));
  });
}
