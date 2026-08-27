import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';
import 'package:test/test.dart';

import 'support/fake_gateway.dart';

void main() {
  group('public/no-auth endpoints', () {
    test('fetchConfig reports signup availability', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );

      final config = await client.fetchConfig();
      expect(config.signupEnabled, isTrue);
      expect(config.channel, 'mychannel');
    });

    test('health returns true on a 200', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );
      expect(await client.health(), isTrue);
    });
  });

  group('register', () {
    test('creates a new account when signup is enabled', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );

      final user = await client.register(
        username: 'newcitizen',
        password: 'hunter2',
        org: 'Org1MSP',
      );

      expect(user.username, 'newcitizen');
      expect(user.org, 'Org1MSP');
    });

    test('throws a 403 LedgerApiException when signup is disabled', () async {
      final gateway = FakeGateway(signupEnabled: false);
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );

      expect(
        () => client.register(username: 'x', password: 'y'),
        throwsA(
          isA<LedgerApiException>().having(
            (e) => e.isForbidden,
            'isForbidden',
            isTrue,
          ),
        ),
      );
    });
  });

  group('isAuthenticated / logout', () {
    test('is false before login, true after, false after logout', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );

      expect(await client.isAuthenticated, isFalse);

      await client.login(username: 'admin', password: 'admin');
      expect(await client.isAuthenticated, isTrue);

      await client.logout();
      expect(await client.isAuthenticated, isFalse);
    });
  });

  group('ledger explorer', () {
    test('fetchChannelInfo parses organizations', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );
      await client.login(username: 'admin', password: 'admin');

      final info = await client.fetchChannelInfo();
      expect(info.channel, 'mychannel');
      expect(info.mspId, 'ACBMSP');
    });

    test('fetchLedgerBlocks parses the blocks array', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );
      await client.login(username: 'admin', password: 'admin');

      final blocks = await client.fetchLedgerBlocks();
      expect(blocks, hasLength(1));
      expect(blocks.single.hash, 'h0');
    });

    test('fetchLedgerActivity parses height and complaint count', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );
      await client.login(username: 'admin', password: 'admin');

      final activity = await client.fetchLedgerActivity();
      expect(activity.height, 1);
      expect(activity.complaintCount, 0);
    });
  });

  group('generic asset ledger', () {
    test('createAsset then listAssetIds/listAssets round-trip', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );
      await client.login(username: 'admin', password: 'admin');

      final result = await client.createAsset(
        assetId: 'asset1',
        color: 'blue',
        size: 5,
        owner: 'Tomoko',
        appraisedValue: 300,
      );
      expect(result.success, isTrue);

      final ids = await client.listAssetIds();
      expect(ids, ['asset1']);

      final assets = await client.listAssets();
      expect(assets, hasLength(1));
      expect(assets.single.owner, 'Tomoko');
    });

    test('createAsset for a duplicate ID reports success:false', () async {
      final gateway = FakeGateway();
      final client = NanayamLedgerClient(
        baseUrl: 'http://localhost:8080',
        httpClient: gateway.asClient(),
      );
      await client.login(username: 'admin', password: 'admin');

      await client.createAsset(
        assetId: 'dup',
        color: 'blue',
        size: 1,
        owner: 'a',
        appraisedValue: 1,
      );
      final second = await client.createAsset(
        assetId: 'dup',
        color: 'blue',
        size: 1,
        owner: 'a',
        appraisedValue: 1,
      );

      expect(second.success, isFalse);
      expect(second.error, contains('already exists'));
    });
  });
}
