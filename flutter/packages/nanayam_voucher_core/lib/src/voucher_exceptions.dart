/// Base type for all voucher-domain failures raised by
/// `VoucherLedgerRepository`. Catch this to handle any voucher error
/// generically, or catch a specific subtype to handle one case.
sealed class VoucherException implements Exception {
  const VoucherException(this.message);

  final String message;

  @override
  String toString() => message;
}

/// No voucher exists on the ledger with the given code.
class VoucherNotFoundException extends VoucherException {
  const VoucherNotFoundException(String code)
      : super('No voucher found with code "$code"');
}

/// `provisionVoucher` was called with a code that is already on the ledger.
class VoucherAlreadyExistsException extends VoucherException {
  const VoucherAlreadyExistsException(String code)
      : super('A voucher with code "$code" already exists');
}

/// A redemption was attempted after the voucher's `expiresAt` date.
class VoucherExpiredException extends VoucherException {
  const VoucherExpiredException(String code, DateTime expiredAt)
      : super('Voucher "$code" expired on $expiredAt');
}

/// A redemption amount exceeds the voucher's remaining balance.
class InsufficientVoucherBalanceException extends VoucherException {
  const InsufficientVoucherBalanceException({
    required this.code,
    required this.requestedCents,
    required this.remainingCents,
  }) : super(
          'Voucher "$code" has $remainingCents cents remaining; '
          'cannot redeem $requestedCents cents',
        );

  final String code;
  final int requestedCents;
  final int remainingCents;
}

/// The gateway rejected a `createAsset` call for a reason other than one of
/// the specific cases above (e.g. a chaincode-level validation error).
/// Carries the raw error string returned by the ledger.
class VoucherLedgerRejectedException extends VoucherException {
  const VoucherLedgerRejectedException(String gatewayError)
      : super('Ledger rejected the request: $gatewayError');
}
