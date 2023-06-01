package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v15/x/gamm/pool-models/balancer"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/protorev/types"
)

func (s *KeeperTestSuite) TestSwapping() {
	type param struct {
		expectedTrades []types.Trade
		executeSwap    func()
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "swap exact amount in",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     1,
						TokenIn:  "akash",
						TokenOut: "Atom",
					},
				},
				executeSwap: func() {
					_, err := s.App.PoolManagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], 1, sdk.NewCoin("akash", sdk.NewInt(100)), "Atom", sdk.NewInt(1))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
		{
			name: "swap route exact amount in",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     1,
						TokenIn:  "akash",
						TokenOut: "Atom",
					},
				},
				executeSwap: func() {

					route := []poolmanagertypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: "Atom"}}

					_, err := s.App.PoolManagerKeeper.RouteExactAmountIn(s.Ctx, s.TestAccs[0], route, sdk.NewCoin("akash", sdk.NewInt(100)), sdk.NewInt(1))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
		{
			name: "swap route exact amount out",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     1,
						TokenIn:  "akash",
						TokenOut: "Atom",
					},
				},
				executeSwap: func() {

					route := []poolmanagertypes.SwapAmountOutRoute{{PoolId: 1, TokenInDenom: "akash"}}

					_, err := s.App.PoolManagerKeeper.RouteExactAmountOut(s.Ctx, s.TestAccs[0], route, sdk.NewInt(10000), sdk.NewCoin("Atom", sdk.NewInt(100)))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
		{
			name: "swap route exact amount in - 2 routes",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     1,
						TokenIn:  "akash",
						TokenOut: "Atom",
					},
					{
						Pool:     1,
						TokenIn:  "Atom",
						TokenOut: "akash",
					},
				},
				executeSwap: func() {

					route := []poolmanagertypes.SwapAmountInRoute{{PoolId: 1, TokenOutDenom: "Atom"}, {PoolId: 1, TokenOutDenom: "akash"}}

					_, err := s.App.PoolManagerKeeper.RouteExactAmountIn(s.Ctx, s.TestAccs[0], route, sdk.NewCoin("akash", sdk.NewInt(100)), sdk.NewInt(1))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
		{
			name: "swap route exact amount in - Concentrated Liquidity",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     49,
						TokenIn:  "uosmo",
						TokenOut: "epochTwo",
					},
				},
				executeSwap: func() {

					route := []poolmanagertypes.SwapAmountInRoute{{PoolId: 49, TokenOutDenom: "epochTwo"}}

					_, err := s.App.PoolManagerKeeper.RouteExactAmountIn(s.Ctx, s.TestAccs[0], route, sdk.NewCoin("uosmo", sdk.NewInt(10)), sdk.NewInt(1))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()
			tc.param.executeSwap()

			routes, err := s.App.ProtoRevKeeper.GetSwapsToBackrun(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.param.expectedTrades, routes.Trades)
		})
	}
}

func (s *KeeperTestSuite) TestLiquidityChanging() {
	type param struct {
		expectedTrades            []types.Trade
		executeLiquidityProviding func()
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "GAMM - Join Swap Exact Amount In",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     1,
						TokenIn:  "akash",
						TokenOut: "Atom",
					},
				},
				executeLiquidityProviding: func() {
					_, err := s.App.GAMMKeeper.JoinSwapExactAmountIn(s.Ctx, s.TestAccs[0], 1, sdk.NewCoins(sdk.NewCoin("akash", sdk.NewInt(100))), sdk.NewInt(1))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
		{
			name: "GAMM - Join Swap Share Amount Out",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     1,
						TokenIn:  "akash",
						TokenOut: "Atom",
					},
				},
				executeLiquidityProviding: func() {
					_, err := s.App.GAMMKeeper.JoinSwapShareAmountOut(s.Ctx, s.TestAccs[0], 1, "akash", sdk.NewInt(1000), sdk.NewInt(10000))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
		{
			name: "GAMM - Exit Swap Exact Amount Out",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     1,
						TokenIn:  "akash",
						TokenOut: "Atom",
					},
				},
				executeLiquidityProviding: func() {
					_, err := s.App.GAMMKeeper.ExitSwapExactAmountOut(s.Ctx, s.TestAccs[0], 1, sdk.NewCoin("Atom", sdk.NewInt(1)), sdk.NewInt(1002141106353159235))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
		{
			name: "GAMM - Exit Swap Share Amount In",
			param: param{
				expectedTrades: []types.Trade{
					{
						Pool:     1,
						TokenIn:  "akash",
						TokenOut: "Atom",
					},
				},
				executeLiquidityProviding: func() {
					_, err := s.App.GAMMKeeper.ExitSwapShareAmountIn(s.Ctx, s.TestAccs[0], 1, "Atom", sdk.NewInt(1000000000000000000), sdk.NewInt(1))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
		{
			name: "GAMM - Exit Swap Share Amount In - Low Shares",
			param: param{
				expectedTrades: []types.Trade(nil),
				executeLiquidityProviding: func() {
					_, err := s.App.GAMMKeeper.ExitSwapShareAmountIn(s.Ctx, s.TestAccs[0], 1, "Atom", sdk.NewInt(1000), sdk.NewInt(0))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			tc.param.executeLiquidityProviding()

			routes, err := s.App.ProtoRevKeeper.GetSwapsToBackrun(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.param.expectedTrades, routes.Trades)

			s.App.ProtoRevKeeper.DeleteSwapsToBackrun(s.Ctx)
		})
	}
}

func (s *KeeperTestSuite) TestPoolCreation() {
	type param struct {
		matchDenom          string
		executePoolCreation func() uint64
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "GAMM - Create Pool",
			param: param{
				matchDenom: "hookGamm",
				executePoolCreation: func() uint64 {
					poolId := s.createGAMMPool([]balancer.PoolAsset{
						{
							Token:  sdk.NewCoin("hookGamm", sdk.NewInt(1000000000)),
							Weight: sdk.NewInt(1),
						},
						{
							Token:  sdk.NewCoin(types.OsmosisDenomination, sdk.NewInt(1000000000)),
							Weight: sdk.NewInt(1),
						},
					},
						sdk.NewDecWithPrec(2, 3),
						sdk.NewDecWithPrec(0, 2))

					return poolId
				},
			},
			expectPass: true,
		},
		{
			name: "Concentrated Liquidity - Create Pool w/ No Liqudity",
			param: param{
				matchDenom: "hookCL",
				executePoolCreation: func() uint64 {
					clPool := s.PrepareConcentratedPool()
					return clPool.GetId()
				},
			},
			expectPass: false,
		},
		{
			name: "Concentrated Liquidity - Create Pool w/ Liqudity",
			param: param{
				matchDenom: "hookCL",
				executePoolCreation: func() uint64 {
					clPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("hookCL", "uosmo")
					return clPool.GetId()
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			poolId := tc.param.executePoolCreation()
			setPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, types.OsmosisDenomination, tc.param.matchDenom)

			if tc.expectPass {
				s.Require().NoError(err)
				s.Require().Equal(poolId, setPoolId)
			} else {
				s.Require().Error(err)
				s.Require().NotEqual(poolId, setPoolId)
			}
		})
	}
}
