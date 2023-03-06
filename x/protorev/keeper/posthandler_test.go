package keeper_test

import (
	"strings"
	"testing"

	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

// BenchmarkBalancerSwapHighestLiquidityArb benchmarks a balancer swap that creates a single three hop arbitrage
// route with only balancer pools created by the highest liquidity method.
func BenchmarkBalancerSwapHighestLiquidityArb(b *testing.B) {
	msgs := []sdk.Msg{
		&poolmanagertypes.MsgSwapExactAmountIn{
			Routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        23,
					TokenOutDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
				},
			},
			TokenIn:           sdk.NewCoin("ibc/0EF15DF2F02480ADE0BB6E85D9EBB5DAEA2836D3860E9F97F9AADE4F57A31AA0", sdk.NewInt(10000)),
			TokenOutMinAmount: sdk.NewInt(10000),
		},
	}
	benchmarkWrapper(b, msgs, 1)
}

// BenchmarkStableSwapHotRouteArb benchmarks a balancer swap that gets back run by a single three hop arbitrage
// with a single stable pool and 2 balancer pools created via the hot routes method.
func BenchmarkStableSwapHotRouteArb(b *testing.B) {
	msgs := []sdk.Msg{
		&poolmanagertypes.MsgSwapExactAmountIn{
			Routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        29,
					TokenOutDenom: types.OsmosisDenomination,
				},
			},
			TokenIn:           sdk.NewCoin("usdc", sdk.NewInt(10000)),
			TokenOutMinAmount: sdk.NewInt(100),
		},
	}
	benchmarkWrapper(b, msgs, 1)
}

// BenchmarkFourHopArb benchmarks a balancer swap that gets back run by a single four hop arbitrage route
// created via the hot routes method.
func BenchmarkFourHopHotRouteArb(b *testing.B) {
	msgs := []sdk.Msg{
		&poolmanagertypes.MsgSwapExactAmountIn{
			Routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        37,
					TokenOutDenom: "test/2",
				},
			},
			TokenIn:           sdk.NewCoin("Atom", sdk.NewInt(10000)),
			TokenOutMinAmount: sdk.NewInt(100),
		},
	}
	benchmarkWrapper(b, msgs, 1)
}

func (suite *KeeperTestSuite) TestAnteHandle() {
	type param struct {
		msgs                []sdk.Msg
		txFee               sdk.Coins
		minGasPrices        sdk.DecCoins
		gasLimit            uint64
		isCheckTx           bool
		baseDenomGas        bool
		expectedNumOfTrades sdk.Int
		expectedProfits     []sdk.Coin
		expectedPoolPoints  uint64
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
				expectedProfits:     []sdk.Coin{},
				expectedPoolPoints:  0,
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
				expectedProfits:     []sdk.Coin{},
				expectedPoolPoints:  0,
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
				expectedProfits: []sdk.Coin{
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(24848),
					},
				},
				expectedPoolPoints: 6,
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
						TokenIn:           sdk.NewCoin("Atom", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(10000),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.NewInt(2),
				expectedProfits: []sdk.Coin{
					{
						Denom:  "Atom",
						Amount: sdk.NewInt(5826),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(24848),
					},
				},
				expectedPoolPoints: 12,
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
				expectedProfits: []sdk.Coin{
					{
						Denom:  "Atom",
						Amount: sdk.NewInt(5826),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(56609900),
					},
				},
				expectedPoolPoints: 21,
			},
			expectPass: true,
		},
		{
			name: "Four Pool Arb Route - Hot Route Build",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        37,
								TokenOutDenom: "test/2",
							},
						},
						TokenIn:           sdk.NewCoin("Atom", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(100),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.NewInt(4),
				expectedProfits: []sdk.Coin{
					{
						Denom:  "Atom",
						Amount: sdk.NewInt(15_767_231),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(56_609_900),
					},
				},
				expectedPoolPoints: 29,
			},
			expectPass: true,
		},
		{
			name: "Two Pool Arb Route - Hot Route Build",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        38,
								TokenOutDenom: "test/3",
							},
						},
						TokenIn:           sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(100),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.NewInt(5),
				expectedProfits: []sdk.Coin{
					{
						Denom:  "Atom",
						Amount: sdk.NewInt(15_767_231),
					},
					{
						Denom:  "test/3",
						Amount: sdk.NewInt(218_149_058),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(56_609_900),
					},
				},
				expectedPoolPoints: 33,
			},
			expectPass: true,
		},
		{ // This test the tx pool points limit caps the number of iterations
			name: "Doomsday Test - Stableswap - Tx Pool Points Limit",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        41,
								TokenOutDenom: "usdc",
							},
						},
						TokenIn:           sdk.NewCoin("busd", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(100),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.NewInt(5),
				expectedProfits: []sdk.Coin{
					{
						Denom:  "Atom",
						Amount: sdk.NewInt(15_767_231),
					},
					{
						Denom:  "test/3",
						Amount: sdk.NewInt(218_149_058),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(56_609_900),
					},
				},
				expectedPoolPoints: 33,
			},
			expectPass: true,
		},
		{ // This test the block pool points limit caps the number of iterations within a tx
			name: "Doomsday Test - Stableswap - Block Pool Points Limit - Within a tx",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        41,
								TokenOutDenom: "usdc",
							},
						},
						TokenIn:           sdk.NewCoin("busd", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(100),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.NewInt(5),
				expectedProfits: []sdk.Coin{
					{
						Denom:  "Atom",
						Amount: sdk.NewInt(15_767_231),
					},
					{
						Denom:  "test/3",
						Amount: sdk.NewInt(218_149_058),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(56_609_900),
					},
				},
				expectedPoolPoints: 33,
			},
			expectPass: true,
		},
		{ // This test the block pool points limit caps the number of txs processed if already reached the limit
			name: "Doomsday Test - Stableswap - Block Pool Points Limit Already Reached - New tx",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        41,
								TokenOutDenom: "usdc",
							},
						},
						TokenIn:           sdk.NewCoin("busd", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(100),
					},
				},
				txFee:               sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))),
				minGasPrices:        sdk.NewDecCoins(),
				gasLimit:            500000,
				isCheckTx:           false,
				baseDenomGas:        true,
				expectedNumOfTrades: sdk.NewInt(5),
				expectedProfits: []sdk.Coin{
					{
						Denom:  "Atom",
						Amount: sdk.NewInt(15_767_231),
					},
					{
						Denom:  "test/3",
						Amount: sdk.NewInt(218_149_058),
					},
					{
						Denom:  types.OsmosisDenomination,
						Amount: sdk.NewInt(56_609_900),
					},
				},
				expectedPoolPoints: 33,
			},
			expectPass: true,
		},
	}

	// Ensure that the max points per tx is enough for the test suite
	suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, 18)
	suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, 100)
	suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, types.PoolWeights{StableWeight: 5, BalancerWeight: 2, ConcentratedWeight: 2})

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.Ctx = suite.Ctx.WithIsCheckTx(tc.params.isCheckTx)
			suite.Ctx = suite.Ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
			suite.Ctx = suite.Ctx.WithMinGasPrices(tc.params.minGasPrices)

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

			var tx authsigning.Tx
			var msgs []sdk.Msg

			// Lower the max points per tx and block if the test cases are doomsday testing
			if strings.Contains(tc.name, "Tx Pool Points Limit") {
				suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, 5)
			} else if strings.Contains(tc.name, "Block Pool Points Limit - Within a tx") {
				suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, 35)
			} else if strings.Contains(tc.name, "Block Pool Points Limit Already Reached") {
				suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, 33)
			}

			if strings.Contains(tc.name, "Doomsday") {
				for i := 0; i < 100; i++ {
					msgs = append(msgs, tc.params.msgs...)
				}

				txBuilder.SetMsgs(msgs...)
				txBuilder.SetSignatures(sigV2)
				txBuilder.SetMemo("")
				txBuilder.SetFeeAmount(tc.params.txFee)
				txBuilder.SetGasLimit(gasLimit)
				tx = txBuilder.GetTx()
			} else {
				msgs = tc.params.msgs
				tx = suite.BuildTx(txBuilder, msgs, sigV2, "", tc.params.txFee, gasLimit)
			}

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

				// Check the current pool point count
				pointCount, err := suite.App.ProtoRevKeeper.GetPointCountForBlock(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.params.expectedPoolPoints, pointCount)

			} else {
				suite.Require().Error(err)
			}

			// Reset the max points per tx and block
			if strings.Contains(tc.name, "Tx Pool Points Limit") {
				suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, 18)
			} else if strings.Contains(tc.name, "Block Pool Points Limit") {
				suite.App.ProtoRevKeeper.SetMaxPointsPerBlock(suite.Ctx, 100)
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
		{
			name: "Single Swap Amount Out Test",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountOut{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountOutRoute{
							{
								PoolId:       28,
								TokenInDenom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
							},
						},
						TokenOut:         sdk.NewCoin("ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858", sdk.NewInt(10000)),
						TokenInMaxAmount: sdk.NewInt(10000),
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
						TokenOutDenom: "ibc/D189335C6E4A68B513C10AB227BF1C1D38C746766278BA3EEB4FB14124F1D858",
						TokenInDenom:  "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
					},
				},
			},
			expectPass: true,
		},
		{
			name: "Single Swap with multiple hops (swapOut)",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountOut{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountOutRoute{
							{
								PoolId:       28,
								TokenInDenom: "atom",
							},
							{
								PoolId:       30,
								TokenInDenom: "weth",
							},
							{
								PoolId:       35,
								TokenInDenom: "bitcoin",
							},
						},
						TokenOut:         sdk.NewCoin("akash", sdk.NewInt(10000)),
						TokenInMaxAmount: sdk.NewInt(10000),
					},
				},
				txFee:              sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000))),
				minGasPrices:       sdk.NewDecCoins(),
				gasLimit:           500000,
				isCheckTx:          false,
				baseDenomGas:       true,
				expectedNumOfPools: 3,
				expectedSwappedPools: []keeper.SwapToBackrun{
					{
						PoolId:        35,
						TokenOutDenom: "akash",
						TokenInDenom:  "bitcoin",
					},
					{
						PoolId:        30,
						TokenOutDenom: "bitcoin",
						TokenInDenom:  "weth",
					},
					{
						PoolId:        28,
						TokenOutDenom: "weth",
						TokenInDenom:  "atom",
					},
				},
			},
			expectPass: true,
		},
		{
			name: "Single Swap with multiple hops (swapIn)",
			params: param{
				msgs: []sdk.Msg{
					&poolmanagertypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        28,
								TokenOutDenom: "atom",
							},
							{
								PoolId:        30,
								TokenOutDenom: "weth",
							},
							{
								PoolId:        35,
								TokenOutDenom: "bitcoin",
							},
							{
								PoolId:        36,
								TokenOutDenom: "juno",
							},
						},
						TokenIn:           sdk.NewCoin("akash", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(1),
					},
				},
				txFee:              sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000))),
				minGasPrices:       sdk.NewDecCoins(),
				gasLimit:           500000,
				isCheckTx:          false,
				baseDenomGas:       true,
				expectedNumOfPools: 4,
				expectedSwappedPools: []keeper.SwapToBackrun{
					{
						PoolId:        28,
						TokenOutDenom: "atom",
						TokenInDenom:  "akash",
					},
					{
						PoolId:        30,
						TokenOutDenom: "weth",
						TokenInDenom:  "atom",
					},
					{
						PoolId:        35,
						TokenOutDenom: "bitcoin",
						TokenInDenom:  "weth",
					},
					{
						PoolId:        36,
						TokenOutDenom: "juno",
						TokenInDenom:  "bitcoin",
					},
				},
			},
			expectPass: true,
		},
		{
			name: "Single Swap with multiple hops (gamm msg swapOut)",
			params: param{
				msgs: []sdk.Msg{
					&gammtypes.MsgSwapExactAmountOut{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountOutRoute{
							{
								PoolId:       28,
								TokenInDenom: "atom",
							},
							{
								PoolId:       30,
								TokenInDenom: "weth",
							},
							{
								PoolId:       35,
								TokenInDenom: "bitcoin",
							},
						},
						TokenOut:         sdk.NewCoin("akash", sdk.NewInt(10000)),
						TokenInMaxAmount: sdk.NewInt(10000),
					},
				},
				txFee:              sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000))),
				minGasPrices:       sdk.NewDecCoins(),
				gasLimit:           500000,
				isCheckTx:          false,
				baseDenomGas:       true,
				expectedNumOfPools: 3,
				expectedSwappedPools: []keeper.SwapToBackrun{
					{
						PoolId:        35,
						TokenOutDenom: "akash",
						TokenInDenom:  "bitcoin",
					},
					{
						PoolId:        30,
						TokenOutDenom: "bitcoin",
						TokenInDenom:  "weth",
					},
					{
						PoolId:        28,
						TokenOutDenom: "weth",
						TokenInDenom:  "atom",
					},
				},
			},
			expectPass: true,
		},
		{
			name: "Single Swap with multiple hops (gamm swapIn)",
			params: param{
				msgs: []sdk.Msg{
					&gammtypes.MsgSwapExactAmountIn{
						Sender: addr0.String(),
						Routes: []poolmanagertypes.SwapAmountInRoute{
							{
								PoolId:        28,
								TokenOutDenom: "atom",
							},
							{
								PoolId:        30,
								TokenOutDenom: "weth",
							},
							{
								PoolId:        35,
								TokenOutDenom: "bitcoin",
							},
							{
								PoolId:        36,
								TokenOutDenom: "juno",
							},
						},
						TokenIn:           sdk.NewCoin("akash", sdk.NewInt(10000)),
						TokenOutMinAmount: sdk.NewInt(1),
					},
				},
				txFee:              sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000))),
				minGasPrices:       sdk.NewDecCoins(),
				gasLimit:           500000,
				isCheckTx:          false,
				baseDenomGas:       true,
				expectedNumOfPools: 4,
				expectedSwappedPools: []keeper.SwapToBackrun{
					{
						PoolId:        28,
						TokenOutDenom: "atom",
						TokenInDenom:  "akash",
					},
					{
						PoolId:        30,
						TokenOutDenom: "weth",
						TokenInDenom:  "atom",
					},
					{
						PoolId:        35,
						TokenOutDenom: "bitcoin",
						TokenInDenom:  "weth",
					},
					{
						PoolId:        36,
						TokenOutDenom: "juno",
						TokenInDenom:  "bitcoin",
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

// benchmarkWrapper is a wrapper function for the benchmark tests. It sets up the suite, accepts the
// messages to be sent, and the expected number of trades. It then runs the benchmark and checks the
// number of trades after the post handler is run.
func benchmarkWrapper(b *testing.B, msgs []sdk.Msg, expectedTrades int) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		suite, tx, postHandler := setUpBenchmarkSuite(msgs)

		b.StartTimer()
		postHandler(suite.Ctx, tx, false)
		b.StopTimer()

		numberTrades, err := suite.App.ProtoRevKeeper.GetNumberOfTrades(suite.Ctx)
		if err != nil {
			if expectedTrades != 0 {
				b.Fatal("error getting number of trades")
			}
		}
		if !numberTrades.Equal(sdk.NewInt(int64(expectedTrades))) {
			b.Fatalf("expected %d trades, got %d", expectedTrades, numberTrades)
		}
	}
}

// setUpBenchmarkSuite sets up a app test suite, tx, and post handler for benchmark tests.
// It returns the app configured to the correct state, a valid tx, and the protorev post handler.
func setUpBenchmarkSuite(msgs []sdk.Msg) (*KeeperTestSuite, authsigning.Tx, sdk.AnteHandler) {
	// Create a new test suite
	suite := new(KeeperTestSuite)
	suite.SetT(&testing.T{})
	suite.SetupTest()

	// Set up the app to the correct state to run the test
	suite.Ctx = suite.Ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	suite.App.ProtoRevKeeper.SetMaxPointsPerTx(suite.Ctx, 40)
	suite.App.ProtoRevKeeper.SetPoolWeights(suite.Ctx, types.PoolWeights{StableWeight: 5, BalancerWeight: 2, ConcentratedWeight: 2})

	// Init a new account and fund it with tokens for gas fees
	priv0, _, addr0 := testdata.KeyTestPubAddr()
	acc1 := suite.App.AccountKeeper.NewAccountWithAddress(suite.Ctx, addr0)
	suite.App.AccountKeeper.SetAccount(suite.Ctx, acc1)
	simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, addr0, sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))))

	// Build the tx
	privs, accNums, accSeqs := []cryptotypes.PrivKey{priv0}, []uint64{0}, []uint64{0}
	signerData := authsigning.SignerData{
		ChainID:       suite.Ctx.ChainID(),
		AccountNumber: accNums[0],
		Sequence:      accSeqs[0],
	}
	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()
	sigV2, _ := clienttx.SignWithPrivKey(
		1,
		signerData,
		txBuilder,
		privs[0],
		suite.clientCtx.TxConfig,
		accSeqs[0],
	)
	tx := suite.BuildTx(txBuilder, msgs, sigV2, "", sdk.NewCoins(sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(10000))), 500000)

	// Set up the post handler
	protoRevDecorator := keeper.NewProtoRevDecorator(*suite.App.ProtoRevKeeper)
	posthandlerProtoRev := sdk.ChainAnteDecorators(protoRevDecorator)

	return suite, tx, posthandlerProtoRev
}
