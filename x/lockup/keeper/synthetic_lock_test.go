package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestSyntheticLockupCreation() {
	s.SetupTest()
	numLocks := 3

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	for i := 0; i < numLocks; i++ {
		s.LockTokens(addr1, coins, time.Second)
	}

	// check locks
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, numLocks)
	for i := 0; i < numLocks; i++ {
		s.Require().Equal(locks[i].Coins, coins)
	}

	// create not unbonding synthetic lockup
	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 1, "suffix1", time.Second, false)
	s.Require().NoError(err)

	// try creating same suffix synthetic lockup for a single lockup
	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 1, "suffix1", time.Second, true)
	s.Require().Error(err)

	// create unbonding synthetic lockup, lock already has a synthetic lock associated with it so it should fail
	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 1, "suffix2", time.Second, true)
	s.Require().Error(err)

	// create unbonding synthetic lockup with new lock id, should succeed
	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 2, "suffix2", time.Second, true)
	s.Require().NoError(err)

	// try creating unbonding synthetic lockup that is long than native lockup unbonding duration
	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 3, "suffix3", time.Second*2, true)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestSyntheticLockupCreateGetDeleteAccumulation() {
	s.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	expectedLocks := []types.PeriodLock{
		{
			ID:                    1,
			Owner:                 addr1.String(),
			RewardReceiverAddress: "",
			Duration:              time.Second,
			EndTime:               time.Time{},
			Coins:                 coins,
		},
	}
	// check locks
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal(locks, expectedLocks)

	// check accumulation store is correctly updated
	accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "10")

	// check queries for native denom before creating synthetic lockup
	locks = s.App.LockupKeeper.GetAccountLockedPastTimeDenom(s.Ctx, addr1, "stake", s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(s.Ctx, addr1, "stake", time.Second)
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountLockedLongerDurationDenom(s.Ctx, addr1, "stake", time.Second)
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetLocksPastTimeDenom(s.Ctx, "stake", s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetLocksLongerThanDurationDenom(s.Ctx, "stake", time.Second)
	s.Require().Equal(locks, expectedLocks)
	amount := s.App.LockupKeeper.GetLockedDenom(s.Ctx, "stake", time.Second)
	s.Require().Equal(amount.String(), "10")

	// check queries for synthetic denom before creating synthetic lockup
	locks = s.App.LockupKeeper.GetAccountLockedPastTimeDenom(s.Ctx, addr1, "synthstakestakedtovalidator1", s.Ctx.BlockTime())
	s.Require().Len(locks, 0)
	locks = s.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(s.Ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	s.Require().Len(locks, 0)
	locks = s.App.LockupKeeper.GetAccountLockedLongerDurationDenom(s.Ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	s.Require().Len(locks, 0)
	locks = s.App.LockupKeeper.GetLocksPastTimeDenom(s.Ctx, "synthstakestakedtovalidator1", s.Ctx.BlockTime())
	s.Require().Len(locks, 0)
	locks = s.App.LockupKeeper.GetLocksLongerThanDurationDenom(s.Ctx, "synthstakestakedtovalidator1", time.Second)
	s.Require().Len(locks, 0)
	amount = s.App.LockupKeeper.GetLockedDenom(s.Ctx, "synthstakestakedtovalidator1", time.Second)
	s.Require().Equal(amount.String(), "0")

	// check non-denom related queries before creating synthetic lockup
	moduleBalance := s.App.LockupKeeper.GetModuleBalance(s.Ctx)
	s.Require().Equal(moduleBalance.String(), "10stake")
	moduleLockedCoins := s.App.LockupKeeper.GetModuleLockedCoins(s.Ctx)
	s.Require().Equal(moduleLockedCoins.String(), "10stake")
	accountUnlockableCoins := s.App.LockupKeeper.GetAccountUnlockableCoins(s.Ctx, addr1)
	s.Require().Equal(accountUnlockableCoins.String(), "")
	accountUnlockingCoins := s.App.LockupKeeper.GetAccountUnlockingCoins(s.Ctx, addr1)
	s.Require().Equal(accountUnlockingCoins.String(), "")
	accountLockedCoins := s.App.LockupKeeper.GetAccountLockedCoins(s.Ctx, addr1)
	s.Require().Equal(accountLockedCoins.String(), "10stake")
	locks = s.App.LockupKeeper.GetAccountLockedPastTime(s.Ctx, addr1, s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(s.Ctx, addr1, s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountUnlockedBeforeTime(s.Ctx, addr1, s.Ctx.BlockTime())
	s.Require().Len(locks, 0)

	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	s.Require().NoError(err)

	synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1")
	s.Require().NoError(err)
	s.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})

	expectedSynthLock := *synthLock
	actualSynthLock, found, err := s.App.LockupKeeper.GetSyntheticLockupByUnderlyingLockId(s.Ctx, 1)
	s.Require().NoError(err)
	s.Require().True(found)
	s.Require().Equal(expectedSynthLock, actualSynthLock)

	allSynthLocks := s.App.LockupKeeper.GetAllSyntheticLockups(s.Ctx)
	s.Require().Equal([]lockuptypes.SyntheticLock{expectedSynthLock}, allSynthLocks)

	// check accumulation store is correctly updated for synthetic lockup
	accum = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator1",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "10")

	// check queries for native denom after creating synthetic lockup
	locks = s.App.LockupKeeper.GetAccountLockedPastTimeDenom(s.Ctx, addr1, "stake", s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(s.Ctx, addr1, "stake", time.Second)
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountLockedLongerDurationDenom(s.Ctx, addr1, "stake", time.Second)
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetLocksPastTimeDenom(s.Ctx, "stake", s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetLocksLongerThanDurationDenom(s.Ctx, "stake", time.Second)
	s.Require().Equal(locks, expectedLocks)
	amount = s.App.LockupKeeper.GetLockedDenom(s.Ctx, "stake", time.Second)
	s.Require().Equal(amount.String(), "10")

	// check queries for synthetic denom after creating synthetic lockup
	locks = s.App.LockupKeeper.GetAccountLockedPastTimeDenom(s.Ctx, addr1, "synthstakestakedtovalidator1", s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountLockedDurationNotUnlockingOnly(s.Ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountLockedLongerDurationDenom(s.Ctx, addr1, "synthstakestakedtovalidator1", time.Second)
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetLocksPastTimeDenom(s.Ctx, "synthstakestakedtovalidator1", s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetLocksLongerThanDurationDenom(s.Ctx, "synthstakestakedtovalidator1", time.Second)
	s.Require().Equal(locks, expectedLocks)
	amount = s.App.LockupKeeper.GetLockedDenom(s.Ctx, "synthstakestakedtovalidator1", time.Second)
	s.Require().Equal(amount.String(), "10")

	// check non-denom related queries after creating synthetic lockup
	moduleBalance = s.App.LockupKeeper.GetModuleBalance(s.Ctx)
	s.Require().Equal(moduleBalance.String(), "10stake")
	moduleLockedCoins = s.App.LockupKeeper.GetModuleLockedCoins(s.Ctx)
	s.Require().Equal(moduleLockedCoins.String(), "10stake")
	accountUnlockableCoins = s.App.LockupKeeper.GetAccountUnlockableCoins(s.Ctx, addr1)
	s.Require().Equal(accountUnlockableCoins.String(), "")
	accountUnlockingCoins = s.App.LockupKeeper.GetAccountUnlockingCoins(s.Ctx, addr1)
	s.Require().Equal(accountUnlockingCoins.String(), "")
	accountLockedCoins = s.App.LockupKeeper.GetAccountLockedCoins(s.Ctx, addr1)
	s.Require().Equal(accountLockedCoins.String(), "10stake")
	locks = s.App.LockupKeeper.GetAccountLockedPastTime(s.Ctx, addr1, s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountLockedPastTimeNotUnlockingOnly(s.Ctx, addr1, s.Ctx.BlockTime())
	s.Require().Equal(locks, expectedLocks)
	locks = s.App.LockupKeeper.GetAccountUnlockedBeforeTime(s.Ctx, addr1, s.Ctx.BlockTime())
	s.Require().Len(locks, 0)

	// try creating synthetic lockup with same lock and suffix
	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	s.Require().Error(err)

	// delete synthetic lockup
	err = s.App.LockupKeeper.DeleteSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1")
	s.Require().NoError(err)

	synthLock, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1")
	s.Require().Error(err)
	s.Require().Nil(synthLock)

	// check accumulation store is correctly updated for synthetic lock
	accum = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator1",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "0")
}

func (s *KeeperTestSuite) TestSyntheticLockupDeleteAllMaturedSyntheticLocks() {
	s.SetupTest()
	numLocks := 2

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	for i := 0; i < numLocks; i++ {
		s.LockTokens(addr1, coins, time.Second)
	}

	// check locks
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, numLocks)
	for i := 0; i < numLocks; i++ {
		s.Require().Equal(locks[i].Coins, coins)
	}

	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1", time.Second, false)
	s.Require().NoError(err)

	err = s.App.LockupKeeper.CreateSyntheticLockup(s.Ctx, 2, "synthstakestakedtovalidator2", time.Second, true)
	s.Require().NoError(err)

	synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1")
	s.Require().NoError(err)
	s.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})
	synthLock, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 2, "synthstakestakedtovalidator2")
	s.Require().NoError(err)
	s.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 2,
		SynthDenom:       "synthstakestakedtovalidator2",
		EndTime:          s.Ctx.BlockTime().Add(time.Second),
		Duration:         time.Second,
	})

	s.App.LockupKeeper.DeleteAllMaturedSyntheticLocks(s.Ctx)

	_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1")
	s.Require().NoError(err)
	_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 2, "synthstakestakedtovalidator2")
	s.Require().NoError(err)

	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second * 2))
	s.App.LockupKeeper.DeleteAllMaturedSyntheticLocks(s.Ctx)

	_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1")
	s.Require().NoError(err)
	_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 2, "synthstakestakedtovalidator2")
	s.Require().Error(err)
}

func (s *KeeperTestSuite) TestResetAllSyntheticLocks() {
	s.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	s.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := s.App.LockupKeeper.GetPeriodLocks(s.Ctx)
	s.Require().NoError(err)
	s.Require().Len(locks, 1)
	s.Require().Equal(locks[0].Coins, coins)

	err = s.App.LockupKeeper.InitializeAllSyntheticLocks(s.Ctx, []types.SyntheticLock{
		{
			UnderlyingLockId: 1,
			SynthDenom:       "synthstakestakedtovalidator1",
			EndTime:          time.Time{},
			Duration:         time.Second,
		},
		{
			UnderlyingLockId: 1,
			SynthDenom:       "synthstakestakedtovalidator2",
			EndTime:          s.Ctx.BlockTime().Add(time.Second),
			Duration:         time.Second,
		},
	})
	s.Require().NoError(err)

	synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator1")
	s.Require().NoError(err)
	s.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator1",
		EndTime:          time.Time{},
		Duration:         time.Second,
	})
	synthLock, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, 1, "synthstakestakedtovalidator2")
	s.Require().NoError(err)
	s.Require().Equal(*synthLock, types.SyntheticLock{
		UnderlyingLockId: 1,
		SynthDenom:       "synthstakestakedtovalidator2",
		EndTime:          s.Ctx.BlockTime().Add(time.Second),
		Duration:         time.Second,
	})

	accum := s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "10")
	accum = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator1",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "10")
	accum = s.App.LockupKeeper.GetPeriodLocksAccumulation(s.Ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "synthstakestakedtovalidator2",
		Duration:      time.Second,
	})
	s.Require().Equal(accum.String(), "10")
}
