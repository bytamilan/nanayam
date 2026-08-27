import 'package:flutter/foundation.dart';
import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';
import 'package:nanayam_voucher_core/nanayam_voucher_core.dart';

/// Builds the [NanayamLedgerClient] used for a given gateway URL. The
/// default just constructs a real client; tests pass one that injects a
/// fake `http.Client` instead, so [SessionController] never needs a live
/// gateway to be unit-testable.
typedef LedgerClientFactory = NanayamLedgerClient Function(String baseUrl);

NanayamLedgerClient _defaultClientFactory(String baseUrl) =>
    NanayamLedgerClient(baseUrl: baseUrl);

/// Holds the app's single [NanayamLedgerClient] connection and the
/// [VoucherLedgerRepository] built on top of it, and tracks whether the app
/// is currently signed in to a Nanayam gateway.
///
/// There is deliberately no dependency-injection framework here — this is a
/// small example app, so a single `ChangeNotifier` exposed via an
/// `InheritedNotifier` ([AppScope]) is enough. The one seam that matters for
/// testing — how the gateway HTTP client gets built — is exposed via
/// [LedgerClientFactory] instead.
class SessionController extends ChangeNotifier {
  SessionController({LedgerClientFactory? clientFactory})
      : _clientFactory = clientFactory ?? _defaultClientFactory;

  final LedgerClientFactory _clientFactory;

  NanayamLedgerClient? _client;
  VoucherLedgerRepository? _vouchers;
  LedgerUser? _currentUser;
  String? _lastError;
  bool _isBusy = false;

  NanayamLedgerClient get client {
    final client = _client;
    if (client == null) {
      throw StateError('Not connected — call connectAndLogin() first');
    }
    return client;
  }

  VoucherLedgerRepository get vouchers {
    final repo = _vouchers;
    if (repo == null) {
      throw StateError('Not connected — call connectAndLogin() first');
    }
    return repo;
  }

  LedgerUser? get currentUser => _currentUser;
  bool get isLoggedIn => _currentUser != null;
  bool get isBusy => _isBusy;
  String? get lastError => _lastError;

  /// Connects to the gateway at [baseUrl] and signs in with [username] /
  /// [password]. Throws [LedgerApiException] on failure (bad credentials,
  /// unreachable gateway); the caller's UI is expected to catch and display
  /// it.
  Future<void> connectAndLogin({
    required String baseUrl,
    required String username,
    required String password,
  }) async {
    _isBusy = true;
    _lastError = null;
    notifyListeners();

    final client = _clientFactory(baseUrl);
    try {
      await client.login(username: username, password: password);
      final user = await client.me();
      _client?.close();
      _client = client;
      _vouchers = VoucherLedgerRepository(client);
      _currentUser = user;
    } catch (e) {
      client.close();
      _lastError = e.toString();
      rethrow;
    } finally {
      _isBusy = false;
      notifyListeners();
    }
  }

  Future<void> logout() async {
    await _client?.logout();
    _client?.close();
    _client = null;
    _vouchers = null;
    _currentUser = null;
    notifyListeners();
  }

  @override
  void dispose() {
    _client?.close();
    super.dispose();
  }
}
