package cli_test

import (
	gocontext "context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting"
	"github.com/osmosis-labs/osmosis/v12/x/lockup/types"

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
			"Query account locked coins",
			func() {
				_, err := s.queryClient.AccountLockedCoins(gocontext.Background(), &types.AccountLockedCoinsRequest{Owner: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query account locked by duration",
			func() {
				_, err := s.queryClient.AccountLockedDuration(gocontext.Background(), &types.AccountLockedDurationRequest{Owner: s.TestAccs[0].String(), Duration: time.Hour * 24})
				s.Require().NoError(err)
			},
		},
		{
			"Query account locked longer than given duration",
			func() {
				_, err := s.queryClient.AccountLockedLongerDuration(gocontext.Background(), &types.AccountLockedLongerDurationRequest{Owner: s.TestAccs[0].String(), Duration: time.Hour})
				s.Require().NoError(err)
			},
		},
		{
			"Query account locked by denom that longer than given duration",
			func() {
				_, err := s.queryClient.AccountLockedLongerDurationDenom(gocontext.Background(), &types.AccountLockedLongerDurationDenomRequest{Owner: s.TestAccs[0].String(), Duration: time.Hour, Denom: "gamm/pool/1"})
				s.Require().NoError(err)
			},
		},
		{
			"Query account locked longer than given duration not unlocking",
			func() {
				_, err := s.queryClient.AccountLockedLongerDurationNotUnlockingOnly(gocontext.Background(), &types.AccountLockedLongerDurationNotUnlockingOnlyRequest{Owner: s.TestAccs[0].String(), Duration: time.Hour})
				s.Require().NoError(err)
			},
		},
		{
			"Query account locked in past time",
			func() {
				_, err := s.queryClient.AccountLockedPastTime(gocontext.Background(), &types.AccountLockedPastTimeRequest{Owner: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query account locked in past time by denom",
			func() {
				_, err := s.queryClient.AccountLockedPastTimeDenom(gocontext.Background(), &types.AccountLockedPastTimeDenomRequest{Owner: s.TestAccs[0].String(), Denom: "gamm/pool/1"})
				s.Require().NoError(err)
			},
		},
		{
			" Query account locked in past time that not unlocking",
			func() {
				_, err := s.queryClient.AccountLockedPastTimeNotUnlockingOnly(gocontext.Background(), &types.AccountLockedPastTimeNotUnlockingOnlyRequest{Owner: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query account unlockable coins",
			func() {
				_, err := s.queryClient.AccountUnlockableCoins(gocontext.Background(), &types.AccountUnlockableCoinsRequest{Owner: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query account unlocked before given time",
			func() {
				_, err := s.queryClient.AccountUnlockedBeforeTime(gocontext.Background(), &types.AccountUnlockedBeforeTimeRequest{Owner: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query account unlocking coins",
			func() {
				_, err := s.queryClient.AccountUnlockingCoins(gocontext.Background(), &types.AccountUnlockingCoinsRequest{Owner: s.TestAccs[0].String()})
				s.Require().NoError(err)
			},
		},
		{
			"Query lock by id",
			func() {
				_, err := s.queryClient.LockedByID(gocontext.Background(), &types.LockedRequest{LockId: 1})
				s.Require().NoError(err)
			},
		},
		{
			"Query lock by denom",
			func() {
				_, err := s.queryClient.LockedDenom(gocontext.Background(), &types.LockedDenomRequest{Duration: time.Hour * 24, Denom: "gamm/pool/1"})
				s.Require().NoError(err)
			},
		},
		{
			"Query module balances",
			func() {
				_, err := s.queryClient.ModuleBalance(gocontext.Background(), &types.ModuleBalanceRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query module locked amount",
			func() {
				_, err := s.queryClient.ModuleLockedAmount(gocontext.Background(), &types.ModuleLockedAmountRequest{})
				s.Require().NoError(err)
			},
		},
		{
			"Query synthetic lock by id",
			func() {
				_, err := s.queryClient.SyntheticLockupsByLockupID(gocontext.Background(), &types.SyntheticLockupsByLockupIDRequest{LockId: 1})
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
