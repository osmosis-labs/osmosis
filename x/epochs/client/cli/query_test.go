package cli_test

import (
	gocontext "context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/epochs/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	// add new epoch
	epoch := types.EpochInfo{
		Identifier:              "weekly",
		StartTime:               time.Time{},
		Duration:                time.Hour,
		CurrentEpoch:            0,
		CurrentEpochStartHeight: 0,
		CurrentEpochStartTime:   time.Time{},
		EpochCountingStarted:    false,
	}
	s.App.EpochsKeeper.AddEpochInfo(s.Ctx, epoch)

	s.Commit()
}

func (s *QueryTestSuite) TestQueriesNeverAlterState() {
	s.SetupSuite()
	testCases := []struct {
		name  string
		query func()
	}{
		{
			"Query current epoch",
			func() {
				res, err := s.queryClient.CurrentEpoch(gocontext.Background(), &types.QueryCurrentEpochRequest{Identifier: "weekly"})
				fmt.Println(res)
				s.Require().NoError(err)
			},
		},
		{
			"Query epochs info",
			func() {
				_, err := s.queryClient.EpochInfos(gocontext.Background(), &types.QueryEpochsInfoRequest{})
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
