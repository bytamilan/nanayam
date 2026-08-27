import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

/// A tiny in-memory stand-in for the Nanayam gateway's generic asset ledger
/// (`CreateAsset` / `QueryAsset` / `ListAssets`), just enough of the real
/// HTTP contract to exercise `VoucherLedgerRepository` without a running
/// Fabric network. Auth is not enforced — these tests only cover the
/// voucher-domain mapping, not the gateway's auth middleware (see
/// `nanayam_ledger_client`'s own tests for that).
class FakeLedger {
  final Map<String, Map<String, dynamic>> _assets = <String, Map<String, dynamic>>{};

  http.Client asClient() => MockClient(_handle);

  Future<http.Response> _handle(http.Request request) async {
    switch ('${request.method} ${request.url.path}') {
      case 'POST /v1/CreateAsset':
        return _createAsset(request);
      case 'GET /v1/QueryAsset':
        return _queryAsset(request);
      case 'GET /v1/ListAssets':
        return _listAssets();
      default:
        return http.Response(jsonEncode({'error': 'not found'}), 404);
    }
  }

  http.Response _createAsset(http.Request request) {
    final body = jsonDecode(request.body) as Map<String, dynamic>;
    final assetId = body['assetId'] as String;
    if (_assets.containsKey(assetId)) {
      return http.Response(
        jsonEncode({
          'success': false,
          'error': 'the asset $assetId already exists',
        }),
        200,
      );
    }
    _assets[assetId] = {
      'ID': assetId,
      'Color': body['color'],
      'Size': body['size'],
      'Owner': body['owner'],
      'AppraisedValue': body['appraisedValue'],
    };
    return http.Response(jsonEncode({'success': true}), 200);
  }

  http.Response _queryAsset(http.Request request) {
    final assetId = request.url.queryParameters['assetId'];
    final asset = _assets[assetId];
    if (asset == null) {
      return http.Response(
        jsonEncode({'error': 'the asset $assetId does not exist'}),
        404,
      );
    }
    return http.Response(jsonEncode({'data': jsonEncode(asset)}), 200);
  }

  http.Response _listAssets() {
    return http.Response(
      jsonEncode({'assetIds': _assets.keys.toList()}),
      200,
    );
  }
}
