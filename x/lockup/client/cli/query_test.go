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
	testCases := []struct {
		name   string
		query  string
		input  interface{}
		output interface{}
	}{
		{
			"Query account locked coins",
			"/osmosis.lockup.Query/AccountLockedCoins",
			&types.AccountLockedCoinsRequest{Owner: s.TestAccs[0].String()},
			&types.AccountLockedCoinsResponse{},
		},
		{
			"Query account locked by duration",
			"/osmosis.lockup.Query/AccountLockedDuration",
			&types.AccountLockedDurationRequest{Owner: s.TestAccs[0].String(), Duration: time.Hour * 24},
			&types.AccountLockedDurationResponse{},
		},
		{
			"Query account locked longer than given duration",
			"/osmosis.lockup.Query/AccountLockedLongerDuration",
			&types.AccountLockedLongerDurationRequest{Owner: s.TestAccs[0].String(), Duration: time.Hour},
			&types.AccountLockedLongerDurationResponse{},
		},
		{
			"Query account locked by denom that longer than given duration",
			"/osmosis.lockup.Query/AccountLockedLongerDurationDenom",
			&types.AccountLockedLongerDurationDenomRequest{Owner: s.TestAccs[0].String(), Duration: time.Hour, Denom: "gamm/pool/1"},
			&types.AccountLockedLongerDurationDenomResponse{},
		},
		{
			"Query account locked longer than given duration not unlocking",
			"/osmosis.lockup.Query/AccountLockedLongerDurationNotUnlockingOnly",
			&types.AccountLockedLongerDurationNotUnlockingOnlyRequest{Owner: s.TestAccs[0].String(), Duration: time.Hour},
			&types.AccountLockedLongerDurationNotUnlockingOnlyResponse{},
		},
		{
			"Query account locked in past time",
			"/osmosis.lockup.Query/AccountLockedPastTime",
			&types.AccountLockedPastTimeRequest{Owner: s.TestAccs[0].String()},
			&types.AccountLockedPastTimeResponse{},
		},
		{
			"Query account locked in past time by denom",
			"/osmosis.lockup.Query/AccountLockedPastTimeDenom",
			&types.AccountLockedPastTimeDenomRequest{Owner: s.TestAccs[0].String(), Denom: "gamm/pool/1"},
			&types.AccountLockedPastTimeDenomResponse{},
		},
		{
			" Query account locked in past time that not unlocking",
			"/osmosis.lockup.Query/AccountLockedPastTimeNotUnlockingOnly",
			&types.AccountLockedPastTimeNotUnlockingOnlyRequest{Owner: s.TestAccs[0].String()},
			&types.AccountLockedPastTimeNotUnlockingOnlyResponse{},
		},
		{
			"Query account unlockable coins",
			"/osmosis.lockup.Query/AccountUnlockableCoins",
			&types.AccountUnlockableCoinsRequest{Owner: s.TestAccs[0].String()},
			&types.AccountUnlockableCoinsResponse{},
		},
		{
			"Query account unlocked before given time",
			"/osmosis.lockup.Query/AccountUnlockedBeforeTime",
			&types.AccountUnlockedBeforeTimeRequest{Owner: s.TestAccs[0].String()},
			&types.AccountUnlockedBeforeTimeResponse{},
		},
		{
			"Query account unlocking coins",
			"/osmosis.lockup.Query/AccountUnlockingCoins",
			&types.AccountUnlockingCoinsRequest{Owner: s.TestAccs[0].String()},
			&types.AccountUnlockingCoinsResponse{},
		},
		{
			"Query lock by id",
			"/osmosis.lockup.Query/LockedByID",
			&types.LockedRequest{LockId: 1},
			&types.LockedResponse{},
		},
		{
			"Query lock by denom",
			"/osmosis.lockup.Query/LockedDenom",
			&types.LockedDenomRequest{Duration: time.Hour * 24, Denom: "gamm/pool/1"},
			&types.LockedDenomResponse{},
		},
		{
			"Query module balances",
			"/osmosis.lockup.Query/ModuleBalance",
			&types.ModuleBalanceRequest{},
			&types.ModuleBalanceResponse{},
		},
		{
			"Query module locked amount",
			"/osmosis.lockup.Query/ModuleLockedAmount",
			&types.ModuleLockedAmountRequest{},
			&types.ModuleLockedAmountResponse{},
		},
		{
			"Query synthetic lock by id",
			"/osmosis.lockup.Query/SyntheticLockupsByLockupID",
			&types.SyntheticLockupsByLockupIDRequest{LockId: 1},
			&types.SyntheticLockupsByLockupIDResponse{},
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
