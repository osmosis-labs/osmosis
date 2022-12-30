package cli_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting"
	swaprouterqueryproto "github.com/osmosis-labs/osmosis/v13/x/swaprouter/client/queryproto"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient swaprouterqueryproto.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = swaprouterqueryproto.NewQueryClient(s.QueryHelper)
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
			"Query num pools",
			"/osmosis.swaprouter.v1beta1.Query/NumPools",
			&swaprouterqueryproto.NumPoolsRequest{},
			&swaprouterqueryproto.NumPoolsResponse{},
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

	// TODO: re-enable this once swaprouter is fully merged.
	t.SkipNow()

	suite.Run(t, new(QueryTestSuite))
}
