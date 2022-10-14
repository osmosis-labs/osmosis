package cli_test

import (
	gocontext "context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/superfluid/types"
)

type QueryTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
	val         sdk.ValAddress
}

func (s *QueryTestSuite) SetupSuite() {
	s.Setup()
	s.queryClient = types.NewQueryClient(s.QueryHelper)

	// set up durations
	s.App.IncentivesKeeper.SetLockableDurations(s.Ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		time.Hour * 24 * 21,
	})
	// set up pool
	s.SetupGammPoolsWithBondDenomMultiplier([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})
	// set up lock with id = 1
	s.LockTokens(s.TestAccs[0], sdk.Coins{sdk.NewCoin("gamm/pool/1", sdk.NewInt(1000000))}, time.Hour*24*21)
	// set up validator
	s.val = s.SetupValidator(stakingtypes.Bonded)
	// set up sfs asset
	err := s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     "gamm/pool/1",
		AssetType: types.SuperfluidAssetTypeLPShare,
	})
	s.Require().NoError(err)
	// set up sfs delegation
	err = s.App.SuperfluidKeeper.SuperfluidDelegate(s.Ctx, s.TestAccs[0].String(), 1, s.val.String())
	fmt.Println(err)
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
			"Query all superfluild assets",
			func() {
				_, err := s.queryClient.AllAssets(gocontext.Background(), &types.AllAssetsRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query all intermediary accounts",
			func() {
				_, err := s.queryClient.AllIntermediaryAccounts(gocontext.Background(), &types.AllIntermediaryAccountsRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query osmo equivalent multiplier of an asset",
			func() {
				_, err := s.queryClient.AssetMultiplier(gocontext.Background(), &types.AssetMultiplierRequest{Denom: "gamm/pool/1"})
				s.Require().NoError(err)
			},
		},
		{
			"Query asset type",
			func() {
				_, err := s.queryClient.AssetType(gocontext.Background(), &types.AssetTypeRequest{Denom: "gamm/pool/1"})
				s.Require().NoError(err)
			},
		},
		{
			"Query connected intermediary account",
			func() {
				_, err := s.queryClient.ConnectedIntermediaryAccount(gocontext.Background(), &types.ConnectedIntermediaryAccountRequest{LockId: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query estimate sfs delegated amount by validator & denom",
			func() {
				_, err := s.queryClient.EstimateSuperfluidDelegatedAmountByValidatorDenom(gocontext.Background(), &types.EstimateSuperfluidDelegatedAmountByValidatorDenomRequest{ValidatorAddress: s.val.String(), Denom: "gamm/pool/1"})
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
		{
			"Query sfs delegation amount",
			func() {
				_, err := s.queryClient.SuperfluidDelegationAmount(gocontext.Background(), &types.SuperfluidDelegationAmountRequest{ValidatorAddress: s.val.String(), Denom: "gamm/pool/1", DelegatorAddress: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query sfs delegation by delegator",
			func() {
				_, err := s.queryClient.SuperfluidDelegationsByDelegator(gocontext.Background(), &types.SuperfluidDelegationsByDelegatorRequest{DelegatorAddress: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query sfs delegation by validator & denom",
			func() {
				_, err := s.queryClient.SuperfluidDelegationsByValidatorDenom(gocontext.Background(), &types.SuperfluidDelegationsByValidatorDenomRequest{ValidatorAddress: s.val.String(), Denom: "gamm/pool/1"})
				s.Require().NoError(err)
			},
		},
		{
			"Query sfs undelegation by delegator",
			func() {
				_, err := s.queryClient.SuperfluidUndelegationsByDelegator(gocontext.Background(), &types.SuperfluidUndelegationsByDelegatorRequest{DelegatorAddress: s.TestAccs[0].String(), Denom: "gamm/pool/1"})
				s.Require().NoError(err)
			},
		},
		{
			"Query total sfs delegation by delegator",
			func() {
				_, err := s.queryClient.TotalDelegationByDelegator(gocontext.Background(), &types.QueryTotalDelegationByDelegatorRequest{DelegatorAddress: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query total sfs delegations",
			func() {
				_, err := s.queryClient.TotalSuperfluidDelegations(gocontext.Background(), &types.TotalSuperfluidDelegationsRequest{})
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
