package keeper_test

import (
	"fmt"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/osmosis-labs/osmosis/v15/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/txfees/types"
)

func (suite *KeeperTestSuite) TestFeeDecorator() {
	suite.SetupTest(false)

	mempoolFeeOpts := types.NewDefaultMempoolFeeOptions()
	mempoolFeeOpts.MinGasPriceForHighGasTx = sdk.MustNewDecFromStr("0.0025")
	baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
	baseGas := uint64(10000)
	consensusMinFeeAmt := int64(25)
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
	succesType := []string{"does", "doesn't"}
	for isCheckTx := 0; isCheckTx <= 1; isCheckTx++ {
		tests = append(tests, []testcase{
			{
				name:       fmt.Sprintf("no min gas price - %s. Fails w/ consensus minimum", txType[isCheckTx]),
				txFee:      sdk.NewCoins(),
				isCheckTx:  isCheckTx == 1,
				expectPass: false,
			},
			{
				name:       fmt.Sprintf("LT Consensus min gas price - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(sdk.NewInt64Coin(baseDenom, consensusMinFeeAmt-1)),
				isCheckTx:  isCheckTx == 1,
				expectPass: false,
			},
			{
				name:       fmt.Sprintf("Consensus min gas price - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(sdk.NewInt64Coin(baseDenom, consensusMinFeeAmt)),
				isCheckTx:  isCheckTx == 1,
				expectPass: true,
			},
			{
				name:       fmt.Sprintf("multiple fee coins - %s", txType[isCheckTx]),
				txFee:      sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1), sdk.NewInt64Coin(uion, 1)),
				isCheckTx:  isCheckTx == 1,
				expectPass: false,
			},
			{
				name:         fmt.Sprintf("works with valid basedenom fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, 1000)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   true,
			},
			{
				name:         fmt.Sprintf("works with valid converted fee - %s", txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1000)),
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   true,
			},
			{
				name:         fmt.Sprintf("%s work with insufficient mempool fee in %s", succesType[isCheckTx], txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(baseDenom, consensusMinFeeAmt)), // consensus minimum
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   isCheckTx != 1,
			},
			{
				name:         fmt.Sprintf("%s work with insufficient converted mempool fee in %s", succesType[isCheckTx], txType[isCheckTx]),
				txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 25)), // consensus minimum
				minGasPrices: point1BaseDenomMinGasPrices,
				isCheckTx:    isCheckTx == 1,
				expectPass:   isCheckTx != 1,
			},
			{
				name:       "invalid fee denom",
				txFee:      sdk.NewCoins(sdk.NewInt64Coin("moooooo", 1000)),
				isCheckTx:  isCheckTx == 1,
				expectPass: false,
			},
		}...)
	}

	custTests := []testcase{
		{
			name:         "min gas price not containing basedenom gets treated as min gas price 0",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1000000)),
			isCheckTx:    true,
			expectPass:   true,
		},
		{
			name:         "tx with gas wanted more than allowed should not pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 100000000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.MaxGasWantedPerTx + 1,
			isCheckTx:    true,
			expectPass:   false,
		},
		{
			name:         "tx with high gas and not enough fee should no pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 1)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.HighGasTxThreshold,
			isCheckTx:    true,
			expectPass:   false,
		},
		{
			name:         "tx with high gas and enough fee should pass",
			txFee:        sdk.NewCoins(sdk.NewInt64Coin(uion, 10*1000)),
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.HighGasTxThreshold,
			isCheckTx:    true,
			expectPass:   true,
		},
		{
			name:         "simulate 0 fee passes",
			txFee:        sdk.Coins{},
			minGasPrices: sdk.NewDecCoins(sdk.NewInt64DecCoin(uion, 1)),
			gasRequested: mempoolFeeOpts.HighGasTxThreshold,
			isCheckTx:    true,
			isSimulate:   true,
			expectPass:   true,
		},
	}
	tests = append(tests, custTests...)

	for _, tc := range tests {
		// reset pool and accounts for each test
		suite.SetupTest(false)
		suite.Run(tc.name, func() {
			// setup uion with 1:1 fee
			uionPoolId := suite.PrepareBalancerPoolWithCoins(
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
				sdk.NewInt64Coin(uion, 500),
			)
			err := suite.ExecuteUpgradeFeeTokenProposal(uion, uionPoolId)
			suite.Require().NoError(err)

			if tc.minGasPrices == nil {
				tc.minGasPrices = sdk.NewDecCoins()
			}
			if tc.gasRequested == 0 {
				tc.gasRequested = baseGas
			}
			suite.Ctx = suite.Ctx.WithIsCheckTx(tc.isCheckTx).WithMinGasPrices(tc.minGasPrices)

			// TODO: Cleanup this code.
			// TxBuilder components reset for every test case
			txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
			priv0, _, addr0 := testdata.KeyTestPubAddr()
			acc1 := suite.App.AccountKeeper.NewAccountWithAddress(suite.Ctx, addr0)
			suite.App.AccountKeeper.SetAccount(suite.Ctx, acc1)
			msgs := []sdk.Msg{testdata.NewTestMsg(addr0)}
			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
			signerData := authsigning.SignerData{
				ChainID:       suite.Ctx.ChainID(),
				AccountNumber: accNums[0],
				Sequence:      accSeqs[0],
			}

			gasLimit := tc.gasRequested
			sigV2, _ := clienttx.SignWithPrivKey(
				1,
				signerData,
				txBuilder,
				privs[0],
				suite.clientCtx.TxConfig,
				accSeqs[0],
			)

			err = simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, tc.txFee)
			suite.Require().NoError(err)

			tx := suite.BuildTx(txBuilder, msgs, sigV2, "", tc.txFee, gasLimit)

			mfd := keeper.NewMempoolFeeDecorator(*suite.App.TxFeesKeeper, mempoolFeeOpts)
			dfd := keeper.NewDeductFeeDecorator(*suite.App.TxFeesKeeper, *suite.App.AccountKeeper, *suite.App.BankKeeper, nil)
			antehandlerMFD := sdk.ChainAnteDecorators(mfd, dfd)
			_, err = antehandlerMFD(suite.Ctx, tx, tc.isSimulate)

			if tc.expectPass {
				// ensure fee was collected
				if !tc.txFee.IsZero() {
					moduleName := types.FeeCollectorName
					if tc.txFee[0].Denom != baseDenom {
						moduleName = types.NonNativeFeeCollectorName
					}
					moduleAddr := suite.App.AccountKeeper.GetModuleAddress(moduleName)
					suite.Require().Equal(tc.txFee[0], suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddr, tc.txFee[0].Denom), tc.name)
				}
				suite.Require().NoError(err, "test: %s", tc.name)
			} else {
				suite.Require().Error(err, "test: %s", tc.name)
			}
		})
	}
}
