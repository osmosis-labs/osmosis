import 'package:flutter/material.dart';
import 'package:starport_template/starport_app.dart';
import 'package:starport_template/stores/wallets_store.dart';
import 'package:starport_template/utils/base_env.dart';
import 'package:transaction_signing_gateway/alan/alan_credentials_serializer.dart';
import 'package:transaction_signing_gateway/alan/alan_transaction_broadcaster.dart';
import 'package:transaction_signing_gateway/alan/alan_transaction_signer.dart';
import 'package:transaction_signing_gateway/gateway/transaction_signing_gateway.dart';
import 'package:transaction_signing_gateway/mobile/mobile_key_info_storage.dart';
import 'package:transaction_signing_gateway/mobile/no_op_transaction_summary_ui.dart';

void main() {
  _buildDependencies();
  runApp(StarportApp());
}

void _buildDependencies() {
  StarportApp.signingGateway = TransactionSigningGateway(
    transactionSummaryUI: NoOpTransactionSummaryUI(),
    signers: [
      AlanTransactionSigner(),
    ],
    broadcasters: [
      AlanTransactionBroadcaster(),
    ],
    infoStorage: MobileKeyInfoStorage(
      serializers: [AlanCredentialsSerializer()],
    ),
  );
  StarportApp.baseEnv = BaseEnv()
    ..setEnv(
      lcdUrl: lcdUrl,
      grpcUrl: grpcUrl,
      lcdPort: lcdPort,
      grpcPort: grpcPort,
      ethUrl: ethUrl,
    );
  StarportApp.walletsStore = WalletsStore(StarportApp.signingGateway, StarportApp.baseEnv);
}
