package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v8/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v8/x/lockup/types"
)

func (suite *KeeperTestSuite) TestRelock() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	lock := types.NewPeriodLock(1, addr1, time.Second, suite.ctx.BlockTime().Add(time.Second), coins)

	// lock with balance
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, coins)
	suite.Require().NoError(err)
	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)

	// lock with balance with same id
	coins2 := sdk.Coins{sdk.NewInt64Coin("stake2", 10)}
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, coins2)
	suite.Require().NoError(err)
	err = keeper.AdminKeeper{*suite.app.LockupKeeper}.Relock(suite.ctx, lock.ID, coins2)
	suite.Require().NoError(err)

	storedLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lock.ID)
	suite.Require().NoError(err)

	suite.Require().Equal(storedLock.Coins, coins2)
}

func (suite *KeeperTestSuite) BreakLock() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	lock := types.NewPeriodLock(1, addr1, time.Second, suite.ctx.BlockTime().Add(time.Second), coins)

	// lock with balance
	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, coins)
	suite.Require().NoError(err)

	err = suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)

	// break lock
	err = keeper.AdminKeeper{*suite.app.LockupKeeper}.BreakLock(suite.ctx, lock.ID)
	suite.Require().NoError(err)

	_, err = suite.app.LockupKeeper.GetLockByID(suite.ctx, lock.ID)
	suite.Require().Error(err)
}
