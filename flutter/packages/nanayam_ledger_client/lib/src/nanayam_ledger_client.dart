import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:nanayam_ledger_models/nanayam_ledger_models.dart';

import 'token_storage.dart';

/// HTTP client for the Nanayam distribution gateway (`services/gateway` in
/// the main repo).
///
/// Wraps the gateway's `/v1/*` REST surface: authentication, the generic
/// asset ledger (`CreateAsset` / `QueryAsset` / `ListAssets`), and the
/// read-only ledger explorer (`ChannelInfo` / `LedgerBlocks` /
/// `LedgerActivity`).
///
/// This client is intentionally domain-agnostic — it knows nothing about
/// vouchers, complaints, or any other business object. Build a
/// domain-specific repository on top of it (see `nanayam_voucher_core`'s
/// `VoucherLedgerRepository` for an example) rather than extending this
/// class per use case.
class NanayamLedgerClient {
  NanayamLedgerClient({
    required String baseUrl,
    http.Client? httpClient,
    TokenStorage? tokenStorage,
  })  : _baseUrl = baseUrl.endsWith('/')
            ? baseUrl.substring(0, baseUrl.length - 1)
            : baseUrl,
        _http = httpClient ?? http.Client(),
        tokenStorage = tokenStorage ?? InMemoryTokenStorage();

  final String _baseUrl;
  final http.Client _http;

  /// Where the bearer token issued by [login] is persisted. Swap in a
  /// secure-storage-backed implementation in a real app.
  final TokenStorage tokenStorage;

  Uri _uri(String path, [Map<String, String>? query]) =>
      Uri.parse('$_baseUrl$path').replace(queryParameters: query);

  Future<Map<String, String>> _authHeaders({bool json = false}) async {
    final headers = <String, String>{
      if (json) 'Content-Type': 'application/json',
    };
    final token = await tokenStorage.read();
    if (token != null && token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  Object? _decode(http.Response response) {
    if (response.statusCode >= 400) {
      var message = response.body;
      try {
        final decoded = jsonDecode(response.body);
        if (decoded is Map && decoded['error'] is String) {
          message = decoded['error'] as String;
        }
      } catch (_) {
        // Body wasn't JSON (e.g. plain-text 500 from http.Error); use as-is.
      }
      throw LedgerApiException(response.statusCode, message);
    }
    if (response.body.isEmpty) return null;
    return jsonDecode(response.body);
  }

  // ---------------------------------------------------------------------
  // Public gateway config / health
  // ---------------------------------------------------------------------

  /// `GET /v1/Config` — no auth required. Check `signupEnabled` before
  /// showing a "create account" affordance.
  Future<GatewayConfig> fetchConfig() async {
    final response = await _http.get(_uri('/v1/Config'));
    return GatewayConfig.fromJson(
      _decode(response)! as Map<String, dynamic>,
    );
  }

  /// `GET /health` — true if the gateway process is up and responding.
  Future<bool> health() async {
    final response = await _http.get(_uri('/health'));
    return response.statusCode == 200;
  }

  // ---------------------------------------------------------------------
  // Auth
  // ---------------------------------------------------------------------

  /// `POST /v1/Register`. Throws [LedgerApiException] with `statusCode 403`
  /// if the gateway was started with signup disabled.
  Future<LedgerUser> register({
    required String username,
    required String password,
    String org = '',
  }) async {
    final response = await _http.post(
      _uri('/v1/Register'),
      headers: await _authHeaders(json: true),
      body: jsonEncode(<String, String>{
        'username': username,
        'password': password,
        'org': org,
      }),
    );
    return LedgerUser.fromJson(_decode(response)! as Map<String, dynamic>);
  }

  /// `POST /v1/Login`. On success, persists the bearer token via
  /// [tokenStorage] so subsequent calls are authenticated automatically.
  Future<LedgerSession> login({
    required String username,
    required String password,
  }) async {
    final response = await _http.post(
      _uri('/v1/Login'),
      headers: await _authHeaders(json: true),
      body: jsonEncode(<String, String>{
        'username': username,
        'password': password,
      }),
    );
    final session =
        LedgerSession.fromJson(_decode(response)! as Map<String, dynamic>);
    await tokenStorage.write(session.token);
    return session;
  }

  /// Clears the locally stored token. The gateway's JWTs are stateless, so
  /// this is purely a client-side sign-out — there is no server call.
  Future<void> logout() => tokenStorage.clear();

  Future<bool> get isAuthenticated async =>
      (await tokenStorage.read())?.isNotEmpty ?? false;

  /// `GET /v1/Me` — the currently authenticated account.
  Future<LedgerUser> me() async {
    final response = await _http.get(
      _uri('/v1/Me'),
      headers: await _authHeaders(),
    );
    return LedgerUser.fromJson(_decode(response)! as Map<String, dynamic>);
  }

  // ---------------------------------------------------------------------
  // Ledger explorer
  // ---------------------------------------------------------------------

  /// `GET /v1/ChannelInfo`.
  Future<ChannelInfo> fetchChannelInfo() async {
    final response = await _http.get(
      _uri('/v1/ChannelInfo'),
      headers: await _authHeaders(),
    );
    return ChannelInfo.fromJson(_decode(response)! as Map<String, dynamic>);
  }

  /// `GET /v1/LedgerBlocks?start=&end=`.
  Future<List<LedgerBlockSummary>> fetchLedgerBlocks({
    int start = 0,
    int end = 10,
  }) async {
    final response = await _http.get(
      _uri('/v1/LedgerBlocks', <String, String>{
        'start': '$start',
        'end': '$end',
      }),
      headers: await _authHeaders(),
    );
    final body = _decode(response)! as Map<String, dynamic>;
    final blocks = (body['blocks'] as List<dynamic>? ?? const <dynamic>[])
        .map((Object? e) =>
            LedgerBlockSummary.fromJson(e! as Map<String, dynamic>))
        .toList();
    return blocks;
  }

  /// `GET /v1/LedgerActivity`.
  Future<LedgerActivity> fetchLedgerActivity() async {
    final response = await _http.get(
      _uri('/v1/LedgerActivity'),
      headers: await _authHeaders(),
    );
    return LedgerActivity.fromJson(_decode(response)! as Map<String, dynamic>);
  }

  // ---------------------------------------------------------------------
  // Generic asset ledger
  // ---------------------------------------------------------------------

  /// `POST /v1/CreateAsset` — submits a new asset to the ledger.
  ///
  /// The chaincode backing this call (`asset-transfer-basic`) has a fixed,
  /// generic schema (`color`, `size`, `owner`, `appraisedValue`); domain
  /// packages repurpose these fields to encode richer records — see
  /// `nanayam_voucher_core` for the voucher example's field mapping.
  Future<CreateAssetResult> createAsset({
    required String assetId,
    required String color,
    required int size,
    required String owner,
    required int appraisedValue,
  }) async {
    final response = await _http.post(
      _uri('/v1/CreateAsset'),
      headers: await _authHeaders(json: true),
      body: jsonEncode(<String, dynamic>{
        'assetId': assetId,
        'color': color,
        'size': size,
        'owner': owner,
        'appraisedValue': appraisedValue,
      }),
    );
    return CreateAssetResult.fromJson(
      _decode(response)! as Map<String, dynamic>,
    );
  }

  /// `GET /v1/QueryAsset?assetId=` — reads a single asset.
  ///
  /// The gateway wraps the chaincode's raw JSON in a `{"data": "..."}`
  /// envelope, so this decodes twice: once for the envelope, once for the
  /// asset itself.
  Future<LedgerAsset> queryAsset(String assetId) async {
    final response = await _http.get(
      _uri('/v1/QueryAsset', <String, String>{'assetId': assetId}),
      headers: await _authHeaders(),
    );
    final envelope = _decode(response)! as Map<String, dynamic>;
    final inner = jsonDecode(envelope['data'] as String? ?? '{}');
    return LedgerAsset.fromJson(inner as Map<String, dynamic>);
  }

  /// `GET /v1/ListAssets` — returns every asset ID currently on the ledger.
  Future<List<String>> listAssetIds() async {
    final response = await _http.get(
      _uri('/v1/ListAssets'),
      headers: await _authHeaders(),
    );
    final body = _decode(response)! as Map<String, dynamic>;
    return (body['assetIds'] as List<dynamic>? ?? const <dynamic>[])
        .map((Object? e) => e! as String)
        .toList();
  }

  /// Convenience helper: fetches every asset ID and then reads each asset in
  /// full. `O(n)` round trips — fine for a demo ledger with a handful of
  /// assets, not meant for production-scale listings.
  Future<List<LedgerAsset>> listAssets() async {
    final ids = await listAssetIds();
    final assets = <LedgerAsset>[];
    for (final id in ids) {
      try {
        assets.add(await queryAsset(id));
      } on LedgerApiException {
        // Skip assets that vanish between the list and read call, or that
        // fail to parse under the caller's expected schema.
      }
    }
    return assets;
  }

  /// Releases the underlying [http.Client]. Only call this if you passed in
  /// your own client via the constructor and own its lifecycle; if this
  /// instance created its own client, call this when the app is disposing
  /// of the [NanayamLedgerClient] itself.
  void close() => _http.close();
}
