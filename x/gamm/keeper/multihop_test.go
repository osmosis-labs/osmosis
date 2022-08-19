package keeper_test

import (
	"github.com/osmosis-labs/osmosis/v11/x/gamm/pool-models/balancer"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func (suite *KeeperTestSuite) TestBalancerPoolSimpleMultihopSwapExactAmountIn() {
	type param struct {
		routes            []types.SwapAmountInRoute
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
			// Prepare 2 pools pairs
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2), // 1%
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 2),
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})

			keeper := suite.App.GAMMKeeper

			if test.expectPass {
				// Calculate the chained spot price.
				calcSpotPrice := func() sdk.Dec {
					dec := sdk.NewDec(1)
					tokenInDenom := test.param.tokenIn.Denom
					for i, route := range test.param.routes {
						if i != 0 {
							tokenInDenom = test.param.routes[i-1].TokenOutDenom
						}
						pool, err := keeper.GetPoolAndPoke(suite.Ctx, route.PoolId)
						suite.NoError(err, "test: %v", test.name)

						sp, err := pool.SpotPrice(suite.Ctx, tokenInDenom, route.TokenOutDenom)
						suite.NoError(err, "test: %v", test.name)
						dec = dec.Mul(sp)
					}
					return dec
				}

				// we create exact the same except swap fee 2 pool pairs
				// use +2 to calc using second pair (w/o swap fee)
				calcOutAmountAsSeparateSwaps := func(poolIdShift uint64) sdk.Coin {
					cacheCtx, _ := suite.Ctx.CacheContext()
					nextTokenIn := test.param.tokenIn
					for _, hop := range test.param.routes {

						tokenOut, err := keeper.SwapExactAmountIn(cacheCtx, suite.TestAccs[0], hop.PoolId+poolIdShift, nextTokenIn, hop.TokenOutDenom, sdk.OneInt())
						suite.Require().NoError(err)
						nextTokenIn = sdk.NewCoin(hop.TokenOutDenom, tokenOut)
					}
					return nextTokenIn
				}

				tokenOutCalculatedAsSeparateSwaps := calcOutAmountAsSeparateSwaps(0)
				tokenOutCalculatedAsSeparateSwapsWithoutFee := calcOutAmountAsSeparateSwaps(2)

				spotPriceBefore := calcSpotPrice()

				tokenOutAmount, err := keeper.MultihopSwapExactAmountIn(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
				suite.NoError(err, "test: %v", test.name)

				spotPriceAfter := calcSpotPrice()

				// Ratio of the token out should be between the before spot price and after spot price.
				sp := test.param.tokenIn.Amount.ToDec().Quo(tokenOutAmount.ToDec())
				suite.True(sp.GT(spotPriceBefore) && sp.LT(spotPriceAfter), "test: %v", test.name)

				if test.reducedFeeApplied {
					// here we do not have exact 50% reduce due to amm math rounding and other staff
					// playing with input values for this test can result in different discount %
					// so lets check, that we have around 50% +-1% reduction
					diffA := tokenOutCalculatedAsSeparateSwapsWithoutFee.Amount.Sub(tokenOutAmount)
					diffB := tokenOutAmount.Sub(tokenOutCalculatedAsSeparateSwaps.Amount)
					diffDistinctionPercent := diffA.Sub(diffB).Abs().ToDec().Quo(diffA.Add(diffB).ToDec())
					suite.Require().True(diffDistinctionPercent.LT(sdk.MustNewDecFromStr("0.01")))
				} else {
					suite.Require().True(tokenOutAmount.Equal(tokenOutCalculatedAsSeparateSwaps.Amount))
				}

			} else {
				_, err := keeper.MultihopSwapExactAmountIn(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
				suite.Error(err, "test: %v", test.name)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestBalancerPoolSimpleMultihopSwapExactAmountOut() {
	type param struct {
		routes           []types.SwapAmountOutRoute
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
			// Prepare 2 pools pairs
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 3), // 1%
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDecWithPrec(1, 3),
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: sdk.NewDec(0),
				ExitFee: sdk.NewDec(0),
			})

			keeper := suite.App.GAMMKeeper

			if test.expectPass {
				// Calculate the chained spot price.
				calcSpotPrice := func() sdk.Dec {
					dec := sdk.NewDec(1)
					for i, route := range test.param.routes {
						tokenOutDenom := test.param.tokenOut.Denom
						if i != len(test.param.routes)-1 {
							tokenOutDenom = test.param.routes[i+1].TokenInDenom
						}

						pool, err := keeper.GetPoolAndPoke(suite.Ctx, route.PoolId)
						suite.NoError(err, "test: %v", test.name)

						sp, err := pool.SpotPrice(suite.Ctx, route.TokenInDenom, tokenOutDenom)
						suite.NoError(err, "test: %v", test.name)
						dec = dec.Mul(sp)
					}
					return dec
				}

				// we create exact the same except swap fee 2 pool pairs
				// use +2 to calc using second pair (w/o swap fee)
				calcInAmountAsSeparateSwaps := func(poolIdShift uint64) sdk.Coin {
					cacheCtx, _ := suite.Ctx.CacheContext()
					nextTokenOut := test.param.tokenOut
					for i := len(test.param.routes) - 1; i >= 0; i-- {
						hop := test.param.routes[i]
						tokenOut, err := keeper.SwapExactAmountOut(cacheCtx, suite.TestAccs[0], hop.PoolId+poolIdShift, hop.TokenInDenom, sdk.NewInt(100000000), nextTokenOut)
						suite.Require().NoError(err)
						nextTokenOut = sdk.NewCoin(hop.TokenInDenom, tokenOut)
					}
					return nextTokenOut
				}

				tokenInCalculatedAsSeparateSwaps := calcInAmountAsSeparateSwaps(0)
				tokenInCalculatedAsSeparateSwapsWithoutFee := calcInAmountAsSeparateSwaps(2)

				spotPriceBefore := calcSpotPrice()
				tokenInAmount, err := keeper.MultihopSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
				suite.Require().NoError(err, "test: %v", test.name)

				spotPriceAfter := calcSpotPrice()
				// Ratio of the token out should be between the before spot price and after spot price.
				// This is because the swap increases the spot price
				sp := tokenInAmount.ToDec().Quo(test.param.tokenOut.Amount.ToDec())
				suite.True(sp.GT(spotPriceBefore) && sp.LT(spotPriceAfter), "multi-hop spot price wrong, test: %v", test.name)

				if test.reducedFeeApplied {
					// here we do not have exact 50% reduce due to amm math rounding and other staff
					// playing with input values for this test can result in different discount %
					// so lets check, that we have around 50% +-1% reduction
					diffA := tokenInCalculatedAsSeparateSwaps.Amount.Sub(tokenInAmount)
					diffB := tokenInAmount.Sub(tokenInCalculatedAsSeparateSwapsWithoutFee.Amount)
					diffDistinctionPercent := diffA.Sub(diffB).Abs().ToDec().Quo(diffA.Add(diffB).ToDec())
					suite.Require().True(diffDistinctionPercent.LT(sdk.MustNewDecFromStr("0.01")))
				} else {
					suite.Require().True(tokenInAmount.Equal(tokenInCalculatedAsSeparateSwaps.Amount))
				}
			} else {
				_, err := keeper.MultihopSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
				suite.Error(err, "test: %v", test.name)
			}
		})
	}
}
