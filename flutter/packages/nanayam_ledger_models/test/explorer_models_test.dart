import 'package:nanayam_ledger_models/nanayam_ledger_models.dart';
import 'package:test/test.dart';

void main() {
  group('LedgerBlockSummary.fromJson', () {
    test('parses a full block summary', () {
      final block = LedgerBlockSummary.fromJson(const <String, dynamic>{
        'number': 3,
        'hash': 'abc',
        'prevHash': 'def',
        'txCount': 2,
        'timestamp': '2024-01-01T00:00:00Z',
        'dataHash': 'ghi',
      });

      expect(block.number, 3);
      expect(block.hash, 'abc');
      expect(block.prevHash, 'def');
      expect(block.txCount, 2);
      expect(block.dataHash, 'ghi');
    });

    test('defaults missing fields to zero/empty', () {
      final block = LedgerBlockSummary.fromJson(const <String, dynamic>{});
      expect(block.number, 0);
      expect(block.hash, '');
      expect(block.txCount, 0);
    });
  });

  group('LedgerActivity.fromJson', () {
    test('parses height and complaint count', () {
      final activity = LedgerActivity.fromJson(const <String, dynamic>{
        'height': 42,
        'complaintCount': 5,
        'channel': 'mychannel',
        'chaincode': 'basic',
      });

      expect(activity.height, 42);
      expect(activity.complaintCount, 5);
      expect(activity.channel, 'mychannel');
      expect(activity.chaincode, 'basic');
    });
  });

  group('ChannelInfo.fromJson', () {
    test('parses nested organizations and orderers', () {
      final info = ChannelInfo.fromJson(const <String, dynamic>{
        'channel': 'mychannel',
        'chaincode': 'basic',
        'mspId': 'ACBMSP',
        'organizations': [
          {
            'mspId': 'ACBMSP',
            'name': 'ACB',
            'peers': ['peer0.acb.nanayam.com:7051'],
            'role': 'Acknowledge',
          },
        ],
        'orderers': ['orderer.nanayam.com:7050'],
      });

      expect(info.channel, 'mychannel');
      expect(info.mspId, 'ACBMSP');
      expect(info.organizations, hasLength(1));
      expect(info.organizations.single.mspId, 'ACBMSP');
      expect(info.organizations.single.peers, ['peer0.acb.nanayam.com:7051']);
      expect(info.orderers, ['orderer.nanayam.com:7050']);
    });

    test('defaults to empty lists when organizations/orderers are absent', () {
      final info = ChannelInfo.fromJson(const <String, dynamic>{
        'channel': 'mychannel',
        'chaincode': 'basic',
        'mspId': 'ACBMSP',
      });

      expect(info.organizations, isEmpty);
      expect(info.orderers, isEmpty);
    });
  });
}
