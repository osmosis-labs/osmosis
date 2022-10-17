package cli_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/gamm/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)
	// create a new pool
	s.PrepareBalancerPool()
	s.Commit()
}

func (s *QueryTestSuite) TestQueriesNeverAlterState() {
	testCases := []struct {
		name   string
		query  string
		input  interface{}
		output interface{}
	}{
		{
			"Query pools",
			"/osmosis.gamm.v1beta1.Query/Pools",
			&types.QueryPoolsRequest{},
			&types.QueryPoolsResponse{},
		},
		{
			"Query single pool",
			"/osmosis.gamm.v1beta1.Query/Pool",
			&types.QueryPoolRequest{PoolId: 1},
			&types.QueryPoolsResponse{},
		},
		{
			"Query num pools",
			"/osmosis.gamm.v1beta1.Query/NumPools",
			&types.QueryNumPoolsRequest{},
			&types.QueryNumPoolsResponse{},
		},
		{
			"Query pool params",
			"/osmosis.gamm.v1beta1.Query/PoolParams",
			&types.QueryPoolParamsRequest{PoolId: 1},
			&types.QueryPoolParamsResponse{},
		},
		{
			"Query pool type",
			"/osmosis.gamm.v1beta1.Query/PoolType",
			&types.QueryPoolTypeRequest{PoolId: 1},
			&types.QueryPoolTypeResponse{},
		},
		{
			"Query spot price",
			"/osmosis.gamm.v1beta1.Query/SpotPrice",
			&types.QuerySpotPriceRequest{PoolId: 1, BaseAssetDenom: "foo", QuoteAssetDenom: "bar"},
			&types.QuerySpotPriceResponse{},
		},
		{
			"Query total liquidity",
			"/osmosis.gamm.v1beta1.Query/TotalLiquidity",
			&types.QueryTotalLiquidityRequest{},
			&types.QueryTotalLiquidityResponse{},
		},
		{
			"Query pool total liquidity",
			"/osmosis.gamm.v1beta1.Query/TotalPoolLiquidity",
			&types.QueryTotalPoolLiquidityRequest{PoolId: 1},
			&types.QueryTotalPoolLiquidityResponse{},
		},
		{
			"Query total shares",
			"/osmosis.gamm.v1beta1.Query/TotalShares",
			&types.QueryTotalSharesRequest{PoolId: 1},
			&types.QueryTotalSharesResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupSuite()
			err := s.QueryHelper.Invoke(gocontext.Background(), tc.query, tc.input, tc.output)
			s.Require().NoError(err)
			s.StateNotAltered()
		})
	}
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}
