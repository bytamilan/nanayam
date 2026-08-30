import 'package:flutter/material.dart';

import '../app_scope.dart';
import 'login_screen.dart';
import 'provision_tab.dart';
import 'redeem_tab.dart';
import 'wallet_tab.dart';

/// Bottom-navigation shell over the three demo personas: citizen wallet,
/// merchant redemption, and government/program provisioning. A real app
/// would gate these behind actual roles; here all three are always visible
/// so the ledger interactions are easy to try out end to end.
class HomeShell extends StatefulWidget {
  const HomeShell({super.key});

  @override
  State<HomeShell> createState() => _HomeShellState();
}

class _HomeShellState extends State<HomeShell> {
  int _index = 0;

  static const _tabs = [
    _TabSpec(label: 'Wallet', icon: Icons.wallet_outlined, child: WalletTab()),
    _TabSpec(
      label: 'Redeem',
      icon: Icons.storefront_outlined,
      child: RedeemTab(),
    ),
    _TabSpec(
      label: 'Provision',
      icon: Icons.add_card_outlined,
      child: ProvisionTab(),
    ),
  ];

  Future<void> _logout() async {
    await AppScope.of(context).logout();
    if (!mounted) return;
    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute(builder: (_) => const LoginScreen()),
      (route) => false,
    );
  }

  @override
  Widget build(BuildContext context) {
    final session = AppScope.of(context);
    return Scaffold(
      appBar: AppBar(
        title: Text(_tabs[_index].label),
        actions: [
          if (session.currentUser != null)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8),
              child: Center(child: Text(session.currentUser!.username)),
            ),
          IconButton(onPressed: _logout, icon: const Icon(Icons.logout)),
        ],
      ),
      body: IndexedStack(
        index: _index,
        children: _tabs.map((t) => t.child).toList(),
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _index,
        onDestinationSelected: (i) => setState(() => _index = i),
        destinations: _tabs
            .map(
              (t) => NavigationDestination(icon: Icon(t.icon), label: t.label),
            )
            .toList(),
      ),
    );
  }
}

class _TabSpec {
  const _TabSpec({
    required this.label,
    required this.icon,
    required this.child,
  });

  final String label;
  final IconData icon;
  final Widget child;
}
