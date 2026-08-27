/// A structured error surfaced by the Nanayam gateway, e.g.
/// `{"error": "invalid username or password"}`.
class LedgerApiException implements Exception {
  const LedgerApiException(this.statusCode, this.message);

  final int statusCode;
  final String message;

  bool get isUnauthorized => statusCode == 401;
  bool get isForbidden => statusCode == 403;
  bool get isNotFound => statusCode == 404;

  @override
  String toString() => 'LedgerApiException($statusCode): $message';
}
