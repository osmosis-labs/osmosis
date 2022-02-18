package keeper_test

import (
	gocontext "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/osmosis-labs/osmosis/v7/x/gamm/types"
)

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
		poolId := suite.prepareBalancerPool()
		poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
			PoolId: poolId,
		})
		suite.Require().NoError(err)
		var pool types.PoolI
		err = suite.app.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
		suite.Require().NoError(err)
		suite.Require().Equal(poolId, pool.GetId())
		suite.Require().Equal(types.NewPoolAddress(poolId).String(), pool.GetAddress().String())
	}
}

func (suite *KeeperTestSuite) TestQueryPools() {
	queryClient := suite.queryClient

	for i := 0; i < 10; i++ {
		poolId := suite.prepareBalancerPool()
		poolRes, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
			PoolId: poolId,
		})
		suite.Require().NoError(err)
		var pool types.PoolI
		err = suite.app.InterfaceRegistry().UnpackAny(poolRes.Pool, &pool)
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
		err = suite.app.InterfaceRegistry().UnpackAny(r, &pool)
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
		err = suite.app.InterfaceRegistry().UnpackAny(r, &pool)
		suite.Require().NoError(err)
		suite.Require().Equal(types.NewPoolAddress(uint64(i+1)).String(), pool.GetAddress().String())
		suite.Require().Equal(uint64(i+1), pool.GetId())
	}
}

func (suite *KeeperTestSuite) TestQueryNumPools1() {
	res, err := suite.queryClient.NumPools(gocontext.Background(), &types.QueryNumPoolsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(0), res.NumPools)
}

func (suite *KeeperTestSuite) TestQueryNumPools2() {
	for i := 0; i < 10; i++ {
		suite.prepareBalancerPool()
	}

	res, err := suite.queryClient.NumPools(gocontext.Background(), &types.QueryNumPoolsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(10), res.NumPools)
}

func (suite *KeeperTestSuite) TestQueryTotalShares() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: 1})
	suite.Require().Error(err)

	poolId := suite.prepareBalancerPool()

	// Share Token would be minted as 100.000000000000000000 share token initially.
	res, err := queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: poolId})
	suite.Require().NoError(err)
	suite.Require().Equal(types.InitPoolSharesSupply.String(), res.TotalShares.Amount.String())

	// Mint more share token.
	pool, err := suite.app.GAMMKeeper.GetPool(suite.ctx, poolId)
	suite.Require().NoError(err)
	err = suite.app.GAMMKeeper.MintPoolShareToAccount(suite.ctx, pool, acc1, types.OneShare.MulRaw(10))
	suite.Require().NoError(err)
	suite.Require().NoError(suite.app.GAMMKeeper.SetPool(suite.ctx, pool))

	res, err = queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: poolId})
	suite.Require().NoError(err)
	suite.Require().Equal(types.InitPoolSharesSupply.Add(types.OneShare.MulRaw(10)).String(), res.TotalShares.Amount.String())
}

func (suite *KeeperTestSuite) TestQueryBalancerPoolTotalLiquidity() {
	queryClient := suite.queryClient

	// Pool not exist
	res, err := queryClient.TotalLiquidity(gocontext.Background(), &types.QueryTotalLiquidityRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal("", sdk.Coins(res.Liquidity).String())

	_ = suite.prepareBalancerPool()

	// create pool
	res, err = queryClient.TotalLiquidity(gocontext.Background(), &types.QueryTotalLiquidityRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal("5000000bar,5000000baz,5000000foo", sdk.Coins(res.Liquidity).String())
}

func (suite *KeeperTestSuite) TestQueryBalancerPoolPoolAssets() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.PoolAssets(gocontext.Background(), &types.QueryPoolAssetsRequest{PoolId: 1})
	suite.Require().Error(err)

	poolId := suite.prepareBalancerPool()

	res, err := queryClient.PoolAssets(gocontext.Background(), &types.QueryPoolAssetsRequest{PoolId: poolId})
	suite.Require().NoError(err)

	/*
		{
			Weight: sdk.NewInt(200 * GuaranteedWeightPrecision),
			Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(300 * GuaranteedWeightPrecision),
			Token:  sdk.NewCoin("baz", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(100 * GuaranteedWeightPrecision),
			Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
		},
	*/
	PoolAssets := res.PoolAssets
	suite.Require().Equal(3, len(PoolAssets))

	suite.Require().Equal(sdk.NewInt(200*types.GuaranteedWeightPrecision), PoolAssets[0].Weight)
	suite.Require().Equal(sdk.NewInt(300*types.GuaranteedWeightPrecision), PoolAssets[1].Weight)
	suite.Require().Equal(sdk.NewInt(100*types.GuaranteedWeightPrecision), PoolAssets[2].Weight)

	suite.Require().Equal("5000000bar", PoolAssets[0].Token.String())
	suite.Require().Equal("5000000baz", PoolAssets[1].Token.String())
	suite.Require().Equal("5000000foo", PoolAssets[2].Token.String())
}

func (suite *KeeperTestSuite) TestQueryBalancerPoolSpotPrice() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{
		PoolId:        1,
		TokenInDenom:  "foo",
		TokenOutDenom: "bar",
	})
	suite.Require().Error(err)

	poolId := suite.prepareBalancerPool()

	// Invalid params
	_, err = queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{PoolId: poolId})
	suite.Require().Error(err)
	_, err = queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{TokenInDenom: "foo"})
	suite.Require().Error(err)
	_, err = queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{TokenOutDenom: "bar"})
	suite.Require().Error(err)

	res, err := queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{
		PoolId:        poolId,
		TokenInDenom:  "foo",
		TokenOutDenom: "bar",
	})
	suite.NoError(err)
	suite.Equal(sdk.NewDec(2).String(), res.SpotPrice)

	res, err = queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{
		PoolId:        poolId,
		TokenInDenom:  "bar",
		TokenOutDenom: "baz",
	})
	suite.NoError(err)
	suite.Equal(sdk.NewDecWithPrec(15, 1).String(), res.SpotPrice)

	res, err = queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{
		PoolId:        poolId,
		TokenInDenom:  "baz",
		TokenOutDenom: "foo",
	})
	suite.NoError(err)
	suite.Equal(sdk.NewDec(1).Quo(sdk.NewDec(3)).String(), res.SpotPrice)
}
