import 'package:nanayam_ledger_models/nanayam_ledger_models.dart';
import 'package:test/test.dart';

void main() {
  group('LedgerAsset.fromJson', () {
    test('parses PascalCase chaincode field names', () {
      final asset = LedgerAsset.fromJson(const <String, dynamic>{
        'ID': 'asset1',
        'Color': 'blue',
        'Size': 5,
        'Owner': 'Tomoko',
        'AppraisedValue': 300,
      });

      expect(asset.id, 'asset1');
      expect(asset.color, 'blue');
      expect(asset.size, 5);
      expect(asset.owner, 'Tomoko');
      expect(asset.appraisedValue, 300);
    });

    test('parses camelCase / gateway field names', () {
      final asset = LedgerAsset.fromJson(const <String, dynamic>{
        'assetId': 'asset2',
        'color': 'red',
        'size': '10',
        'owner': 'Alice',
        'appraisedValue': '500',
      });

      expect(asset.id, 'asset2');
      expect(asset.size, 10);
      expect(asset.appraisedValue, 500);
    });
  });

  group('LedgerApiException', () {
    test('classifies status codes', () {
      const err = LedgerApiException(401, 'unauthorized');
      expect(err.isUnauthorized, isTrue);
      expect(err.isForbidden, isFalse);
    });
  });
}
