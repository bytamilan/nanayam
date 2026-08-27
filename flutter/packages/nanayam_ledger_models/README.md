# nanayam_ledger_models

Plain-Dart data classes for the Nanayam distribution gateway's REST API
(`services/gateway` in the main repo): auth (`LedgerUser`, `LedgerSession`,
`GatewayConfig`), generic ledger assets (`LedgerAsset`, `CreateAssetResult`),
ledger-explorer responses (`ChannelInfo`, `LedgerBlockSummary`,
`LedgerActivity`), and `LedgerApiException`.

No Flutter dependency, no code generation — hand-written `fromJson`/`toJson`
so it works in any Dart project (CLI tools, servers, Flutter apps) with a
single `dart pub get`.

This package has no I/O of its own; pair it with [`nanayam_ledger_client`]
for an HTTP client that returns these types.

[`nanayam_ledger_client`]: ../nanayam_ledger_client
