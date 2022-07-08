package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v10/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestSyntheticLockupCreation() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	// create not unbonding synthetic lockup
	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "suffix1", time.Second, false)
	suite.Require().NoError(err)

	// try creating same suffix synthetic lockup for a single lockup
	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "suffix1", time.Second, true)
	suite.Require().Error(err)

	// create unbonding synthetic lockup
	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "suffix2", time.Second, true)
	suite.Require().NoError(err)

	// try creating unbonding synthetic lockup that is long than native lockup unbonding duration
	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "suffix3", time.Second*2, true)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestSyntheticLockupCreateGetDeleteAccumulation() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	expectedLocks := []types.PeriodLock{
		{
			ID:       1,
			Owner:    addr1.String(),
			Duration: time.Second,
			EndTime:  time.Time{},
			Coins:    coins,
		},
	}
	// check locks
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(locks, expectedLocks)

	// check accumulation store is correctly updated
	accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// check queries for native denom before creating synthetic lockup
	locks = suite.App.LockupKeeper.GetAccountLockedPastTimeDenom(suite.Ctx, addr1, "stake", suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.Ctx, addr1, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.Ctx, addr1, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetLocksPastTimeDenom(suite.Ctx, "stake", suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetLocksLongerThanDurationDenom(suite.Ctx, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	amount := suite.App.LockupKeeper.GetLockedDenom(suite.Ctx, "stake", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check queries for synthetic denom before creating synthetic lockup
	locks = suite.App.LockupKeeper.GetAccountLockedPastTimeDenom(suite.Ctx, addr1, "synthstakestakedtovalidator1", suite.Ctx.BlockTime())
	suite.Require().Len(locks, 0)
	locks = suite.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.Ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	locks = suite.App.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.Ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	locks = suite.App.LockupKeeper.GetLocksPastTimeDenom(suite.Ctx, "synthstakestakedtovalidator1", suite.Ctx.BlockTime())
	suite.Require().Len(locks, 0)
	locks = suite.App.LockupKeeper.GetLocksLongerThanDurationDenom(suite.Ctx, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	amount = suite.App.LockupKeeper.GetLockedDenom(suite.Ctx, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(amount.String(), "0")

	// check non-denom related queries before creating synthetic lockup
	moduleBalance := suite.App.LockupKeeper.GetModuleBalance(suite.Ctx)
	suite.Require().Equal(moduleBalance.String(), "10stake")
	moduleLockedCoins := suite.App.LockupKeeper.GetModuleLockedCoins(suite.Ctx)
	suite.Require().Equal(moduleLockedCoins.String(), "10stake")
	accountUnlockableCoins := suite.App.LockupKeeper.GetAccountUnlockableCoins(suite.Ctx, addr1)
	suite.Require().Equal(accountUnlockableCoins.String(), "")
	accountUnlockingCoins := suite.App.LockupKeeper.GetAccountUnlockingCoins(suite.Ctx, addr1)
	suite.Require().Equal(accountUnlockingCoins.String(), "")
	accountLockedCoins := suite.App.LockupKeeper.GetAccountLockedCoins(suite.Ctx, addr1)
	suite.Require().Equal(accountLockedCoins.String(), "10stake")
	locks = suite.App.LockupKeeper.GetAccountLockedPastTime(suite.Ctx, addr1, suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(suite.Ctx, addr1, suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountUnlockedBeforeTime(suite.Ctx, addr1, suite.Ctx.BlockTime())
	suite.Require().Len(locks, 0)

	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	suite.Require().NoError(err)

	synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})

	expectedSynthLocks := []types.SyntheticLock{*synthLock}
	synthLocks := suite.App.LockupKeeper.GetAllSyntheticLockupsByLockup(suite.Ctx, 1)
	suite.Require().Equal(synthLocks, expectedSynthLocks)

	synthLocks = suite.App.LockupKeeper.GetAllSyntheticLockups(suite.Ctx)
	suite.Require().Equal(synthLocks, expectedSynthLocks)

	// check accumulation store is correctly updated for synthetic lockup
	accum = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// check queries for native denom after creating synthetic lockup
	locks = suite.App.LockupKeeper.GetAccountLockedPastTimeDenom(suite.Ctx, addr1, "stake", suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.Ctx, addr1, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.Ctx, addr1, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetLocksPastTimeDenom(suite.Ctx, "stake", suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetLocksLongerThanDurationDenom(suite.Ctx, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	amount = suite.App.LockupKeeper.GetLockedDenom(suite.Ctx, "stake", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check queries for synthetic denom after creating synthetic lockup
	locks = suite.App.LockupKeeper.GetAccountLockedPastTimeDenom(suite.Ctx, addr1, "synthstakestakedtovalidator1", suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.Ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.Ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetLocksPastTimeDenom(suite.Ctx, "synthstakestakedtovalidator1", suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetLocksLongerThanDurationDenom(suite.Ctx, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	amount = suite.App.LockupKeeper.GetLockedDenom(suite.Ctx, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check non-denom related queries after creating synthetic lockup
	moduleBalance = suite.App.LockupKeeper.GetModuleBalance(suite.Ctx)
	suite.Require().Equal(moduleBalance.String(), "10stake")
	moduleLockedCoins = suite.App.LockupKeeper.GetModuleLockedCoins(suite.Ctx)
	suite.Require().Equal(moduleLockedCoins.String(), "10stake")
	accountUnlockableCoins = suite.App.LockupKeeper.GetAccountUnlockableCoins(suite.Ctx, addr1)
	suite.Require().Equal(accountUnlockableCoins.String(), "")
	accountUnlockingCoins = suite.App.LockupKeeper.GetAccountUnlockingCoins(suite.Ctx, addr1)
	suite.Require().Equal(accountUnlockingCoins.String(), "")
	accountLockedCoins = suite.App.LockupKeeper.GetAccountLockedCoins(suite.Ctx, addr1)
	suite.Require().Equal(accountLockedCoins.String(), "10stake")
	locks = suite.App.LockupKeeper.GetAccountLockedPastTime(suite.Ctx, addr1, suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(suite.Ctx, addr1, suite.Ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.App.LockupKeeper.GetAccountUnlockedBeforeTime(suite.Ctx, addr1, suite.Ctx.BlockTime())
	suite.Require().Len(locks, 0)

	// try creating synthetic lockup with same lock and suffix
	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	suite.Require().Error(err)

	// delete synthetic lockup
	err = suite.App.LockupKeeper.DeleteSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)

	synthLock, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().Error(err)
	suite.Require().Nil(synthLock)

	// check accumulation store is correctly updated for synthetic lock
	accum = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "0")
}

func (suite *KeeperTestSuite) TestSyntheticLockupDeleteAllMaturedSyntheticLocks() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	suite.Require().NoError(err)

	err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator2", time.Second, true)
	suite.Require().NoError(err)

	synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})
	synthLock, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator2")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator2",
		EndTime:          suite.Ctx.BlockTime().Add(time.Second),
		Duration:         time.Second,
	})

	suite.App.LockupKeeper.DeleteAllMaturedSyntheticLocks(suite.Ctx)

	_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator2")
	suite.Require().NoError(err)

	suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(time.Second * 2))
	suite.App.LockupKeeper.DeleteAllMaturedSyntheticLocks(suite.Ctx)

	_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator2")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestResetAllSyntheticLocks() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	suite.App.LockupKeeper.ResetAllSyntheticLocks(suite.Ctx, []types.SyntheticLock{
		{
			UnderlyingLockId: 1,
			SynthDenom:       "synthstakestakedtovalidator1",
			EndTime:          time.Time{},
			Duration:         time.Second,
		},
		{
			UnderlyingLockId: 1,
			SynthDenom:       "synthstakestakedtovalidator2",
			EndTime:          suite.Ctx.BlockTime().Add(time.Second),
			Duration:         time.Second,
		},
	})

	synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})
	synthLock, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, 1, "synthstakestakedtovalidator2")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator2",
		EndTime:          suite.Ctx.BlockTime().Add(time.Second),
		Duration:         time.Second,
	})

	accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
	accum = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
	accum = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator2",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
}
