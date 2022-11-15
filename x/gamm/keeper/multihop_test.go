package keeper_test

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v12/x/pool-incentives/types"
)

func (suite *KeeperTestSuite) TestBalancerPoolSimpleMultihopSwapExactAmountIn() {
	type param struct {
		routes            []types.SwapAmountInRoute
		poolFee           []sdk.Dec
		tokenIn           sdk.Coin
		tokenOutMinAmount sdk.Int
	}

	poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1% pool swap fee default

	tests := []struct {
		name              string
		param             param
		expectPass        bool
		coinA             sdk.Coin
		coinB             sdk.Coin
		coinC             sdk.Coin
		reducedFeeApplied bool
	}{
		{
			name: "Swap - [foo -> bar](pool 1) - [bar -> baz](pool 2), both pools 1 percent fee",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: "bar",
					},
					{
						PoolId:        2,
						TokenOutDenom: "baz",
					},
				},
				poolFee:           []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			coinA:      sdk.NewCoin("foo", sdk.NewInt(1000000000)),
			coinB:      sdk.NewCoin("bar", sdk.NewInt(1000000000)),
			coinC:      sdk.NewCoin("baz", sdk.NewInt(1000000000)),
			expectPass: true,
		},
		{
			name: "Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, both pools 1 percent fee",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: "uosmo",
					},
					{
						PoolId:        2,
						TokenOutDenom: "baz",
					},
				},
				poolFee:           []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			coinA:             sdk.NewCoin("foo", sdk.NewInt(1000000000)),
			coinB:             sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
			coinC:             sdk.NewCoin("baz", sdk.NewInt(1000000000)),
			reducedFeeApplied: true,
			expectPass:        true,
		},
		{
			name: "Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, (pool 1) 1 percent fee, (pool 2) 10 percent fee",
			param: param{
				routes: []types.SwapAmountInRoute{
					{
						PoolId:        1,
						TokenOutDenom: "uosmo",
					},
					{
						PoolId:        2,
						TokenOutDenom: "baz",
					},
				},
				poolFee:           []sdk.Dec{poolDefaultSwapFee, sdk.NewDecWithPrec(1, 1)},
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			coinA:             sdk.NewCoin("foo", sdk.NewInt(1000000000)),
			coinB:             sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
			coinC:             sdk.NewCoin("baz", sdk.NewInt(1000000000)),
			reducedFeeApplied: true,
			expectPass:        true,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		suite.Run(test.name, func() {
			keeper := suite.App.GAMMKeeper

			// Create pools 1 and 2 with desired swap fee
			suite.PrepareBalancerPoolWithCoins(test.coinA, test.coinB)
			suite.updatePoolSwapFee(suite.Ctx, 1, test.param.poolFee[0])
			suite.PrepareBalancerPoolWithCoins(test.coinB, test.coinC)
			suite.updatePoolSwapFee(suite.Ctx, 2, test.param.poolFee[1])

			// Fund test account with all three assets that pools 1 and 2 consist of
			suite.FundAcc(suite.TestAccs[0], sdk.NewCoins(test.coinA, test.coinB, test.coinC))

			// if we expect a reduced fee to apply, we set both pools in DistrInfo to replicate it being an incentivized pool
			// each pool has three gauges, hence 6 gauges for 2 pools
			if test.reducedFeeApplied {
				test := poolincentivestypes.DistrInfo{
					TotalWeight: sdk.NewInt(6),
					Records: []poolincentivestypes.DistrRecord{
						{GaugeId: 1, Weight: sdk.OneInt()}, {GaugeId: 2, Weight: sdk.OneInt()}, {GaugeId: 3, Weight: sdk.OneInt()},
						{GaugeId: 4, Weight: sdk.OneInt()}, {GaugeId: 5, Weight: sdk.OneInt()}, {GaugeId: 6, Weight: sdk.OneInt()}},
				}
				suite.App.PoolIncentivesKeeper.SetDistrInfo(suite.Ctx, test)
			}

			// calcOutAmountAsSeparateSwaps calculates the multi-hop swap as separate swaps, while also
			// utilizing a cacheContext so the state does not change
			calcOutAmountAsSeparateSwaps := func(osmoFeeReduced bool) sdk.Coin {
				cacheCtx, _ := suite.Ctx.CacheContext()

				if osmoFeeReduced {
					for _, hop := range test.param.routes {
						// determine the expected reduced swap fee
						hopPool, _ := keeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
						currentPoolSwapFee := hopPool.GetSwapFee(cacheCtx)
						route := types.SwapAmountInRoutes(test.param.routes)
						routeSwapFee, sumOfSwapFees, err := keeper.GetOsmoRoutedMultihopTotalSwapFee(suite.Ctx, route)
						suite.Require().NoError(err)
						swapFee := routeSwapFee.Mul((currentPoolSwapFee.Quo(sumOfSwapFees)))
						// update the pools swap fee directly
						err = suite.updatePoolSwapFee(cacheCtx, hop.PoolId, swapFee)
						suite.Require().NoError(err)
					}
				}

				nextTokenIn := test.param.tokenIn
				// we then do individual swaps until we reach the end of the swap route
				for _, hop := range test.param.routes {
					tokenOut, err := keeper.SwapExactAmountIn(cacheCtx, suite.TestAccs[0], hop.PoolId, nextTokenIn, hop.TokenOutDenom, sdk.OneInt())
					suite.Require().NoError(err)
					nextTokenIn = sdk.NewCoin(hop.TokenOutDenom, tokenOut)
				}
				return nextTokenIn
			}

			if test.expectPass {
				var expectedMultihopTokenOutAmount sdk.Coin
				if test.reducedFeeApplied {
					// calculate the swap as separate swaps with a reduced swap fee
					expectedMultihopTokenOutAmount = calcOutAmountAsSeparateSwaps(true)
				} else {
					// calculate the swap as separate swaps with the default swap fee
					expectedMultihopTokenOutAmount = calcOutAmountAsSeparateSwaps(false)
				}
				// compare the expected tokenOut to the actual tokenOut
				multihopTokenOutAmount, err := keeper.MultihopSwapExactAmountIn(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
				suite.Require().NoError(err)
				suite.Require().True(multihopTokenOutAmount.Equal(expectedMultihopTokenOutAmount.Amount))
			} else {
				_, err := keeper.MultihopSwapExactAmountIn(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBalancerPoolSimpleMultihopSwapExactAmountOut() {
	type param struct {
		routes           []types.SwapAmountOutRoute
		poolFee          []sdk.Dec
		tokenInMaxAmount sdk.Int
		tokenOut         sdk.Coin
	}

	poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1% pool swap fee default

	tests := []struct {
		name              string
		param             param
		coinA             sdk.Coin
		coinB             sdk.Coin
		coinC             sdk.Coin
		expectPass        bool
		reducedFeeApplied bool
	}{
		{
			name: "Swap - [foo -> bar](pool 1) - [bar -> baz](pool 2), both pools 1 percent fee",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: "foo",
					},
					{
						PoolId:       2,
						TokenInDenom: "bar",
					},
				},
				poolFee:          []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			coinA:      sdk.NewCoin("foo", sdk.NewInt(1000000000)),
			coinB:      sdk.NewCoin("bar", sdk.NewInt(1000000000)),
			coinC:      sdk.NewCoin("baz", sdk.NewInt(1000000000)),
			expectPass: true,
		},
		{
			name: "Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, both pools 1 percent fee",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: "foo",
					},
					{
						PoolId:       2,
						TokenInDenom: "uosmo",
					},
				},
				poolFee:          []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			coinA:             sdk.NewCoin("foo", sdk.NewInt(1000000000)),
			coinB:             sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
			coinC:             sdk.NewCoin("baz", sdk.NewInt(1000000000)),
			expectPass:        true,
			reducedFeeApplied: true,
		},
		{
			name: "Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) with a half fee applied, (pool 1) 1 percent fee, (pool 2) 10 percent fee",
			param: param{
				routes: []types.SwapAmountOutRoute{
					{
						PoolId:       1,
						TokenInDenom: "foo",
					},
					{
						PoolId:       2,
						TokenInDenom: "uosmo",
					},
				},
				poolFee:          []sdk.Dec{poolDefaultSwapFee, sdk.NewDecWithPrec(1, 1)},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			coinA:             sdk.NewCoin("foo", sdk.NewInt(1000000000)),
			coinB:             sdk.NewCoin("uosmo", sdk.NewInt(1000000000)),
			coinC:             sdk.NewCoin("baz", sdk.NewInt(1000000000)),
			expectPass:        true,
			reducedFeeApplied: true,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			keeper := suite.App.GAMMKeeper

			// Create pools 1 and 2 with desired swap fee
			suite.PrepareBalancerPoolWithCoins(test.coinA, test.coinB)
			suite.updatePoolSwapFee(suite.Ctx, 1, test.param.poolFee[0])
			suite.PrepareBalancerPoolWithCoins(test.coinB, test.coinC)
			suite.updatePoolSwapFee(suite.Ctx, 2, test.param.poolFee[1])

			// Fund test account with all three assets that pools 1 and 2 consist of
			suite.FundAcc(suite.TestAccs[0], sdk.NewCoins(test.coinA, test.coinB, test.coinC))

			// if we expect a reduced fee to apply, we set both pools in DistrInfo to replicate it being an incentivized pool
			// each pool has three gauges, hence 6 gauges for 2 pools
			if test.reducedFeeApplied {
				test := poolincentivestypes.DistrInfo{
					TotalWeight: sdk.NewInt(2),
					Records: []poolincentivestypes.DistrRecord{
						{GaugeId: 1, Weight: sdk.OneInt()}, {GaugeId: 2, Weight: sdk.OneInt()}, {GaugeId: 3, Weight: sdk.OneInt()},
						{GaugeId: 4, Weight: sdk.OneInt()}, {GaugeId: 5, Weight: sdk.OneInt()}, {GaugeId: 6, Weight: sdk.OneInt()}},
				}
				suite.App.PoolIncentivesKeeper.SetDistrInfo(suite.Ctx, test)
			}

			// calcInAmountAsSeparateSwaps calculates the multi-hop swap as separate swaps, while also
			// utilizing a cacheContext so the state does not change
			calcInAmountAsSeparateSwaps := func(osmoFeeReduced bool) sdk.Coin {
				cacheCtx, _ := suite.Ctx.CacheContext()

				if osmoFeeReduced {
					for _, hop := range test.param.routes {
						// determine the expected reduced swap fee
						hopPool, _ := keeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
						currentPoolSwapFee := hopPool.GetSwapFee(cacheCtx)
						route := types.SwapAmountOutRoutes(test.param.routes)
						routeSwapFee, sumOfSwapFees, err := keeper.GetOsmoRoutedMultihopTotalSwapFee(suite.Ctx, route)
						suite.Require().NoError(err)
						swapFee := routeSwapFee.Mul((currentPoolSwapFee.Quo(sumOfSwapFees)))
						// update the pools swap fee directly
						err = suite.updatePoolSwapFee(cacheCtx, hop.PoolId, swapFee)
						suite.Require().NoError(err)
					}
				}

				nextTokenOut := test.param.tokenOut
				// we then do individual swaps until we reach the end of the swap route
				for i := len(test.param.routes) - 1; i >= 0; i-- {
					hop := test.param.routes[i]
					tokenOut, err := keeper.SwapExactAmountOut(cacheCtx, suite.TestAccs[0], hop.PoolId, hop.TokenInDenom, sdk.NewInt(100000000), nextTokenOut)
					suite.Require().NoError(err)
					nextTokenOut = sdk.NewCoin(hop.TokenInDenom, tokenOut)
				}
				return nextTokenOut
			}

			if test.expectPass {
				var expectedMultihopTokenOutAmount sdk.Coin
				if test.reducedFeeApplied {
					// calculate the swap as separate swaps with the new reduced swap fee
					expectedMultihopTokenOutAmount = calcInAmountAsSeparateSwaps(true)
				} else {
					// calculate the swap as separate swaps with the default swap fee
					expectedMultihopTokenOutAmount = calcInAmountAsSeparateSwaps(false)
				}
				// compare the expected tokenOut to the actual tokenOut
				multihopTokenOutAmount, err := keeper.MultihopSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
				suite.Require().NoError(err)
				suite.Require().True(multihopTokenOutAmount.Equal(expectedMultihopTokenOutAmount.Amount))
			} else {
				_, err := keeper.MultihopSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
				suite.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) updatePoolSwapFee(ctx sdk.Context, poolId uint64, adjustedPoolSwapFee sdk.Dec) error {
	pool, err := s.App.GAMMKeeper.GetPoolAndPoke(ctx, poolId)
	if err != nil {
		return err
	}

	balancerPool, ok := pool.(*balancer.Pool)
	if !ok {
		return errors.New("can't update swap fee on non-balancer pool")
	}
	balancerPool.PoolParams.SwapFee = adjustedPoolSwapFee
	return s.App.GAMMKeeper.SetPool(ctx, pool)
}
