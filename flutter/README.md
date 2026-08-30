# Nanayam Flutter Workspace

A [Melos](https://melos.invertase.dev/)-managed monorepo of Flutter/Dart
packages for building clients against the Nanayam distribution gateway
(`services/gateway`), plus one example app: a **voucher provisioning &
usage** demo modeled after schemes like Singapore's CDC vouchers.

```
flutter/
├── melos.yaml
├── pubspec.yaml                  # workspace root (melos dev-dependency only)
├── packages/
│   ├── nanayam_ledger_models/    # shared DTOs for the gateway REST API
│   ├── nanayam_ledger_client/    # HTTP client for the gateway (pure Dart)
│   ├── nanayam_ui_kit/           # shared Flutter theme + widgets
│   └── nanayam_voucher_core/     # voucher domain model + ledger mapping
└── apps/
    └── voucher_wallet/           # example app built on the packages above
```

## Why packages first

The ask behind this workspace was "make something reusable, not just one
app". Three of the four packages have nothing to do with vouchers at all:

| Package | Reusable for | Depends on |
|---|---|---|
| `nanayam_ledger_models` | Any Dart/Flutter code that talks to the Nanayam gateway | nothing but `meta` |
| `nanayam_ledger_client` | Any Dart/Flutter app or CLI needing gateway auth + the generic asset ledger | `nanayam_ledger_models` |
| `nanayam_ui_kit` | Any Nanayam Flutter app wanting a consistent theme/loading/error/status UI | Flutter only |
| `nanayam_voucher_core` | The voucher example specifically | `nanayam_ledger_client` |

A new Nanayam Flutter app — a complaint-tracking companion app, an
operator console mobile client, anything else — starts from the first
three and only needs its own domain package (like `nanayam_voucher_core`)
for whatever it's adding.

## Getting started

Install melos globally, then bootstrap the workspace from `flutter/`:

```bash
dart pub global activate melos
cd flutter
melos bootstrap    # fetches dependencies for every package and app
melos run analyze  # static analysis across every package
melos run format   # format check across every package
melos run test      # run tests in every package that has one
```

Each package can also be developed independently — `cd` into it and run
plain `flutter pub get` / `dart pub get` — melos is a convenience for
working across all of them at once, not a requirement.

## Testing

Every package ships its own test suite, and none of them need a running
Fabric network or gateway:

- `nanayam_ledger_models` / `nanayam_ledger_client` test against
  hand-rolled fake gateways (`test/support/fake_gateway.dart` in the client
  package) — a `package:http` `MockClient` standing in for
  `services/gateway`, so the HTTP contract is exercised without a live
  server.
- `nanayam_voucher_core` does the same one layer up, against a
  `FakeLedger` that only understands the generic asset endpoints —
  covering provisioning, redemption, balance/status derivation, and every
  documented failure mode (duplicate code, expired voucher, over-redemption,
  a rejected ledger write).
- `nanayam_ui_kit` has plain widget tests for its theme and shared states.
- `voucher_wallet` injects a fake gateway through `SessionController`'s
  `LedgerClientFactory` seam, so both the controller and the actual login
  screen get driven through real login/failure/logout flows in tests
  without touching the network.

New code in any package should follow the same shape: write the test
against the nearest fake first, then the implementation. `melos run test`
runs everything at once (see [Getting started](#getting-started) above).

## The voucher example

See [`apps/voucher_wallet/README.md`](apps/voucher_wallet/README.md) for
how to run the example app, and the main repo's
[`docs/flutter-voucher-example.md`](../docs/flutter-voucher-example.md#quickstart)
for a step-by-step guide covering the Fabric network, the gateway, and the
app together, plus the full design write-up: why a generic asset ledger
stands in for a dedicated chaincode, the exact field encoding used, and its
known limitations as a demo.
