package cli_test

import (
	gocontext "context"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/txfees/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	// set up pool
	poolAssets := []sdk.Coin{
		sdk.NewInt64Coin("uosmo", 1000000),
		sdk.NewInt64Coin("stake", 120000000),
	}
	s.PrepareBalancerPoolWithCoins(poolAssets...)

	// set up fee token
	upgradeProp := types.NewUpdateFeeTokenProposal(
		"Test Proposal",
		"test",
		types.FeeToken{
			Denom:  "uosmo",
			PoolID: 1,
		},
	)
	err := s.App.TxFeesKeeper.HandleUpdateFeeTokenProposal(s.Ctx, &upgradeProp)
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
			"Query base denom",
			func() {
				_, err := s.queryClient.BaseDenom(gocontext.Background(), &types.QueryBaseDenomRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query poolID by denom",
			func() {
				_, err := s.queryClient.DenomPoolId(gocontext.Background(), &types.QueryDenomPoolIdRequest{Denom: "uosmo"})
				s.Require().NoError(err)
			},
		},
		{
			"Query spot price by denom",
			func() {
				_, err := s.queryClient.DenomSpotPrice(gocontext.Background(), &types.QueryDenomSpotPriceRequest{Denom: "uosmo"})
				s.Require().NoError(err)
			},
		},
		{
			"Query fee tokens",
			func() {
				_, err := s.queryClient.FeeTokens(gocontext.Background(), &types.QueryFeeTokensRequest{})
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
