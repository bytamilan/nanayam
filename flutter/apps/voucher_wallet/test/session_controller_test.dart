import 'package:flutter_test/flutter_test.dart';
import 'package:nanayam_ledger_client/nanayam_ledger_client.dart';
import 'package:voucher_wallet/src/session_controller.dart';

import 'support/fake_gateway.dart';

void main() {
  late FakeGateway gateway;
  late SessionController session;

  setUp(() {
    gateway = FakeGateway();
    session = SessionController(
      clientFactory: (baseUrl) => NanayamLedgerClient(
        baseUrl: baseUrl,
        httpClient: gateway.asClient(),
      ),
    );
  });

  tearDown(() => session.dispose());

  test('starts signed out', () {
    expect(session.isLoggedIn, isFalse);
    expect(session.currentUser, isNull);
  });

  test('connectAndLogin signs in and exposes the current user', () async {
    await session.connectAndLogin(
      baseUrl: 'http://localhost:8080',
      username: 'admin',
      password: 'admin',
    );

    expect(session.isLoggedIn, isTrue);
    expect(session.currentUser?.username, 'admin');
    expect(session.isBusy, isFalse);
  });

  test('connectAndLogin with bad credentials throws and stays signed out', () async {
    await expectLater(
      session.connectAndLogin(
        baseUrl: 'http://localhost:8080',
        username: 'admin',
        password: 'wrong',
      ),
      throwsA(isA<LedgerApiException>()),
    );

    expect(session.isLoggedIn, isFalse);
    expect(session.lastError, isNotNull);
  });

  test('logout clears the session', () async {
    await session.connectAndLogin(
      baseUrl: 'http://localhost:8080',
      username: 'admin',
      password: 'admin',
    );
    expect(session.isLoggedIn, isTrue);

    await session.logout();

    expect(session.isLoggedIn, isFalse);
    expect(session.currentUser, isNull);
  });

  test('vouchers/client getters throw before a successful login', () {
    expect(() => session.client, throwsStateError);
    expect(() => session.vouchers, throwsStateError);
  });

  test('notifies listeners on login and logout', () async {
    var notifications = 0;
    session.addListener(() => notifications++);

    await session.connectAndLogin(
      baseUrl: 'http://localhost:8080',
      username: 'admin',
      password: 'admin',
    );
    expect(notifications, greaterThan(0));

    final afterLogin = notifications;
    await session.logout();
    expect(notifications, greaterThan(afterLogin));
  });
}
