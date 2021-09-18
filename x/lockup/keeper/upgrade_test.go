package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/tendermint/tendermint/abci/types"
)

func (suite *KeeperTestSuite) LegacyLockTokens(addr sdk.AccAddress, coins sdk.Coins, duration time.Duration) {
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	_, err := suite.app.LockupKeeper.LegacyLockTokens(suite.ctx, addr, coins, duration)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestUpgradeStoreManagement() {
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	addr3 := sdk.AccAddress([]byte("addr3---------------"))

	testCases := []struct {
		msg         string
		pre_update  func()
		update      func()
		post_update func()
		expPass     bool
	}{
		{
			"with current upgrade plan",
			func() {
				coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

				// lock coins
				suite.LegacyLockTokens(addr1, coins, 10*time.Second)
				suite.LegacyLockTokens(addr2, coins, 200*time.Second)
				suite.LegacyLockTokens(addr3, coins, 50*time.Second)

				// check locks
				locks, err := suite.app.LockupKeeper.GetLegacyPeriodLocks(suite.ctx)
				suite.Require().NoError(err)
				suite.Require().Len(locks, 3)

				// begin unlock
				err = suite.app.LockupKeeper.LegacyBeginUnlock(suite.ctx, locks[0])
				suite.Require().NoError(err)

				err = suite.app.LockupKeeper.LegacyBeginUnlock(suite.ctx, locks[2])
				suite.Require().NoError(err)
			},
			func() {
				// run block 20 seconds into future
				suite.app.BeginBlocker(suite.ctx, types.RequestBeginBlock{})
				suite.app.EndBlocker(suite.ctx, types.RequestEndBlock{suite.ctx.BlockHeight()})
				suite.ctx = suite.ctx.WithBlockTime(
					suite.ctx.BlockTime().Add(20 * time.Second))
				suite.app.BeginBlocker(suite.ctx, types.RequestBeginBlock{})
				suite.app.EndBlocker(suite.ctx, types.RequestEndBlock{suite.ctx.BlockHeight()})

				// mint coins to distribution module / community pool so prop12 upgrade doesn't panic
				var bal = int64(1000000000000)
				coin := sdk.NewInt64Coin("uosmo", bal)
				coins := sdk.NewCoins(coin)
				suite.app.BankKeeper.MintCoins(suite.ctx, "mint", coins)
				suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, "mint", "distribution", coins)
				feePool := suite.app.DistrKeeper.GetFeePool(suite.ctx)
				feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinFromCoin(coin))
				suite.app.DistrKeeper.SetFeePool(suite.ctx, feePool)

				// run upgrades
				plan := upgradetypes.Plan{Name: "v4", Height: 5}
				suite.app.UpgradeKeeper.ScheduleUpgrade(suite.ctx, plan)
				plan, exists := suite.app.UpgradeKeeper.GetUpgradePlan(suite.ctx)
				suite.Require().True(exists)
				suite.Assert().NotPanics(func() {
					suite.app.UpgradeKeeper.ApplyUpgrade(suite.ctx.WithBlockHeight(5), plan)
				})
			},
			func() {
				// check all queries just after upgrade
				locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
				suite.Require().NoError(err)
				suite.Require().Len(locks, 3)

				// run a next block
				suite.ctx = suite.ctx.WithBlockHeight(6).WithBlockTime(suite.ctx.BlockTime().Add(5 * time.Second))
				suite.app.BeginBlocker(suite.ctx, types.RequestBeginBlock{})
				suite.app.EndBlocker(suite.ctx, types.RequestEndBlock{suite.ctx.BlockHeight()})

				// check all remainings
				locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
				suite.Require().NoError(err)
				suite.Require().Len(locks, 2)

				// TODO: Update the rest of these queries
				locks = suite.app.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(suite.ctx, addr1, suite.ctx.BlockTime())
				suite.Require().Len(locks, 0)

				locks = suite.app.LockupKeeper.GetAccountUnlockedBeforeTime(suite.ctx, addr1, suite.ctx.BlockTime())
				suite.Require().Len(locks, 0)

				locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "stake", suite.ctx.BlockTime())
				suite.Require().Len(locks, 0)

				locks = suite.app.LockupKeeper.GetAccountLockedLongerDuration(suite.ctx, addr2, time.Second)
				suite.Require().Len(locks, 1)

				locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationNotUnlockingOnly(suite.ctx, addr1, time.Second)
				suite.Require().Len(locks, 0)

				locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr2, "stake", time.Second)
				suite.Require().Len(locks, 1)

				locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", suite.ctx.BlockTime())
				suite.Require().Len(locks, 2)

				locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", time.Second)
				suite.Require().Len(locks, 2)

				_, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, 1)
				suite.Require().Error(err)

				_, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, 2)
				suite.Require().NoError(err)

				locks = suite.app.LockupKeeper.GetAccountPeriodLocks(suite.ctx, addr1)
				suite.Require().Len(locks, 0)

				// accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, lockuptypes.QueryCondition{
				// 	LockQueryType: lockuptypes.ByDuration,
				// 	Denom:         "stake",
				// 	Duration:      time.Second,
				// })
				// suite.Require().Equal(accum.String(), "20")

				// accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, lockuptypes.QueryCondition{
				// 	LockQueryType: lockuptypes.ByDuration,
				// 	Denom:         "stake",
				// 	Duration:      50 * time.Second,
				// })
				// suite.Require().Equal(accum.String(), "20")

				// accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, lockuptypes.QueryCondition{
				// 	LockQueryType: lockuptypes.ByDuration,
				// 	Denom:         "stake",
				// 	Duration:      200 * time.Second,
				// })
				// suite.Require().Equal(accum.String(), "10")
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.pre_update()
			tc.update()
			tc.post_update()

		})
	}
}
