# voucher_wallet (example app)

An example Flutter app for the **Nanayam voucher provisioning & usage**
sample: a CDC-voucher-style demo where a program provisions vouchers to
citizens/households, businesses redeem against them, and every transaction
is recorded on the Nanayam sample ledger.

It exists to show the packages in `flutter/packages/` working together end
to end, not to be a polished product. Three tabs, all visible after any
sign-in (there's no separate citizen/merchant/admin login in this sample —
see the per-tab doc comments for why):

- **Wallet** — look up every voucher provisioned to a holder ID, see
  balance/status, drill into full redemption history.
- **Redeem** — a business redeems an amount against a voucher code, and
  sees its own recent redemptions.
- **Provision** — issue a new voucher to a holder ID with a category,
  program name, face value, and expiry date.

For a single numbered walkthrough covering the Fabric network, the gateway,
and this app together, see
[`docs/flutter-voucher-example.md`](../../../docs/flutter-voucher-example.md#quickstart).

## Running it

1. Start a Nanayam gateway against a running Fabric network (see the main
   repo's [`README`](../../../README.md) / `nanayam network up` /
   `nanayam gateway`), or point at wherever your gateway is already running.
2. This directory ships only `lib/`, `pubspec.yaml`, and `test/` —
   platform runner folders (`android/`, `ios/`, `web/`, ...) are
   intentionally **not** checked in, since they're best generated fresh by
   whatever Flutter SDK version you have installed rather than committed and
   left to bit-rot. From the repo root:

   ```bash
   cd flutter
   melos bootstrap        # or: flutter pub get, run per package
   cd apps/voucher_wallet
   flutter create --org com.nanayam --platforms=android,ios,web .
   flutter run
   ```

3. Sign in with the gateway URL (defaults to `http://localhost:8080`) and
   the username/password of a user seeded or registered on that gateway
   (the gateway seeds an `admin` / `admin` account on first run — see
   `services/gateway`'s `AuthStore.SeedAdmin`).
4. On the **Provision** tab, issue a voucher to any holder ID you like.
   Switch to **Wallet**, enter that same holder ID, and load it. Switch to
   **Redeem** and redeem part of its balance — then reload the wallet to see
   the updated balance and redemption history.

## Testing

No Fabric network or gateway needed — everything is exercised against an
in-memory fake gateway (`test/support/fake_gateway.dart`):

```bash
flutter test
```

`session_controller_test.dart` drives `SessionController` directly (login,
bad credentials, logout, listener notifications); `login_flow_test.dart`
drives the same paths through the actual `LoginScreen` widget tree, using
`VoucherWalletApp(session: ...)` to inject a `SessionController` built with
a fake `LedgerClientFactory` instead of a real gateway connection.

## How it's built

- `SessionController` (`lib/src/session_controller.dart`) owns the single
  `NanayamLedgerClient` connection and the `VoucherLedgerRepository` built
  on it, plus the signed-in user. It builds that client through an
  injectable `LedgerClientFactory` (real gateway by default) specifically so
  tests can swap in a fake `http.Client` without touching the rest of the
  app.
- `AppScope` (`lib/src/app_scope.dart`) is a small `InheritedNotifier` that
  makes `SessionController` available to every screen — there's no
  state-management package dependency for an app this size.
- Everything voucher-specific (models, ledger encoding, exceptions) lives in
  `nanayam_voucher_core`, not here; this app is purely the UI layer.
