package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

func (suite *KeeperTestSuite) TestBalancerPoolSimpleMultihopSwapExactAmountIn() {
	type param struct {
		routes            []types.SwapAmountInRoute
		tokenIn           sdk.Coin
		tokenOutMinAmount sdk.Int
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
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
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		// Prepare 2 pools
		suite.prepareBalancerPool()
		suite.prepareBalancerPool()

		keeper := suite.app.GAMMKeeper

		if test.expectPass {
			// Calculate the chained spot price.
			spotPriceBefore := func() sdk.Dec {
				dec := sdk.NewDec(1)
				tokenInDenom := test.param.tokenIn.Denom
				for i, route := range test.param.routes {
					if i != 0 {
						tokenInDenom = test.param.routes[i-1].TokenOutDenom
					}

					sp, err := keeper.CalculateSpotPriceWithSwapFee(suite.ctx, route.PoolId, tokenInDenom, route.TokenOutDenom)
					suite.NoError(err, "test: %v", test.name)
					dec = dec.Mul(sp)
				}
				return dec
			}()

			tokenOutAmount, err := keeper.MultihopSwapExactAmountIn(suite.ctx, acc1, test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
			suite.NoError(err, "test: %v", test.name)

			// Calculate the chained spot price.
			spotPriceAfter := func() sdk.Dec {
				dec := sdk.NewDec(1)
				tokenInDenom := test.param.tokenIn.Denom
				for i, route := range test.param.routes {
					if i != 0 {
						tokenInDenom = test.param.routes[i-1].TokenOutDenom
					}

					sp, err := keeper.CalculateSpotPriceWithSwapFee(suite.ctx, route.PoolId, tokenInDenom, route.TokenOutDenom)
					suite.NoError(err, "test: %v", test.name)
					dec = dec.Mul(sp)
				}
				return dec
			}()

			// Ratio of the token out should be between the before spot price and after spot price.
			sp := test.param.tokenIn.Amount.ToDec().Quo(tokenOutAmount.ToDec())
			suite.True(sp.GT(spotPriceBefore) && sp.LT(spotPriceAfter), "test: %v", test.name)
		} else {
			_, err := keeper.MultihopSwapExactAmountIn(suite.ctx, acc1, test.param.routes, test.param.tokenIn, test.param.tokenOutMinAmount)
			suite.Error(err, "test: %v", test.name)
		}
	}
}

func (suite *KeeperTestSuite) TestBalancerPoolSimpleMultihopSwapExactAmountOut() {
	type param struct {
		routes           []types.SwapAmountOutRoute
		tokenInMaxAmount sdk.Int
		tokenOut         sdk.Coin
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "Proper swap - foo -> bar(pool 1) - bar(pool 2) -> baz",
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
	}

	for _, test := range tests {
		// Init suite for each test.
		suite.SetupTest()

		// Prepare 2 pools
		suite.prepareBalancerPool()
		suite.prepareBalancerPool()

		keeper := suite.app.GAMMKeeper

		if test.expectPass {
			// Calculate the chained spot price.
			spotPriceBefore := func() sdk.Dec {
				dec := sdk.NewDec(1)
				for i, route := range test.param.routes {
					tokenOutDenom := test.param.tokenOut.Denom
					if i != len(test.param.routes)-1 {
						tokenOutDenom = test.param.routes[i+1].TokenInDenom
					}

					sp, err := keeper.CalculateSpotPriceWithSwapFee(suite.ctx, route.PoolId, route.TokenInDenom, tokenOutDenom)
					suite.NoError(err, "test: %v", test.name)
					dec = dec.Mul(sp)
				}
				return dec
			}()

			tokenInAmount, err := keeper.MultihopSwapExactAmountOut(suite.ctx, acc1, test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
			suite.NoError(err, "test: %v", test.name)

			// Calculate the chained spot price.
			spotPriceAfter := func() sdk.Dec {
				dec := sdk.NewDec(1)
				for i, route := range test.param.routes {
					tokenOutDenom := test.param.tokenOut.Denom
					if i != len(test.param.routes)-1 {
						tokenOutDenom = test.param.routes[i+1].TokenInDenom
					}

					sp, err := keeper.CalculateSpotPriceWithSwapFee(suite.ctx, route.PoolId, route.TokenInDenom, tokenOutDenom)
					suite.NoError(err, "test: %v", test.name)
					dec = dec.Mul(sp)
				}
				return dec
			}()

			// Ratio of the token out should be between the before spot price and after spot price.
			sp := tokenInAmount.ToDec().Quo(test.param.tokenOut.Amount.ToDec())
			suite.True(sp.GT(spotPriceBefore) && sp.LT(spotPriceAfter), "test: %v", test.name)
		} else {
			_, err := keeper.MultihopSwapExactAmountOut(suite.ctx, acc1, test.param.routes, test.param.tokenInMaxAmount, test.param.tokenOut)
			suite.Error(err, "test: %v", test.name)
		}
	}
}
