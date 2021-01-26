package keeper_test

import (
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestModuleBalance() {
	// test for module balance check
	suite.SetupTest()

	// initial module balance check
	res, err := suite.app.LockupKeeper.ModuleBalance(sdk.WrapSDKContext(suite.ctx), &types.ModuleBalanceRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	duration := time.Second
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	ID := suite.app.LockupKeeper.GetLastLockID(suite.ctx) + 1
	lock := types.NewPeriodLock(ID, addr1, duration, time.Now().Add(duration), coins)
	suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)

	// final module balance check
	res, err = suite.app.LockupKeeper.ModuleBalance(sdk.WrapSDKContext(suite.ctx), &types.ModuleBalanceRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}

func (suite *KeeperTestSuite) TestModuleLockedAmount() {
	// TODO: write test for LockupKeeper.ModuleLockedAmount
}

func (suite *KeeperTestSuite) TestAccountUnlockableCoins() {
	// TODO: write test for LockupKeeper.AccountUnlockableCoins
}

func (suite *KeeperTestSuite) TestAccountLockedCoins() {
	// TODO: write test for LockupKeeper.AccountLockedCoins
}

func (suite *KeeperTestSuite) TestAccountLockedPastTime() {
	// TODO: write test for LockupKeeper.AccountLockedPastTime
}

func (suite *KeeperTestSuite) TestAccountUnlockedBeforeTime() {
	// TODO: write test for LockupKeeper.AccountUnlockedBeforeTime
}

func (suite *KeeperTestSuite) TestAccountLockedPastTimeDenom() {
	// TODO: write test for LockupKeeper.AccountLockedPastTimeDenom
}

func (suite *KeeperTestSuite) TestLocked() {
	// TODO: write test for LockupKeeper.Locked
}

func (suite *KeeperTestSuite) TestAccountLockedLongerThanDuration() {
	// TODO: write test for LockupKeeper.AccountLockedLongerThanDuration
}

func (suite *KeeperTestSuite) TestAccountLockedLongerThanDurationDenom() {
	// TODO: write test for LockupKeeper.AccountLockedLongerThanDurationDenom
}
