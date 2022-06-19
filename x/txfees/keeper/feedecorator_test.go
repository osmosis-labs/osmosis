package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	appParams "github.com/osmosis-labs/osmosis/v7/app/params"
	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"
	txKeeper "github.com/osmosis-labs/osmosis/v7/x/txfees/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/txfees/types"
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
	}
}

func (suite *KeeperTestSuite) TestIsSufficientFee() {
	suite.SetupTest(false)

	createSybil := func(after func(s txKeeper.Sybil) txKeeper.Sybil) txKeeper.Sybil {
		properSybil := txKeeper.Sybil{
			GasPrice: sdk.MustNewDecFromStr("1"),
			FeesPaid: sdk.NewCoin("test", sdk.NewInt(1)),
		}

		return after(properSybil)
	}

	sybilStruct := createSybil(func(s txKeeper.Sybil) txKeeper.Sybil {
		// do nothing
		return s
	})

	suite.Require().Equal(sdk.MustNewDecFromStr("1"), sybilStruct.GasPrice)
	suite.Require().Equal(sdk.NewCoin("test", sdk.NewInt(1)), sybilStruct.FeesPaid)

	baseDenom, err := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
	suite.Require().NoError(err, "base denom error for test is sufficient fee")

	tests := []struct {
		name         string
		sybil        txKeeper.Sybil
		gasRequested uint64
		feeCoin      sdk.Coin
		expectPass   bool
	}{
		{
			name: "exactly sufficient fee without feesPaid",
			sybil: createSybil(func(sybil txKeeper.Sybil) txKeeper.Sybil {
				sybil.FeesPaid = sdk.NewCoin(baseDenom, sdk.ZeroInt())

				return sybil
			}),
			gasRequested: 1,
			feeCoin:      sdk.NewCoin(baseDenom, sdk.NewInt(1)),
			expectPass:   true,
		},
		{
			name: "excess fees",
			sybil: createSybil(func(sybil txKeeper.Sybil) txKeeper.Sybil {
				sybil.FeesPaid = sdk.NewCoin(baseDenom, sdk.ZeroInt())

				return sybil
			}),
			gasRequested: 1,
			feeCoin:      sdk.NewCoin(baseDenom, sdk.NewInt(1000)),
			expectPass:   true,
		},
		{
			name: "0 amount fee coin but fees paid cover gas cost",
			sybil: createSybil(func(sybil txKeeper.Sybil) txKeeper.Sybil {
				sybil.FeesPaid = sdk.NewCoin(baseDenom, sdk.NewInt(100))
				sybil.GasPrice = sdk.MustNewDecFromStr("10")

				return sybil
			}),
			gasRequested: 1,
			feeCoin:      sdk.NewCoin(baseDenom, sdk.ZeroInt()),
			expectPass:   true,
		},
		{
			name: "null fee coin, sufficient swap fees",
			sybil: createSybil(func(sybil txKeeper.Sybil) txKeeper.Sybil {
				sybil.FeesPaid = sdk.NewCoin(baseDenom, sdk.NewInt(10))
				sybil.GasPrice = sdk.MustNewDecFromStr("1")

				return sybil
			}),
			gasRequested: 10,
			feeCoin:      sdk.Coin{},
			expectPass:   false,
		},
		{
			name: "insufficient fee coin and feees paid",
			sybil: createSybil(func(sybil txKeeper.Sybil) txKeeper.Sybil {
				sybil.FeesPaid = sdk.NewCoin(baseDenom, sdk.NewInt(5))
				sybil.GasPrice = sdk.MustNewDecFromStr("1")

				return sybil
			}),
			gasRequested: 10,
			feeCoin:      sdk.NewCoin(baseDenom, sdk.NewInt(4)),
			expectPass:   false,
		},
		{
			name: "fee coin not a fee token",
			sybil: createSybil(func(sybil txKeeper.Sybil) txKeeper.Sybil {
				sybil.FeesPaid = sdk.NewCoin(baseDenom, sdk.NewInt(1000))
				sybil.GasPrice = sdk.MustNewDecFromStr("10")

				return sybil
			}),
			gasRequested: 10,
			feeCoin:      sdk.NewCoin("test", sdk.NewInt(10)),
			expectPass:   false,
		},
		{
			name: "sybil fees paid is not a fee token",
			sybil: createSybil(func(sybil txKeeper.Sybil) txKeeper.Sybil {
				sybil.FeesPaid = sdk.NewCoin("test", sdk.NewInt(10))
				sybil.GasPrice = sdk.MustNewDecFromStr("10")

				return sybil
			}),
			gasRequested: 1,
			feeCoin:      sdk.NewCoin(baseDenom, sdk.NewInt(1)),
			expectPass:   false,
		},
	}

	for _, test := range tests {
		err := suite.App.TxFeesKeeper.IsSufficientFee(suite.Ctx, test.sybil, test.gasRequested, test.feeCoin)

		if test.expectPass {
			suite.Require().NoError(err, "test: %s", test.name)
		} else {
			suite.Require().Error(err, "test: %s", test.name)
		}
	}
}

func (suite *KeeperTestSuite) TestGetMinBaseGasPriceForTxExactAmountIn() {
	appParams.SetAddressPrefixes()
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	//	invalidAddr := sdk.AccAddress("invalid")

	createMsg := func(after func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
		properMsg := gammtypes.MsgSwapExactAmountIn{
			Sender: addr1,
			Routes: []gammtypes.SwapAmountInRoute{{
				PoolId:        1,
				TokenOutDenom: "test1",
			}, {
				PoolId:        2,
				TokenOutDenom: "test2",
			}},
			TokenIn:           sdk.NewCoin("test", sdk.NewInt(100)),
			TokenOutMinAmount: sdk.NewInt(200),
		}

		return after(properMsg)
	}

	msg := createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
		// Do nothing
		return msg
	})

	suite.Require().Equal(msg.Route(), gammtypes.RouterKey)
	suite.Require().Equal(msg.Type(), "swap_exact_amount_in")
	signersSybilFee := msg.GetSigners()
	suite.Require().Equal(len(signersSybilFee), 1)
	suite.Require().Equal(signersSybilFee[0].String(), addr1)

	createSybil := func(after func(s txKeeper.Sybil) txKeeper.Sybil) txKeeper.Sybil {
		properSybil := txKeeper.Sybil{
			GasPrice: sdk.MustNewDecFromStr("1"),
			FeesPaid: sdk.NewCoin("test", sdk.NewInt(1)),
		}

		return after(properSybil)
	}

	sybilStruct := createSybil(func(s txKeeper.Sybil) txKeeper.Sybil {
		// do nothing
		return s
	})

	suite.Require().Equal(sdk.MustNewDecFromStr("1"), sybilStruct.GasPrice)
	suite.Require().Equal(sdk.NewCoin("test", sdk.NewInt(1)), sybilStruct.FeesPaid)

	baseDenom, err := suite.App.TxFeesKeeper.GetBaseDenom(suite.Ctx)
	suite.Require().NoError(err, "base denom %v error %v for test get min gas price", baseDenom, err)
	mempoolFeeOpts := types.NewDefaultMempoolFeeOptions()
	mempoolFeeOpts.MinGasPriceForHighGasTx = sdk.MustNewDecFromStr("0.0025")

	tests := []struct {
		name           string
		msg            gammtypes.MsgSwapExactAmountIn
		txFee          sdk.Coin
		expectGasPrice sdk.Dec
		expectFeesPaid sdk.Coin
		gasRequested   uint64
		expectPass     bool
	}{
		{
			name: "msg does not need sybil fees",
			msg: createMsg(func(msg gammtypes.MsgSwapExactAmountIn) gammtypes.MsgSwapExactAmountIn {
				// Do nothing
				return msg
			}),
			txFee:          sdk.NewCoin(baseDenom, sdk.NewInt(10000)),
			expectGasPrice: suite.Ctx.MinGasPrices().AmountOf(baseDenom),
			expectFeesPaid: sdk.NewCoin(baseDenom, sdk.ZeroInt()),
			expectPass:     true,
		},
	}

	for _, test := range tests {
		suite.SetupTest(false)

		// TxBuilder components reset for every test case
		txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
		priv0, _, addr0 := testdata.KeyTestPubAddr()
		acc1 := suite.App.AccountKeeper.NewAccountWithAddress(suite.Ctx, addr0)
		suite.App.AccountKeeper.SetAccount(suite.Ctx, acc1)
		msgs := []sdk.Msg{&test.msg}
		privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
		signerData := authsigning.SignerData{
			ChainID:       suite.Ctx.ChainID(),
			AccountNumber: accNums[0],
			Sequence:      accSeqs[0],
		}

		gasLimit := test.gasRequested
		sigV2, _ := clienttx.SignWithPrivKey(
			1,
			signerData,
			txBuilder,
			privs[0],
			suite.clientCtx.TxConfig,
			accSeqs[0],
		)

		// build transaction with swap extern amount out msg
		tx := suite.BuildTx(txBuilder, msgs, sigV2, "", sdk.NewCoins(test.txFee), gasLimit)

		// create mempool fee decorator
		mfd := keeper.NewMempoolFeeDecorator(*suite.App.TxFeesKeeper, mempoolFeeOpts)

		// get sybil fee struct
		sybil, err := mfd.GetMinBaseGasPriceForTx(suite.Ctx, baseDenom, tx)

		// check if sybil fee struct is as expected
		if test.expectPass {
			suite.Require().Equal(sybil.GasPrice, test.expectGasPrice)
			suite.Require().Equal(sybil.FeesPaid, test.expectFeesPaid)
			suite.Require().NoError(err, "test: %s", test.name)
		} else {
			suite.Require().Error(err, "test: %s", test.name)
		}
	}
}
