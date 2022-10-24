package swaprouter_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter"
	"github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
	swaproutertypes "github.com/osmosis-labs/osmosis/v12/x/swaprouter/types"
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
		routes                []swaproutertypes.SwapAmountInRoute
		tokenIn               sdk.Coin
		tokenOutMinAmount     sdk.Int
		expectError           bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:                []swaproutertypes.SwapAmountInRoute{},
			tokenIn:               sdk.NewCoin("foo", sdk.NewInt(tokenIn)),
			tokenOutMinAmount:     sdk.NewInt(tokenInMinAmount),
			expectedMessageEvents: 1, // 1 gamm.
		},
		"one hop": {
			routes: []swaproutertypes.SwapAmountInRoute{
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
			routes: []swaproutertypes.SwapAmountInRoute{
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
			routes: []swaproutertypes.SwapAmountInRoute{
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

			msgServer := swaprouter.NewMsgServerImpl(suite.App.SwapRouterKeeper)

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

			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtTokenSwapped, tc.expectedSwapEvents)
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
		routes                []swaproutertypes.SwapAmountOutRoute
		tokenOut              sdk.Coin
		tokenInMaxAmount      sdk.Int
		expectError           bool
		expectedSwapEvents    int
		expectedMessageEvents int
	}{
		"zero hops": {
			routes:                []swaproutertypes.SwapAmountOutRoute{},
			tokenOut:              sdk.NewCoin("foo", sdk.NewInt(tokenOut)),
			tokenInMaxAmount:      sdk.NewInt(tokenInMaxAmount),
			expectedMessageEvents: 1, // 1 gamm.
		},
		"one hop": {
			routes: []swaproutertypes.SwapAmountOutRoute{
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
			routes: []swaproutertypes.SwapAmountOutRoute{
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
			routes: []swaproutertypes.SwapAmountOutRoute{
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

			msgServer := swaprouter.NewMsgServerImpl(suite.App.SwapRouterKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			suite.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SwapExactAmountOut(sdk.WrapSDKContext(ctx), &swaproutertypes.MsgSwapExactAmountOut{
				Sender:           suite.TestAccs[0].String(),
				Routes:           tc.routes,
				TokenOut:         tc.tokenOut,
				TokenInMaxAmount: tc.tokenInMaxAmount,
			})

			if !tc.expectError {
				suite.NoError(err)
				suite.NotNil(response)
			}

			suite.AssertEventEmitted(ctx, gammtypes.TypeEvtTokenSwapped, tc.expectedSwapEvents)
			suite.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
		})
	}
}
