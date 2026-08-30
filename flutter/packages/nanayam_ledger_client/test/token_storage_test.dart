import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';
import 'package:test/test.dart';

void main() {
  group('InMemoryTokenStorage', () {
    test('starts empty', () async {
      final storage = InMemoryTokenStorage();
      expect(await storage.read(), isNull);
    });

    test('round-trips a written token', () async {
      final storage = InMemoryTokenStorage();
      await storage.write('token-123');
      expect(await storage.read(), 'token-123');
    });

    test('clear() forgets the token', () async {
      final storage = InMemoryTokenStorage();
      await storage.write('token-123');
      await storage.clear();
      expect(await storage.read(), isNull);
    });
  });
}
