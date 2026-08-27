import 'package:nanayam_ledger_models/nanayam_ledger_models.dart';
import 'package:test/test.dart';

void main() {
  group('LedgerUser.fromJson', () {
    test('parses a well-formed user', () {
      final user = LedgerUser.fromJson(const <String, dynamic>{
        'id': 'usr-1',
        'username': 'admin',
        'org': 'ACBMSP',
        'role': 'admin',
        'createdAt': '2024-01-01T00:00:00Z',
      });

      expect(user.id, 'usr-1');
      expect(user.username, 'admin');
      expect(user.org, 'ACBMSP');
      expect(user.isAdmin, isTrue);
      expect(user.createdAt, DateTime.utc(2024));
    });

    test('defaults role to "user" semantics when missing', () {
      final user = LedgerUser.fromJson(const <String, dynamic>{
        'id': 'usr-2',
        'username': 'alice',
        'org': 'Org1MSP',
      });

      expect(user.role, 'user');
      expect(user.isAdmin, isFalse);
    });

    test('falls back to the epoch for an unparsable createdAt', () {
      final user = LedgerUser.fromJson(const <String, dynamic>{
        'id': 'usr-3',
        'username': 'bob',
        'org': 'Org1MSP',
        'role': 'user',
        'createdAt': 'not-a-date',
      });

      expect(
        user.createdAt,
        DateTime.fromMillisecondsSinceEpoch(0, isUtc: true),
      );
    });

    test('round-trips through toJson', () {
      final original = LedgerUser(
        id: 'usr-1',
        username: 'admin',
        org: 'ACBMSP',
        role: 'admin',
        createdAt: DateTime.utc(2024, 1, 1),
      );
      final json = original.toJson();
      final restored = LedgerUser.fromJson(json);

      expect(restored.id, original.id);
      expect(restored.username, original.username);
      expect(restored.org, original.org);
      expect(restored.role, original.role);
      expect(restored.createdAt, original.createdAt);
    });
  });

  group('LedgerSession.fromJson', () {
    test('extracts the bearer token', () {
      final session =
          LedgerSession.fromJson(const <String, dynamic>{'token': 'abc.def'});
      expect(session.token, 'abc.def');
    });
  });

  group('GatewayConfig.fromJson', () {
    test('parses signup flag and channel metadata', () {
      final config = GatewayConfig.fromJson(const <String, dynamic>{
        'signupEnabled': true,
        'channel': 'mychannel',
        'chaincode': 'basic',
      });

      expect(config.signupEnabled, isTrue);
      expect(config.channel, 'mychannel');
      expect(config.chaincode, 'basic');
    });

    test('defaults signupEnabled to false when absent', () {
      final config = GatewayConfig.fromJson(const <String, dynamic>{
        'channel': 'mychannel',
        'chaincode': 'basic',
      });
      expect(config.signupEnabled, isFalse);
    });
  });
}
