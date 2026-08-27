import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

/// A tiny in-memory stand-in for the Nanayam gateway, covering just enough
/// of `services/gateway`'s REST surface to exercise `NanayamLedgerClient`
/// without a running Fabric network: auth, the generic asset ledger, and
/// the read-only ledger explorer endpoints.
class FakeGateway {
  FakeGateway({this.signupEnabled = true});

  final bool signupEnabled;
  final Map<String, String> _users = <String, String>{'admin': 'admin'};
  final Map<String, Map<String, dynamic>> _assets =
      <String, Map<String, dynamic>>{};

  http.Client asClient() => MockClient(_handle);

  Future<http.Response> _handle(http.Request request) async {
    switch ('${request.method} ${request.url.path}') {
      case 'GET /v1/Config':
        return _json({
          'signupEnabled': signupEnabled,
          'channel': 'mychannel',
          'chaincode': 'basic',
        });
      case 'GET /health':
        return _json({'status': 'ok'});
      case 'POST /v1/Register':
        return _register(request);
      case 'POST /v1/Login':
        return _login(request);
      case 'GET /v1/Me':
        return _me(request);
      case 'GET /v1/ChannelInfo':
        return _json({
          'channel': 'mychannel',
          'chaincode': 'basic',
          'mspId': 'ACBMSP',
          'organizations': <dynamic>[],
          'orderers': <dynamic>[],
        });
      case 'GET /v1/LedgerBlocks':
        return _json({
          'blocks': [
            {
              'number': 0,
              'hash': 'h0',
              'prevHash': '',
              'txCount': 1,
              'timestamp': '2024-01-01T00:00:00Z',
              'dataHash': 'd0',
            },
          ],
        });
      case 'GET /v1/LedgerActivity':
        return _json({
          'height': 1,
          'complaintCount': 0,
          'channel': 'mychannel',
          'chaincode': 'basic',
        });
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

  bool _authorized(http.Request request) =>
      request.headers['Authorization'] == 'Bearer test-token';

  http.Response _register(http.Request request) {
    if (!signupEnabled) {
      return _error(403, 'registration is disabled');
    }
    final body = jsonDecode(request.body) as Map<String, dynamic>;
    final username = body['username'] as String;
    if (_users.containsKey(username)) {
      return _error(400, 'username already exists');
    }
    _users[username] = body['password'] as String;
    return http.Response(
      jsonEncode({
        'id': 'usr-$username',
        'username': username,
        'org': body['org'],
        'role': 'user',
        'createdAt': '2024-01-01T00:00:00Z',
      }),
      201,
      headers: {'content-type': 'application/json'},
    );
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
    if (!_authorized(request)) return _error(401, 'unauthorized');
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
