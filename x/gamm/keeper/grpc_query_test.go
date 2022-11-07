package keeper_test

import (
	gocontext "context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

func (suite *KeeperTestSuite) TestCalcExitPoolCoinsFromShares() {
	queryClient := suite.queryClient
	ctx := suite.Ctx
	poolId := suite.PrepareBalancerPool()
	exitFee := sdk.ZeroDec()

	testCases := []struct {
		name          string
		poolId        uint64
		shareInAmount sdk.Int
		expectedErr   error
	}{
		{
			"valid test case",
			poolId,
			sdk.NewInt(1000000000000000000),
			nil,
		},
		{
			"pool id does not exist",
			poolId + 1,
			sdk.NewInt(1000000000000000000),
			types.ErrPoolNotFound,
		},
		{
			"zero share in amount",
			poolId,
			sdk.ZeroInt(),
			sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative"),
		},
		{
			"negative share in amount",
			poolId,
			sdk.NewInt(-10000),
			sdkerrors.Wrapf(types.ErrInvalidMathApprox, "share ratio is zero or negative"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			out, err := queryClient.CalcExitPoolCoinsFromShares(gocontext.Background(), &types.QueryCalcExitPoolCoinsFromSharesRequest{
				PoolId:        tc.poolId,
				ShareInAmount: tc.shareInAmount,
			})
			if tc.expectedErr == nil {
				poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
					PoolId: tc.poolId,
				})
				suite.Require().NoError(err)

				var pool types.PoolI
				err = suite.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
				suite.Require().NoError(err)

				exitCoins, err := pool.CalcExitPoolCoinsFromShares(ctx, tc.shareInAmount, exitFee)
				suite.Require().NoError(err)

				// For each coin in exitCoins we are looking for a match in our response
				// We need to find exactly len(out) such matches
				coins_checked := 0
				for _, coin := range exitCoins {
					for _, actual_coin := range out.TokensOut {
						if coin.Denom == actual_coin.Denom {
							suite.Require().Equal(coin.Amount, actual_coin.Amount)
							coins_checked++
						}
					}
				}
				suite.Require().Equal(out.TokensOut, exitCoins)
			} else {
				suite.Require().ErrorIs(err, tc.expectedErr)
			}
		})
	}
}
func (suite *KeeperTestSuite) TestCalcJoinPoolShares() {
	queryClient := suite.queryClient
	ctx := suite.Ctx
	poolId := suite.PrepareBalancerPool()
	swapFee := sdk.ZeroDec()

	testCases := []struct {
		name        string
		poolId      uint64
		tokensIn    sdk.Coins
		expectedErr error
	}{
		{
			"valid uneven multi asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(5000000)), sdk.NewCoin("bar", sdk.NewInt(5000000)), sdk.NewCoin("baz", sdk.NewInt(5000000)), sdk.NewCoin("uosmo", sdk.NewInt(5000000))),
			nil,
		},
		{
			"valid even multi asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(500000)), sdk.NewCoin("bar", sdk.NewInt(1000000)), sdk.NewCoin("baz", sdk.NewInt(1500000)), sdk.NewCoin("uosmo", sdk.NewInt(2000000))),
			nil,
		},
		{
			"valid single asset join test case",
			poolId,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000))),
			nil,
		},
		{
			"pool id does not exist",
			poolId + 1,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(1000000))),
			types.PoolDoesNotExistError{PoolId: poolId + 1},
		},
		{
			"token in denom does not exist",
			poolId,
			sdk.NewCoins(sdk.NewCoin("random", sdk.NewInt(10000))),
			sdkerrors.Wrapf(types.ErrDenomNotFoundInPool, "input denoms must already exist in the pool (%s)", "random"),
		},
		{
			"join pool with incorrect amount of assets",
			poolId,
			sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(10000)), sdk.NewCoin("bar", sdk.NewInt(10000))),
			errors.New("balancer pool only supports LP'ing with one asset or all assets in pool"),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			out, err := queryClient.CalcJoinPoolShares(gocontext.Background(), &types.QueryCalcJoinPoolSharesRequest{
				PoolId:   tc.poolId,
				TokensIn: tc.tokensIn,
			})
			if tc.expectedErr == nil {
				poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
					PoolId: tc.poolId,
				})
				suite.Require().NoError(err)

				var pool types.PoolI
				err = suite.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
				suite.Require().NoError(err)

				numShares, numLiquidity, err := pool.CalcJoinPoolShares(ctx, tc.tokensIn, swapFee)
				suite.Require().NoError(err)
				suite.Require().Equal(numShares, out.ShareOutAmount)
				suite.Require().Equal(numLiquidity, out.TokensOut)
			} else {
				suite.Require().EqualError(err, tc.expectedErr.Error())
			}
		})
	}

}
func (suite *KeeperTestSuite) TestQueryPool() {
	queryClient := suite.queryClient

	// Invalid param
	_, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{})
	suite.Require().Error(err)

	// Pool not exist
	_, err = queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
		PoolId: 1,
	})
	suite.Require().Error(err)

	for i := 0; i < 10; i++ {
		poolId := suite.PrepareBalancerPool()
		poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
			PoolId: poolId,
		})
		suite.Require().NoError(err)
		var pool types.PoolI
		err = suite.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
		suite.Require().NoError(err)
		suite.Require().Equal(poolId, pool.GetId())
		suite.Require().Equal(types.NewPoolAddress(poolId).String(), pool.GetAddress().String())
	}
}

func (suite *KeeperTestSuite) TestQueryPools() {
	queryClient := suite.queryClient

	for i := 0; i < 10; i++ {
		poolId := suite.PrepareBalancerPool()
		poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
			PoolId: poolId,
		})
		suite.Require().NoError(err)
		var pool types.PoolI
		err = suite.App.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
		suite.Require().NoError(err)
		suite.Require().Equal(poolId, pool.GetId())
		suite.Require().Equal(types.NewPoolAddress(poolId).String(), pool.GetAddress().String())
	}

	res, err := queryClient.Pools(gocontext.Background(), &types.QueryPoolsRequest{
		Pagination: &query.PageRequest{
			Key:        nil,
			Limit:      1,
			CountTotal: false,
		},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(1, len(res.Pools))
	for _, r := range res.Pools {
		var pool types.PoolI
		err = suite.App.InterfaceRegistry().UnpackAny(r, &pool)
		suite.Require().NoError(err)
		suite.Require().Equal(types.NewPoolAddress(uint64(1)).String(), pool.GetAddress().String())
		suite.Require().Equal(uint64(1), pool.GetId())
	}

	res, err = queryClient.Pools(gocontext.Background(), &types.QueryPoolsRequest{
		Pagination: &query.PageRequest{
			Key:        nil,
			Limit:      5,
			CountTotal: false,
		},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(5, len(res.Pools))
	for i, r := range res.Pools {
		var pool types.PoolI
		err = suite.App.InterfaceRegistry().UnpackAny(r, &pool)
		suite.Require().NoError(err)
		suite.Require().Equal(types.NewPoolAddress(uint64(i+1)).String(), pool.GetAddress().String())
		suite.Require().Equal(uint64(i+1), pool.GetId())
	}
}

func (suite *KeeperTestSuite) TestPoolType() {
	poolId := suite.PrepareBalancerPool()

	// error when querying invalid pool ID
	_, err := suite.queryClient.PoolType(gocontext.Background(), &types.QueryPoolTypeRequest{PoolId: poolId + 1})
	suite.Require().Error(err)

	res, err := suite.queryClient.PoolType(gocontext.Background(), &types.QueryPoolTypeRequest{PoolId: poolId})
	suite.Require().NoError(err)
	suite.Require().Equal("Balancer", res.PoolType)
}

func (suite *KeeperTestSuite) TestQueryNumPools1() {
	res, err := suite.queryClient.NumPools(gocontext.Background(), &types.QueryNumPoolsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), res.NumPools)
}

func (suite *KeeperTestSuite) TestQueryNumPools2() {
	for i := 0; i < 10; i++ {
		suite.PrepareBalancerPool()
	}

	res, err := suite.queryClient.NumPools(gocontext.Background(), &types.QueryNumPoolsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(10), res.NumPools)
}

func (suite *KeeperTestSuite) TestQueryTotalPoolLiquidity() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.TotalPoolLiquidity(gocontext.Background(), &types.QueryTotalPoolLiquidityRequest{PoolId: 1})
	suite.Require().Error(err)

	poolId := suite.PrepareBalancerPool()

	res, err := queryClient.TotalPoolLiquidity(gocontext.Background(), &types.QueryTotalPoolLiquidityRequest{PoolId: poolId})
	suite.Require().NoError(err)
	expectedCoins := sdk.NewCoins(sdk.NewCoin("foo", sdk.NewInt(5000000)), sdk.NewCoin("bar", sdk.NewInt(5000000)), sdk.NewCoin("baz", sdk.NewInt(5000000)), sdk.NewCoin("uosmo", sdk.NewInt(5000000)))
	suite.Require().Equal(res.Liquidity, expectedCoins)
}

func (suite *KeeperTestSuite) TestQueryTotalShares() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: 1})
	suite.Require().Error(err)

	poolId := suite.PrepareBalancerPool()

	// Share Token would be minted as 100.000000000000000000 share token initially.
	res, err := queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: poolId})
	suite.Require().NoError(err)
	suite.Require().Equal(types.InitPoolSharesSupply.String(), res.TotalShares.Amount.String())

	// Mint more share token.
	// TODO: Change this test structure. perhaps JoinPoolExactShareAmountOut can be used once written
	// pool, err := suite.App.GAMMKeeper.GetPool(suite.Ctx, poolId)
	// suite.Require().NoError(err)
	// err = suite.App.GAMMKeeper.MintPoolShareToAccount(suite.Ctx, pool, acc1, types.OneShare.MulRaw(10))
	// suite.Require().NoError(err)
	// suite.Require().NoError(suite.App.GAMMKeeper.SetPool(suite.Ctx, pool))

	// res, err = queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: poolId})
	// suite.Require().NoError(err)
	// suite.Require().Equal(types.InitPoolSharesSupply.Add(types.OneShare.MulRaw(10)).String(), res.TotalShares.Amount.String())
}

func (suite *KeeperTestSuite) TestQueryBalancerPoolTotalLiquidity() {
	queryClient := suite.queryClient

	// Pool not exist
	res, err := queryClient.TotalLiquidity(gocontext.Background(), &types.QueryTotalLiquidityRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal("", sdk.Coins(res.Liquidity).String())

	_ = suite.PrepareBalancerPool()

	// create pool
	res, err = queryClient.TotalLiquidity(gocontext.Background(), &types.QueryTotalLiquidityRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal("5000000bar,5000000baz,5000000foo,5000000uosmo", sdk.Coins(res.Liquidity).String())
}

// TODO: Come fix
// func (suite *KeeperTestSuite) TestQueryBalancerPoolPoolAssets() {
// 	queryClient := suite.queryClient

// 	// Pool not exist
// 	_, err := queryClient.PoolAssets(gocontext.Background(), &types.QueryPoolAssetsRequest{PoolId: 1})
// 	suite.Require().Error(err)

// 	poolId := suite.PrepareBalancerPool()

// 	res, err := queryClient.PoolAssets(gocontext.Background(), &types.QueryPoolAssetsRequest{PoolId: poolId})
// 	suite.Require().NoError(err)

// 	/*
// 		{
// 			Weight: sdk.NewInt(200 * GuaranteedWeightPrecision),
// 			Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(300 * GuaranteedWeightPrecision),
// 			Token:  sdk.NewCoin("baz", sdk.NewInt(5000000)),
// 		},
// 		{
// 			Weight: sdk.NewInt(100 * GuaranteedWeightPrecision),
// 			Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
// 		},
// 	*/
// 	PoolAssets := res.PoolAssets
// 	suite.Require().Equal(3, len(PoolAssets))

// 	suite.Require().Equal(sdk.NewInt(200*types.GuaranteedWeightPrecision), PoolAssets[0].Weight)
// 	suite.Require().Equal(sdk.NewInt(300*types.GuaranteedWeightPrecision), PoolAssets[1].Weight)
// 	suite.Require().Equal(sdk.NewInt(100*types.GuaranteedWeightPrecision), PoolAssets[2].Weight)

// 	suite.Require().Equal("5000000bar", PoolAssets[0].Token.String())
// 	suite.Require().Equal("5000000baz", PoolAssets[1].Token.String())
// 	suite.Require().Equal("5000000foo", PoolAssets[2].Token.String())
// }

func (suite *KeeperTestSuite) TestQueryBalancerPoolSpotPrice() {
	queryClient := suite.queryClient
	poolID := suite.PrepareBalancerPool()

	testCases := []struct {
		name      string
		req       *types.QuerySpotPriceRequest
		expectErr bool
		result    string
	}{
		{
			name: "non-existant pool",
			req: &types.QuerySpotPriceRequest{
				PoolId:          0,
				BaseAssetDenom:  "foo",
				QuoteAssetDenom: "bar",
			},
			expectErr: true,
		},
		{
			name: "missing asset denoms",
			req: &types.QuerySpotPriceRequest{
				PoolId: poolID,
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and quote denom",
			req: &types.QuerySpotPriceRequest{
				BaseAssetDenom: "foo",
			},
			expectErr: true,
		},
		{
			name: "missing pool ID and base denom",
			req: &types.QuerySpotPriceRequest{
				QuoteAssetDenom: "bar",
			},
			expectErr: true,
		},
		{
			name: "valid request for foo/bar",
			req: &types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "foo",
				QuoteAssetDenom: "bar",
			},
			result: sdk.NewDec(2).String(),
		},
		{
			name: "valid request for bar/baz",
			req: &types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "bar",
				QuoteAssetDenom: "baz",
			},
			result: sdk.NewDecWithPrec(15, 1).String(),
		},
		{
			name: "valid request for baz/foo",
			req: &types.QuerySpotPriceRequest{
				PoolId:          poolID,
				BaseAssetDenom:  "baz",
				QuoteAssetDenom: "foo",
			},
			result: sdk.MustNewDecFromStr("0.333333330000000000").String(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			result, err := queryClient.SpotPrice(gocontext.Background(), tc.req)
			if tc.expectErr {
				suite.Require().Error(err, "expected error")
			} else {
				suite.Require().NoError(err, "unexpected error")
				suite.Require().Equal(tc.result, result.SpotPrice)
			}
		})
	}
}
