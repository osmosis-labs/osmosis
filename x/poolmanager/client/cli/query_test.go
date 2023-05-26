package cli_test

import (
	gocontext "context"
	"testing"

	"github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	poolmanagerqueryproto "github.com/osmosis-labs/osmosis/v15/x/poolmanager/client/queryproto"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient poolmanagerqueryproto.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = poolmanagerqueryproto.NewQueryClient(s.QueryHelper)
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
			"/osmosis.poolmanager.v1beta1.Query/NumPools",
			&poolmanagerqueryproto.NumPoolsRequest{},
			&poolmanagerqueryproto.NumPoolsResponse{},
		},
		{
			"Query estimate swap in",
			"/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountIn",
			&poolmanagerqueryproto.EstimateSwapExactAmountInRequest{
				PoolId:  1,
				TokenIn: "10bar",
				Routes:  types.SwapAmountInRoutes{{PoolId: 1, TokenOutDenom: "baz"}},
			},
			&poolmanagerqueryproto.EstimateSwapExactAmountInResponse{},
		},
		{
			"Query estimate swap out",
			"/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountOut",
			&poolmanagerqueryproto.EstimateSwapExactAmountOutRequest{
				PoolId:   1,
				TokenOut: "6baz",
				Routes:   types.SwapAmountOutRoutes{{PoolId: 1, TokenInDenom: "bar"}},
			},
			&poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{},
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

func (s *QueryTestSuite) TestSimplifiedQueries() {
	swapIn := &poolmanagerqueryproto.EstimateSwapExactAmountInRequest{
		PoolId:  1,
		TokenIn: "10bar",
		Routes:  types.SwapAmountInRoutes{{PoolId: 1, TokenOutDenom: "baz"}},
	}
	swapOut := &poolmanagerqueryproto.EstimateSwapExactAmountOutRequest{
		PoolId:   1,
		TokenOut: "6baz",
		Routes:   types.SwapAmountOutRoutes{{PoolId: 1, TokenInDenom: "bar"}},
	}
	simplifiedSwapIn := &poolmanagerqueryproto.EstimateSinglePoolSwapExactAmountInRequest{
		PoolId:        1,
		TokenIn:       "10bar",
		TokenOutDenom: "baz",
	}
	simplifiedSwapOut := &poolmanagerqueryproto.EstimateSinglePoolSwapExactAmountOutRequest{
		PoolId:       1,
		TokenOut:     "6baz",
		TokenInDenom: "bar",
	}
	s.SetupSuite()
	output1 := &poolmanagerqueryproto.EstimateSwapExactAmountInResponse{}
	output2 := &poolmanagerqueryproto.EstimateSwapExactAmountInResponse{}
	err := s.QueryHelper.Invoke(gocontext.Background(),
		"/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountIn", swapIn, output1)
	s.Require().NoError(err)
	err = s.QueryHelper.Invoke(gocontext.Background(),
		"/osmosis.poolmanager.v1beta1.Query/EstimateSinglePoolSwapExactAmountIn", simplifiedSwapIn, output2)
	s.Require().NoError(err)
	s.Require().Equal(output1, output2)

	output3 := &poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{}
	output4 := &poolmanagerqueryproto.EstimateSwapExactAmountOutResponse{}
	err = s.QueryHelper.Invoke(gocontext.Background(),
		"/osmosis.poolmanager.v1beta1.Query/EstimateSwapExactAmountOut", swapOut, output3)
	s.Require().NoError(err)
	err = s.QueryHelper.Invoke(gocontext.Background(),
		"/osmosis.poolmanager.v1beta1.Query/EstimateSinglePoolSwapExactAmountOut", simplifiedSwapOut, output4)
	s.Require().NoError(err)
	s.Require().Equal(output3, output4)
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}
