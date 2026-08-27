# nanayam_ledger_client

A small, dependency-light Dart HTTP client for the Nanayam distribution
gateway (`services/gateway` in the main repo). Use it from any Flutter app,
CLI, or server that needs to:

- authenticate (`register` / `login` / `me` / `logout`)
- read/write the generic asset ledger (`createAsset` / `queryAsset` /
  `listAssets`)
- inspect the channel and ledger explorer (`fetchChannelInfo` /
  `fetchLedgerBlocks` / `fetchLedgerActivity`)

```dart
import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';

final client = NanayamLedgerClient(baseUrl: 'http://localhost:8080');

await client.login(username: 'admin', password: 'admin');

await client.createAsset(
  assetId: 'asset1',
  color: 'blue',
  size: 5,
  owner: 'Tomoko',
  appraisedValue: 300,
);

final assets = await client.listAssets();
```

## Design notes

- **Pure Dart.** No Flutter dependency, so it also works from Dart CLIs and
  server code, not just apps. A Flutter app supplies its own [`TokenStorage`]
  implementation (e.g. backed by `flutter_secure_storage`); the package ships
  an in-memory default for tests and short-lived scripts.
- **Domain-agnostic.** This client only knows the gateway's generic
  endpoints. It does not know about vouchers, complaints, or any other
  business object — build a domain repository on top of it instead of
  subclassing it. See [`nanayam_voucher_core`] for a worked example that maps
  a voucher-provisioning domain onto `createAsset` / `queryAsset` /
  `listAssets`.
- **Errors** surface as `LedgerApiException` (from `nanayam_ledger_models`),
  carrying the HTTP status code and the gateway's `{"error": "..."}` message
  when present.

[`TokenStorage`]: lib/src/token_storage.dart
[`nanayam_voucher_core`]: ../nanayam_voucher_core
