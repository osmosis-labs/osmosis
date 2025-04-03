package apptesting

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	customante "github.com/osmosis-labs/osmosis/v27/ante"

	"github.com/cosmos/cosmos-sdk/client"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/txfees/types"
)

var baseGas = uint64(10000)

func (s *KeeperTestHelper) ExecuteUpgradeFeeTokenProposal(feeToken string, poolId uint64) error {
	upgradeProp := types.NewUpdateFeeTokenProposal(
		"Test Proposal",
		"test",
		[]types.FeeToken{
			{
				Denom:  feeToken,
				PoolID: poolId,
			},
		},
	)
	return s.App.TxFeesKeeper.HandleUpdateFeeTokenProposal(s.Ctx, &upgradeProp)
}

func (s *KeeperTestHelper) SetupTxFeeAnteHandlerAndChargeFee(clientCtx client.Context, minGasPrices sdk.DecCoins, gasRequested uint64, isCheckTx, isSimulate bool, txFee sdk.Coins) error {
	mempoolFeeOpts := types.NewDefaultMempoolFeeOptions()
	mempoolFeeOpts.MinGasPriceForHighGasTx = osmomath.MustNewDecFromStr("0.0025")

	uionPoolId := s.PrepareBalancerPoolWithCoins(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin("uion", 500),
	)
	err := s.ExecuteUpgradeFeeTokenProposal("uion", uionPoolId)
	s.Require().NoError(err)

	if gasRequested == 0 {
		gasRequested = baseGas
	}
	s.Ctx = s.Ctx.WithIsCheckTx(isCheckTx).WithMinGasPrices(minGasPrices)

	// TODO: Cleanup this code.
	// TxBuilder components reset for every test case
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	priv0, _, addr0 := testdata.KeyTestPubAddr()
	acc1 := s.App.AccountKeeper.NewAccountWithAddress(s.Ctx, addr0)
	s.App.AccountKeeper.SetAccount(s.Ctx, acc1)
	msgs := []sdk.Msg{testdata.NewTestMsg(addr0)}
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
	signerData := authsigning.SignerData{
		ChainID:       s.Ctx.ChainID(),
		AccountNumber: accNums[0],
		Sequence:      accSeqs[0],
	}

	gasLimit := gasRequested
	sigV2, _ := clienttx.SignWithPrivKey(
		s.Ctx,
		1,
		signerData,
		txBuilder,
		privs[0],
		clientCtx.TxConfig,
		accSeqs[0],
	)

	err = testutil.FundAccount(s.Ctx, s.App.BankKeeper, addr0, txFee)
	s.Require().NoError(err)

	tx := s.BuildTx(txBuilder, msgs, sigV2, "", txFee, gasLimit)

	mfd := keeper.NewMempoolFeeDecorator(*s.App.TxFeesKeeper, mempoolFeeOpts)
	dfd := customante.NewDeductFeeDecorator(*s.App.TxFeesKeeper, *s.App.AccountKeeper, s.App.BankKeeper, nil,
		s.App.TreasuryKeeper, s.App.OracleKeeper)
	antehandlerMFD := sdk.ChainAnteDecorators(mfd, dfd)
	_, err = antehandlerMFD(s.Ctx, tx, isSimulate)
	return err
}
