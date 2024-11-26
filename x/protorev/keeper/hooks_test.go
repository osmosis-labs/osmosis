package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/stableswap"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/protorev/types"
)

// Tests the hook implementation that is called after swapping
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
					_, _, err := s.App.PoolManagerKeeper.SwapExactAmountIn(s.Ctx, s.TestAccs[0], 1, sdk.NewCoin("akash", osmomath.NewInt(100)), "Atom", osmomath.NewInt(1))
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

					_, err := s.App.PoolManagerKeeper.RouteExactAmountIn(s.Ctx, s.TestAccs[0], route, sdk.NewCoin("akash", osmomath.NewInt(100)), osmomath.NewInt(1))
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

					_, err := s.App.PoolManagerKeeper.RouteExactAmountOut(s.Ctx, s.TestAccs[0], route, osmomath.NewInt(10000), sdk.NewCoin("Atom", osmomath.NewInt(100)))
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

					_, err := s.App.PoolManagerKeeper.RouteExactAmountIn(s.Ctx, s.TestAccs[0], route, sdk.NewCoin("akash", osmomath.NewInt(100)), osmomath.NewInt(1))
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
						Pool:     50,
						TokenIn:  appparams.BaseCoinUnit,
						TokenOut: "epochTwo",
					},
				},
				executeSwap: func() {

					route := []poolmanagertypes.SwapAmountInRoute{{PoolId: 50, TokenOutDenom: "epochTwo"}}

					_, err := s.App.PoolManagerKeeper.RouteExactAmountIn(s.Ctx, s.TestAccs[0], route, sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10)), osmomath.NewInt(1))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupPoolsTest()
			tc.param.executeSwap()

			routes, err := s.App.ProtoRevKeeper.GetSwapsToBackrun(s.Ctx)
			s.Require().NoError(err)
			s.Require().Equal(tc.param.expectedTrades, routes.Trades)
		})
	}
}

// Tests the hook implementation that is called after liquidity providing
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
					_, err := s.App.GAMMKeeper.JoinSwapExactAmountIn(s.Ctx, s.TestAccs[0], 1, sdk.NewCoins(sdk.NewCoin("akash", osmomath.NewInt(100))), osmomath.NewInt(1))
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
					_, err := s.App.GAMMKeeper.JoinSwapShareAmountOut(s.Ctx, s.TestAccs[0], 1, "akash", osmomath.NewInt(1000), osmomath.NewInt(10000))
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
					_, err := s.App.GAMMKeeper.ExitSwapExactAmountOut(s.Ctx, s.TestAccs[0], 1, sdk.NewCoin("Atom", osmomath.NewInt(1)), osmomath.NewInt(1002141106353159235))
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
					_, err := s.App.GAMMKeeper.ExitSwapShareAmountIn(s.Ctx, s.TestAccs[0], 1, "Atom", osmomath.NewInt(1000000000000000000), osmomath.NewInt(1))
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
					_, err := s.App.GAMMKeeper.ExitSwapShareAmountIn(s.Ctx, s.TestAccs[0], 1, "Atom", osmomath.NewInt(1000), osmomath.NewInt(0))
					s.Require().NoError(err)
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupPoolsTest()
			tc.param.executeLiquidityProviding()

			routes, err := s.App.ProtoRevKeeper.GetSwapsToBackrun(s.Ctx)
			if tc.expectPass {
				s.Require().NoError(err)
				s.Require().Equal(tc.param.expectedTrades, routes.Trades)
			}
		})
	}
}

// Tests the hook implementation that is called after pool creation with coins
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
							Token:  sdk.NewCoin("hookGamm", osmomath.NewInt(1000000000)),
							Weight: osmomath.NewInt(1),
						},
						{
							Token:  sdk.NewCoin(types.OsmosisDenomination, osmomath.NewInt(1000000000)),
							Weight: osmomath.NewInt(1),
						},
					},
						osmomath.NewDecWithPrec(2, 3),
						osmomath.NewDecWithPrec(0, 2))

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
					clPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("hookCL", appparams.BaseCoinUnit)
					return clPool.GetId()
				},
			},
			expectPass: true,
		},
		{
			name: "Create Balancer Pool First, Then Concentrated Liquidity w/ Liquidity - CL with more liquidity so should be stored",
			param: param{
				matchDenom: "hook",
				executePoolCreation: func() uint64 {
					// Create balancer pool first with a new denom pair
					balancerPoolId := s.createGAMMPool([]balancer.PoolAsset{
						{
							Token:  sdk.NewCoin("hook", osmomath.NewInt(1)),
							Weight: osmomath.NewInt(1),
						},
						{
							Token:  sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(1)),
							Weight: osmomath.NewInt(1),
						},
					},
						osmomath.NewDecWithPrec(1, 1),
						osmomath.NewDecWithPrec(0, 2),
					)

					// Ensure that the balancer pool is stored since no other pool exists for the denom pair
					setPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, "hook")
					s.Require().NoError(err)
					s.Require().Equal(balancerPoolId, setPoolId)

					// Create Concentrated Liquidity pool with the same denom pair and more liquidity
					// The returned pool id should be what is finally stored in the protorev keeper
					clPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("hook", appparams.BaseCoinUnit)
					return clPool.GetId()
				},
			},
			expectPass: true,
		},
		{
			name: "Create Concentrated Liquidity Pool w/ Liquidity First, Then Balancer Pool - Balancer with more liquidity so should be stored",
			param: param{
				matchDenom: "hook",
				executePoolCreation: func() uint64 {
					// Create Concentrated Liquidity pool with a denom pair not already stored
					clPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition("hook", appparams.BaseCoinUnit)

					// Ensure that the concentrated pool is stored since no other pool exists for the denom pair
					setPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, "hook")
					s.Require().NoError(err)
					s.Require().Equal(clPool.GetId(), setPoolId)

					// Get clPool liquidity
					clPoolLiquidity, err := s.App.PoolManagerKeeper.GetTotalPoolLiquidity(s.Ctx, clPool.GetId())
					s.Require().NoError(err)

					// Create balancer pool with the same denom pair and more liquidity
					balancerPoolId := s.createGAMMPool([]balancer.PoolAsset{
						{
							Token:  sdk.NewCoin(clPoolLiquidity[0].Denom, clPoolLiquidity[0].Amount.Add(osmomath.NewInt(1))),
							Weight: osmomath.NewInt(1),
						},
						{
							Token:  sdk.NewCoin(clPoolLiquidity[1].Denom, clPoolLiquidity[1].Amount.Add(osmomath.NewInt(1))),
							Weight: osmomath.NewInt(1),
						},
					},
						osmomath.NewDecWithPrec(1, 1),
						osmomath.NewDecWithPrec(0, 2),
					)

					// The returned pool id should be what is finally stored in the protorev keeper since it has more liquidity
					return balancerPoolId
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupPoolsTest()
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

// Helper function tests

// Tests StoreSwap stores a swap properly
func (s *KeeperTestSuite) TestStoreSwap() {
	type param struct {
		expectedSwap           types.Trade
		prepareState           func()
		expectedSwapsStoredLen int
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Store Swap Without An Existing Swap Stored",
			param: param{
				expectedSwap: types.Trade{
					Pool:     1,
					TokenIn:  "akash",
					TokenOut: "Atom",
				},
				prepareState: func() {
					s.App.ProtoRevKeeper.DeleteSwapsToBackrun(s.Ctx)
				},
				expectedSwapsStoredLen: 1,
			},
			expectPass: true,
		},
		{
			name: "Store Swap With With An Existing Swap Stored",
			param: param{
				expectedSwap: types.Trade{
					Pool:     2,
					TokenIn:  appparams.BaseCoinUnit,
					TokenOut: "test",
				},
				prepareState: func() {
					s.App.ProtoRevKeeper.SetSwapsToBackrun(s.Ctx, types.Route{
						Trades: []types.Trade{
							{
								Pool:     1,
								TokenIn:  "Atom",
								TokenOut: "akash",
							},
						},
						StepSize: osmomath.NewInt(1),
					})
				},
				expectedSwapsStoredLen: 2,
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupPoolsTest()

			// Run any state preparation
			tc.param.prepareState()

			// Store the swap
			s.App.ProtoRevKeeper.StoreSwap(s.Ctx, tc.param.expectedSwap.Pool, tc.param.expectedSwap.TokenIn, tc.param.expectedSwap.TokenOut)

			routes, err := s.App.ProtoRevKeeper.GetSwapsToBackrun(s.Ctx)
			if tc.expectPass {
				s.Require().NoError(err)
				s.Require().Equal(tc.param.expectedSwapsStoredLen, len(routes.Trades))
				s.Require().Equal(tc.param.expectedSwap, routes.Trades[len(routes.Trades)-1])
			}
		})
	}
}

// Tests GetComparablePoolLiquidity gets the comparable liquidity of a pool by multiplying the amounts of the pool coins.
func (s *KeeperTestSuite) TestGetComparablePoolLiquidity() {
	type param struct {
		executePoolCreation         func() uint64
		expectedComparableLiquidity osmomath.Int
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Get Balancer Pool Comparable Liquidity",
			param: param{
				executePoolCreation: func() uint64 {
					return s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10)), sdk.NewCoin("juno", osmomath.NewInt(10)))
				},
				expectedComparableLiquidity: osmomath.NewInt(100),
			},
			expectPass: true,
		},
		{
			name: "Get Stable Swap Pool Comparable Liquidity",
			param: param{
				executePoolCreation: func() uint64 {
					return s.createStableswapPool(
						sdk.NewCoins(
							sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10)),
							sdk.NewCoin("juno", osmomath.NewInt(10)),
						),
						stableswap.PoolParams{
							SwapFee: osmomath.NewDecWithPrec(1, 4),
							ExitFee: osmomath.NewDecWithPrec(0, 2),
						},
						[]uint64{1, 1},
					)
				},
				expectedComparableLiquidity: osmomath.NewInt(100),
			},
			expectPass: true,
		},
		{
			name: "Get Concentrated Liquidity Pool Comparable Liquidity",
			param: param{
				executePoolCreation: func() uint64 {
					pool := s.PrepareConcentratedPool()
					err := s.App.BankKeeper.SendCoins(
						s.Ctx,
						s.TestAccs[0],
						pool.GetAddress(),
						sdk.NewCoins(sdk.NewCoin("eth", osmomath.NewInt(10)), sdk.NewCoin("usdc", osmomath.NewInt(10))))
					s.Require().NoError(err)
					return pool.GetId()
				},
				expectedComparableLiquidity: osmomath.NewInt(100),
			},
			expectPass: true,
		},
		{
			name: "Catch overflow error on liquidity multiplication",
			param: param{
				executePoolCreation: func() uint64 {
					return s.PrepareBalancerPoolWithCoins(
						sdk.NewCoin(appparams.BaseCoinUnit, osmomath.Int(osmomath.NewUintFromString("999999999999999999999999999999999999999"))),
						sdk.NewCoin("juno", osmomath.Int(osmomath.NewUintFromString("999999999999999999999999999999999999999"))))
				},
				expectedComparableLiquidity: osmomath.Int{},
			},
			expectPass: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupPoolsTest()

			// Create the pool
			poolId := tc.param.executePoolCreation()

			// Get the comparable liquidity
			comparableLiquidity, err := s.App.ProtoRevKeeper.GetComparablePoolLiquidity(s.Ctx, poolId)

			if tc.expectPass {
				s.Require().NoError(err)
				s.Require().Equal(tc.param.expectedComparableLiquidity, comparableLiquidity)
			} else {
				s.Require().Error(err)
				s.Require().Equal(tc.param.expectedComparableLiquidity, comparableLiquidity)
			}
		})
	}
}

// Tests StoreJoinExitPoolSwaps stores the swaps associated with GAMM join/exit pool messages in the store, depending on if it is a join or exit.
func (s *KeeperTestSuite) TestStoreJoinExitPoolSwaps() {
	type param struct {
		poolId       uint64
		denom        string
		isJoin       bool
		expectedSwap types.Trade
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Store Join Pool Swap",
			param: param{
				poolId: 1,
				denom:  "akash",
				isJoin: true,
				expectedSwap: types.Trade{
					Pool:     1,
					TokenIn:  "akash",
					TokenOut: "Atom",
				},
			},
			expectPass: true,
		},
		{
			name: "Store Exit Pool Swap",
			param: param{
				poolId: 1,
				denom:  "akash",
				isJoin: false,
				expectedSwap: types.Trade{
					Pool:     1,
					TokenIn:  "Atom",
					TokenOut: "akash",
				},
			},
			expectPass: true,
		},
		{
			name: "Non-Gamm Pool, Return Early Do Not Store Any Swaps",
			param: param{
				poolId:       50,
				denom:        appparams.BaseCoinUnit,
				isJoin:       true,
				expectedSwap: types.Trade{},
			},
			expectPass: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupPoolsTest()

			// All pools are already created in the setup
			s.App.ProtoRevKeeper.StoreJoinExitPoolSwaps(s.Ctx, s.TestAccs[0], tc.param.poolId, tc.param.denom, tc.param.isJoin)

			// Get the swaps to backrun after storing the swap via StoreJoinExitPoolSwaps
			swapsToBackrun, err := s.App.ProtoRevKeeper.GetSwapsToBackrun(s.Ctx)

			if tc.expectPass {
				s.Require().NoError(err)
				s.Require().Equal(tc.param.expectedSwap, swapsToBackrun.Trades[len(swapsToBackrun.Trades)-1])
			} else {
				s.Require().Equal(swapsToBackrun, types.Route{})
			}
		})
	}
}

// Tests CompareAndStorePool compares the comparable liquidity of a pool with the stored pool, storing the new pool if it has higher comparable liquidity.
// Note: This test calls DeleteAllPoolsForBaseDenom in the prepareStateAndGetPoolIdToCompare function because the
// hooks are triggered by default and we want to test the CompareAndStorePool on the state before the hooks are triggered.
func (s *KeeperTestSuite) TestCompareAndStorePool() {
	type param struct {
		baseDenom                         string
		matchDenom                        string
		prepareStateAndGetPoolIdToCompare func() (uint64, uint64)
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Nothing Stored, Store Balancer",
			param: param{
				baseDenom:  appparams.BaseCoinUnit,
				matchDenom: "juno",
				prepareStateAndGetPoolIdToCompare: func() (uint64, uint64) {
					// Prepare a balancer pool with coins
					poolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10)), sdk.NewCoin("juno", osmomath.NewInt(10)))

					// Delete all pools for the base denom uosmo so that all tests start with a clean slate
					s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, appparams.BaseCoinUnit)

					return poolId, poolId
				},
			},
			expectPass: true,
		},
		{
			name: "Nothing Stored, Store Concentrated Liquidity Pool w/ Coins",
			param: param{
				baseDenom:  appparams.BaseCoinUnit,
				matchDenom: "stake",
				prepareStateAndGetPoolIdToCompare: func() (uint64, uint64) {
					// Prepare a concentrated liquidity pool with coins
					poolId := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(appparams.BaseCoinUnit, "stake").GetId()

					// Delete all pools for the base denom uosmo so that all tests start with a clean slate
					s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, appparams.BaseCoinUnit)

					return poolId, poolId
				},
			},
			expectPass: true,
		},
		{
			name: "Balancer Previously Stored w/ Less liquidity, Compare Concentrated Liquidity Pool w/ More liqudidity, Ensure CL Gets Stored",
			param: param{
				baseDenom:  appparams.BaseCoinUnit,
				matchDenom: "stake",
				prepareStateAndGetPoolIdToCompare: func() (uint64, uint64) {
					// Create a concentrated liquidity pool with more liquidity
					clPoolId := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(appparams.BaseCoinUnit, "stake").GetId()

					// Delete all pools for the base denom uosmo so that all tests start with a clean slate
					s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, appparams.BaseCoinUnit)

					preparedPoolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10)), sdk.NewCoin("stake", osmomath.NewInt(10)))
					storedPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, "stake")
					s.Require().NoError(err)
					s.Require().Equal(preparedPoolId, storedPoolId)

					// Return the cl pool id as expected since it has higher liquidity, compare the cl pool id
					return clPoolId, clPoolId
				},
			},
			expectPass: true,
		},
		{
			name: "Balancer Previously Stored w/ More liquidity, Compare Concentrated Liquidity Pool w/ Less liqudidity, Ensure Balancer Stays Stored",
			param: param{
				baseDenom:  appparams.BaseCoinUnit,
				matchDenom: "stake",
				prepareStateAndGetPoolIdToCompare: func() (uint64, uint64) {
					// Create a concentrated liquidity pool with more liquidity
					clPoolId := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(appparams.BaseCoinUnit, "stake").GetId()

					// Delete all pools for the base denom uosmo so that all tests start with a clean slate
					s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, appparams.BaseCoinUnit)

					// Prepare a balancer pool with more liquidity
					balancerPoolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(2000000000000000000)), sdk.NewCoin("stake", osmomath.NewInt(1000000000000000000)))
					storedPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, "stake")
					s.Require().NoError(err)
					s.Require().Equal(balancerPoolId, storedPoolId)

					// Return the balancer pool id as expected since it has higher liquidity, compare the cl pool id
					return balancerPoolId, clPoolId
				},
			},
			expectPass: true,
		},
		{
			name: "Concentrated Liquidity Previously Stored w/ Less liquidity, Compare Balancer Pool w/ More liqudidity, Ensure Balancer Gets Stored",
			param: param{
				baseDenom:  appparams.BaseCoinUnit,
				matchDenom: "stake",
				prepareStateAndGetPoolIdToCompare: func() (uint64, uint64) {
					// Prepare a balancer pool with more liquidity
					balancerPoolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(2000000000000000000)), sdk.NewCoin("stake", osmomath.NewInt(1000000000000000000)))

					// Delete all pools for the base denom uosmo so that all tests start with a clean slate
					s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, appparams.BaseCoinUnit)

					// Prepare a concentrated liquidity pool with less liquidity, should be stored since nothing is stored
					clPoolId := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(appparams.BaseCoinUnit, "stake").GetId()
					storedPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, "stake")
					s.Require().NoError(err)
					s.Require().Equal(clPoolId, storedPoolId)

					// Return the balancer pool id as expected since it has higher liquidity, compare the balancer pool id
					return balancerPoolId, balancerPoolId
				},
			},
			expectPass: true,
		},
		{
			name: "Concentrated Liquidity Previously Stored w/ More liquidity, Compare Balancer Pool w/ Less liqudidity, Ensure CL Stays Stored",
			param: param{
				baseDenom:  appparams.BaseCoinUnit,
				matchDenom: "stake",
				prepareStateAndGetPoolIdToCompare: func() (uint64, uint64) {
					// Prepare a balancer pool with less liquidity
					balancerPoolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(500000000000000000)), sdk.NewCoin("stake", osmomath.NewInt(1000000000000000000)))

					// Delete all pools for the base denom uosmo so that all tests start with a clean slate
					s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, appparams.BaseCoinUnit)

					// Prepare a concentrated liquidity pool with less liquidity, should be stored since nothing is stored
					clPoolId := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(appparams.BaseCoinUnit, "stake").GetId()
					storedPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, appparams.BaseCoinUnit, "stake")
					s.Require().NoError(err)
					s.Require().Equal(clPoolId, storedPoolId)

					// Return the cl pool id as expected since it has higher liquidity, compare the balancer pool id
					return clPoolId, balancerPoolId
				},
			},
			expectPass: true,
		},
		{
			name: "Catch overflow error when getting newPoolLiquidity - Ensure test doesn't panic",
			param: param{
				baseDenom:  appparams.BaseCoinUnit,
				matchDenom: "stake",
				prepareStateAndGetPoolIdToCompare: func() (uint64, uint64) {
					// Prepare a balancer pool with liquidity levels that will overflow when multiplied
					overflowPoolId := s.PrepareBalancerPoolWithCoins(
						sdk.NewCoin(appparams.BaseCoinUnit, osmomath.Int(osmomath.NewUintFromString("999999999999999999999999999999999999999"))),
						sdk.NewCoin("stake", osmomath.Int(osmomath.NewUintFromString("999999999999999999999999999999999999999"))))

					// Delete all pools for the base denom uosmo so that all tests start with a clean slate
					s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, appparams.BaseCoinUnit)

					// Prepare a balancer pool with normal liquidity
					poolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10)), sdk.NewCoin("stake", osmomath.NewInt(10)))

					// The normal liquidity pool should be stored since the function will return early when catching the overflow error
					return poolId, overflowPoolId
				},
			},
			expectPass: true,
		},
		{
			name: "Catch overflow error when getting storedPoolLiquidity - Ensure test doesn't panic",
			param: param{
				baseDenom:  appparams.BaseCoinUnit,
				matchDenom: "stake",
				prepareStateAndGetPoolIdToCompare: func() (uint64, uint64) {
					// Prepare a balancer pool with normal liquidity
					poolId := s.PrepareBalancerPoolWithCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(10)), sdk.NewCoin("stake", osmomath.NewInt(10)))

					// Delete all pools for the base denom uosmo so that all tests start with a clean slate
					s.App.ProtoRevKeeper.DeleteAllPoolsForBaseDenom(s.Ctx, appparams.BaseCoinUnit)

					// Prepare a balancer pool with liquidity levels that will overflow when multiplied
					overflowPoolId := s.PrepareBalancerPoolWithCoins(
						sdk.NewCoin(appparams.BaseCoinUnit, osmomath.Int(osmomath.NewUintFromString("999999999999999999999999999999999999999"))),
						sdk.NewCoin("stake", osmomath.Int(osmomath.NewUintFromString("999999999999999999999999999999999999999"))))

					// The overflow pool should be stored since the function will return early when catching the overflow error
					return overflowPoolId, poolId
				},
			},
			expectPass: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupPoolsTest()

			// Run any state preparation and get the pool id to compare to the stored pool
			expectedStoredPoolId, comparePoolId := tc.param.prepareStateAndGetPoolIdToCompare()

			// Compare and store the pool
			s.App.ProtoRevKeeper.CompareAndStorePool(s.Ctx, comparePoolId, tc.param.baseDenom, tc.param.matchDenom)

			// Get the stored pool id for the highest liquidity pool in protorev
			storedPoolId, err := s.App.ProtoRevKeeper.GetPoolForDenomPair(s.Ctx, tc.param.baseDenom, tc.param.matchDenom)

			if tc.expectPass {
				s.Require().NoError(err)
				s.Require().Equal(expectedStoredPoolId, storedPoolId)
			}
		})
	}
}
