package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	poolmanagerKeeper "github.com/osmosis-labs/osmosis/v19/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
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
			expectedMessageEvents:       16, // 4 pool creation + 12 events in SplitRouteExactAmountIn keeper methods
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
			expectedMessageEvents:       17, // 4 pool creation + 13 events in SplitRouteExactAmountOut keeper methods
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

func (s *KeeperTestSuite) TestSetDenomPairTakerFee() {
	adminAcc := s.TestAccs[0].String()
	nonAdminAcc := s.TestAccs[1].String()
	testcases := map[string]struct {
		denomPairTakerFeeMessage types.MsgSetDenomPairTakerFee

		expectedSetDenomPairTakerFeeEvent int
		expectedMessageEvents             int
		expectedError                     bool
	}{
		"valid case: two pairs": {
			denomPairTakerFeeMessage: types.MsgSetDenomPairTakerFee{
				Sender: adminAcc,
				DenomPairTakerFee: []types.DenomPairTakerFee{
					{
						Denom0:   "denom0",
						Denom1:   "denom1",
						TakerFee: sdk.MustNewDecFromStr("0.0013"),
					},
					{
						Denom0:   "denom0",
						Denom1:   "denom2",
						TakerFee: sdk.MustNewDecFromStr("0.0016"),
					},
				},
			},

			expectedSetDenomPairTakerFeeEvent: 2,
		},
		"valid case: one pair": {
			denomPairTakerFeeMessage: types.MsgSetDenomPairTakerFee{
				Sender: adminAcc,
				DenomPairTakerFee: []types.DenomPairTakerFee{
					{
						Denom0:   "denom0",
						Denom1:   "denom1",
						TakerFee: sdk.MustNewDecFromStr("0.0013"),
					},
				},
			},

			expectedSetDenomPairTakerFeeEvent: 1,
		},
		"error: not admin account": {
			denomPairTakerFeeMessage: types.MsgSetDenomPairTakerFee{
				Sender: nonAdminAcc,
				DenomPairTakerFee: []types.DenomPairTakerFee{
					{
						Denom0:   "denom0",
						Denom1:   "denom1",
						TakerFee: sdk.MustNewDecFromStr("0.0013"),
					},
				},
			},

			expectedError: true,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Setup()
			msgServer := poolmanagerKeeper.NewMsgServerImpl(s.App.PoolManagerKeeper)

			// Add the admin address to the pool manager params.
			poolManagerParams := s.App.PoolManagerKeeper.GetParams(s.Ctx)
			poolManagerParams.TakerFeeParams.AdminAddresses = []string{adminAcc}
			s.App.PoolManagerKeeper.SetParams(s.Ctx, poolManagerParams)

			// Reset event counts to 0 by creating a new manager.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(s.Ctx.EventManager().Events()))

			response, err := msgServer.SetDenomPairTakerFee(sdk.WrapSDKContext(s.Ctx), &types.MsgSetDenomPairTakerFee{
				Sender:            tc.denomPairTakerFeeMessage.Sender,
				DenomPairTakerFee: tc.denomPairTakerFeeMessage.DenomPairTakerFee,
			})
			if tc.expectedError {
				s.Require().Error(err)
				s.Require().Nil(response)
			} else {
				s.Require().NoError(err)
				s.AssertEventEmitted(s.Ctx, types.TypeMsgSetDenomPairTakerFee, tc.expectedSetDenomPairTakerFeeEvent)
				s.AssertEventEmitted(s.Ctx, sdk.EventTypeMessage, 1)
			}
		})
	}
}
