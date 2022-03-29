package keeper_test

import (
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
)

func (suite *KeeperTestSuite) TestFeeDecorator() {
	suite.SetupTest(false)

	mempoolFeeOpts := types.NewDefaultMempoolFeeOptions()
	mempoolFeeOpts.MinGasPriceForHighGasTx = sdk.MustNewDecFromStr("0.0025")
	baseDenom, _ := suite.app.TxFeesKeeper.GetBaseDenom(suite.ctx)

	uion := "uion"

	uionPoolId := suite.PreparePoolWithAssets(
		sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
		sdk.NewInt64Coin(uion, 500),
	)
	suite.ExecuteUpgradeFeeTokenProposal(uion, uionPoolId)

	tests := []struct {
		name         string
		txFee        sdk.Coins
		minGasPrices sdk.DecCoins
		gasRequested uint64
		isCheckTx    bool
		expectPass   bool
		baseDenomGas bool
	}{
		{
			name:         "no min gas price - checktx",
			txFee:        sdk.NewCoins(),
			minGasPrices: sdk.NewDecCoins(),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   true,
			baseDenomGas: true,
		},
		{
			name:         "no min gas price - delivertx",
			txFee:        sdk.NewCoins(),
			minGasPrices: sdk.NewDecCoins(),
			gasRequested: 10000,
			isCheckTx:    false,
			expectPass:   true,
			baseDenomGas: true,
		},
		{
			name:  "works with valid basedenom fee",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   true,
			baseDenomGas: true,
		},
		{
			name:  "doesn't work with not enough fee in checktx",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   false,
			baseDenomGas: true,
		},
		{
			name:  "works with not enough fee in delivertx",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 10000,
			isCheckTx:    false,
			expectPass:   true,
			baseDenomGas: true,
		},
		{
			name:  "works with valid converted fee",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(uion, 1000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   true,
			baseDenomGas: false,
		},
		{
			name:  "doesn't work with not enough converted fee in checktx",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(uion, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   false,
			baseDenomGas: false,
		},
		{
			name:  "works with not enough converted fee in delivertx",
			txFee: sdk.NewCoins(sdk.NewInt64Coin(uion, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
				sdk.MustNewDecFromStr("0.1"))),
			gasRequested: 10000,
			isCheckTx:    false,
			expectPass:   true,
			baseDenomGas: false,
		},
		{
			name:         "multiple fee coins - checktx",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1), sdk.NewInt64Coin(uion, 1)),
			minGasPrices: sdk.NewDecCoins(),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   false,
			baseDenomGas: false,
		},
		{
			name:         "multiple fee coins - delivertx",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1), sdk.NewInt64Coin(uion, 1)),
			minGasPrices: sdk.NewDecCoins(),
			gasRequested: 10000,
			isCheckTx:    false,
			expectPass:   false,
			baseDenomGas: false,
		},
		{
			name:         "invalid fee denom",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin("moo", 1)),
			minGasPrices: sdk.NewDecCoins(),
			gasRequested: 10000,
			isCheckTx:    false,
			expectPass:   false,
			baseDenomGas: false,
		},
		{
			name:         "mingasprice not containing basedenom gets treated as min gas price 0",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 100000000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: 10000,
			isCheckTx:    true,
			expectPass:   true,
			baseDenomGas: false,
		},
		{
			name:         "tx with gas wanted more than allowed should not pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 100000000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.MaxGasWantedPerTx + 1,
			isCheckTx:    true,
			expectPass:   false,
			baseDenomGas: false,
		},
		{
			name:         "tx with high gas and not enough fee should no pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.HighGasTxThreshold,
			isCheckTx:    true,
			expectPass:   false,
			baseDenomGas: false,
		},
		{
			name:         "tx with high gas and enough fee should pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 10*1000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.HighGasTxThreshold,
			isCheckTx:    true,
			expectPass:   true,
			baseDenomGas: false,
		},
	}

	for _, tc := range tests {
		suite.ctx = suite.ctx.WithIsCheckTx(tc.isCheckTx)
		suite.ctx = suite.ctx.WithMinGasPrices(tc.minGasPrices)

		txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
		priv0, _, addr0 := testdata.KeyTestPubAddr()
		acc1 := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr0)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc1)
		msgs := []sdk.Msg{testdata.NewTestMsg(addr0)}
		privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
		signerData := authsigning.SignerData{
			ChainID: suite.ctx.ChainID(), 
			AccountNumber: accNums[0], 
			Sequence: accSeqs[0]}

		gasLimit := tc.gasRequested
		sigV2, _ := clienttx.SignWithPrivKey(
			1, 
			signerData,
			txBuilder, 
			privs[0], 
			suite.clientCtx.TxConfig,
			accSeqs[0])

		simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr0, tc.txFee)
		
		txBuilder.SetMsgs(msgs[0])
		txBuilder.SetSignatures(sigV2)
		txBuilder.SetMemo("")
		txBuilder.SetFeeAmount(tc.txFee)
		txBuilder.SetGasLimit(gasLimit)

		tx := txBuilder.GetTx()

		mfd := keeper.NewMempoolFeeDecorator(*suite.app.TxFeesKeeper, mempoolFeeOpts)
		dfd := keeper.NewDeductFeeDecorator(*suite.app.TxFeesKeeper, *suite.app.AccountKeeper, *suite.app.BankKeeper, *suite.app.FeeGrantKeeper)
		antehandlerMFD := sdk.ChainAnteDecorators(mfd, dfd)
		_, err := antehandlerMFD(suite.ctx, tx, false)

		if tc.expectPass {
			if tc.baseDenomGas && !tc.txFee.IsZero() {
				moduleAddr := suite.app.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
				suite.Require().Equal(tc.txFee[0], suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddr, baseDenom), tc.name)
				suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.FeeCollectorName, addr0, tc.txFee)
			} else if !tc.txFee.IsZero() {
				moduleAddr := suite.app.AccountKeeper.GetModuleAddress(types.FooCollectorName)
				suite.Require().Equal(tc.txFee[0], suite.app.BankKeeper.GetBalance(suite.ctx, moduleAddr, tc.txFee[0].Denom), tc.name)
				suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, types.FooCollectorName, addr0, tc.txFee)
			}
			suite.Require().NoError(err, "test: %s", tc.name)
		} else {
			suite.Require().Error(err, "test: %s", tc.name)
		}
	}
}
