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
	s.SetupSuite()
	testCases := []struct {
		name  string
		query func()
	}{
		{
			"Query pools",
			func() {
				_, err := s.queryClient.Pools(gocontext.Background(), &types.QueryPoolsRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query single pool",
			func() {
				_, err := s.queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{PoolId: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query single pool",
			func() {
				_, err := s.queryClient.Pool(gocontext.Background(), &types.QueryPoolRequest{PoolId: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query num pools",
			func() {
				_, err := s.queryClient.NumPools(gocontext.Background(), &types.QueryNumPoolsRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query pool params",
			func() {
				_, err := s.queryClient.PoolParams(gocontext.Background(), &types.QueryPoolParamsRequest{PoolId: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query pool type",
			func() {
				_, err := s.queryClient.PoolType(gocontext.Background(), &types.QueryPoolTypeRequest{PoolId: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query spot price",
			func() {
				_, err := s.queryClient.SpotPrice(gocontext.Background(), &types.QuerySpotPriceRequest{PoolId: 1, BaseAssetDenom: "foo", QuoteAssetDenom: "bar"})
				s.Require().NoError(err)
			},
		},
		{
			"Query total liquidity",
			func() {
				_, err := s.queryClient.TotalLiquidity(gocontext.Background(), &types.QueryTotalLiquidityRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query pool total liquidity",
			func() {
				_, err := s.queryClient.TotalPoolLiquidity(gocontext.Background(), &types.QueryTotalPoolLiquidityRequest{PoolId: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query spot price",
			func() {
				_, err := s.queryClient.TotalShares(gocontext.Background(), &types.QueryTotalSharesRequest{PoolId: 1})
				s.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			tc.query()
			s.StateNotAltered()
		})
	}
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}
