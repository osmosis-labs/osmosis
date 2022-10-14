package cli_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/pool-incentives/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	// set up pool
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
			"Query distribution info",
			func() {
				_, err := s.queryClient.DistrInfo(gocontext.Background(), &types.QueryDistrInfoRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query external incentive gauges",
			func() {
				_, err := s.queryClient.ExternalIncentiveGauges(gocontext.Background(), &types.QueryExternalIncentiveGaugesRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query all gauge ids",
			func() {
				_, err := s.queryClient.GaugeIds(gocontext.Background(), &types.QueryGaugeIdsRequest{PoolId: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query all incentivized pools",
			func() {
				_, err := s.queryClient.IncentivizedPools(gocontext.Background(), &types.QueryIncentivizedPoolsRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query lockable durations",
			func() {
				_, err := s.queryClient.LockableDurations(gocontext.Background(), &types.QueryLockableDurationsRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query params",
			func() {
				_, err := s.queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
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
