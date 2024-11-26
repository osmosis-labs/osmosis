package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

const (
	doesNotExistDenom = "nodenom"
	// Max positive int64.
	int64Max = int64(^uint64(0) >> 1)
)

// TestSwapExactAmountIn_Events tests that events are correctly emitted
// when calling SwapExactAmountIn.
func (s *KeeperTestSuite) TestSwapExactAmountIn_Events() {
	const (
		tokenInMinAmount = 1
		tokenIn          = 10
	)

	testcases := map[string]struct {
		routes                []poolmanagertypes.SwapAmountInRoute
		tokenIn               sdk.Coin
		tokenOutMinAmount     osmomath.Int
		expectError           bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:            []poolmanagertypes.SwapAmountInRoute{},
			tokenIn:           sdk.NewCoin("foo", osmomath.NewInt(tokenIn)),
			tokenOutMinAmount: osmomath.NewInt(tokenInMinAmount),
			expectError:       true,
		},
		"one hop": {
			routes: []poolmanagertypes.SwapAmountInRoute{
				{
					PoolId:        1,
					TokenOutDenom: "bar",
				},
			},
			tokenIn:               sdk.NewCoin("foo", osmomath.NewInt(tokenIn)),
			tokenOutMinAmount:     osmomath.NewInt(tokenInMinAmount),
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
			tokenIn:               sdk.NewCoin("foo", osmomath.NewInt(tokenIn)),
			tokenOutMinAmount:     osmomath.NewInt(tokenInMinAmount),
			expectedSwapEvents:    2,
			expectedMessageEvents: 6, // 1 gamm + 5 events emitted by other keeper methods.
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
			tokenIn:           sdk.NewCoin(doesNotExistDenom, osmomath.NewInt(tokenIn)),
			tokenOutMinAmount: osmomath.NewInt(tokenInMinAmount),
			expectError:       true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Setup()
			ctx := s.Ctx

			poolManagerParams := s.App.PoolManagerKeeper.GetParams(ctx)
			poolManagerParams.TakerFeeParams.DefaultTakerFee = osmomath.MustNewDecFromStr("0.01")
			s.App.PoolManagerKeeper.SetParams(ctx, poolManagerParams)

			s.PrepareBalancerPool()
			s.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(s.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountIn(ctx, &types.MsgSwapExactAmountIn{
				Sender:            s.TestAccs[0].String(),
				Routes:            tc.routes,
				TokenIn:           tc.tokenIn,
				TokenOutMinAmount: tc.tokenOutMinAmount,
			})

			if !tc.expectError {
				s.NoError(err)
				s.NotNil(response)
			}

			s.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestSwapExactAmountOut_Events tests that events are correctly emitted
// when calling SwapExactAmountOut.
func (s *KeeperTestSuite) TestSwapExactAmountOut_Events() {
	const (
		tokenInMaxAmount = int64Max
		tokenOut         = 10
	)

	testcases := map[string]struct {
		routes                []poolmanagertypes.SwapAmountOutRoute
		tokenOut              sdk.Coin
		tokenInMaxAmount      osmomath.Int
		expectError           bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:           []poolmanagertypes.SwapAmountOutRoute{},
			tokenOut:         sdk.NewCoin("foo", osmomath.NewInt(tokenOut)),
			tokenInMaxAmount: osmomath.NewInt(tokenInMaxAmount),
			expectError:      true,
		},
		"one hop": {
			routes: []poolmanagertypes.SwapAmountOutRoute{
				{
					PoolId:       1,
					TokenInDenom: "bar",
				},
			},
			tokenOut:              sdk.NewCoin("foo", osmomath.NewInt(tokenOut)),
			tokenInMaxAmount:      osmomath.NewInt(tokenInMaxAmount),
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
			tokenOut:              sdk.NewCoin("foo", osmomath.NewInt(tokenOut)),
			tokenInMaxAmount:      osmomath.NewInt(tokenInMaxAmount),
			expectedSwapEvents:    2,
			expectedMessageEvents: 6, // 1 gamm + 5 events emitted by other keeper methods.
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
			tokenOut:         sdk.NewCoin(doesNotExistDenom, osmomath.NewInt(tokenOut)),
			tokenInMaxAmount: osmomath.NewInt(tokenInMaxAmount),
			expectError:      true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Reset()
			ctx := s.Ctx

			poolManagerParams := s.App.PoolManagerKeeper.GetParams(ctx)
			poolManagerParams.TakerFeeParams.DefaultTakerFee = osmomath.MustNewDecFromStr("0.01")
			s.App.PoolManagerKeeper.SetParams(ctx, poolManagerParams)

			s.PrepareBalancerPool()
			s.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(s.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountOut(ctx, &types.MsgSwapExactAmountOut{
				Sender:           s.TestAccs[0].String(),
				Routes:           tc.routes,
				TokenOut:         tc.tokenOut,
				TokenInMaxAmount: tc.tokenInMaxAmount,
			})

			if !tc.expectError {
				s.NoError(err)
				s.NotNil(response)
			}

			s.AssertEventEmitted(ctx, types.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestJoinPool_Events tests that events are correctly emitted
// when calling JoinPool.
func (s *KeeperTestSuite) TestJoinPool_Events() {
	const (
		tokenInMaxAmount = int64Max
		shareOut         = 110
	)

	testcases := map[string]struct {
		poolId                     uint64
		shareOutAmount             osmomath.Int
		tokenInMaxs                sdk.Coins
		expectError                bool
		expectedAddLiquidityEvents int
		expectedMessageEvents      int
	}{
		"successful join": {
			poolId:         1,
			shareOutAmount: osmomath.NewInt(shareOut),
			tokenInMaxs: sdk.NewCoins(
				sdk.NewCoin("foo", osmomath.NewInt(tokenInMaxAmount)),
				sdk.NewCoin("bar", osmomath.NewInt(tokenInMaxAmount)),
				sdk.NewCoin("baz", osmomath.NewInt(tokenInMaxAmount)),
				sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(tokenInMaxAmount)),
			),
			expectedAddLiquidityEvents: 1,
			expectedMessageEvents:      2, // 1 gamm + 1 event emitted by other keeper methods.
		},
		"tokenInMaxs do not match all tokens in pool - invalid join": {
			poolId:         1,
			shareOutAmount: osmomath.NewInt(shareOut),
			tokenInMaxs:    sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(tokenInMaxAmount))),
			expectError:    true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Reset()
			ctx := s.Ctx

			s.PrepareBalancerPool()

			msgServer := keeper.NewMsgServerImpl(s.App.GAMMKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.JoinPool(ctx, &types.MsgJoinPool{
				Sender:         s.TestAccs[0].String(),
				PoolId:         tc.poolId,
				ShareOutAmount: tc.shareOutAmount,
				TokenInMaxs:    tc.tokenInMaxs,
			})

			if !tc.expectError {
				s.Require().NoError(err)
				s.Require().NotNil(response)
			}

			s.AssertEventEmitted(ctx, types.TypeEvtPoolJoined, tc.expectedAddLiquidityEvents)
			s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}

// TestExitPool_Events tests that events are correctly emitted
// when calling ExitPool.
func (s *KeeperTestSuite) TestExitPool_Events() {
	const (
		tokenOutMinAmount = 1
		shareIn           = 110
	)

	testcases := map[string]struct {
		poolId                        uint64
		shareInAmount                 osmomath.Int
		tokenOutMins                  sdk.Coins
		expectError                   bool
		expectedRemoveLiquidityEvents int
		expectedMessageEvents         int
	}{
		"successful exit": {
			poolId:                        1,
			shareInAmount:                 osmomath.NewInt(shareIn),
			tokenOutMins:                  sdk.NewCoins(),
			expectedRemoveLiquidityEvents: 1,
			expectedMessageEvents:         2, // 1 gamm + 1 event emitted by other keeper methods.
		},
		"invalid tokenOutMins": {
			poolId:        1,
			shareInAmount: osmomath.NewInt(shareIn),
			tokenOutMins:  sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(tokenOutMinAmount))),
			expectError:   true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Reset()
			ctx := s.Ctx

			s.PrepareBalancerPool()
			msgServer := keeper.NewMsgServerImpl(s.App.GAMMKeeper)

			sender := s.TestAccs[0].String()

			// Pre-join pool to be able to ExitPool.
			joinPoolResponse, err := msgServer.JoinPool(ctx, &types.MsgJoinPool{
				Sender:         sender,
				PoolId:         tc.poolId,
				ShareOutAmount: osmomath.NewInt(shareIn),
				TokenInMaxs: sdk.NewCoins(
					sdk.NewCoin("foo", osmomath.NewInt(int64Max)),
					sdk.NewCoin("bar", osmomath.NewInt(int64Max)),
					sdk.NewCoin("baz", osmomath.NewInt(int64Max)),
					sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(int64Max)),
				),
			})
			s.Require().NoError(err)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Require().Equal(0, len(ctx.EventManager().Events()))

			// System under test.
			response, err := msgServer.ExitPool(ctx, &types.MsgExitPool{
				Sender:        sender,
				PoolId:        tc.poolId,
				ShareInAmount: joinPoolResponse.ShareOutAmount,
				TokenOutMins:  tc.tokenOutMins,
			})

			if !tc.expectError {
				s.Require().NoError(err)
				s.Require().NotNil(response)
			}

			s.AssertEventEmitted(ctx, types.TypeEvtPoolExited, tc.expectedRemoveLiquidityEvents)
			s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}
