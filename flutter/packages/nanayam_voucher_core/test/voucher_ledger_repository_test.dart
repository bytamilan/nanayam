import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';
import 'package:test/test.dart';

import 'support/fake_ledger.dart';

void main() {
  late FakeLedger ledger;
  late NanayamLedgerClient client;
  late VoucherLedgerRepository repo;

  setUp(() {
    ledger = FakeLedger();
    client = NanayamLedgerClient(
      baseUrl: 'http://localhost:8080',
      httpClient: ledger.asClient(),
    );
    repo = VoucherLedgerRepository(client);
  });

  DateTime future() => DateTime.now().toUtc().add(const Duration(days: 30));
  DateTime past() => DateTime.now().toUtc().subtract(const Duration(days: 1));

  group('provisionVoucher', () {
    test('issues a voucher that can be read back', () async {
      final voucher = await repo.provisionVoucher(
        code: 'CDC-001',
        holderId: 'citizen-1',
        category: 'groceries',
        program: 'CDC Vouchers 2026',
        faceValueCents: 30000,
        expiresAt: future(),
      );

      expect(voucher.status, VoucherStatus.issued);
      expect(voucher.remainingCents, 30000);

      final reloaded = await repo.getVoucher('CDC-001');
      expect(reloaded.holderId, 'citizen-1');
      expect(reloaded.category, 'groceries');
      expect(reloaded.faceValueCents, 30000);
      expect(reloaded.remainingCents, 30000);
      expect(reloaded.redemptions, isEmpty);
    });

    test('rejects a duplicate code', () async {
      await repo.provisionVoucher(
        code: 'CDC-DUP',
        holderId: 'citizen-1',
        category: 'groceries',
        program: 'CDC Vouchers 2026',
        faceValueCents: 10000,
        expiresAt: future(),
      );

      expect(
        () => repo.provisionVoucher(
          code: 'CDC-DUP',
          holderId: 'citizen-2',
          category: 'groceries',
          program: 'CDC Vouchers 2026',
          faceValueCents: 10000,
          expiresAt: future(),
        ),
        throwsA(isA<VoucherAlreadyExistsException>()),
      );
    });
  });

  group('getVoucher', () {
    test('throws VoucherNotFoundException for an unknown code', () {
      expect(
        () => repo.getVoucher('NOPE'),
        throwsA(isA<VoucherNotFoundException>()),
      );
    });
  });

  group('redeemVoucher', () {
    test('partial redemption updates remaining balance and status',
        () async {
      await repo.provisionVoucher(
        code: 'CDC-002',
        holderId: 'citizen-1',
        category: 'groceries',
        program: 'CDC Vouchers 2026',
        faceValueCents: 30000,
        expiresAt: future(),
      );

      final redemption = await repo.redeemVoucher(
        code: 'CDC-002',
        merchantId: 'merchant-1',
        amountCents: 5000,
      );
      expect(redemption.remainingAfterCents, 25000);

      final voucher = await repo.getVoucher('CDC-002');
      expect(voucher.remainingCents, 25000);
      expect(voucher.redeemedCents, 5000);
      expect(voucher.status, VoucherStatus.partiallyRedeemed);
      expect(voucher.redemptions, hasLength(1));
    });

    test('redeeming the full balance marks the voucher fully redeemed',
        () async {
      await repo.provisionVoucher(
        code: 'CDC-003',
        holderId: 'citizen-1',
        category: 'groceries',
        program: 'CDC Vouchers 2026',
        faceValueCents: 10000,
        expiresAt: future(),
      );

      await repo.redeemVoucher(
        code: 'CDC-003',
        merchantId: 'merchant-1',
        amountCents: 10000,
      );

      final voucher = await repo.getVoucher('CDC-003');
      expect(voucher.remainingCents, 0);
      expect(voucher.status, VoucherStatus.fullyRedeemed);
    });

    test('rejects redemption above the remaining balance', () async {
      await repo.provisionVoucher(
        code: 'CDC-004',
        holderId: 'citizen-1',
        category: 'groceries',
        program: 'CDC Vouchers 2026',
        faceValueCents: 10000,
        expiresAt: future(),
      );

      expect(
        () => repo.redeemVoucher(
          code: 'CDC-004',
          merchantId: 'merchant-1',
          amountCents: 20000,
        ),
        throwsA(isA<InsufficientVoucherBalanceException>()),
      );
    });

    test('rejects redemption of an expired voucher', () async {
      await repo.provisionVoucher(
        code: 'CDC-005',
        holderId: 'citizen-1',
        category: 'groceries',
        program: 'CDC Vouchers 2026',
        faceValueCents: 10000,
        expiresAt: past(),
      );

      expect(
        () => repo.redeemVoucher(
          code: 'CDC-005',
          merchantId: 'merchant-1',
          amountCents: 1000,
        ),
        throwsA(isA<VoucherExpiredException>()),
      );
    });
  });

  group('listVouchersForHolder', () {
    test('finds only this holder\'s vouchers and ignores foreign assets',
        () async {
      await repo.provisionVoucher(
        code: 'CDC-010',
        holderId: 'citizen-a',
        category: 'groceries',
        program: 'CDC Vouchers 2026',
        faceValueCents: 10000,
        expiresAt: future(),
      );
      await repo.provisionVoucher(
        code: 'CDC-011',
        holderId: 'citizen-b',
        category: 'hawker',
        program: 'CDC Vouchers 2026',
        faceValueCents: 20000,
        expiresAt: future(),
      );
      // Simulate a pre-seeded, unrelated asset-transfer-basic sample asset.
      await client.createAsset(
        assetId: 'asset1',
        color: 'blue',
        size: 5,
        owner: 'citizen-a',
        appraisedValue: 300,
      );

      final vouchers = await repo.listVouchersForHolder('citizen-a');
      expect(vouchers.map((v) => v.code), ['CDC-010']);
    });
  });

  group('listMerchantRedemptions', () {
    test('returns redemptions for the given merchant, most recent first',
        () async {
      await repo.provisionVoucher(
        code: 'CDC-020',
        holderId: 'citizen-a',
        category: 'groceries',
        program: 'CDC Vouchers 2026',
        faceValueCents: 30000,
        expiresAt: future(),
      );

      await repo.redeemVoucher(
        code: 'CDC-020',
        merchantId: 'merchant-x',
        amountCents: 1000,
      );
      await repo.redeemVoucher(
        code: 'CDC-020',
        merchantId: 'merchant-x',
        amountCents: 2000,
      );
      await repo.redeemVoucher(
        code: 'CDC-020',
        merchantId: 'merchant-y',
        amountCents: 500,
      );

      final xRedemptions = await repo.listMerchantRedemptions('merchant-x');
      expect(xRedemptions, hasLength(2));
      expect(xRedemptions.every((r) => r.merchantId == 'merchant-x'), isTrue);
    });
  });
}
