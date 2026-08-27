# nanayam_ui_kit

Shared Flutter presentation layer for Nanayam sample apps — not specific to
the voucher example, so any new Nanayam Flutter app can start from it:

- `NanayamTheme.light()` / `.dark()` — a consistent Material 3 theme.
- `formatCents(int)` — renders an integer amount of minor currency units
  (cents) as `$10.50`. All ledger amounts in these samples are stored as
  cents because the underlying chaincode fields are `int32` with no decimal
  type.
- `StatusBadge` — a colored pill for a domain status string; the widget
  doesn't know what the statuses mean, callers map their own enum to a color.
- `LoadingView`, `ErrorView`, `EmptyView` — the three states almost every
  list/detail screen needs, with a consistent look.

This package has no dependency on `nanayam_ledger_client` or
`nanayam_voucher_core` — it is pure presentation and can be reused even by
apps that don't talk to the ledger at all.
