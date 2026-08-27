import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';
import 'package:test/test.dart';

void main() {
  group('NanayamLedgerClient', () {
    test('login stores the returned token and me() sends it as a bearer',
        () async {
      String? capturedAuthHeader;

      final mock = MockClient((http.Request request) async {
        if (request.url.path == '/v1/Login') {
          return http.Response(jsonEncode(<String, String>{
            'token': 'test-token',
          }), 200);
        }
        if (request.url.path == '/v1/Me') {
          capturedAuthHeader = request.headers['Authorization'];
          return http.Response(
            jsonEncode(<String, dynamic>{
              'id': 'usr-1',
              'username': 'admin',
              'org': 'ACBMSP',
              'role': 'admin',
              'createdAt': '2024-01-01T00:00:00Z',
            }),
            200,
          );
        }
        return http.Response('not found', 404);
      });

      final client =
          NanayamLedgerClient(baseUrl: 'http://localhost:8080', httpClient: mock);

      await client.login(username: 'admin', password: 'admin');
      final me = await client.me();

      expect(capturedAuthHeader, 'Bearer test-token');
      expect(me.username, 'admin');
      expect(me.isAdmin, isTrue);
    });

    test('queryAsset decodes the nested chaincode JSON envelope', () async {
      final mock = MockClient((http.Request request) async {
        expect(request.url.path, '/v1/QueryAsset');
        expect(request.url.queryParameters['assetId'], 'asset1');
        return http.Response(
          jsonEncode(<String, String>{
            'data': jsonEncode(<String, dynamic>{
              'ID': 'asset1',
              'Color': 'blue',
              'Size': 5,
              'Owner': 'Tomoko',
              'AppraisedValue': 300,
            }),
          }),
          200,
        );
      });

      final client =
          NanayamLedgerClient(baseUrl: 'http://localhost:8080', httpClient: mock);
      final asset = await client.queryAsset('asset1');

      expect(asset.owner, 'Tomoko');
      expect(asset.appraisedValue, 300);
    });

    test('non-2xx responses throw LedgerApiException with parsed message',
        () async {
      final mock = MockClient((http.Request request) async {
        return http.Response(
          jsonEncode(<String, String>{'error': 'invalid username or password'}),
          401,
        );
      });

      final client =
          NanayamLedgerClient(baseUrl: 'http://localhost:8080', httpClient: mock);

      expect(
        () => client.login(username: 'nope', password: 'nope'),
        throwsA(
          isA<LedgerApiException>()
              .having((e) => e.statusCode, 'statusCode', 401)
              .having((e) => e.message, 'message', 'invalid username or password'),
        ),
      );
    });
  });
}
