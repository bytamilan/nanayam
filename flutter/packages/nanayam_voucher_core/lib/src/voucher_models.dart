/// Lifecycle state of a [Voucher], derived from its face value, its
/// redemption history, and the current time — never stored directly on the
/// ledger.
enum VoucherStatus {
  /// No redemptions yet and not expired.
  issued,

  /// Some, but not all, of the face value has been redeemed.
  partiallyRedeemed,

  /// The full face value has been redeemed.
  fullyRedeemed,

  /// The expiry date has passed with balance still remaining.
  expired,
}

/// A single redemption of a [Voucher] by a business.
///
/// Mirrors one `TXN-<code>-<...>` entry on the ledger (see
/// `VoucherLedgerRepository` for the encoding).
class VoucherRedemption {
  const VoucherRedemption({
    required this.transactionId,
    required this.voucherCode,
    required this.merchantId,
    required this.amountCents,
    required this.remainingAfterCents,
    required this.redeemedAt,
  });

  /// The ledger asset ID this redemption is stored under.
  final String transactionId;

  final String voucherCode;

  /// The business identity that redeemed this amount.
  final String merchantId;

  final int amountCents;

  /// The voucher's remaining balance immediately after this redemption, as
  /// computed by the client that submitted it. Because the demo ledger has
  /// no atomic "decrement" operation (see the repository docs), this is a
  /// best-effort snapshot rather than a value guaranteed consistent under
  /// concurrent redemptions.
  final int remainingAfterCents;

  final DateTime redeemedAt;
}

/// A voucher provisioned to a citizen/household, redeemable in whole or in
/// part by one or more businesses before [expiresAt].
///
/// Modeled after schemes like Singapore's CDC vouchers: a program issues a
/// fixed-value voucher to a holder, and merchants redeem against it until
/// either the balance or the expiry date is exhausted.
class Voucher {
  const Voucher({
    required this.code,
    required this.holderId,
    required this.category,
    required this.program,
    required this.faceValueCents,
    required this.issuedAt,
    required this.expiresAt,
    this.redemptions = const <VoucherRedemption>[],
  });

  /// Human-facing voucher code, e.g. `CDC-7F3A9B`. Also the ledger key
  /// (as `VCH-<code>`).
  final String code;

  /// The citizen/household identity this voucher was provisioned to.
  final String holderId;

  /// Spending category, e.g. `groceries`, `hawker`, `general`.
  final String category;

  /// The scheme/program name, e.g. `CDC Vouchers 2026`.
  final String program;

  final int faceValueCents;
  final DateTime issuedAt;
  final DateTime expiresAt;

  /// Every redemption recorded against this voucher, oldest first.
  final List<VoucherRedemption> redemptions;

  int get redeemedCents =>
      redemptions.fold(0, (sum, r) => sum + r.amountCents);

  int get remainingCents => faceValueCents - redeemedCents;

  bool get isExpired => DateTime.now().toUtc().isAfter(expiresAt);

  VoucherStatus get status {
    if (isExpired && remainingCents > 0) return VoucherStatus.expired;
    if (remainingCents <= 0) return VoucherStatus.fullyRedeemed;
    if (redeemedCents > 0) return VoucherStatus.partiallyRedeemed;
    return VoucherStatus.issued;
  }

  Voucher copyWith({List<VoucherRedemption>? redemptions}) => Voucher(
        code: code,
        holderId: holderId,
        category: category,
        program: program,
        faceValueCents: faceValueCents,
        issuedAt: issuedAt,
        expiresAt: expiresAt,
        redemptions: redemptions ?? this.redemptions,
      );
}
