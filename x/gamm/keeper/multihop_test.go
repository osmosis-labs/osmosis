package keeper_test

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/pool-models/balancer"

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
			keeper := suite.App.GAMMKeeper
			poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1%
			poolZeroSwapFee := sdk.ZeroDec()

			// Prepare pools
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})

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

			calcOutAmountAsSeparateSwaps := func(adjustedPoolSwapFee sdk.Dec) sdk.Coin {
				cacheCtx, _ := suite.Ctx.CacheContext()

				if adjustedPoolSwapFee != poolDefaultSwapFee {
					for _, hop := range test.param.routes {
						err := suite.updatePoolSwapFee(cacheCtx, hop.PoolId, adjustedPoolSwapFee)
						suite.NoError(err, "test: %v", test.name)
					}
				}

				nextTokenIn := test.param.tokenIn
				for _, hop := range test.param.routes {
					tokenOut, err := keeper.SwapExactAmountIn(cacheCtx, suite.TestAccs[0], hop.PoolId, nextTokenIn, hop.TokenOutDenom, sdk.OneInt())
					suite.Require().NoError(err)
					nextTokenIn = sdk.NewCoin(hop.TokenOutDenom, tokenOut)
				}
				return nextTokenIn
			}

			if test.expectPass {
				tokenOutCalculatedAsSeparateSwaps := calcOutAmountAsSeparateSwaps(poolDefaultSwapFee)
				tokenOutCalculatedAsSeparateSwapsWithoutFee := calcOutAmountAsSeparateSwaps(poolZeroSwapFee)

				spotPriceBefore := calcSpotPrice()

				multihopTokenOutAmount, err := keeper.MultihopSwapExactAmountIn(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
				suite.NoError(err, "test: %v", test.name)

				spotPriceAfter := calcSpotPrice()

				// Ratio of the token out should be between the before spot price and after spot price.
				sp := test.param.tokenIn.Amount.ToDec().Quo(multihopTokenOutAmount.ToDec())
				suite.True(sp.GT(spotPriceBefore) && sp.LT(spotPriceAfter), "test: %v", test.name)

				if test.reducedFeeApplied {
					// we have 3 values:
					// ---------------- tokenOutCalculatedAsSeparateSwapsWithoutFee
					//    diffA
					// ---------------- multihopTokenOutAmount (with half fees)
					//    diffB
					// ---------------- tokenOutCalculatedAsSeparateSwaps (with full fees)
					// here we want to test, that difference between this 3 values around 50% (fee is actually halved)
					// we do not have exact 50% reduce due to amm math rounding
					// playing with input values for this test can result in different discount %
					// so lets check, that we have around 50% +-1% reduction
					diffA := tokenOutCalculatedAsSeparateSwapsWithoutFee.Amount.Sub(multihopTokenOutAmount)
					diffB := multihopTokenOutAmount.Sub(tokenOutCalculatedAsSeparateSwaps.Amount)
					diffDistinctionPercent := diffA.Sub(diffB).Abs().ToDec().Quo(diffA.Add(diffB).ToDec())
					suite.Require().True(diffDistinctionPercent.LT(sdk.MustNewDecFromStr("0.01")))
				} else {
					suite.Require().True(multihopTokenOutAmount.Equal(tokenOutCalculatedAsSeparateSwaps.Amount))
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
			keeper := suite.App.GAMMKeeper
			poolDefaultSwapFee := sdk.NewDecWithPrec(1, 2) // 1%
			poolZeroSwapFee := sdk.ZeroDec()

			// Prepare pools
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee, // 1%
				ExitFee: sdk.NewDec(0),
			})
			suite.PrepareBalancerPoolWithPoolParams(balancer.PoolParams{
				SwapFee: poolDefaultSwapFee,
				ExitFee: sdk.NewDec(0),
			})

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

			calcInAmountAsSeparateSwaps := func(adjustedPoolSwapFee sdk.Dec) sdk.Coin {
				cacheCtx, _ := suite.Ctx.CacheContext()

				if adjustedPoolSwapFee != poolDefaultSwapFee {
					for _, hop := range test.param.routes {
						err := suite.updatePoolSwapFee(cacheCtx, hop.PoolId, adjustedPoolSwapFee)
						suite.NoError(err, "test: %v", test.name)
					}
				}

				nextTokenOut := test.param.tokenOut
				for i := len(test.param.routes) - 1; i >= 0; i-- {
					hop := test.param.routes[i]
					tokenOut, err := keeper.SwapExactAmountOut(cacheCtx, suite.TestAccs[0], hop.PoolId, hop.TokenInDenom, sdk.NewInt(100000000), nextTokenOut)
					suite.Require().NoError(err)
					nextTokenOut = sdk.NewCoin(hop.TokenInDenom, tokenOut)
				}
				return nextTokenOut
			}

			if test.expectPass {
				tokenInCalculatedAsSeparateSwaps := calcInAmountAsSeparateSwaps(poolDefaultSwapFee)
				tokenInCalculatedAsSeparateSwapsWithoutFee := calcInAmountAsSeparateSwaps(poolZeroSwapFee)

				spotPriceBefore := calcSpotPrice()
				multihopTokenInAmount, err := keeper.MultihopSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
				suite.Require().NoError(err, "test: %v", test.name)

				spotPriceAfter := calcSpotPrice()
				// Ratio of the token out should be between the before spot price and after spot price.
				// This is because the swap increases the spot price
				sp := multihopTokenInAmount.ToDec().Quo(test.param.tokenOut.Amount.ToDec())
				suite.True(sp.GT(spotPriceBefore) && sp.LT(spotPriceAfter), "multi-hop spot price wrong, test: %v", test.name)

				if test.reducedFeeApplied {
					// we have 3 values:
					// ---------------- tokenInCalculatedAsSeparateSwaps (with full fees)
					//    diffA
					// ---------------- multihopTokenInAmount (with half fees)
					//    diffB
					// ---------------- tokenInCalculatedAsSeparateSwapsWithoutFee
					// here we want to test, that difference between this 3 values around 50% (fee is actually halved)
					// we do not have exact 50% reduce due to amm math rounding
					// playing with input values for this test can result in different discount %
					// so lets check, that we have around 50% +-1% reduction
					diffA := tokenInCalculatedAsSeparateSwaps.Amount.Sub(multihopTokenInAmount)
					diffB := multihopTokenInAmount.Sub(tokenInCalculatedAsSeparateSwapsWithoutFee.Amount)
					diffDistinctionPercent := diffA.Sub(diffB).Abs().ToDec().Quo(diffA.Add(diffB).ToDec())
					suite.Require().True(diffDistinctionPercent.LT(sdk.MustNewDecFromStr("0.01")))
				} else {
					suite.Require().True(multihopTokenInAmount.Equal(tokenInCalculatedAsSeparateSwaps.Amount))
				}
			} else {
				_, err := keeper.MultihopSwapExactAmountOut(suite.Ctx, suite.TestAccs[0], test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
				suite.Error(err, "test: %v", test.name)
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
