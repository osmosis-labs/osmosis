package cli_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	"github.com/osmosis-labs/osmosis/v27/x/tokenfactory/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	// fund acc
	fundAccsAmount := sdk.NewCoins(sdk.NewCoin(apptesting.SecondaryDenom, apptesting.SecondaryAmount))
	s.FundAcc(s.TestAccs[0], fundAccsAmount)
	// create new token
	_, err := s.App.TokenFactoryKeeper.CreateDenom(s.Ctx, s.TestAccs[0].String(), "tokenfactory")
	s.Require().NoError(err)

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
			"Query denom authority metadata",
			"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata",
			&types.QueryDenomAuthorityMetadataRequest{Denom: "tokenfactory"},
			&types.QueryDenomAuthorityMetadataResponse{},
		},
		{
			"Query denom with encoded values",
			"/osmosis.tokenfactory.v1beta1.Query/DenomAuthorityMetadata",
			&types.QueryDenomAuthorityMetadataRequest{Denom: "factory%2Fosmo1zs0txy03pv5crj2rvty8wemd3zhrka2ne8u05n%2Fdenom"},
			&types.QueryDenomAuthorityMetadataResponse{},
		},
		{
			"Query denoms by creator",
			"/osmosis.tokenfactory.v1beta1.Query/DenomsFromCreator",
			&types.QueryDenomsFromCreatorRequest{Creator: s.TestAccs[0].String()},
			&types.QueryDenomsFromCreatorResponse{},
		},
		{
			"Query params",
			"/osmosis.tokenfactory.v1beta1.Query/Params",
			&types.QueryParamsRequest{},
			&types.QueryParamsResponse{},
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
