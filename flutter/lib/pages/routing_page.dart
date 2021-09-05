import 'package:cosmos_ui_components/cosmos_ui_components.dart';
import 'package:flutter/material.dart';
import 'package:starport_template/pages/mnemonic_onboarding_page.dart';
import 'package:starport_template/pages/wallets_list_page.dart';
import 'package:starport_template/starport_app.dart';

class RoutingPage extends StatefulWidget {
  const RoutingPage({Key? key}) : super(key: key);

  @override
  _RoutingPageState createState() => _RoutingPageState();
}

class _RoutingPageState extends State<RoutingPage> {
  @override
  void initState() {
    super.initState();
    _loadWallets();
  }

  Future<void> _loadWallets() async {
    final store = StarportApp.walletsStore;
    await store.loadWallets();
    if (store.loadWalletsFailure.value == null) {
      if (!mounted) {
        return;
      }
      if (store.wallets.value.isEmpty) {
        Navigator.of(context).push(MaterialPageRoute(builder: (_) => const MnemonicOnboardingPage()));
      } else {
        Navigator.of(context).push(MaterialPageRoute(builder: (_) => const WalletsListPage()));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: ContentStateSwitcher(
        isLoading: StarportApp.walletsStore.areWalletsLoading,
        isError: StarportApp.walletsStore.loadWalletsFailure.value != null,
        errorChild: const CosmosErrorView(
          title: "Something went wrong",
          message: "We had problems retrieving wallets from secure storage.",
        ),
        contentChild: const SizedBox(),
      ),
    );
  }
}
