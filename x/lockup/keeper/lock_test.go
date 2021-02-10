package keeper_test

import (
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestGetPeriodLocks() {
	// test for module locked balance check
	suite.SetupTest()

	// initial module locked balance check
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
	// test for all unlockable coins
	suite.SetupTest()
	now := suite.ctx.BlockTime()

	// initial module locked balance check
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

	// unlock locks after 1s
	unlocks2, ucoins2, err := suite.app.LockupKeeper.UnlockAllUnlockableCoins(suite.ctx.WithBlockTime(now.Add(time.Second)), addr1)
	suite.Require().Equal(ucoins2, coins)
	suite.Require().Len(unlocks2, 1)

	// check addr1 locks, no lock now
	locks, err = suite.app.LockupKeeper.GetAccountPeriodLocks(suite.ctx, addr1)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 0)

	// check addr2 locks, still 1 as noone unlocked it yet
	locks, err = suite.app.LockupKeeper.GetAccountPeriodLocks(suite.ctx, addr2)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)

	// totally 1 lock
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
}

func (suite *KeeperTestSuite) TestUnlockPeriodLockByID() {
	// test for all unlockable coins
	suite.SetupTest()
	now := suite.ctx.BlockTime()

	// initial locks check
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

	// unlock lock after 1s
	lock2, err := suite.app.LockupKeeper.UnlockPeriodLockByID(suite.ctx.WithBlockTime(now.Add(time.Second)), 1)
	suite.Require().NoError(err)
	suite.Require().Equal(lock2.ID, uint64(1))

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

	// unlock with lock object
	err = suite.app.LockupKeeper.Unlock(suite.ctx.WithBlockTime(now.Add(time.Second)), lock)
	suite.Require().NoError(err)
}
