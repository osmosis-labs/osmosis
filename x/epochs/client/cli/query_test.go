package cli_test

// import (
// 	"testing"
// 	"time"

// 	"github.com/osmosis-labs/osmosis/x/epochs/keeper"
// 	"github.com/osmosis-labs/osmosis/x/epochs/types"
// 	"github.com/stretchr/testify/suite"
// )

// type QueryTestSuite struct {
// 	keeper.Keeper
// 	queryClient types.QueryClient
// }

// func (s *QueryTestSuite) SetupSuite() {
// 	s.Setup()
// 	s.queryClient = types.NewQueryClient(s.QueryHelper)

// 	// add new epoch
// 	epoch := types.EpochInfo{
// 		Identifier:              "weekly",
// 		StartTime:               time.Time{},
// 		Duration:                time.Hour,
// 		CurrentEpoch:            0,
// 		CurrentEpochStartHeight: 0,
// 		CurrentEpochStartTime:   time.Time{},
// 		EpochCountingStarted:    false,
// 	}
// 	s.App.EpochsKeeper.AddEpochInfo(s.Ctx, epoch)

// 	s.Commit()
// }

// func (s *QueryTestSuite) TestQueriesNeverAlterState() {
// 	testCases := []struct {
// 		name   string
// 		query  string
// 		input  interface{}
// 		output interface{}
// 	}{
// 		{
// 			"Query current epoch",
// 			"/osmosis.epochs.v1beta1.Query/CurrentEpoch",
// 			&types.QueryCurrentEpochRequest{Identifier: "weekly"},
// 			&types.QueryCurrentEpochResponse{},
// 		},
// 		{
// 			"Query epochs info",
// 			"/osmosis.epochs.v1beta1.Query/EpochInfos",
// 			&types.QueryEpochsInfoRequest{},
// 			&types.QueryEpochsInfoResponse{},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc

// 		s.Run(tc.name, func() {
// 			s.SetupSuite()
// 			err := s.QueryHelper.Invoke(gocontext.Background(), tc.query, tc.input, tc.output)
// 			s.Require().NoError(err)
// 			s.StateNotAltered()
// 		})
// 	}
// }

// func TestQueryTestSuite(t *testing.T) {
// 	suite.Run(t, new(QueryTestSuite))
// }
