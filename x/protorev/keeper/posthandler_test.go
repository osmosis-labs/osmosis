package keeper_test

import (
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v14/x/protorev/types"
)

func (suite *KeeperTestSuite) TestAnteHandle() {

	type param struct {
		msgs                []sdk.Msg
		txFee               sdk.Coins
		minGasPrices        sdk.DecCoins
		gasLimit            uint64
		isCheckTx           bool
		baseDenomGas        bool
		expectedNumOfTrades sdk.Int
		expectedProfits     []*sdk.Coin
		expectedRouteCount  uint64
	}

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	priv0, _, addr0 := testdata.KeyTestPubAddr()
	acc1 := suite.App.AccountKeeper.NewAccountWithAddress(suite.Ctx, addr0)
	suite.App.AccountKeeper.SetAccount(suite.Ctx, acc1)

	// Keep testing order consistent to make adding tests easier
	// Add all tests that are not expected to execute a trade first
	// Then track number of trades and profits for the rest of the tests
	tests := []struct {
		name       string
		params     param
		expectPass bool
	}{
		{
			name: "Random Msg - Expect Nothing to Happen",
			params: param{
				msgs:                []sdk.Msg{testdata.NewTestMsg(addr0)},
				txFee:               sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.ZeroInt(),
				expectedProfits:     []*sdk.Coin{},
				expectedRouteCount:  0,
			},
			expectPass: true,
		},
		{
			name: "No Arb",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        12,
								TokenOutDenom: "akash",
							},
						},
						TokenIn:           sdk.NewCoin("juno", sdk.NewInt(10)),
						TokenOutMinAmount: sdk.NewInt(1),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.ZeroInt(),
				expectedProfits:     []*sdk.Coin{},
				expectedRouteCount:  4,
			},
			expectPass: true,
		},
		{
			name: "Mainnet Arb (Block: 5905150) - Highest Liquidity Pool Build",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        23,
								TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
							},
						},
						TokenIn:           sdk.NewCoin("ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(10000),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.OneInt(),
				expectedProfits: []*sdk.Coin{
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(24848),
					},
				},
				expectedRouteCount: 6,
			},
			expectPass: true,
		},
		{
			name: "Mainnet Arb Route - Multi Asset, Same Weights (Block: 6906570) - Hot Route Build - Atom Arb",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        33,
								TokenOutDenom: "ibc/A0CC0CF735BFB30E730C70019D4218A1244FF383503FF7579C9201AB93CA9293",
							},
						},
						TokenIn:           sdk.NewCoin(types.AtomDenomination, sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(10000),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.NewInt(2),
				expectedProfits: []*sdk.Coin{
					{
						Denom:  types.AtomDenomination,
						Amount: sdk.NewInt(5826),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(24848),
					},
				},
				expectedRouteCount: 8,
			},
			expectPass: true,
		},
		{
			name: "Stableswap Test Arb Route - Hot Route Build",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        29,
								TokenOutDenom: types.OsmosisDenomination,
							},
						},
						TokenIn:           sdk.NewCoin("usdc", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(100),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.NewInt(3),
				expectedProfits: []*sdk.Coin{
					{
						Denom:  types.AtomDenomination,
						Amount: sdk.NewInt(5826),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(56609900),
					},
				},
				expectedRouteCount: 13,
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Ctx = suite.Ctx.WithIsCheckTx(tc.params.isCheckTx)
			suite.Ctx = suite.Ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
			suite.Ctx = suite.Ctx.WithMinGasPrices(tc.params.minGasPrices)
			msgs := tc.params.msgs

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
			signerData := authsigning.SignerData{
				ChainID:       suite.Ctx.ChainID(),
				AccountNumber: accNums[0],
				Sequence:      accSeqs[0],
			}
			gasLimit := tc.params.gasLimit
			sigV2, _ := clienttx.SignWithPrivKey(
				1,
				signerData,
				txBuilder,
				privs[0],
				suite.clientCtx.TxConfig,
				accSeqs[0],
			)
			simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, tc.params.txFee)
			tx := suite.BuildTx(txBuilder, msgs, sigV2, "", tc.params.txFee, gasLimit)
			protoRevDecorator := keeper.NewProtoRevDecorator(*suite.App.ProtoRevKeeper)
			posthandlerProtoRev := sdk.ChainAnteDecorators(protoRevDecorator)

			// Added so we can check the gas consumed during the posthandler
			suite.Ctx = suite.Ctx.WithGasMeter(sdk.NewGasMeter(tc.params.gasLimit))
			halfGas := tc.params.gasLimit / 2
			suite.Ctx.GasMeter().ConsumeGas(halfGas, "consume half gas")
			gasBefore := suite.Ctx.GasMeter().GasConsumed()
			gasLimitBefore := suite.Ctx.GasMeter().Limit()

			_, err := posthandlerProtoRev(suite.Ctx, tx, false)

			gasAfter := suite.Ctx.GasMeter().GasConsumed()
			gasLimitAfter := suite.Ctx.GasMeter().Limit()

			if tc.expectPass {
				suite.Require().NoError(err)
				// Check that the gas consumed is the same before and after the posthandler
				suite.Require().Equal(gasBefore, gasAfter)
				// Check that the gas limit is the same before and after the posthandler
				suite.Require().Equal(gasLimitBefore, gasLimitAfter)

				suite.Ctx = suite.Ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

				// Check that the number of trades is correct
				numOfTrades, _ := suite.App.ProtoRevKeeper.GetNumberOfTrades(suite.Ctx)
				suite.Require().Equal(tc.params.expectedNumOfTrades, numOfTrades)

				// Check that the profits are correct
				profits := suite.App.ProtoRevKeeper.GetAllProfits(suite.Ctx)
				suite.Require().Equal(tc.params.expectedProfits, profits)

				// Check the current route count
				routeCount, err := suite.App.ProtoRevKeeper.GetRouteCountForBlock(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.params.expectedRouteCount, routeCount)

			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestExtractSwappedPools() {
	type param struct {
		msgs                 []sdk.Msg
		txFee                sdk.Coins
		minGasPrices         sdk.DecCoins
		gasLimit             uint64
		isCheckTx            bool
		baseDenomGas         bool
		expectedNumOfPools   int
		expectedSwappedPools []keeper.SwapToBackrun
	}

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	priv0, _, addr0 := testdata.KeyTestPubAddr()
	acc1 := suite.App.AccountKeeper.NewAccountWithAddress(suite.Ctx, addr0)
	suite.App.AccountKeeper.SetAccount(suite.Ctx, acc1)

	tests := []struct {
		name       string
		params     param
		expectPass bool
	}{
		{
			name: "Single Swap",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        28,
								TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
							},
						},
						TokenIn:           sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(10000),
					},
				},
				txFee:              sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000))),
				minGasPrices:       sdk.NewDecCoins(),
				gasLimit:           500000,
				isCheckTx:          false,
				baseDenomGas:       true,
				expectedNumOfPools: 1,
				expectedSwappedPools: []keeper.SwapToBackrun{
					{
						PoolId:        28,
						TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
						TokenInDenom:  "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
					},
				},
			},
			expectPass: true,
		},
		{
			name: "Two Swaps",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        28,
								TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
							},
						},
						TokenIn:           sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(10000),
					},
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        22,
								TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
							},
						},
						TokenIn:           sdk.NewCoin("uosmo", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(10000),
					},
				},
				txFee:              sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000))),
				minGasPrices:       sdk.NewDecCoins(),
				gasLimit:           500000,
				isCheckTx:          false,
				baseDenomGas:       true,
				expectedNumOfPools: 2,
				expectedSwappedPools: []keeper.SwapToBackrun{
					{
						PoolId:        28,
						TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
						TokenInDenom:  "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
					},
					{
						PoolId:        22,
						TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
						TokenInDenom:  "uosmo",
					},
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {

			suite.Ctx = suite.Ctx.WithIsCheckTx(tc.params.isCheckTx)
			suite.Ctx = suite.Ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
			suite.Ctx = suite.Ctx.WithMinGasPrices(tc.params.minGasPrices)
			msgs := tc.params.msgs

			privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
			signerData := authsigning.SignerData{
				ChainID:       suite.Ctx.ChainID(),
				AccountNumber: accNums[0],
				Sequence:      accSeqs[0],
			}
			gasLimit := tc.params.gasLimit
			sigV2, _ := clienttx.SignWithPrivKey(
				1,
				signerData,
				txBuilder,
				privs[0],
				suite.clientCtx.TxConfig,
				accSeqs[0],
			)
			simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, tc.params.txFee)

			// Can't use test suite BuildTx because it doesn't allow for multiple msgs
			txBuilder.SetMsgs(msgs...)
			txBuilder.SetSignatures(sigV2)
			txBuilder.SetMemo("")
			txBuilder.SetFeeAmount(tc.params.txFee)
			txBuilder.SetGasLimit(gasLimit)

			tx := txBuilder.GetTx()

			swappedPools := keeper.ExtractSwappedPools(tx)
			if tc.expectPass {
				suite.Require().Equal(tc.params.expectedNumOfPools, len(swappedPools))
				suite.Require().Equal(tc.params.expectedSwappedPools, swappedPools)
			}
		})
	}
}
