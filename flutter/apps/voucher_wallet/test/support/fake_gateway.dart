import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

/// A minimal in-memory stand-in for the Nanayam gateway, just enough to
/// drive [SessionController] and the login screen through a real
/// login/me/logout cycle plus a couple of ledger calls, without a running
/// Fabric network. See `nanayam_ledger_client`'s and `nanayam_voucher_core`'s
/// own test suites for more exhaustive gateway-contract coverage — this one
/// only needs to be as complete as the app's own tests require.
class FakeGateway {
  final Map<String, String> _users = <String, String>{'admin': 'admin'};
  final Map<String, Map<String, dynamic>> _assets =
      <String, Map<String, dynamic>>{};

  http.Client asClient() => MockClient(_handle);

  Future<http.Response> _handle(http.Request request) async {
    switch ('${request.method} ${request.url.path}') {
      case 'POST /v1/Login':
        return _login(request);
      case 'GET /v1/Me':
        return _me(request);
      case 'POST /v1/CreateAsset':
        return _createAsset(request);
      case 'GET /v1/QueryAsset':
        return _queryAsset(request);
      case 'GET /v1/ListAssets':
        return _json({'assetIds': _assets.keys.toList()});
      default:
        return _error(404, 'not found');
    }
  }

  http.Response _login(http.Request request) {
    final body = jsonDecode(request.body) as Map<String, dynamic>;
    final username = body['username'] as String;
    final password = body['password'] as String;
    if (_users[username] != password) {
      return _error(401, 'invalid username or password');
    }
    return _json({'token': 'test-token'});
  }

  http.Response _me(http.Request request) {
    if (request.headers['Authorization'] != 'Bearer test-token') {
      return _error(401, 'unauthorized');
    }
    return _json({
      'id': 'usr-admin',
      'username': 'admin',
      'org': 'ACBMSP',
      'role': 'admin',
      'createdAt': '2024-01-01T00:00:00Z',
    });
  }

  http.Response _createAsset(http.Request request) {
    final body = jsonDecode(request.body) as Map<String, dynamic>;
    final assetId = body['assetId'] as String;
    if (_assets.containsKey(assetId)) {
      return _json({
        'success': false,
        'error': 'the asset $assetId already exists',
      });
    }
    _assets[assetId] = {
      'ID': assetId,
      'Color': body['color'],
      'Size': body['size'],
      'Owner': body['owner'],
      'AppraisedValue': body['appraisedValue'],
    };
    return _json({'success': true});
  }

  http.Response _queryAsset(http.Request request) {
    final assetId = request.url.queryParameters['assetId'];
    final asset = _assets[assetId];
    if (asset == null) {
      return _error(404, 'the asset $assetId does not exist');
    }
    return _json({'data': jsonEncode(asset)});
  }

  http.Response _json(Map<String, dynamic> body) => http.Response(
        jsonEncode(body),
        200,
        headers: {'content-type': 'application/json'},
      );

  http.Response _error(int status, String message) => http.Response(
        jsonEncode({'error': message}),
        status,
        headers: {'content-type': 'application/json'},
      );
}
