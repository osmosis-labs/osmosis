package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v11/x/lockup/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestRelock() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	// lock with balance
	suite.FundAcc(addr1, coins)
	lock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr1, coins, time.Second)
	suite.Require().NoError(err)

	// lock with balance with same id
	coins2 := sdk.Coins{sdk.NewInt64Coin("stake2", 10)}
	suite.FundAcc(addr1, coins2)
	err = keeper.AdminKeeper{*suite.App.LockupKeeper}.Relock(suite.Ctx, lock.ID, coins2)
	suite.Require().NoError(err)

	storedLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
	suite.Require().NoError(err)

	suite.Require().Equal(storedLock.Coins, coins2)
}

func (suite *KeeperTestSuite) BreakLock() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	// lock with balance
	suite.FundAcc(addr1, coins)

	lock, err := suite.App.LockupKeeper.CreateLock(suite.Ctx, addr1, coins, time.Second)

	suite.Require().NoError(err)

	// break lock
	err = keeper.AdminKeeper{*suite.App.LockupKeeper}.BreakLock(suite.Ctx, lock.ID)
	suite.Require().NoError(err)

	_, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
	suite.Require().Error(err)
}
