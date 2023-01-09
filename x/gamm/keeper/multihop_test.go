package keeper_test

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v14/x/gamm/pool-models/balancer"

	"github.com/osmosis-labs/osmosis/v14/x/gamm/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v14/x/pool-incentives/types"
)

var (
	fooCoin   = sdk.NewCoin("foo", sdk.NewInt(1000000000))
	barCoin   = sdk.NewCoin("bar", sdk.NewInt(1000000000))
	bazCoin   = sdk.NewCoin("baz", sdk.NewInt(1000000000))
	uosmoCoin = sdk.NewCoin("uosmo", sdk.NewInt(1000000000))
)

func (suite *KeeperTestSuite) TestBalancerPoolMultihopSwapExactAmountIn() {
	type param struct {
		routes             []types.SwapAmountInRoute
		incentivizedGauges []uint64
		fourAssetPools     int
		poolAssets         []sdk.Coins
		poolFee            []sdk.Dec
		tokenIn            sdk.Coin
		tokenOutMinAmount  sdk.Int
	}

	poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1% pool swap fee default

	tests := []struct {
		name                    string
		param                   param
		expectPass              bool
		expectReducedFeeApplied bool
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
				incentivizedGauges: []uint64{},
				poolAssets:         []sdk.Coins{sdk.NewCoins(fooCoin, barCoin), sdk.NewCoins(barCoin, bazCoin)},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenIn:            sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount:  sdk.NewInt(1),
			},
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
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolAssets:         []sdk.Coins{sdk.NewCoins(fooCoin, uosmoCoin), sdk.NewCoins(uosmoCoin, bazCoin)},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenIn:            sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount:  sdk.NewInt(1),
			},
			expectReducedFeeApplied: true,
			expectPass:              true,
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
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolAssets:         []sdk.Coins{sdk.NewCoins(fooCoin, uosmoCoin), sdk.NewCoins(uosmoCoin, bazCoin)},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, sdk.NewDecWithPrec(1, 1)},
				tokenIn:            sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount:  sdk.NewInt(1),
			},
			expectReducedFeeApplied: true,
			expectPass:              true,
		},
		{
			name: "Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
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
					{
						PoolId:        3,
						TokenOutDenom: "bar",
					},
				},
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolAssets:         []sdk.Coins{sdk.NewCoins(fooCoin, uosmoCoin), sdk.NewCoins(uosmoCoin, bazCoin), sdk.NewCoins(bazCoin, barCoin)},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee, poolDefaultSwapFee},
				tokenIn:            sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount:  sdk.NewInt(1),
			},
			expectReducedFeeApplied: false,
			expectPass:              true,
		},
		{
			name: "Swap between four asset pools - [foo -> bar](pool 1) - [bar -> baz](pool 2), all pools 1 percent fee",
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
				fourAssetPools:     2,
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenIn:            sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount:  sdk.NewInt(1),
			},
			expectReducedFeeApplied: false,
			expectPass:              true,
		},
		{
			name: "Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2), with a half fee applied, both pools 1 percent fee",
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
				fourAssetPools:     2,
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenIn:            sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount:  sdk.NewInt(1),
			},
			expectReducedFeeApplied: true,
			expectPass:              true,
		},
		{
			name: "Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
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
					{
						PoolId:        3,
						TokenOutDenom: "bar",
					},
				},
				fourAssetPools:     3,
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee, poolDefaultSwapFee},
				tokenIn:            sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount:  sdk.NewInt(1),
			},
			expectReducedFeeApplied: false,
			expectPass:              true,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		suite.Run(test.name, func() {
			keeper := suite.App.GAMMKeeper

			// Create two asset pools with desired swap fee, if specified. Fund acc with respective coins
			for i, coin := range test.param.poolAssets {
				suite.PrepareBalancerPoolWithCoins(coin[0], coin[1])
				suite.FundAcc(suite.TestAccs[0], sdk.NewCoins(coin[0], coin[1]))
				suite.updatePoolSwapFee(suite.Ctx, uint64(i+1), test.param.poolFee[i])
			}

			// Create four asset pools with desired swap fee, if specified.
			for i := 0; i < test.param.fourAssetPools; i++ {
				suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
					SwapFee: test.param.poolFee[i],
					ExitFee: sdk.NewDec(0),
				})
			}

			// if test specifies incentivized gauges, set them here
			if len(test.param.incentivizedGauges) > 0 {
				var records []poolincentivestypes.DistrRecord
				totalWeight := sdk.NewInt(int64(len(test.param.incentivizedGauges)))
				for _, gauge := range test.param.incentivizedGauges {
					records = append(records, poolincentivestypes.DistrRecord{GaugeId: gauge, Weight: sdk.OneInt()})
				}
				distInfo := poolincentivestypes.DistrInfo{
					TotalWeight: totalWeight,
					Records:     records,
				}
				suite.App.PoolIncentivesKeeper.SetDistrInfo(suite.Ctx, distInfo)
			}

			// calcOutAmountAsSeparateSwaps calculates the multi-hop swap as separate swaps, while also
			// utilizing a cacheContext so the state does not change
			calcOutAmountAsSeparateSwaps := func(osmoFeeReduced bool) sdk.Coin {
				cacheCtx, _ := suite.Ctx.CacheContext()

				if osmoFeeReduced {
					// extract route from swap
					route := types.SwapAmountInRoutes(test.param.routes)
					// utilizing the extracted route, determine the routeSwapFee and sumOfSwapFees
					// these two variables are used to calculate the overall swap fee utilizing the following formula
					// swapFee = routeSwapFee * ((pool_fee) / (sumOfSwapFees))
					routeSwapFee, sumOfSwapFees, err := keeper.GetOsmoRoutedMultihopTotalSwapFee(suite.Ctx, route)
					suite.Require().NoError(err)
					for _, hop := range test.param.routes {
						// extract the current pool's swap fee
						hopPool, err := keeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
						suite.Require().NoError(err)
						currentPoolSwapFee := hopPool.GetSwapFee(cacheCtx)
						// utilize the routeSwapFee, sumOfSwapFees, and current pool swap fee to calculate the new reduced swap fee
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
				// calculate the swap as separate swaps with either the reduced swap fee or normal fee
				expectedMultihopTokenOutAmount = calcOutAmountAsSeparateSwaps(test.expectReducedFeeApplied)
				// compare the expected tokenOut to the actual tokenOut
				multihopTokenOutAmount, err := keeper.MultihopSwapExactAmountIn(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
				suite.Require().NoError(err)
				suite.Require().Equal(expectedMultihopTokenOutAmount.Amount.String(), multihopTokenOutAmount.String())
			} else {
				_, err := keeper.MultihopSwapExactAmountIn(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBalancerPoolMultihopSwapExactAmountOut() {
	type param struct {
		routes             []types.SwapAmountOutRoute
		incentivizedGauges []uint64
		fourAssetPools     int
		poolAssets         []sdk.Coins
		poolFee            []sdk.Dec
		tokenInMaxAmount   sdk.Int
		tokenOut           sdk.Coin
	}

	poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1% pool swap fee default

	tests := []struct {
		name                    string
		param                   param
		expectPass              bool
		expectReducedFeeApplied bool
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
				incentivizedGauges: []uint64{},
				poolAssets:         []sdk.Coins{sdk.NewCoins(fooCoin, barCoin), sdk.NewCoins(barCoin, bazCoin)},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenInMaxAmount:   sdk.NewInt(90000000),
				tokenOut:           sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
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
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolAssets:         []sdk.Coins{sdk.NewCoins(fooCoin, uosmoCoin), sdk.NewCoins(uosmoCoin, bazCoin)},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenInMaxAmount:   sdk.NewInt(90000000),
				tokenOut:           sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			expectPass:              true,
			expectReducedFeeApplied: true,
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
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolAssets:         []sdk.Coins{sdk.NewCoins(fooCoin, uosmoCoin), sdk.NewCoins(uosmoCoin, bazCoin)},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, sdk.NewDecWithPrec(1, 1)},
				tokenInMaxAmount:   sdk.NewInt(90000000),
				tokenOut:           sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			expectPass:              true,
			expectReducedFeeApplied: true,
		},
		{
			name: "Swap - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
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
					{
						PoolId:       3,
						TokenInDenom: "baz",
					},
				},
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolAssets:         []sdk.Coins{sdk.NewCoins(fooCoin, uosmoCoin), sdk.NewCoins(uosmoCoin, bazCoin), sdk.NewCoins(bazCoin, barCoin)},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee, poolDefaultSwapFee},
				tokenInMaxAmount:   sdk.NewInt(90000000),
				tokenOut:           sdk.NewCoin("bar", sdk.NewInt(100000)),
			},
			expectReducedFeeApplied: false,
			expectPass:              true,
		},
		{
			name: "Swap between four asset pools - [foo -> bar](pool 1) - [bar -> baz](pool 2), all pools 1 percent fee",
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
				fourAssetPools:     2,
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenOut:           sdk.NewCoin("baz", sdk.NewInt(100000)),
				tokenInMaxAmount:   sdk.NewInt(90000000),
			},
			expectReducedFeeApplied: false,
			expectPass:              true,
		},
		{
			name: "Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2), with a half fee applied, both pools 1 percent fee",
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
				fourAssetPools:     2,
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee},
				tokenOut:           sdk.NewCoin("baz", sdk.NewInt(100000)),
				tokenInMaxAmount:   sdk.NewInt(90000000),
			},
			expectReducedFeeApplied: true,
			expectPass:              true,
		},
		{
			name: "Swap between four asset pools - [foo -> uosmo](pool 1) - [uosmo -> baz](pool 2) - [baz -> bar](pool 3), all pools 1 percent fee",
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
					{
						PoolId:       3,
						TokenInDenom: "baz",
					},
				},
				fourAssetPools:     3,
				incentivizedGauges: []uint64{1, 2, 3, 4, 5, 6, 7, 8, 9},
				poolFee:            []sdk.Dec{poolDefaultSwapFee, poolDefaultSwapFee, poolDefaultSwapFee},
				tokenOut:           sdk.NewCoin("bar", sdk.NewInt(100000)),
				tokenInMaxAmount:   sdk.NewInt(90000000),
			},
			expectReducedFeeApplied: false,
			expectPass:              true,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		suite.Run(test.name, func() {
			keeper := suite.App.GAMMKeeper

			// Create two asset pools with desired swap fee, if specified. Fund acc with respective coins
			for i, coin := range test.param.poolAssets {
				suite.PrepareBalancerPoolWithCoins(coin[0], coin[1])
				suite.FundAcc(suite.TestAccs[0], sdk.NewCoins(coin[0], coin[1]))
				suite.updatePoolSwapFee(suite.Ctx, uint64(i+1), test.param.poolFee[i])
			}

			// Create four asset pools with desired swap fee, if specified.
			for i := 0; i < test.param.fourAssetPools; i++ {
				suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
					SwapFee: test.param.poolFee[i],
					ExitFee: sdk.NewDec(0),
				})
			}

			// if test specifies incentivized gauges, set them here
			if len(test.param.incentivizedGauges) > 0 {
				var records []poolincentivestypes.DistrRecord
				totalWeight := sdk.NewInt(int64(len(test.param.incentivizedGauges)))
				for _, gauge := range test.param.incentivizedGauges {
					records = append(records, poolincentivestypes.DistrRecord{GaugeId: gauge, Weight: sdk.OneInt()})
				}
				distInfo := poolincentivestypes.DistrInfo{
					TotalWeight: totalWeight,
					Records:     records,
				}
				suite.App.PoolIncentivesKeeper.SetDistrInfo(suite.Ctx, distInfo)
			}

			// calcInAmountAsSeparateSwaps calculates the multi-hop swap as separate swaps, while also
			// utilizing a cacheContext so the state does not change
			calcInAmountAsSeparateSwaps := func(osmoFeeReduced bool) sdk.Coin {
				cacheCtx, _ := suite.Ctx.CacheContext()

				if osmoFeeReduced {
					// extract route from swap
					route := types.SwapAmountOutRoutes(test.param.routes)
					// utilizing the extracted route, determine the routeSwapFee and sumOfSwapFees
					// these two variables are used to calculate the overall swap fee utilizing the following formula
					// swapFee = routeSwapFee * ((pool_fee) / (sumOfSwapFees))
					routeSwapFee, sumOfSwapFees, err := keeper.GetOsmoRoutedMultihopTotalSwapFee(suite.Ctx, route)
					suite.Require().NoError(err)
					for _, hop := range test.param.routes {
						// extract the current pool's swap fee
						hopPool, err := keeper.GetPoolAndPoke(cacheCtx, hop.PoolId)
						suite.Require().NoError(err)
						currentPoolSwapFee := hopPool.GetSwapFee(cacheCtx)
						// utilize the routeSwapFee, sumOfSwapFees, and current pool swap fee to calculate the new reduced swap fee
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
				// calculate the swap as separate swaps with either the reduced swap fee or normal fee
				expectedMultihopTokenOutAmount = calcInAmountAsSeparateSwaps(test.expectReducedFeeApplied)
				// compare the expected tokenOut to the actual tokenOut
				multihopTokenOutAmount, err := keeper.MultihopSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
				suite.Require().NoError(err)
				suite.Require().Equal(expectedMultihopTokenOutAmount.Amount.String(), multihopTokenOutAmount.String())
			} else {
				_, err := keeper.MultihopSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
				suite.Require().NoError(err)
			}
		})
	}
}

// TestEstimateMultihopSwapExactAmountIn tests that the estimation done via `EstimateSwapExactAmountIn`
// results in the same amount of token out as the actual swap.
func (suite *KeeperTestSuite) TestEstimateMultihopSwapExactAmountIn() {
	type param struct {
		routes            []types.SwapAmountInRoute
		estimateRoutes    []types.SwapAmountInRoute
		tokenIn           sdk.Coin
		tokenOutMinAmount sdk.Int
	}

	tests := []struct {
		name              string
		param             param
		expectPass        bool
		reducedFeeApplied bool
	}{
		{
			name: "Proper swap - foo -> bar(pool 1) - bar(pool 2) -> baz",
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
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: "bar",
					},
					{
						PoolId:        4,
						TokenOutDenom: "baz",
					},
				},
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
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
				estimateRoutes: []types.SwapAmountInRoute{
					{
						PoolId:        3,
						TokenOutDenom: "uosmo",
					},
					{
						PoolId:        4,
						TokenOutDenom: "baz",
					},
				},
				tokenIn:           sdk.NewCoin("foo", sdk.NewInt(100000)),
				tokenOutMinAmount: sdk.NewInt(1),
			},
			reducedFeeApplied: true,
			expectPass:        true,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			keeper := suite.App.GAMMKeeper
			poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1%

			// Prepare 4 pools,
			// Two pools for calculating `MultihopSwapExactAmountIn`
			// and two pools for calculating `EstimateMultihopSwapExactAmountIn`
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})

			firstEstimatePoolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})
			secondEstimatePoolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})

			firstEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			// calculate token out amount using `MultihopSwapExactAmountIn`
			multihopTokenOutAmount, err := keeper.MultihopSwapExactAmountIn(
				suite.Ctx,
				suite.TestAccs[0],
				test.param.routes,
				test.param.tokenIn,
				test.param.tokenOutMinAmount)
			suite.Require().NoError(err)

			// calculate token out amount using `EstimateMultihopSwapExactAmountIn`
			estimateMultihopTokenOutAmount, err := keeper.MultihopEstimateOutGivenExactAmountIn(
				suite.Ctx,
				test.param.estimateRoutes,
				test.param.tokenIn)
			suite.Require().NoError(err)

			// ensure that the token out amount is same
			suite.Require().Equal(multihopTokenOutAmount, estimateMultihopTokenOutAmount)

			// ensure that pool state has not been altered after estimation
			firstEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			suite.Require().Equal(firstEstimatePool, firstEstimatePoolAfterSwap)
			suite.Require().Equal(secondEstimatePool, secondEstimatePoolAfterSwap)
		})
	}
}

// TestEstimateMultihopSwapExactAmountOut tests that the estimation done via `EstimateSwapExactAmountOut`
// results in the same amount of token in as the actual swap.
func (suite *KeeperTestSuite) TestEstimateMultihopSwapExactAmountOut() {
	type param struct {
		routes           []types.SwapAmountOutRoute
		estimateRoutes   []types.SwapAmountOutRoute
		tokenInMaxAmount sdk.Int
		tokenOut         sdk.Coin
	}

	tests := []struct {
		name              string
		param             param
		expectPass        bool
		reducedFeeApplied bool
	}{
		{
			name: "Proper swap: foo -> bar (pool 1), bar -> baz (pool 2)",
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
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: "foo",
					},
					{
						PoolId:       4,
						TokenInDenom: "bar",
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			expectPass: true,
		},
		{
			name: "Swap - foo -> uosmo(pool 1) - uosmo(pool 2) -> baz with a half fee applied",
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
				estimateRoutes: []types.SwapAmountOutRoute{
					{
						PoolId:       3,
						TokenInDenom: "foo",
					},
					{
						PoolId:       4,
						TokenInDenom: "uosmo",
					},
				},
				tokenInMaxAmount: sdk.NewInt(90000000),
				tokenOut:         sdk.NewCoin("baz", sdk.NewInt(100000)),
			},
			expectPass:        true,
			reducedFeeApplied: true,
		},
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		suite.Run(test.name, func() {
			keeper := suite.App.GAMMKeeper
			poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1%

			// Prepare 4 pools,
			// Two pools for calculating `MultihopSwapExactAmountOut`
			// and two pools for calculating `EstimateMultihopSwapExactAmountOut`
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee, // 1%
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})
			firstEstimatePoolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee, // 1%
				ExitFee: sdk.NewDec(0),
			})
			secondEstimatePoolId := suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})
			firstEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			multihopTokenInAmount, err := keeper.MultihopSwapExactAmountOut(
				suite.Ctx,
				suite.TestAccs[0],
				test.param.routes,
				test.param.tokenInMaxAmount,
				test.param.tokenOut)
			suite.Require().NoError(err, "test: %v", test.name)

			estimateMultihopTokenInAmount, err := keeper.MultihopEstimateInGivenExactAmountOut(
				suite.Ctx,
				test.param.estimateRoutes,
				test.param.tokenOut)
			suite.Require().NoError(err, "test: %v", test.name)

			suite.Require().Equal(multihopTokenInAmount, estimateMultihopTokenInAmount)

			// ensure that pool state has not been altered after estimation
			firstEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, firstEstimatePoolId)
			suite.Require().NoError(err)
			secondEstimatePoolAfterSwap, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, secondEstimatePoolId)
			suite.Require().NoError(err)

			suite.Require().Equal(firstEstimatePool, firstEstimatePoolAfterSwap)
			suite.Require().Equal(secondEstimatePool, secondEstimatePoolAfterSwap)
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
