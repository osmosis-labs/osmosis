package cli_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/tokenfactory/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	// fund acc
	fundAccsAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().DenomCreationFee[0].Denom, types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100)), sdk.NewCoin(apptesting.SecondaryDenom, apptesting.SecondaryAmount))
	s.FundAcc(s.TestAccs[0], fundAccsAmount)
	// create new token
	_, err := s.App.TokenFactoryKeeper.CreateDenom(s.Ctx, s.TestAccs[0].String(), "tokenfactory")
	s.Require().NoError(err)

	s.Commit()
}

func (s *QueryTestSuite) TestQueriesNeverAlterState() {
	s.SetupSuite()
	testCases := []struct {
		name  string
		query func()
	}{
		{
			"Query denom authority metadata",
			func() {
				_, err := s.queryClient.DenomAuthorityMetadata(gocontext.Background(), &types.QueryDenomAuthorityMetadataRequest{Denom: "tokenfactory"})
				s.Require().NoError(err)
			},
		},
		{
			"Query denoms by creator",
			func() {
				_, err := s.queryClient.DenomsFromCreator(gocontext.Background(), &types.QueryDenomsFromCreatorRequest{Creator: s.TestAccs[0].String()})
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
