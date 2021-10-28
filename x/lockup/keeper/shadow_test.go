package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (suite *KeeperTestSuite) TestShadowLockupCreateGetDeleteAccumulation() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	// check accumulation store is correctly updated
	accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// check queries for native denom before shadow
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "stake", suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr1, "stake", time.Second)
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "stake", time.Second)
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", time.Second)
	suite.Require().Len(locks, 1)
	amount := suite.app.LockupKeeper.GetLockedDenom(suite.ctx, "stake", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check queries for shadow denom before shadow
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "stakestakedtovalidator1", suite.ctx.BlockTime())
	suite.Require().Len(locks, 0)
	locks = suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr1, "stakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "stakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stakestakedtovalidator1", suite.ctx.BlockTime())
	suite.Require().Len(locks, 0)
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	amount = suite.app.LockupKeeper.GetLockedDenom(suite.ctx, "stakestakedtovalidator1", time.Second)
	suite.Require().Equal(amount.String(), "0")

	// check non-denom related queries before shadow
	moduleBalance := suite.app.LockupKeeper.GetModuleBalance(suite.ctx)
	suite.Require().Equal(moduleBalance.String(), "10stake")
	moduleLockedCoins := suite.app.LockupKeeper.GetModuleLockedCoins(suite.ctx)
	suite.Require().Equal(moduleLockedCoins.String(), "10stake")
	accountUnlockableCoins := suite.app.LockupKeeper.GetAccountUnlockableCoins(suite.ctx, addr1)
	suite.Require().Equal(accountUnlockableCoins.String(), "")
	accountUnlockingCoins := suite.app.LockupKeeper.GetAccountUnlockingCoins(suite.ctx, addr1)
	suite.Require().Equal(accountUnlockingCoins.String(), "")
	accountLockedCoins := suite.app.LockupKeeper.GetAccountLockedCoins(suite.ctx, addr1)
	suite.Require().Equal(accountLockedCoins.String(), "10stake")
	locks = suite.app.LockupKeeper.GetAccountLockedPastTime(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountUnlockedBeforeTime(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Len(locks, 0)

	err = suite.app.LockupKeeper.CreateShadowLockup(suite.ctx, 1, "stakedtovalidator1", false)
	suite.Require().NoError(err)

	shadowLock, err := suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator1",
		EndTime: time.Time{},
	})

	shadowLocks := suite.app.LockupKeeper.GetAllShadowsByLockup(suite.ctx, 1)
	suite.Require().Len(shadowLocks, 1)
	suite.Require().Equal(*shadowLock, shadowLocks[0])

	shadowLocks = suite.app.LockupKeeper.GetAllShadows(suite.ctx)
	suite.Require().Len(shadowLocks, 1)
	suite.Require().Equal(*shadowLock, shadowLocks[0])

	// check accumulation store is correctly updated for shadow lock
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// check queries for native denom after shadow
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "stake", suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr1, "stake", time.Second)
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "stake", time.Second)
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", time.Second)
	suite.Require().Len(locks, 1)
	amount = suite.app.LockupKeeper.GetLockedDenom(suite.ctx, "stake", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check queries for shadow denom after shadow
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "stakestakedtovalidator1", suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr1, "stakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "stakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stakestakedtovalidator1", suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 1)
	amount = suite.app.LockupKeeper.GetLockedDenom(suite.ctx, "stakestakedtovalidator1", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check non-denom related queries after shadow
	moduleBalance = suite.app.LockupKeeper.GetModuleBalance(suite.ctx)
	suite.Require().Equal(moduleBalance.String(), "10stake")
	moduleLockedCoins = suite.app.LockupKeeper.GetModuleLockedCoins(suite.ctx)
	suite.Require().Equal(moduleLockedCoins.String(), "10stake")
	accountUnlockableCoins = suite.app.LockupKeeper.GetAccountUnlockableCoins(suite.ctx, addr1)
	suite.Require().Equal(accountUnlockableCoins.String(), "")
	accountUnlockingCoins = suite.app.LockupKeeper.GetAccountUnlockingCoins(suite.ctx, addr1)
	suite.Require().Equal(accountUnlockingCoins.String(), "")
	accountLockedCoins = suite.app.LockupKeeper.GetAccountLockedCoins(suite.ctx, addr1)
	suite.Require().Equal(accountLockedCoins.String(), "10stake")
	locks = suite.app.LockupKeeper.GetAccountLockedPastTime(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Len(locks, 1)
	locks = suite.app.LockupKeeper.GetAccountUnlockedBeforeTime(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Len(locks, 0)

	err = suite.app.LockupKeeper.DeleteShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)

	shadowLock, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().Error(err)
	suite.Require().Nil(shadowLock)

	// check accumulation store is correctly updated for shadow lock
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "0")
}

func (suite *KeeperTestSuite) TestShadowLockupDeleteAllShadowsByLockup() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	err = suite.app.LockupKeeper.CreateShadowLockup(suite.ctx, 1, "stakedtovalidator1", false)
	suite.Require().NoError(err)

	shadowLock, err := suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator1",
		EndTime: time.Time{},
	})

	err = suite.app.LockupKeeper.DeleteAllShadowsByLockup(suite.ctx, 1)
	suite.Require().NoError(err)

	shadowLock, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().Error(err)
	suite.Require().Nil(shadowLock)
}

func (suite *KeeperTestSuite) TestShadowLockupDeleteAllMaturedShadowLocks() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	err = suite.app.LockupKeeper.CreateShadowLockup(suite.ctx, 1, "stakedtovalidator1", false)
	suite.Require().NoError(err)

	err = suite.app.LockupKeeper.CreateShadowLockup(suite.ctx, 1, "stakedtovalidator2", true)
	suite.Require().NoError(err)

	shadowLock, err := suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator1",
		EndTime: time.Time{},
	})
	shadowLock, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator2")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator2",
		EndTime: suite.ctx.BlockTime().Add(time.Second),
	})

	suite.app.LockupKeeper.DeleteAllMaturedShadowLocks(suite.ctx)

	_, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator2")
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 2))
	suite.app.LockupKeeper.DeleteAllMaturedShadowLocks(suite.ctx)

	_, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator2")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestResetAllShadowLocks() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	suite.app.LockupKeeper.ResetAllShadowLocks(suite.ctx, []types.ShadowLock{
		{
			LockId:  1,
			Shadow:  "stakedtovalidator1",
			EndTime: time.Time{},
		},
		{
			LockId:  1,
			Shadow:  "stakedtovalidator2",
			EndTime: suite.ctx.BlockTime().Add(time.Second),
		},
	})

	shadowLock, err := suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator1",
		EndTime: time.Time{},
	})
	shadowLock, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator2")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator2",
		EndTime: suite.ctx.BlockTime().Add(time.Second),
	})

	accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stakestakedtovalidator2",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
}
