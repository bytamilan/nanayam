import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';
import 'package:test/test.dart';

Voucher _voucher({
  int faceValueCents = 10000,
  DateTime? expiresAt,
  List<VoucherRedemption> redemptions = const [],
}) {
  return Voucher(
    code: 'CDC-TEST',
    holderId: 'citizen-1',
    category: 'groceries',
    program: 'CDC Vouchers 2026',
    faceValueCents: faceValueCents,
    issuedAt: DateTime.now().toUtc(),
    expiresAt: expiresAt ?? DateTime.now().toUtc().add(const Duration(days: 1)),
    redemptions: redemptions,
  );
}

VoucherRedemption _redemption(int amountCents, int remainingAfterCents) {
  return VoucherRedemption(
    transactionId: 'TXN-CDC-TEST-1-a',
    voucherCode: 'CDC-TEST',
    merchantId: 'merchant-1',
    amountCents: amountCents,
    remainingAfterCents: remainingAfterCents,
    redeemedAt: DateTime.now().toUtc(),
  );
}

void main() {
  group('Voucher derived state', () {
    test('is "issued" with no redemptions and not expired', () {
      final voucher = _voucher();
      expect(voucher.status, VoucherStatus.issued);
      expect(voucher.remainingCents, 10000);
      expect(voucher.redeemedCents, 0);
      expect(voucher.isExpired, isFalse);
    });

    test('is "partiallyRedeemed" once some balance is spent', () {
      final voucher = _voucher(redemptions: [_redemption(2000, 8000)]);
      expect(voucher.status, VoucherStatus.partiallyRedeemed);
      expect(voucher.remainingCents, 8000);
      expect(voucher.redeemedCents, 2000);
    });

    test('is "fullyRedeemed" once the full face value is spent', () {
      final voucher = _voucher(redemptions: [_redemption(10000, 0)]);
      expect(voucher.status, VoucherStatus.fullyRedeemed);
      expect(voucher.remainingCents, 0);
    });

    test('sums multiple redemptions', () {
      final voucher = _voucher(
        redemptions: [_redemption(3000, 7000), _redemption(4000, 3000)],
      );
      expect(voucher.redeemedCents, 7000);
      expect(voucher.remainingCents, 3000);
    });

    test('is "expired" once past expiresAt with balance still remaining', () {
      final voucher = _voucher(
        expiresAt: DateTime.now().toUtc().subtract(const Duration(days: 1)),
      );
      expect(voucher.isExpired, isTrue);
      expect(voucher.status, VoucherStatus.expired);
    });

    test('a fully redeemed voucher is not reported as expired', () {
      final voucher = _voucher(
        expiresAt: DateTime.now().toUtc().subtract(const Duration(days: 1)),
        redemptions: [_redemption(10000, 0)],
      );
      // Fully spent takes priority over the expiry check: there's no
      // remaining balance left for expiry to matter for.
      expect(voucher.status, VoucherStatus.fullyRedeemed);
    });
  });

  group('Voucher.copyWith', () {
    test('replaces only the redemptions list, keeping everything else', () {
      final original = _voucher();
      final updated = original.copyWith(redemptions: [_redemption(1000, 9000)]);

      expect(updated.code, original.code);
      expect(updated.holderId, original.holderId);
      expect(updated.faceValueCents, original.faceValueCents);
      expect(updated.redemptions, hasLength(1));
      expect(original.redemptions, isEmpty);
    });
  });
}
