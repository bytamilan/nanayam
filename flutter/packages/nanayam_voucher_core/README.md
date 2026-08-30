# nanayam_voucher_core

Domain layer for the Nanayam **voucher provisioning & usage** example — a
demo modeled after schemes like Singapore's CDC vouchers: a program
provisions a fixed-value voucher to a citizen/household, and businesses
redeem against it (in whole or in part) until the balance or expiry date is
exhausted. Every provision and redemption is recorded on the Nanayam sample
ledger.

```dart
import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';

final client = NanayamLedgerClient(baseUrl: 'http://localhost:8080');
await client.login(username: 'admin', password: 'admin');

final vouchers = VoucherLedgerRepository(client);

final voucher = await vouchers.provisionVoucher(
  code: 'CDC-7F3A9B',
  holderId: 'citizen-042',
  category: 'groceries',
  program: 'CDC Vouchers 2026',
  faceValueCents: 30000, // $300.00
  expiresAt: DateTime.utc(2026, 12, 31),
);

final redemption = await vouchers.redeemVoucher(
  code: voucher.code,
  merchantId: 'merchant-fairprice-01',
  amountCents: 5000, // $50.00
);

final current = await vouchers.getVoucher(voucher.code);
print(current.status);          // VoucherStatus.partiallyRedeemed
print(current.remainingCents);  // 25000
```

## Why a generic ledger, not a dedicated `Voucher` chaincode?

The main Nanayam repo has exactly one place a new chaincode would live
(`services/gateway` + a Fabric chaincode directory fetched at setup time,
which is not checked into this repo), and adding one is a Fabric-chaincode
task, not a Flutter one. Instead, `VoucherLedgerRepository` shows how a rich
domain (provisioning, categories, expiry, partial redemption, per-merchant
history) can be layered on the gateway's existing generic asset ledger
(`createAsset` / `queryAsset` / `listAssets`) — the same pattern you'd reach
for with any ledger whose write surface is narrower than your domain.

See the doc comment on `VoucherLedgerRepository` for the exact field
encoding and its known limitations (no cross-redemption atomicity, `O(n)`
listing) — both clearly acceptable for a sample app, both called out so
nobody mistakes this for production-ready.

## What's in this package

- `Voucher`, `VoucherRedemption`, `VoucherStatus` — plain domain models with
  derived `remainingCents` / `status` getters (nothing here is ever stored
  as a "current balance" field; it's always recomputed from the ledger).
- `VoucherLedgerRepository` — `provisionVoucher`, `getVoucher`,
  `listVouchersForHolder`, `redeemVoucher`, `listMerchantRedemptions`.
- Typed exceptions (`VoucherNotFoundException`,
  `VoucherAlreadyExistsException`, `VoucherExpiredException`,
  `InsufficientVoucherBalanceException`, `VoucherLedgerRejectedException`).

Pure Dart, no Flutter dependency — the `voucher_wallet` example app
(`flutter/apps/voucher_wallet`) is the UI built on top of it.
