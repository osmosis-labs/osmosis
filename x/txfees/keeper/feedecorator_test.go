package keeper_test

import (
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	gomock "github.com/golang/mock/gomock"

	gammtypes "github.com/osmosis-labs/osmosis/v10/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v10/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v10/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/txfees/keeper/mocks"
	"github.com/osmosis-labs/osmosis/v10/x/txfees/types"
)

func (suite *KeeperTestSuite) TestFeeDecorator() {
	suite.SetupTest(false)

	mempoolFeeOpts := types.NewDefaultMempoolFeeOptions()
	mempoolFeeOpts.MinGasPriceForHighGasTx = sdk.MustNewDecFromStr("0.0025")
	baseDenom, _ := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)

	uion := "uion"

	uionPoolId := suite.PrepareUni2PoolWithAssets(
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
		// reset pool and accounts for each test
		suite.SetupTest(false)
		suite.Run(tc.name, func() {
			uionPoolId := suite.PrepareUni2PoolWithAssets(
				sdk.NewInt64Coin(sdk.DefaultBondDenom, 500),
				sdk.NewInt64Coin(uion, 500),
			)
			suite.ExecuteUpgradeFeeTokenProposal(uion, uionPoolId)

			suite.Ctx = suite.Ctx.WithIsCheckTx(tc.isCheckTx).WithMinGasPrices(tc.minGasPrices)
			suite.Ctx = suite.Ctx.WithMinGasPrices(tc.minGasPrices)

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

			simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, tc.txFee)

			tx := suite.BuildTx(txBuilder, msgs, sigV2, "", tc.txFee, gasLimit)

			mfd := keeper.NewMempoolFeeDecorator(*suite.App.TxFeesKeeper, mempoolFeeOpts)
			dfd := keeper.NewDeductFeeDecorator(*suite.App.TxFeesKeeper, *suite.App.AccountKeeper, *suite.App.BankKeeper, nil)
			antehandlerMFD := sdk.ChainAnteDecorators(mfd, dfd)
			_, err := antehandlerMFD(suite.Ctx, tx, false)

			if tc.expectPass {
				if tc.baseDenomGas && !tc.txFee.IsZero() {
					moduleAddr := suite.App.AccountKeeper.GetModuleAddress(types.FeeCollectorName)
					suite.Require().Equal(tc.txFee[0], suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddr, baseDenom), tc.name)
				} else if !tc.txFee.IsZero() {
					moduleAddr := suite.App.AccountKeeper.GetModuleAddress(types.NonNativeFeeCollectorName)
					suite.Require().Equal(tc.txFee[0], suite.App.BankKeeper.GetBalance(suite.Ctx, moduleAddr, tc.txFee[0].Denom), tc.name)
				}
				suite.Require().NoError(err, "test: %s", tc.name)
			} else {
				suite.Require().Error(err, "test: %s", tc.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestIsSufficientFee() {
	suite.SetupTest(false)
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	ctx := suite.Ctx
	txfeesKeeper := suite.App.TxFeesKeeper

	createGaugeFee := incentivestypes.MsgCreateGauge{}.GetRequiredMinBaseFee().RoundInt64()

	testcases := map[string]struct {
		txFee  int64
		gasFee int64
		msg    []sdk.Msg

		minGasFee   int64
		expectError bool
	}{
		"tx fee above gas fee and msg fee - no error": {
			txFee:  createGaugeFee + 1,
			gasFee: createGaugeFee - 1,
			msg:    []sdk.Msg{&incentivestypes.MsgCreateGauge{}},
		},
		"tx fee above gas fee but below msg fee - error": {
			txFee:     createGaugeFee - 1,
			minGasFee: createGaugeFee - 2,
			msg:       []sdk.Msg{&incentivestypes.MsgCreateGauge{}},

			expectError: true,
		},
		"tx fee above msg fee but below gas fee - error": {
			txFee:     createGaugeFee + 2,
			minGasFee: createGaugeFee + 5,
			msg:       []sdk.Msg{&incentivestypes.MsgCreateGauge{}},

			expectError: true,
		},
		"tx fee above gas fee and msg 1 but below msg 2 fee - error": {
			txFee:     createGaugeFee - 1,
			minGasFee: createGaugeFee - 3,
			msg:       []sdk.Msg{&incentivestypes.MsgAddToGauge{}, &incentivestypes.MsgCreateGauge{}},

			expectError: true,
		},
		"tx fee above gas fee and msg with no fee - success": {
			txFee:     createGaugeFee,
			minGasFee: createGaugeFee - 3,
			msg:       []sdk.Msg{&gammtypes.MsgExitPool{}},
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			minBaseGasPrice := sdk.NewDec(tc.minGasFee)

			txMock := mocks.NewMockFeeTx(ctrl)

			txMock.EXPECT().GetFee().Return(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(tc.txFee)))).AnyTimes()
			txMock.EXPECT().GetGas().Return(uint64(1)).AnyTimes() // minBaseGasPrice = 1 * tc.MinGasFee, for easy comparison with msg fee.
			txMock.EXPECT().GetMsgs().Return([]sdk.Msg{
				&incentivestypes.MsgCreateGauge{},
			}).AnyTimes()

			err := txfeesKeeper.IsSufficientFee(ctx, minBaseGasPrice, txMock)

			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}
