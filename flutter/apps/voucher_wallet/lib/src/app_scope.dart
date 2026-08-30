import 'package:flutter/widgets.dart';

import 'session_controller.dart';

/// Makes the app's single [SessionController] available to the whole widget
/// tree via [AppScope.of], without pulling in a state-management package for
/// an app this small.
class AppScope extends InheritedNotifier<SessionController> {
  const AppScope({
    required SessionController session,
    required super.child,
    super.key,
  }) : super(notifier: session);

  static SessionController of(BuildContext context) {
    final scope = context.dependOnInheritedWidgetOfExactType<AppScope>();
    assert(scope != null, 'No AppScope found in context');
    return scope!.notifier!;
  }
}
