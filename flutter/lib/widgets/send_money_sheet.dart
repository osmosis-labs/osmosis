import 'package:cosmos_ui_components/components/cosmos_elevated_button.dart';
import 'package:cosmos_ui_components/components/template/cosmos_wallets_list_view.dart';
import 'package:flutter/material.dart';
import 'package:starport_template/entities/amount.dart';
import 'package:starport_template/entities/balance.dart';
import 'package:starport_template/entities/denom.dart';
import 'package:starport_template/starport_app.dart';
import 'package:transaction_signing_gateway/model/wallet_public_info.dart';

class SendMoneySheet extends StatefulWidget {
  final Denom denom;
  final WalletInfo walletInfo;

  const SendMoneySheet({
    Key? key,
    required this.denom,
    required this.walletInfo,
  }) : super(key: key);

  @override
  _SendMoneySheetState createState() => _SendMoneySheetState();
}

class _SendMoneySheetState extends State<SendMoneySheet> {
  String _toAddress = '';
  String _amount = '';

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Padding(padding: EdgeInsets.only(top: 16)),
        Text(
          widget.denom.text,
          style: Theme.of(context).textTheme.headline6,
        ),
        ListTile(
          title: TextFormField(
            decoration: const InputDecoration(
              labelText: 'Enter wallet address',
              border: OutlineInputBorder(),
            ),
            onChanged: (value) => _toAddress = value,
          ),
        ),
        ListTile(
          title: TextFormField(
            decoration: const InputDecoration(
              labelText: 'Enter amount',
              border: OutlineInputBorder(),
            ),
            onChanged: (value) => _amount = value,
          ),
        ),
        CosmosElevatedButton(
          onTap: _onSendMoneyClicked,
          text: 'Send money',
        ),
      ],
    );
  }

  void _onSendMoneyClicked() {
    final amount = Amount.fromString(_amount);
    final info = WalletPublicInfo(
      name: widget.walletInfo.name,
      publicAddress: widget.walletInfo.address,
      walletId: widget.walletInfo.walletId,
      chainId: 'cosmos',
    );
    final balance = Balance(denom: widget.denom, amount: amount);
    StarportApp.walletsStore.sendCosmosMoney(info, balance, _toAddress);
  }
}
