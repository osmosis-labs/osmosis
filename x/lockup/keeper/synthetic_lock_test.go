package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v8/x/lockup/types"
)

func (suite *KeeperTestSuite) TestSyntheticLockupCreation() {
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

	// create not unbonding synthetic lockup
	err = suite.app.LockupKeeper.CreateSyntheticLockup(suite.ctx, 1, "suffix1", time.Second, false)
	suite.Require().NoError(err)

	// try creating same suffix synthetic lockup for a single lockup
	err = suite.app.LockupKeeper.CreateSyntheticLockup(suite.ctx, 1, "suffix1", time.Second, true)
	suite.Require().Error(err)

	// create unbonding synthetic lockup
	err = suite.app.LockupKeeper.CreateSyntheticLockup(suite.ctx, 1, "suffix2", time.Second, true)
	suite.Require().NoError(err)

	// try creating unbonding synthetic lockup that is long than native lockup unbonding duration
	err = suite.app.LockupKeeper.CreateSyntheticLockup(suite.ctx, 1, "suffix3", time.Second*2, true)
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
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(locks, expectedLocks)

	// check accumulation store is correctly updated
	accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// check queries for native denom before creating synthetic lockup
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "stake", suite.ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr1, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", suite.ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	amount := suite.app.LockupKeeper.GetLockedDenom(suite.ctx, "stake", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check queries for synthetic denom before creating synthetic lockup
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "synthstakestakedtovalidator1", suite.ctx.BlockTime())
	suite.Require().Len(locks, 0)
	locks = suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "synthstakestakedtovalidator1", suite.ctx.BlockTime())
	suite.Require().Len(locks, 0)
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Len(locks, 0)
	amount = suite.app.LockupKeeper.GetLockedDenom(suite.ctx, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(amount.String(), "0")

	// check non-denom related queries before creating synthetic lockup
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
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountUnlockedBeforeTime(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Len(locks, 0)

	err = suite.app.LockupKeeper.CreateSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	suite.Require().NoError(err)

	synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})

	expectedSynthLocks := []types.SyntheticLock{*synthLock}
	synthLocks := suite.app.LockupKeeper.GetAllSyntheticLockupsByLockup(suite.ctx, 1)
	suite.Require().Equal(synthLocks, expectedSynthLocks)

	synthLocks = suite.app.LockupKeeper.GetAllSyntheticLockups(suite.ctx)
	suite.Require().Equal(synthLocks, expectedSynthLocks)

	// check accumulation store is correctly updated for synthetic lockup
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// check queries for native denom after creating synthetic lockup
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "stake", suite.ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr1, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", suite.ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	amount = suite.app.LockupKeeper.GetLockedDenom(suite.ctx, "stake", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check queries for synthetic denom after creating synthetic lockup
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeDenom(suite.ctx, addr1, "synthstakestakedtovalidator1", suite.ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(suite.ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountLockedLongerDurationDenom(suite.ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "synthstakestakedtovalidator1", suite.ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(locks, expectedLocks)
	amount = suite.app.LockupKeeper.GetLockedDenom(suite.ctx, "synthstakestakedtovalidator1", time.Second)
	suite.Require().Equal(amount.String(), "10")

	// check non-denom related queries after creating synthetic lockup
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
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Equal(locks, expectedLocks)
	locks = suite.app.LockupKeeper.GetAccountUnlockedBeforeTime(suite.ctx, addr1, suite.ctx.BlockTime())
	suite.Require().Len(locks, 0)

	// try creating synthetic lockup with same lock and suffix
	err = suite.app.LockupKeeper.CreateSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	suite.Require().Error(err)

	// delete synthetic lockup
	err = suite.app.LockupKeeper.DeleteSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)

	synthLock, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().Error(err)
	suite.Require().Nil(synthLock)

	// check accumulation store is correctly updated for synthetic lock
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
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
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	err = suite.app.LockupKeeper.CreateSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	suite.Require().NoError(err)

	err = suite.app.LockupKeeper.CreateSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator2", time.Second, true)
	suite.Require().NoError(err)

	synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})
	synthLock, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator2")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator2",
		EndTime:          suite.ctx.BlockTime().Add(time.Second),
		Duration:         time.Second,
	})

	suite.app.LockupKeeper.DeleteAllMaturedSyntheticLocks(suite.ctx)

	_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator2")
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 2))
	suite.app.LockupKeeper.DeleteAllMaturedSyntheticLocks(suite.ctx)

	_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator2")
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestResetAllSyntheticLocks() {
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

	suite.app.LockupKeeper.ResetAllSyntheticLocks(suite.ctx, []types.SyntheticLock{
		{
			UnderlyingLockId: 1,
			SynthDenom:       "synthstakestakedtovalidator1",
			EndTime:          time.Time{},
			Duration:         time.Second,
		},
		{
			UnderlyingLockId: 1,
			SynthDenom:       "synthstakestakedtovalidator2",
			EndTime:          suite.ctx.BlockTime().Add(time.Second),
			Duration:         time.Second,
		},
	})

	synthLock, err := suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})
	synthLock, err = suite.app.LockupKeeper.GetSyntheticLockup(suite.ctx, 1, "synthstakestakedtovalidator2")
	suite.Require().NoError(err)
	suite.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator2",
		EndTime:          suite.ctx.BlockTime().Add(time.Second),
		Duration:         time.Second,
	})

	accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator2",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")
}
