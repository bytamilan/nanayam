# Flutter Voucher Example

A worked example of building a Flutter client against the Nanayam
distribution gateway: **voucher provisioning & usage**, modeled loosely on
schemes like Singapore's CDC (Community Development Council) vouchers — a
program provisions a fixed-value voucher to a citizen/household, and
businesses redeem against it (in whole or in part) until the balance or an
expiry date is exhausted, with every step recorded on the ledger.

The code lives in [`flutter/`](../flutter) at the repo root, as a
[Melos](https://melos.invertase.dev/)-managed monorepo:

```
flutter/
├── packages/
│   ├── nanayam_ledger_models/    # shared DTOs for the gateway REST API
│   ├── nanayam_ledger_client/    # HTTP client for the gateway (pure Dart)
│   ├── nanayam_ui_kit/           # shared Flutter theme + widgets
│   └── nanayam_voucher_core/     # voucher domain model + ledger mapping
└── apps/
    └── voucher_wallet/           # the example app
```

See [`flutter/README.md`](../flutter/README.md) for setup instructions and
[`flutter/apps/voucher_wallet/README.md`](../flutter/apps/voucher_wallet/README.md)
for how to run the app end to end. This document is the design write-up:
*why* it's built this way, and what corners were knowingly cut because this
is a sample, not a production system.

## Quickstart

Everything below assumes [Docker](https://docs.docker.com/get-docker/) and
the [Flutter SDK](https://docs.flutter.dev/get-started/install) are already
installed. Six steps, run from the repo root:

1. **Install the Nanayam CLI** (skip if you already have it):

   ```bash
   curl -fsSL https://raw.githubusercontent.com/bytamilan/nanayam/main/install.sh | bash
   ```

2. **Bring up the Fabric network and deploy the sample chaincode:**

   ```bash
   nanayam prerequisites --auto
   nanayam network up
   nanayam channel create --name mychannel --profile TwoOrgsChannel
   nanayam chaincode package --path ./chaincode/asset-transfer-basic --name basic
   nanayam chaincode install --package basic.tar.gz
   nanayam chaincode approve --name basic --channel mychannel --package-id <id from install>
   nanayam chaincode commit --name basic --channel mychannel
   ```

   (Or use the classic scripts — `./scripts/setup-fabric.sh && ./scripts/start-fabric.sh
   && ./scripts/deploy-chaincode.sh` — see the root [`README`](../README.md)
   for both paths in full.)

3. **Start the gateway:**

   ```bash
   nanayam gateway
   ```

   It listens on `:8080` and seeds a default `admin` / `admin` account on
   first run (`AuthStore.SeedAdmin` in `services/gateway/auth.go`) — that's
   what you'll sign into the app with below. Confirm it's up:

   ```bash
   curl http://localhost:8080/health   # {"status":"ok"}
   ```

4. **Bootstrap the Flutter workspace:**

   ```bash
   cd flutter
   dart pub global activate melos   # once per machine
   melos bootstrap
   ```

5. **Generate the app's platform folders and run it** (not checked in —
   see [`flutter/apps/voucher_wallet/README.md`](../flutter/apps/voucher_wallet/README.md#running-it)
   for why):

   ```bash
   cd apps/voucher_wallet
   flutter create --org com.nanayam --platforms=android,ios,web .
   flutter run -d chrome   # or an emulator/simulator/device of your choice
   ```

6. **Sign in** with `admin` / `admin` against `http://localhost:8080`
   (pre-filled). On the **Provision** tab, issue a voucher to any holder ID;
   switch to **Wallet** and load that holder ID to see it; switch to
   **Redeem** and redeem part of its balance; reload the wallet to see the
   updated balance and history.

Running the packages' own test suites doesn't need any of the above — no
Fabric network, no gateway, no device/emulator:

```bash
cd flutter
melos run test
```

## Packages before app

Three of the four packages are intentionally not voucher-specific, so any
future Nanayam Flutter app — an operator console companion, a
complaint-tracking app for citizens, anything else — can start from them:

- **`nanayam_ledger_models`** mirrors the gateway's REST types
  (`services/gateway/http.go`, `handler.go`): auth (`LedgerUser`,
  `LedgerSession`, `GatewayConfig`), the generic asset ledger (`LedgerAsset`,
  `CreateAssetResult`), and the ledger explorer (`ChannelInfo`,
  `LedgerBlockSummary`, `LedgerActivity`). Hand-written `fromJson`/`toJson`,
  no code generation, no Flutter dependency.
- **`nanayam_ledger_client`** wraps the gateway's `/v1/*` endpoints —
  `register` / `login` / `me` / `logout`, `createAsset` / `queryAsset` /
  `listAssets`, `fetchChannelInfo` / `fetchLedgerBlocks` /
  `fetchLedgerActivity`. Pure Dart (works from a CLI or server, not just
  Flutter); a `TokenStorage` interface lets a real app plug in secure
  storage instead of the in-memory default.
- **`nanayam_ui_kit`** is a Material 3 theme plus the handful of widgets
  almost every screen needs: `formatCents`, `StatusBadge`, `LoadingView`,
  `ErrorView`, `EmptyView`.
- **`nanayam_voucher_core`** is the one voucher-specific package: domain
  models (`Voucher`, `VoucherRedemption`, `VoucherStatus`) and
  `VoucherLedgerRepository`, which is where the interesting design
  decisions live.

The `voucher_wallet` app is a thin UI layer on top of all four — see its
own README for the screen-by-screen tour.

## Why a generic ledger, not a dedicated `Voucher` chaincode

The main Nanayam repo has exactly one sample chaincode
(`asset-transfer-basic`, fetched at setup time by `scripts/setup-fabric.sh`
— its source is not checked into this repo) and one bespoke one (the
anti-corruption complaint workflow, wired through its own gRPC methods in
`services/gateway/handler.go` and REST routes in `http.go`). Giving vouchers
the same treatment as complaints — a new chaincode plus new
`ProvisionVoucher` / `RedeemVoucher` gRPC methods — would mean designing and
deploying new Fabric chaincode, which is a Hyperledger Fabric task, not a
Flutter one, and there's nowhere in this repo to put that chaincode's
source today.

Instead, `VoucherLedgerRepository` shows the pattern you reach for whenever
a ledger's write surface is narrower than your domain: encode the domain
onto what's already there. Every voucher event becomes one call to the
existing `CreateAsset` RPC; nothing about vouchers required touching
`services/gateway` or Fabric configuration at all.

## Ledger encoding

The gateway exposes exactly three ledger operations to a Flutter client:
`createAsset`, `queryAsset`, `listAssets`. There is no `updateAsset`, so
`VoucherLedgerRepository` never mutates a ledger entry — it only appends new
ones. This turns out to fit an immutable ledger better than in-place
updates would: it's an event-sourcing model, not a limitation being worked
around.

| Event | `assetId` | `owner` | `color` (JSON) | `size` | `appraisedValue` |
|---|---|---|---|---|---|
| Provision | `VCH-<code>` | holder ID | `{"kind":"issue","category":...,"program":...,"issuedAt":...,"expiresAt":...}` | face value, in cents | face value, in cents |
| Redemption | `TXN-<code>-<microsecond timestamp>-<random suffix>` | merchant ID | `{"kind":"redeem","voucherCode":<code>}` | amount redeemed, in cents | balance remaining after this redemption, in cents |

A voucher's balance and status are **never** stored directly — they're
always recomputed by folding the issuance record with every redemption
record that references it (`Voucher.remainingCents`, `Voucher.status` in
`nanayam_voucher_core/lib/src/voucher_models.dart`). The `color` field
(a free-text string in `asset-transfer-basic`'s schema) carries a small JSON
object rather than a single value, because it's the only field in the fixed
`{color, size, owner, appraisedValue}` schema flexible enough to hold
metadata. Amounts are integer cents because the chaincode's numeric fields
are `int32` with no decimal type.

Every asset already on a shared ledger — including `asset-transfer-basic`'s
own seeded sample data (`asset1`, `blue`, `5`, ...) — is safe to coexist
with: the repository tries to `jsonDecode` each asset's `color` field and
silently ignores anything that isn't the expected shape, rather than
assuming every asset on the channel is a voucher.

## Known limitations (this is a demo)

- **No cross-redemption atomicity.** `redeemVoucher` reads the current
  balance, checks it client-side, then writes a new redemption record. Two
  redemptions submitted concurrently against the same voucher can both read
  the same balance and both succeed, over-redeeming it. A production system
  would enforce this the way `UpdateComplaint` enforces the complaint
  workflow's state machine today: inside chaincode, where the read and the
  write are one atomic transaction.
- **`O(n)` discovery.** "Every voucher for holder X" and "every redemption
  by merchant Y" both list every asset on the ledger and inspect each one,
  because the gateway has no query-by-owner index. Fine for a demo with a
  handful of vouchers; a real deployment would add a CouchDB rich query or a
  dedicated index — the same tradeoff `ListComplaints` already accepts for
  the complaint workflow.
- **No role enforcement.** Any signed-in user can provision, redeem, or
  browse any wallet in the example app — there's no notion of "citizen" vs.
  "merchant" vs. "issuing officer" accounts, only ledger-level holder/
  merchant IDs typed into a form. A real deployment would gate provisioning
  and redemption behind the gateway's existing auth roles/orgs.

None of these are bugs to fix in the sample — they're the corners a
production voucher system would need to close, called out so the example
doesn't get mistaken for one.
