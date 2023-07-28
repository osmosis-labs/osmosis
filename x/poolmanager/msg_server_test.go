package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagerKeeper "github.com/osmosis-labs/osmosis/v17/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v17/x/poolmanager/types"
)

var (
	amount     = sdk.NewInt(100)
	min_amount = sdk.ZeroInt()
	max_amount = sdk.NewInt(10000000)

	pool1_in = types.SwapAmountInRoute{PoolId: 1, TokenOutDenom: "bar"}
	pool2_in = types.SwapAmountInRoute{PoolId: 2, TokenOutDenom: "baz"}
	pool3_in = types.SwapAmountInRoute{PoolId: 3, TokenOutDenom: "uosmo"}
	pool4_in = types.SwapAmountInRoute{PoolId: 4, TokenOutDenom: "baz"}

	pool1_out = types.SwapAmountOutRoute{PoolId: 1, TokenInDenom: "bar"}
	pool2_out = types.SwapAmountOutRoute{PoolId: 2, TokenInDenom: "baz"}
	pool3_out = types.SwapAmountOutRoute{PoolId: 3, TokenInDenom: "bar"}
	pool4_out = types.SwapAmountOutRoute{PoolId: 4, TokenInDenom: "baz"}
)

func (s *KeeperTestSuite) TestSplitRouteSwapExactAmountIn() {
	testcases := map[string]struct {
		routes            []types.SwapAmountInSplitRoute
		tokenInDenom      string
		tokenoutMinAmount sdk.Int

		expectedSplitRouteSwapEvent int
		expectedMessageEvents       int
		expectedError               bool
	}{
		"valid case: two routes": {
			routes: []types.SwapAmountInSplitRoute{
				{
					Pools:         []types.SwapAmountInRoute{pool1_in, pool2_in},
					TokenInAmount: amount,
				},
				{
					Pools:         []types.SwapAmountInRoute{pool3_in, pool4_in},
					TokenInAmount: amount,
				},
			},
			tokenInDenom:      "baz",
			tokenoutMinAmount: min_amount,

			expectedSplitRouteSwapEvent: 1,
			expectedMessageEvents:       9, // 4 pool creation + 5 events in SplitRouteExactAmountIn keeper methods
		},
		"error: empty route": {
			routes:            []types.SwapAmountInSplitRoute{},
			tokenInDenom:      "baz",
			tokenoutMinAmount: min_amount,
			expectedError:     true,
		},
		"error path: denom doesnot exist routes": {
			routes: []types.SwapAmountInSplitRoute{
				{
					Pools:         []types.SwapAmountInRoute{pool1_in, pool2_in},
					TokenInAmount: amount,
				},
				{
					Pools:         []types.SwapAmountInRoute{pool1_in, pool2_in},
					TokenInAmount: amount,
				},
			},
			tokenInDenom:      "baz",
			tokenoutMinAmount: min_amount,

			expectedError: true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Setup()
			ctx := s.Ctx

			s.PrepareBalancerPool()
			s.PrepareBalancerPool()
			s.PrepareBalancerPool()
			s.PrepareBalancerPool()

			msgServer := poolmanagerKeeper.NewMsgServerImpl(s.App.PoolManagerKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SplitRouteSwapExactAmountIn(sdk.WrapSDKContext(ctx), &types.MsgSplitRouteSwapExactAmountIn{
				Sender:            s.TestAccs[0].String(),
				Routes:            tc.routes,
				TokenInDenom:      tc.tokenInDenom,
				TokenOutMinAmount: tc.tokenoutMinAmount,
			})
			if tc.expectedError {
				s.Require().Error(err)
				s.Require().Nil(response)
			} else {
				s.Require().NoError(err)
				s.AssertEventEmitted(ctx, types.TypeMsgSplitRouteSwapExactAmountIn, tc.expectedSplitRouteSwapEvent)
				s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			}

		})
	}
}

func (s *KeeperTestSuite) TestSplitRouteSwapExactAmountOut() {
	testcases := map[string]struct {
		routes            []types.SwapAmountOutSplitRoute
		tokenOutDenom     string
		tokenoutMaxAmount sdk.Int

		expectedSplitRouteSwapEvent int
		expectedMessageEvents       int
		expectedError               bool
	}{
		"valid case: two routes": {
			routes: []types.SwapAmountOutSplitRoute{
				{
					Pools:          []types.SwapAmountOutRoute{pool1_out, pool2_out},
					TokenOutAmount: amount,
				},
				{
					Pools:          []types.SwapAmountOutRoute{pool3_out, pool4_out},
					TokenOutAmount: amount,
				},
			},
			tokenOutDenom:     "uosmo",
			tokenoutMaxAmount: max_amount,

			expectedSplitRouteSwapEvent: 1,
			expectedMessageEvents:       9, // 4 pool creation + 5 events in SplitRouteExactAmountOut keeper methods
		},
		"error: empty route": {
			routes:            []types.SwapAmountOutSplitRoute{},
			tokenOutDenom:     "baz",
			tokenoutMaxAmount: max_amount,

			expectedError: true,
		},
		"error path: denom duplicate route": {
			routes: []types.SwapAmountOutSplitRoute{
				{
					Pools:          []types.SwapAmountOutRoute{pool1_out, pool2_out},
					TokenOutAmount: amount,
				},
				{
					Pools:          []types.SwapAmountOutRoute{pool1_out, pool2_out},
					TokenOutAmount: amount,
				},
			},
			tokenOutDenom:     "baz",
			tokenoutMaxAmount: max_amount,

			expectedError: true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Setup()
			ctx := s.Ctx

			s.PrepareBalancerPool()
			s.PrepareBalancerPool()
			s.PrepareBalancerPool()
			s.PrepareBalancerPool()

			msgServer := poolmanagerKeeper.NewMsgServerImpl(s.App.PoolManagerKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SplitRouteSwapExactAmountOut(sdk.WrapSDKContext(ctx), &types.MsgSplitRouteSwapExactAmountOut{
				Sender:           s.TestAccs[0].String(),
				Routes:           tc.routes,
				TokenOutDenom:    tc.tokenOutDenom,
				TokenInMaxAmount: tc.tokenoutMaxAmount,
			})
			if tc.expectedError {
				s.Require().Error(err)
				s.Require().Nil(response)
			} else {

				s.Require().NoError(err)
				s.AssertEventEmitted(ctx, types.TypeMsgSplitRouteSwapExactAmountOut, tc.expectedSplitRouteSwapEvent)
				s.AssertEventEmitted(ctx, sdk.EventTypeMessage, tc.expectedMessageEvents)
			}

		})
	}
}
