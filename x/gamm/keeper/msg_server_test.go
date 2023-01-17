package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/gamm/keeper"
	balancer "github.com/osmosis-labs/osmosis/v14/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v14/x/poolmanager/types"
)

const (
	doesNotExistDenom = "nodenom"
	// Max positive int64.
	int64Max = int64(^uint64(0) >> 1)
)

// TestSwapExactAmountIn_Events tests that events are correctly emitted
// when calling SwapExactAmountIn.
func (suite *KeeperTestSuite) TestSwapExactAmountIn_Events() {
	const (
		tokenInMinAmount = 1
		tokenIn          = 5
	)

	testcases := map[string]struct {
		routes                []poolmanagertypes.SwapAmountInRoute
		tokenIn               sdk.Coin
		tokenOutMinAmount     sdk.Int
		expectError           bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:            []poolmanagertypes.SwapAmountInRoute{},
			tokenIn:           sdk.NewCoin("foo", sdk.NewInt(tokenIn)),
			tokenOutMinAmount: sdk.NewInt(tokenInMinAmount),
			expectError:       true,
		},
		"one hop": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "bar",
				},
			},
			tokenIn:               sdk.NewCoin("foo", sdk.NewInt(tokenIn)),
			tokenOutMinAmount:     sdk.NewInt(tokenInMinAmount),
			expectedSwapEvents:    1,
			expectedMessageEvents: 3, // 1 gamm + 2 events emitted by other keeper methods.
		},
		"two hops": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "bar",
				},
				{
					PoolId:        2,
					TokenOutDenom: "baz",
				},
			},
			tokenIn:               sdk.NewCoin("foo", sdk.NewInt(tokenIn)),
			tokenOutMinAmount:     sdk.NewInt(tokenInMinAmount),
			expectedSwapEvents:    2,
			expectedMessageEvents: 5, // 1 gamm + 4 events emitted by other keeper methods.
		},
		"invalid - two hops, denom does not exist": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "bar",
				},
				{
					PoolId:        2,
					TokenOutDenom: "baz",
				},
			},
			tokenIn:           sdk.NewCoin(doesNotExistDenom, sdk.NewInt(tokenIn)),
			tokenOutMinAmount: sdk.NewInt(tokenInMinAmount),
			expectError:       true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			suite.PrepareBalancerPool()
			suite.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountIn(sdk.WrapSDKContext(ctx), &types.MsgSwapExactAmountIn{
				Sender:            suite.TestAccs[0].String(),
				Routes:            tc.routes,
				TokenIn:           tc.tokenIn,
				TokenOutMinAmount: tc.tokenOutMinAmount,
			})

			if !tc.expectError {
				suite.NoError(err)
				suite.NotNil(response)
			}

			suite.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestSwapExactAmountOut_Events tests that events are correctly emitted
// when calling SwapExactAmountOut.
func (suite *KeeperTestSuite) TestSwapExactAmountOut_Events() {
	const (
		tokenInMaxAmount = int64Max
		tokenOut         = 5
	)

	testcases := map[string]struct {
		routes                []poolmanagertypes.SwapAmountOutRoute
		tokenOut              sdk.Coin
		tokenInMaxAmount      sdk.Int
		expectError           bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:           []poolmanagertypes.SwapAmountOutRoute{},
			tokenOut:         sdk.NewCoin("foo", sdk.NewInt(tokenOut)),
			tokenInMaxAmount: sdk.NewInt(tokenInMaxAmount),
			expectError:      true,
		},
		"one hop": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "bar",
				},
			},
			tokenOut:              sdk.NewCoin("foo", sdk.NewInt(tokenOut)),
			tokenInMaxAmount:      sdk.NewInt(tokenInMaxAmount),
			expectedSwapEvents:    1,
			expectedMessageEvents: 3, // 1 gamm + 2 events emitted by other keeper methods.
		},
		"two hops": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "bar",
				},
				{
					PoolId:       2,
					TokenInDenom: "baz",
				},
			},
			tokenOut:              sdk.NewCoin("foo", sdk.NewInt(tokenOut)),
			tokenInMaxAmount:      sdk.NewInt(tokenInMaxAmount),
			expectedSwapEvents:    2,
			expectedMessageEvents: 5, // 1 gamm + 4 events emitted by other keeper methods.
		},
		"invalid - two hops, denom does not exist": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "bar",
				},
				{
					PoolId:       2,
					TokenInDenom: "baz",
				},
			},
			tokenOut:         sdk.NewCoin(doesNotExistDenom, sdk.NewInt(tokenOut)),
			tokenInMaxAmount: sdk.NewInt(tokenInMaxAmount),
			expectError:      true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			suite.PrepareBalancerPool()
			suite.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountOut(sdk.WrapSDKContext(ctx), &types.MsgSwapExactAmountOut{
				Sender:           suite.TestAccs[0].String(),
				Routes:           tc.routes,
				TokenOut:         tc.tokenOut,
				TokenInMaxAmount: tc.tokenInMaxAmount,
			})

			if !tc.expectError {
				suite.NoError(err)
				suite.NotNil(response)
			}

			suite.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestJoinPool_Events tests that events are correctly emitted
// when calling JoinPool.
func (suite *KeeperTestSuite) TestJoinPool_Events() {
	const (
		tokenInMaxAmount = int64Max
		shareOut         = 110
	)

	testcases := map[string]struct {
		poolId                     uint64
		shareOutAmount             sdk.Int
		tokenInMaxs                sdk.Coins
		expectError                bool
		expectedAddLiquidityEvents int
		expectedMessageEvents      int
	}{
		"successful join": {
			poolId:         1,
			shareOutAmount: sdk.NewInt(shareOut),
			tokenInMaxs: sdk.NewCoins(
				sdk.NewCoin("foo", sdk.NewInt(tokenInMaxAmount)),
				sdk.NewCoin("bar", sdk.NewInt(tokenInMaxAmount)),
				sdk.NewCoin("baz", sdk.NewInt(tokenInMaxAmount)),
				sdk.NewCoin("uosmo", sdk.NewInt(tokenInMaxAmount)),
			),
			expectedAddLiquidityEvents: 1,
			expectedMessageEvents:      3, // 1 gamm + 2 events emitted by other keeper methods.
		},
		"tokenInMaxs do not match all tokens in pool - invalid join": {
			poolId:         1,
			shareOutAmount: sdk.NewInt(shareOut),
			tokenInMaxs:    sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tokenInMaxAmount))),
			expectError:    true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			suite.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.JoinPool(sdk.WrapSDKContext(ctx), &types.MsgJoinPool{
				Sender:         suite.TestAccs[0].String(),
				PoolId:         tc.poolId,
				ShareOutAmount: tc.shareOutAmount,
				TokenInMaxs:    tc.tokenInMaxs,
			})

			if !tc.expectError {
				suite.Require().NoError(err)
				suite.Require().NotNil(response)
			}

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, tc.expectedAddLiquidityEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestExitPool_Events tests that events are correctly emitted
// when calling ExitPool.
func (suite *KeeperTestSuite) TestExitPool_Events() {
	const (
		tokenOutMinAmount = 1
		shareIn           = 110
	)

	testcases := map[string]struct {
		poolId                        uint64
		shareInAmount                 sdk.Int
		tokenOutMins                  sdk.Coins
		expectError                   bool
		expectedRemoveLiquidityEvents int
		expectedMessageEvents         int
	}{
		"successful exit": {
			poolId:                        1,
			shareInAmount:                 sdk.NewInt(shareIn),
			tokenOutMins:                  sdk.NewCoins(),
			expectedRemoveLiquidityEvents: 1,
			expectedMessageEvents:         3, // 1 gamm + 2 events emitted by other keeper methods.
		},
		"invalid tokenOutMins": {
			poolId:        1,
			shareInAmount: sdk.NewInt(shareIn),
			tokenOutMins:  sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(tokenOutMinAmount))),
			expectError:   true,
		},
	}

	for name, tc := range testcases {
		suite.Run(name, func() {
			suite.Setup()
			ctx := suite.Ctx

			suite.PrepareBalancerPool()
			msgServer := keeper.NewMsgServerImpl(suite.App.GAMMKeeper)

			sender := suite.TestAccs[0].String()

			// Pre-join pool to be able to ExitPool.
			joinPoolResponse, err := msgServer.JoinPool(sdk.WrapSDKContext(ctx), &types.MsgJoinPool{
				Sender:         sender,
				PoolId:         tc.poolId,
				ShareOutAmount: sdk.NewInt(shareIn),
				TokenInMaxs: sdk.NewCoins(
					sdk.NewCoin("foo", sdk.NewInt(int64Max)),
					sdk.NewCoin("bar", sdk.NewInt(int64Max)),
					sdk.NewCoin("baz", sdk.NewInt(int64Max)),
					sdk.NewCoin("uosmo", sdk.NewInt(int64Max)),
				),
			})
			suite.Require().NoError(err)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Require().Equal(0, len(ctx.EventManager().Events()))

			// System under test.
			response, err := msgServer.ExitPool(sdk.WrapSDKContext(ctx), &types.MsgExitPool{
				Sender:        sender,
				PoolId:        tc.poolId,
				ShareInAmount: joinPoolResponse.ShareOutAmount,
				TokenOutMins:  tc.tokenOutMins,
			})

			if !tc.expectError {
				suite.Require().NoError(err)
				suite.Require().NotNil(response)
			}

			suite.AssertEventEmitted(ctx, types.TypeEvtPoolExited, tc.expectedRemoveLiquidityEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

func (suite *KeeperTestSuite) TestMsgMigrateShares() {
	defaultAccount := suite.TestAccs[0]
	defaultAccountFunds := sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(100000000000)), sdk.NewCoin("usdc", sdk.NewInt(100000000000)))
	joinPoolLiquidity := sdk.MustNewDecFromStr("100000000000000000000").RoundInt()

	type param struct {
		sender            sdk.AccAddress
		sharesToMigrate   sdk.Coin
		poolIdLeaving     uint64
		poolIdEntering    uint64
		minTick           int64
		maxTick           int64
		withdrawLiquidity sdk.Dec
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "happy path",
			param: param{
				sender:            defaultAccount,
				sharesToMigrate:   sdk.NewCoin("gamm/pool/1", joinPoolLiquidity),
				poolIdLeaving:     1,
				poolIdEntering:    2,
				minTick:           -1620000,
				maxTick:           3420000,
				withdrawLiquidity: sdk.MustNewDecFromStr("100000000000.000000000000000000"),
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		suite.SetupTest()
		msgServer := keeper.NewBalancerMsgServerImpl(suite.App.GAMMKeeper)
		ctx := sdk.WrapSDKContext(suite.Ctx)

		// Prepare both balancer and concentrated pools
		suite.FundAcc(test.param.sender, defaultAccountFunds)
		suite.PrepareBalancerPoolWithCoins(sdk.NewCoin("eth", sdk.NewInt(100000000000)), sdk.NewCoin("usdc", sdk.NewInt(100000000000)))
		clPool := suite.PrepareConcentratedPool()

		// Note balance before
		ethBalanceBefore := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, ETH)
		usdcBalanceBefore := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, USDC)
		fmt.Printf("user eth balance before joining pool 1: %v \n", ethBalanceBefore)
		fmt.Printf("user usdc balance before joining pool 1: %v \n", usdcBalanceBefore)

		_, _, err := suite.App.GAMMKeeper.JoinPoolNoSwap(suite.Ctx, test.param.sender, test.param.poolIdLeaving, joinPoolLiquidity, sdk.NewCoins(sdk.NewCoin("eth", sdk.NewInt(999999999999999)), sdk.NewCoin("usdc", sdk.NewInt(999999999999999))))
		suite.Require().NoError(err)

		ethBalanceAfter := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, ETH)
		usdcBalanceAfter := suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, USDC)
		fmt.Printf("user eth balance after joining pool 1: %v \n", ethBalanceAfter)
		fmt.Printf("user usdc balance after joining pool 1: %v \n", usdcBalanceAfter)

		fmt.Printf("test sender %v \n", test.param.sender.String())
		fmt.Printf("test receiver %v \n", clPool.GetAddress().String())

		msg := &balancer.MsgMigrateSharesToFullRangeConcentratedPosition{
			Sender:          test.param.sender.String(),
			SharesToMigrate: test.param.sharesToMigrate,
			PoolIdLeaving:   test.param.poolIdLeaving,
			PoolIdEntering:  test.param.poolIdEntering,
		}

		_, err = msgServer.MigrateSharesToFullRangeConcentratedPosition(ctx, msg)
		suite.Require().NoError(err)

		ethBalanceAfter = suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, ETH)
		usdcBalanceAfter = suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, USDC)
		fmt.Printf("user eth balance after joining pool 2: %v \n", ethBalanceAfter)
		fmt.Printf("user usdc balance after joining pool 2: %v \n", usdcBalanceAfter)

		clPoolAddress := clPool.GetAddress()

		ethBalanceAfter = suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
		usdcBalanceAfter = suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)
		fmt.Printf("pool eth balance after user joining pool 2: %v \n", ethBalanceAfter)
		fmt.Printf("pool usdc balance after user joining pool 2: %v \n", usdcBalanceAfter)

		// withdraw concentrated liquidity position
		amt0, amt1, err := suite.App.ConcentratedLiquidityKeeper.WithdrawPosition(suite.Ctx, test.param.poolIdEntering, test.param.sender, test.param.minTick, test.param.maxTick, test.param.withdrawLiquidity)
		suite.Require().NoError(err)
		fmt.Printf("amt0: %v, amt1: %v \n", amt0, amt1)

		ethBalanceAfter = suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, ETH)
		usdcBalanceAfter = suite.App.BankKeeper.GetBalance(suite.Ctx, test.param.sender, USDC)
		fmt.Printf("user eth balance after leaving pool 2: %v \n", ethBalanceAfter)
		fmt.Printf("user usdc balance after leaving pool 2: %v \n", usdcBalanceAfter)

		ethBalanceAfter = suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, ETH)
		usdcBalanceAfter = suite.App.BankKeeper.GetBalance(suite.Ctx, clPoolAddress, USDC)
		fmt.Printf("pool eth balance after user leaving pool 2: %v \n", ethBalanceAfter)
		fmt.Printf("pool usdc balance after user leaving pool 2: %v \n", usdcBalanceAfter)
	}
}
