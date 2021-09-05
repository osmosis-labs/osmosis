import 'package:alan/alan.dart' as alan;
import 'package:mobx/mobx.dart';
import 'package:starport_template/entities/balance.dart';
import 'package:starport_template/utils/base_env.dart';
import 'package:starport_template/utils/cosmos_balances.dart';
import 'package:starport_template/utils/token_sender.dart';
import 'package:transaction_signing_gateway/gateway/transaction_signing_gateway.dart';
import 'package:transaction_signing_gateway/model/credentials_storage_failure.dart';
import 'package:transaction_signing_gateway/model/wallet_public_info.dart';
import 'package:transaction_signing_gateway/transaction_signing_gateway.dart';
import 'package:uuid/uuid.dart';

class WalletsStore {
  final TransactionSigningGateway _transactionSigningGateway;
  final BaseEnv baseEnv;

  WalletsStore(this._transactionSigningGateway, this.baseEnv);

  final Observable<bool> _areWalletsLoading = Observable(false);

  final Observable<bool> _isSendMoneyLoading = Observable(false);
  final Observable<bool> _isSendMoneyError = Observable(false);
  final Observable<bool> _isBalancesLoading = Observable(false);
  final Observable<bool> _isError = Observable(false);

  bool get areWalletsLoading => _areWalletsLoading.value;

  set areWalletsLoading(bool val) => Action(() => _areWalletsLoading.value = val)();

  bool get isSendMoneyError => _isSendMoneyError.value;

  set isSendMoneyError(bool val) => Action(() => _isSendMoneyError.value = val)();

  bool get isSendMoneyLoading => _isSendMoneyLoading.value;

  set isSendMoneyLoading(bool val) => Action(() => _isSendMoneyLoading.value = val)();

  bool get isError => _isError.value;

  set isError(bool val) => Action(() => _isError.value = val)();

  bool get isBalancesLoading => _isBalancesLoading.value;

  set isBalancesLoading(bool val) => Action(() => _isBalancesLoading.value = val)();

  final Observable<List<Balance>> balancesList = Observable([]);

  final Observable<CredentialsStorageFailure?> loadWalletsFailure = Observable(null);

  Observable<List<WalletPublicInfo>> wallets = Observable([]);

  Future<void> loadWallets() async {
    areWalletsLoading = true;
    (await _transactionSigningGateway.getWalletsList()).fold(
      (fail) => loadWalletsFailure.value = fail,
      (newWallets) => wallets.value = newWallets,
    );
    areWalletsLoading = false;
  }

  Future<void> getBalances(String walletAddress) async {
    isError = false;
    isBalancesLoading = true;
    try {
      balancesList.value = await CosmosBalances(baseEnv).getBalances(walletAddress);
    } catch (error) {
      isError = false;
    }
    isBalancesLoading = false;
  }

  Future<WalletPublicInfo> importAlanWallet(
    String mnemonic,
    String password,
  ) async {
    final wallet = alan.Wallet.derive(mnemonic.split(" "), baseEnv.networkInfo);
    final creds = AlanPrivateWalletCredentials(
      publicInfo: WalletPublicInfo(
        chainId: 'cosmos',
        walletId: const Uuid().v4(),
        name: 'First wallet',
        publicAddress: wallet.bech32Address,
      ),
      mnemonic: mnemonic,
      networkInfo: baseEnv.networkInfo,
    );
    await _transactionSigningGateway.storeWalletCredentials(
      credentials: creds,
      password: password,
    );
    wallets.value.add(creds.publicInfo);
    return creds.publicInfo;
  }

  Future<void> sendCosmosMoney(
    WalletPublicInfo info,
    Balance balance,
    String toAddress,
  ) async {
    isSendMoneyLoading = true;
    isSendMoneyError = false;
    try {
      await TokenSender(_transactionSigningGateway).sendCosmosMoney(
        info,
        balance,
        toAddress,
      );
    } catch (ex) {
      isError = true;
    }
    isSendMoneyLoading = false;
  }
}
