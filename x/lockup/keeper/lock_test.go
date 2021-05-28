package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (suite *KeeperTestSuite) TestBeginUnlocking() { // test for all unlockable coins
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), false)

	// begin unlock
	locks, unlockCoins, err := suite.app.LockupKeeper.BeginUnlockAllNotUnlockings(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(unlockCoins, coins)
	suite.Require().Equal(locks[0].ID, uint64(1))

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().NotEqual(locks[0].EndTime, time.Time{})
	suite.Require().NotEqual(locks[0].IsUnlocking(), false)
}

func (suite *KeeperTestSuite) TestBeginUnlockPeriodLock() {
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), false)

	// begin unlock
	lock1, err := suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx, 1)
	suite.Require().NoError(err)
	suite.Require().Equal(lock1.ID, uint64(1))

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().NotEqual(locks[0].EndTime, time.Time{})
	suite.Require().NotEqual(locks[0].IsUnlocking(), false)
}

func (suite *KeeperTestSuite) TestGetPeriodLocks() {
	suite.SetupTest()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestUnlockAllUnlockableCoins() {
	suite.SetupTest()
	now := suite.ctx.BlockTime()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)
	suite.LockTokens(addr2, coins, time.Second)

	// unlock locks just now
	unlocks1, ucoins1, err := suite.app.LockupKeeper.UnlockAllUnlockableCoins(suite.ctx, addr1)
	suite.Require().Equal(ucoins1, sdk.Coins{})
	suite.Require().Len(unlocks1, 0)

	// unlock locks after 1s before starting unlock
	unlocks2, ucoins2, err := suite.app.LockupKeeper.UnlockAllUnlockableCoins(suite.ctx.WithBlockTime(now.Add(time.Second)), addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(ucoins2, sdk.Coins{})
	suite.Require().Len(unlocks2, 0)

	// begin unlock after 1s
	unlocks3, ucoins3, err := suite.app.LockupKeeper.BeginUnlockAllNotUnlockings(suite.ctx.WithBlockTime(now.Add(time.Second)), addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(ucoins3, coins)
	suite.Require().Len(unlocks3, 1)

	// unlock after 1s begin unlock
	unlocks4, ucoins4, err := suite.app.LockupKeeper.UnlockAllUnlockableCoins(suite.ctx.WithBlockTime(now.Add(time.Second*2)), addr1)
	suite.Require().NoError(err)
	suite.Require().Equal(ucoins4, coins)
	suite.Require().Len(unlocks4, 1)

	// check addr1 locks, no lock now
	locks = suite.app.LockupKeeper.GetAccountPeriodLocks(suite.ctx, addr1)
	suite.Require().Len(locks, 0)

	// check addr2 locks, still 1 as noone unlocked it yet
	locks = suite.app.LockupKeeper.GetAccountPeriodLocks(suite.ctx, addr2)
	suite.Require().Len(locks, 1)

	// totally 1 lock
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestUnlockPeriodLockByID() {
	suite.SetupTest()
	now := suite.ctx.BlockTime()

	// initial check
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// unlock lock just now
	lock1, err := suite.app.LockupKeeper.UnlockPeriodLockByID(suite.ctx, 1)
	suite.Require().Error(err)
	suite.Require().Equal(lock1.ID, uint64(1))

	// unlock lock after 1s before starting unlock
	lock2, err := suite.app.LockupKeeper.UnlockPeriodLockByID(suite.ctx.WithBlockTime(now.Add(time.Second)), 1)
	suite.Require().Error(err)
	suite.Require().Equal(lock2.ID, uint64(1))

	// begin unlock
	lock3, err := suite.app.LockupKeeper.BeginUnlockPeriodLockByID(suite.ctx.WithBlockTime(now.Add(time.Second)), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(lock3.ID, uint64(1))

	// unlock 1s after begin unlock
	lock4, err := suite.app.LockupKeeper.UnlockPeriodLockByID(suite.ctx.WithBlockTime(now.Add(time.Second*2)), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(lock4.ID, uint64(1))

	// check locks
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)
}

func (suite *KeeperTestSuite) TestLock() {
	// test for coin locking
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	lock := types.NewPeriodLock(1, addr1, time.Second, suite.ctx.BlockTime().Add(time.Second), coins)

	// try lock without balance
	err := suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().Error(err)

	// lock with balance
	suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)

	// lock with balance with same id
	suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().Error(err)

	// lock with balance with different id
	lock = types.NewPeriodLock(2, addr1, time.Second, suite.ctx.BlockTime().Add(time.Second), coins)
	suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestUnlock() {
	// test for coin unlocking
	suite.SetupTest()
	now := suite.ctx.BlockTime()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	lock := types.NewPeriodLock(1, addr1, time.Second, now.Add(time.Second), coins)

	// lock with balance
	suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	err := suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)

	// begin unlock with lock object
	err = suite.app.LockupKeeper.BeginUnlock(suite.ctx, lock)
	suite.Require().NoError(err)

	// unlock with lock object
	err = suite.app.LockupKeeper.Unlock(suite.ctx.WithBlockTime(now.Add(time.Second)), lock)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestModuleLockedCoins() {
	suite.SetupTest()

	// initial check
	lockedCoins := suite.app.LockupKeeper.GetModuleLockedCoins(suite.ctx)
	suite.Require().Equal(lockedCoins, sdk.Coins(nil))

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// final check
	lockedCoins = suite.app.LockupKeeper.GetModuleLockedCoins(suite.ctx)
	suite.Require().Equal(lockedCoins, coins)
}

func (suite *KeeperTestSuite) TestLocksPastTimeDenom() {
	suite.SetupTest()

	now := time.Now()
	suite.ctx = suite.ctx.WithBlockTime(now)

	// initial check
	locks := suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", now)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// final check
	locks = suite.app.LockupKeeper.GetLocksPastTimeDenom(suite.ctx, "stake", now)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestLocksLongerThanDurationDenom() {
	suite.SetupTest()

	// initial check
	duration := time.Second
	locks := suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", duration)
	suite.Require().Len(locks, 0)

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// final check
	locks = suite.app.LockupKeeper.GetLocksLongerThanDurationDenom(suite.ctx, "stake", duration)
	suite.Require().Len(locks, 1)
}
