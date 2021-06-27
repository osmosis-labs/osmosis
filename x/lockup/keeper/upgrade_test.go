package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (suite *KeeperTestSuite) LegacyLockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	_, err := suite.app.LockupKeeper.LegacyLockTokens(suite.ctx, addr, coins, duration)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestUpgradeStoreManagement() {
	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"with current upgrade plan",
			func() {
				coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

				// lock coins
				addr1 := sdk.AccAddress([]byte("addr1---------------"))
				suite.LegacyLockTokens(addr1, coins, time.Second)
				addr2 := sdk.AccAddress([]byte("addr2---------------"))
				suite.LegacyLockTokens(addr2, coins, time.Second)
				addr3 := sdk.AccAddress([]byte("addr3---------------"))
				suite.LegacyLockTokens(addr3, coins, time.Second)

				// check locks
				locks, err := suite.app.LockupKeeper.GetLegacyPeriodLocks(suite.ctx)
				suite.Require().NoError(err)
				suite.Require().Len(locks, 3)

				// begin unlock
				err = suite.app.LockupKeeper.LegacyBeginUnlock(suite.ctx, locks[0])
				suite.Require().NoError(err)

				// run upgrades
				plan := upgradetypes.Plan{Name: "v2", Height: 5}
				suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)
				plan, exists := suite.app.UpgradeKeeper.GetUpgradePlan(suite.ctx)
				suite.Require().True(exists)
				suite.Assert().NotPanics(func() {
					suite.app.UpgradeKeeper.ApplyUpgrade(suite.ctx.WithBlockHeight(5), plan)
				})

				// check all queries
				locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
				suite.Require().NoError(err)
				suite.Require().Len(locks, 3)

				locks = suite.app.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(suite.ctx, addr1, suite.ctx.BlockTime())
				suite.Require().Len(locks, 0)

				locks = suite.app.LockupKeeper.GetAccountUnlockedBeforeTime(suite.ctx, addr1, suite.ctx.BlockTime())
				suite.Require().Len(locks, 0)

				locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "stake", suite.ctx.BlockTime())
				suite.Require().Len(locks, 1)

				locks = suite.app.LockupKeeper.GetAccountLockedLongerDuration(suite.ctx, addr1, time.Second)
				suite.Require().Len(locks, 1)

				locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationNotUnlockingOnly(suite.ctx, addr1, time.Second)
				suite.Require().Len(locks, 0)

				locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "stake", time.Second)
				suite.Require().Len(locks, 1)

				locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", suite.ctx.BlockTime())
				suite.Require().Len(locks, 3)

				locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", time.Second)
				suite.Require().Len(locks, 3)

				_, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
				suite.Require().NoError(err)

				locks = suite.app.LockupKeeper.GetAccountPeriodLocks(suite.ctx, addr1)
				suite.Require().Len(locks, 1)

				accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         "stake",
					Duration:      time.Second,
				})
				suite.Require().Equal(accum.String(), "20")
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()

		})
	}
}
