package keeper_test

import (
	gocontext "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/c-osmosis/osmosis/x/gamm/types"
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
		poolId := suite.preparePool()
		pool, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
			PoolId: poolId,
		})
		suite.Require().NoError(err)
		suite.Require().Equal(poolId, pool.Pool.Id)
	}
}

func (suite *KeeperTestSuite) TestQueryPools() {
	queryClient := suite.queryClient

	for i := 0; i < 10; i++ {
		poolId := suite.preparePool()
		pool, err := queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{
			PoolId: poolId,
		})
		suite.Require().NoError(err)
		suite.Require().Equal(poolId, pool.Pool.Id)
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
		suite.Require().Equal(uint64(1), r.Id)
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
		suite.Require().Equal(uint64(i+1), r.Id)
	}
}

func (suite *KeeperTestSuite) TestQueryPoolParams() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.PoolParams(gocontext.Background(), &types.QueryPoolParamsRequest{PoolId: 1})
	suite.Require().Error(err)

	poolId1 := suite.preparePoolWithPoolParams(types.PoolParams{
		Lock:    false,
		SwapFee: sdk.NewDecWithPrec(1, 2),
		ExitFee: sdk.NewDecWithPrec(15, 2),
	})

	poolId2 := suite.preparePoolWithPoolParams(types.PoolParams{
		Lock:    false,
		SwapFee: sdk.NewDecWithPrec(1, 1),
		ExitFee: sdk.NewDecWithPrec(15, 3),
	})

	params1, err := queryClient.PoolParams(gocontext.Background(), &types.QueryPoolParamsRequest{PoolId: poolId1})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewDecWithPrec(1, 2).String(), params1.Params.SwapFee.String())
	suite.Require().Equal(sdk.NewDecWithPrec(15, 2).String(), params1.Params.ExitFee.String())

	params2, err := queryClient.PoolParams(gocontext.Background(), &types.QueryPoolParamsRequest{PoolId: poolId2})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewDecWithPrec(1, 1).String(), params2.Params.SwapFee.String())
	suite.Require().Equal(sdk.NewDecWithPrec(15, 3).String(), params2.Params.ExitFee.String())
}

func (suite *KeeperTestSuite) TestQueryTotalShare() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.TotalShare(gocontext.Background(), &types.QueryTotalShareRequest{PoolId: 1})
	suite.Require().Error(err)

	poolId := suite.preparePool()

	// Share Token would be minted as 100.000000 share token initially.
	res, err := queryClient.TotalShare(gocontext.Background(), &types.QueryTotalShareRequest{PoolId: poolId})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewIntWithDecimal(100, 6).String(), res.TotalShare.Amount.String())

	// Mint more share token.
	poolAcc, err := suite.app.GAMMKeeper.GetPool(suite.ctx, poolId)
	suite.Require().NoError(err)
	err = suite.app.GAMMKeeper.MintPoolShareToAccount(suite.ctx, poolAcc, acc1, sdk.NewIntWithDecimal(10, 6))
	suite.Require().NoError(err)
	suite.Require().NoError(suite.app.GAMMKeeper.SetPool(suite.ctx, poolAcc))

	res, err = queryClient.TotalShare(gocontext.Background(), &types.QueryTotalShareRequest{PoolId: poolId})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewIntWithDecimal(110, 6).String(), res.TotalShare.Amount.String())
}

func (suite *KeeperTestSuite) TestQueryRecords() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.Records(gocontext.Background(), &types.QueryRecordsRequest{PoolId: 1})
	suite.Require().Error(err)

	poolId := suite.preparePool()

	res, err := queryClient.Records(gocontext.Background(), &types.QueryRecordsRequest{PoolId: poolId})
	suite.Require().NoError(err)

	/*
		{
			Weight: sdk.NewInt(200),
			Token:  sdk.NewCoin("bar", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(300),
			Token:  sdk.NewCoin("baz", sdk.NewInt(5000000)),
		},
		{
			Weight: sdk.NewInt(100),
			Token:  sdk.NewCoin("foo", sdk.NewInt(5000000)),
		},
	*/
	records := res.Records
	suite.Require().Equal(3, len(records))

	suite.Require().Equal("200", records[0].Weight.String())
	suite.Require().Equal("300", records[1].Weight.String())
	suite.Require().Equal("100", records[2].Weight.String())

	suite.Require().Equal("5000000bar", records[0].Token.String())
	suite.Require().Equal("5000000baz", records[1].Token.String())
	suite.Require().Equal("5000000foo", records[2].Token.String())
}

func (suite *KeeperTestSuite) TestQuerySpotPrice() {
	queryClient := suite.queryClient

	// Pool not exist
	_, err := queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{PoolId: 1})
	suite.Require().Error(err)

	poolId := suite.preparePool()

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
