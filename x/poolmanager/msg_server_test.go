package poolmanager_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	poolmanagerKeeper "github.com/osmosis-labs/osmosis/v27/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

var (
	amount     = osmomath.NewInt(100)
	min_amount = osmomath.ZeroInt()
	max_amount = osmomath.NewInt(10000000)

	pool1_in = types.SwapAmountInRoute{PoolId: 1, TokenOutDenom: "bar"}
	pool2_in = types.SwapAmountInRoute{PoolId: 2, TokenOutDenom: "baz"}
	pool3_in = types.SwapAmountInRoute{PoolId: 3, TokenOutDenom: appparams.BaseCoinUnit}
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
		tokenoutMinAmount osmomath.Int

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
			expectedMessageEvents:       12, // 4 pool creation + 8 events in SplitRouteExactAmountIn keeper methods
		},
		"error: empty route": {
			routes:            []types.SwapAmountInSplitRoute{},
			tokenInDenom:      "baz",
			tokenoutMinAmount: min_amount,
			expectedError:     true,
		},
		"error path: denom does not exist routes": {
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

			poolManagerParams := s.App.PoolManagerKeeper.GetParams(ctx)
			poolManagerParams.TakerFeeParams.DefaultTakerFee = osmomath.MustNewDecFromStr("0.01")
			s.App.PoolManagerKeeper.SetParams(ctx, poolManagerParams)

			s.PrepareBalancerPool()
			s.PrepareBalancerPool()
			s.PrepareBalancerPool()
			s.PrepareBalancerPool()

			msgServer := poolmanagerKeeper.NewMsgServerImpl(s.App.PoolManagerKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SplitRouteSwapExactAmountIn(ctx, &types.MsgSplitRouteSwapExactAmountIn{
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
		tokenoutMaxAmount osmomath.Int

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
			tokenOutDenom:     appparams.BaseCoinUnit,
			tokenoutMaxAmount: max_amount,

			expectedSplitRouteSwapEvent: 1,
			expectedMessageEvents:       12, // 4 pool creation + 8 events in SplitRouteExactAmountOut keeper methods
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

			poolManagerParams := s.App.PoolManagerKeeper.GetParams(ctx)
			poolManagerParams.TakerFeeParams.DefaultTakerFee = osmomath.MustNewDecFromStr("0.01")
			s.App.PoolManagerKeeper.SetParams(ctx, poolManagerParams)

			s.PrepareBalancerPool()
			s.PrepareBalancerPool()
			s.PrepareBalancerPool()
			s.PrepareBalancerPool()

			msgServer := poolmanagerKeeper.NewMsgServerImpl(s.App.PoolManagerKeeper)

			// Reset event counts to 0 by creating a new manager.
			ctx = ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(ctx.EventManager().Events()))

			response, err := msgServer.SplitRouteSwapExactAmountOut(ctx, &types.MsgSplitRouteSwapExactAmountOut{
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
		"valid case: two pairs, single direction": {
			denomPairTakerFeeMessage: types.MsgSetDenomPairTakerFee{
				Sender: adminAcc,
				DenomPairTakerFee: []types.DenomPairTakerFee{
					{
						TokenInDenom:  "denom0",
						TokenOutDenom: "denom1",
						TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
					},
					{
						TokenInDenom:  "denom0",
						TokenOutDenom: "denom2",
						TakerFee:      osmomath.MustNewDecFromStr("0.0016"),
					},
				},
			},

			expectedSetDenomPairTakerFeeEvent: 2,
		},
		"valid case: two pairs, both directions": {
			denomPairTakerFeeMessage: types.MsgSetDenomPairTakerFee{
				Sender: adminAcc,
				DenomPairTakerFee: []types.DenomPairTakerFee{
					{
						TokenInDenom:  "denom0",
						TokenOutDenom: "denom1",
						TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
					},
					{
						TokenInDenom:  "denom1",
						TokenOutDenom: "denom0",
						TakerFee:      osmomath.MustNewDecFromStr("0.0014"),
					},
					{
						TokenInDenom:  "denom0",
						TokenOutDenom: "denom2",
						TakerFee:      osmomath.MustNewDecFromStr("0.0016"),
					},
					{
						TokenInDenom:  "denom2",
						TokenOutDenom: "denom0",
						TakerFee:      osmomath.MustNewDecFromStr("0.0015"),
					},
				},
			},

			expectedSetDenomPairTakerFeeEvent: 4,
		},
		"valid case: one pair, single direction": {
			denomPairTakerFeeMessage: types.MsgSetDenomPairTakerFee{
				Sender: adminAcc,
				DenomPairTakerFee: []types.DenomPairTakerFee{
					{
						TokenInDenom:  "denom0",
						TokenOutDenom: "denom1",
						TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
					},
				},
			},

			expectedSetDenomPairTakerFeeEvent: 1,
		},
		"valid case: one pair, both directions": {
			denomPairTakerFeeMessage: types.MsgSetDenomPairTakerFee{
				Sender: adminAcc,
				DenomPairTakerFee: []types.DenomPairTakerFee{
					{
						TokenInDenom:  "denom0",
						TokenOutDenom: "denom1",
						TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
					},
					{
						TokenInDenom:  "denom1",
						TokenOutDenom: "denom0",
						TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
					},
				},
			},

			expectedSetDenomPairTakerFeeEvent: 2,
		},
		"error: not admin account": {
			denomPairTakerFeeMessage: types.MsgSetDenomPairTakerFee{
				Sender: nonAdminAcc,
				DenomPairTakerFee: []types.DenomPairTakerFee{
					{
						TokenInDenom:  "denom0",
						TokenOutDenom: "denom1",
						TakerFee:      osmomath.MustNewDecFromStr("0.0013"),
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

			response, err := msgServer.SetDenomPairTakerFee(s.Ctx, &types.MsgSetDenomPairTakerFee{
				Sender:            tc.denomPairTakerFeeMessage.Sender,
				DenomPairTakerFee: tc.denomPairTakerFeeMessage.DenomPairTakerFee,
			})
			if tc.expectedError {
				s.Require().Error(err)
				s.Require().Nil(response)
			} else {
				s.Require().NoError(err)
				s.AssertEventEmitted(s.Ctx, types.TypeMsgSetDenomPairTakerFee, tc.expectedSetDenomPairTakerFeeEvent)
				s.AssertEventEmitted(s.Ctx, sdk.EventTypeMessage, 0)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetTakerFeeShareAgreementForDenomMsg() {
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName).String()
	nonGovAddr := s.TestAccs[0].String()
	skimAddress := s.TestAccs[1].String()

	testcases := map[string]struct {
		takerFeeShareAgreementMessage types.MsgSetTakerFeeShareAgreementForDenom
		expectedError                 error
	}{
		"valid case": {
			takerFeeShareAgreementMessage: types.MsgSetTakerFeeShareAgreementForDenom{
				Sender:      govAddr,
				Denom:       "nBTC",
				SkimPercent: osmomath.MustNewDecFromStr("0.01"),
				SkimAddress: skimAddress,
			},
		},
		"error: not gov account": {
			takerFeeShareAgreementMessage: types.MsgSetTakerFeeShareAgreementForDenom{
				Sender:      nonGovAddr,
				Denom:       "nBTC",
				SkimPercent: osmomath.MustNewDecFromStr("0.01"),
				SkimAddress: skimAddress,
			},
			expectedError: types.ErrUnauthorizedGov,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Setup()

			msgServer := poolmanagerKeeper.NewMsgServerImpl(s.App.PoolManagerKeeper)

			response, err := msgServer.SetTakerFeeShareAgreementForDenom(s.Ctx, &tc.takerFeeShareAgreementMessage)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError, err)
				s.Require().Nil(response)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetRegisteredAlloyedPoolMsg() {
	govAddr := s.App.AccountKeeper.GetModuleAddress(govtypes.ModuleName).String()
	nonGovAddr := s.TestAccs[0].String()

	testcases := map[string]struct {
		registeredAlloyedPoolMessage types.MsgSetRegisteredAlloyedPool
		useAlloyedPool               bool
		useConcentratedPool          bool
		expectedError                error
	}{
		"valid sender, valid pool": {
			registeredAlloyedPoolMessage: types.MsgSetRegisteredAlloyedPool{
				Sender: govAddr,
			},
			useAlloyedPool: true,
		},
		"valid sender, invalid pool": {
			registeredAlloyedPoolMessage: types.MsgSetRegisteredAlloyedPool{
				Sender: govAddr,
			},
			useConcentratedPool: true,
			expectedError:       types.NotCosmWasmPoolError{PoolId: 1},
		},
		"invalid sender, valid pool": {
			registeredAlloyedPoolMessage: types.MsgSetRegisteredAlloyedPool{
				Sender: nonGovAddr,
			},
			useAlloyedPool: true,
			expectedError:  types.ErrUnauthorizedGov,
		},
		"invalid sender, invalid pool": {
			registeredAlloyedPoolMessage: types.MsgSetRegisteredAlloyedPool{
				Sender: nonGovAddr,
			},
			useConcentratedPool: true,
			expectedError:       types.ErrUnauthorizedGov,
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.Setup()
			msgServer := poolmanagerKeeper.NewMsgServerImpl(s.App.PoolManagerKeeper)

			allSupportedPoolInfo := s.PrepareAllSupportedPools()

			if tc.useAlloyedPool {
				tc.registeredAlloyedPoolMessage.PoolId = allSupportedPoolInfo.AlloyedPoolID
			} else if tc.useConcentratedPool {
				tc.registeredAlloyedPoolMessage.PoolId = allSupportedPoolInfo.ConcentratedPoolID
			}

			response, err := msgServer.SetRegisteredAlloyedPool(s.Ctx, &tc.registeredAlloyedPoolMessage)
			if tc.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(tc.expectedError, err)
				s.Require().Nil(response)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
