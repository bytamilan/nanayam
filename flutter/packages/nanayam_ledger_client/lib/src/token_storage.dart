/// Persists the bearer token issued by `/v1/Login`.
///
/// This package is pure Dart and has no opinion on *where* the token lives,
/// so it defines this small interface rather than depending on Flutter or a
/// specific secure-storage plugin. A Flutter app should provide an
/// implementation backed by `flutter_secure_storage` (or similar); tests and
/// CLIs can use [InMemoryTokenStorage].
abstract class TokenStorage {
  Future<String?> read();
  Future<void> write(String token);
  Future<void> clear();
}

/// Default, non-persistent [TokenStorage] — the token only lives for the
/// process lifetime. Fine for tests and short-lived scripts; a real app
/// should supply its own persistent implementation.
class InMemoryTokenStorage implements TokenStorage {
  String? _token;

  @override
  Future<String?> read() async => _token;

  @override
  Future<void> write(String token) async => _token = token;

  @override
  Future<void> clear() async => _token = null;
}
