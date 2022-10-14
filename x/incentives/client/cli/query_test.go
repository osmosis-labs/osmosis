package cli_test

import (
	gocontext "context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/incentives/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	// create a pool
	s.PrepareBalancerPool()
	// set up lock with id = 1
	s.LockTokens(s.TestAccs[0], sdk.Coins{sdk.NewCoin("gamm/pool/1", sdk.NewInt(1000000))}, time.Hour*24)

	s.Commit()
}

func (s *QueryTestSuite) TestQueriesNeverAlterState() {
	s.SetupSuite()
	testCases := []struct {
		name  string
		query func()
	}{
		{
			"Query active gauges",
			func() {
				_, err := s.queryClient.ActiveGauges(gocontext.Background(), &types.ActiveGaugesRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query active gauges per denom",
			func() {
				_, err := s.queryClient.ActiveGaugesPerDenom(gocontext.Background(), &types.ActiveGaugesPerDenomRequest{Denom: "stake"})
				s.Require().NoError(err)
			},
		},
		{
			"Query gauge by id",
			func() {
				_, err := s.queryClient.GaugeByID(gocontext.Background(), &types.GaugeByIDRequest{Id: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query all gauges",
			func() {
				_, err := s.queryClient.Gauges(gocontext.Background(), &types.GaugesRequest{})
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
			"Query module to distibute coins",
			func() {
				_, err := s.queryClient.ModuleToDistributeCoins(gocontext.Background(), &types.ModuleToDistributeCoinsRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query reward estimate",
			func() {
				_, err := s.queryClient.RewardsEst(gocontext.Background(), &types.RewardsEstRequest{Owner: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query upcoming gauges",
			func() {
				_, err := s.queryClient.UpcomingGauges(gocontext.Background(), &types.UpcomingGaugesRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query upcoming gauges",
			func() {
				_, err := s.queryClient.UpcomingGaugesPerDenom(gocontext.Background(), &types.UpcomingGaugesPerDenomRequest{Denom: "stake"})
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
