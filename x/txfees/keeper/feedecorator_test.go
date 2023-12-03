package keeper_test

import (
	"fmt"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/osmosis-labs/osmosis/v15/x/txfees/ante"
	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

func (suite *KeeperTestSuite) TestFeeDecorator() {
	baseDenom := sdk.DefaultBondDenom
	baseGas := uint64(10000)
	point1BaseDenomMinGasPrices := sdk.NewDecCoins(sdk.NewDecCoinFromDec(baseDenom,
		sdk.MustNewDecFromStr("0.1")))

	// uion is setup with a relative price of 1:1
	uion := "uion"

	type testcase struct {
		name         string
		txFee        sdk.Coins
		minGasPrices sdk.DecCoins // if blank, set to 0
		gasRequested uint64       // if blank, set to base gas
		isCheckTx    bool
		isSimulate   bool // if blank, is false
		expectPass   bool
	}

	tests := []testcase{}
	txType := []string{"delivertx", "checktx"}
	for isCheckTx := 0; isCheckTx <= 1; isCheckTx++ {
		tests = append(tests, []testcase{
			{
				name:       fmt.Sprintf("no min gas price - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1000)),
				isCheckTx:  isCheckTx == 1,
				expectPass: true,
			},
			{
				name:       fmt.Sprintf("no min gas price, no fee - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(),
				isCheckTx:  isCheckTx == 1,
				expectPass: true,
			},
			{
				name:       fmt.Sprintf("no min gas price, invalid fee token - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(sdk.NewInt64Coin("uatom", 1000)),
				isCheckTx:  isCheckTx == 1,
				expectPass: true,
			},
			{
				name:         fmt.Sprintf("multiple fee coins - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1), sdk.NewInt64Coin(uion, 1)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   false,
			},
			{
				name:         fmt.Sprintf("no fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   isCheckTx != 1, //should pass on deliverTx, fail on checkTx
			},
			{
				name:         fmt.Sprintf("works with valid basedenom fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1000)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   true,
			},
			{
				name:         fmt.Sprintf("insufficient valid basedenom fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 10)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   isCheckTx != 1, //should pass on deliverTx, fail on checkTx
			},
			{
				name:         fmt.Sprintf("works with valid converted fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1000)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   true,
			},
			{
				name:         fmt.Sprintf("insufficient valid converted fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 10)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   isCheckTx != 1,
			},
			{
				name:         fmt.Sprintf("invalid fee denom - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin("moooooo", 1000)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   isCheckTx != 1, //should pass on deliverTx, fail on checkTx,
			},
			{
				name:         "min gas price not containing basedenom gets treated as min gas price 0",
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1000)),
				minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1000000)),
				isCheckTx:    isCheckTx == 1,
				expectPass:   true,
			},
		}...)
	}

	for _, tc := range tests {
		// reset pool and accounts for each test
		suite.SetupTest(false)

		// setup uion with 1:1 fee
		suite.PrepareBalancerPoolWithCoins(
			sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
			sdk.NewInt64Coin(uion, 500),
		)

		if tc.minGasPrices == nil {
			tc.minGasPrices = sdk.NewDecCoins()
		}
		if tc.gasRequested == 0 {
			tc.gasRequested = baseGas
		}
		suite.Ctx = suite.Ctx.WithIsCheckTx(tc.isCheckTx).WithMinGasPrices(tc.minGasPrices)

		// TxBuilder components reset for every test case
		txconfig := suite.App.GetTxConfig()
		txBuilder := txconfig.NewTxBuilder()
		priv0, _, addr0 := testdata.KeyTestPubAddr()
		acc1 := suite.App.AccountKeeper.NewAccountWithAddress(suite.Ctx, addr0)
		suite.App.AccountKeeper.SetAccount(suite.Ctx, acc1)
		msgs := []sdk.Msg{testdata.NewTestMsg(addr0)}
		signerData := authsigning.SignerData{
			ChainID:       suite.Ctx.ChainID(),
			AccountNumber: 0,
			Sequence:      0,
		}

		gasLimit := tc.gasRequested

		sigV2, err := clienttx.SignWithPrivKey(
			txconfig.SignModeHandler().DefaultMode(), signerData,
			txBuilder, priv0, txconfig, 0)
		suite.Require().NoError(err, "test: %s", tc.name)
		err = txBuilder.SetSignatures(sigV2)
		suite.Require().NoError(err, "test: %s", tc.name)

		bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, tc.txFee)
		tx := suite.BuildTx(txBuilder, msgs, sigV2, "", tc.txFee, gasLimit)

		mfd := ante.NewMempoolFeeDecorator(*suite.App.TxFeesKeeper)
		dfd := ante.NewDeductFeeDecorator(*suite.App.TxFeesKeeper, suite.App.AccountKeeper, suite.App.BankKeeper, nil)
		antehandlerMFD := sdk.ChainAnteDecorators(mfd, dfd)
		_, err = antehandlerMFD(suite.Ctx, tx, tc.isSimulate)

		if tc.expectPass {
			suite.Require().NoError(err, "test: %s", tc.name)
			// ensure fee was collected
			if !tc.txFee.IsZero() {
				var moduleName string
				//check dym in the fee collector
				if tc.txFee[0].Denom == baseDenom {
					moduleName = types.FeeCollectorName
				} else {
					moduleName = types.ModuleName
				}

				moduleAddr := suite.App.AccountKeeper.GetModuleAddress(moduleName)
				suite.Require().Equal(tc.txFee[0], suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddr, tc.txFee[0].Denom), tc.name)
			} else {
				// ensure no fee was collected
				moduleAddr := suite.App.AccountKeeper.GetModuleAddress(types.ModuleName)
				suite.Require().Equal(sdk.NewCoins(), suite.App.BankKeeper.GetAllBalances(suite.Ctx, moduleAddr), tc.name)
			}
		} else {
			suite.Require().Error(err, "test: %s", tc.name)
		}
	}
}
