package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"

	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"

	"github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"
)

func (suite *KeeperTestSuite) TestFeeDecorator() {
	suite.SetupTest(false)

	/*
	// Note: gammKeeper is expected to satisfy the SpotPriceCalculator interface parameter
	txFeesKeeper := txfeeskeeper.NewKeeper(
		appCodec,
		suite.app.AccountKeeper,
		suite.app.BankKeeper,
		suite.app.EpochsKeeper,
		keys[txfeestypes.StoreKey],
		app.GAMMKeeper,
		app.GAMMKeeper,
		txfeestypes.FeeCollectorName,
		txfeestypes.FooCollectorName,
	)
	app.TxFeesKeeper = &txFeesKeeper
	*/

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

		tx := legacytx.NewStdTx([]sdk.Msg{}, legacytx.NewStdFee(
			tc.gasRequested,
			tc.txFee,
		), []legacytx.StdSignature{}, "")

		mfd := keeper.NewMempoolFeeDecorator(*suite.app.TxFeesKeeper, mempoolFeeOpts)
		antehandlerMFD := sdk.ChainAnteDecorators(mfd)
		_, err := antehandlerMFD(suite.ctx, tx, false)
		if tc.expectPass {
			suite.Require().NoError(err, "test: %s", tc.name)
		} else {
			suite.Require().Error(err, "test: %s", tc.name)
		}

		// DeductFeeDecorator tests:

		// Does FeeGrantKeeper need to be added to the TxFeesKeeper struct?

		dfd := keeper.NewDeductFeeDecorator(*suite.app.TxFeesKeeper, *suite.app.AccountKeeper, *suite.app.BankKeeper, *suite.app.FeeGrantKeeper)
		antehandlerDFD := sdk.ChainAnteDecorators(dfd)
		_, err = antehandlerDFD(suite.ctx, tx, false)
		if tc.expectPass {
			if tc.baseDenomGas {
				// check main module account osmo balance & require no non-OSMO
				// also require that the balance of the second module account (just get AllBalances) has not changed (suite.Require().Equalf(123, 123, "error message %s", "formatted"))
			} else {
				// check second module account to make sure the right fee amount has flowed there
				// also require that the balance of the main module account (just get AllBalances) has not changed (suite.Require().Equalf(123, 123, "error message %s", "formatted"))
			}
			suite.Require().NoError(err, "test: %s", tc.name)
		} else {
			suite.Require().Error(err, "test: %s", tc.name)
		}
	}
}
