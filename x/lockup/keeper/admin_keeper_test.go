package keeper_test

import (
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/keeper"
	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestRageQuit() {
	suite.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	lock := types.NewPeriodLock(1, addr1, time.Second, suite.ctx.BlockTime().Add(time.Second), coins)

	// lock with balance
	suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	err := suite.app.LockupKeeper.Lock(suite.ctx, lock)
	suite.Require().NoError(err)

	// lock with balance with same id
	coins2 := sdk.Coins{sdk.NewInt64Coin("stake2", 10)}
	suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins2)
	err = keeper.AdminKeeper{suite.app.LockupKeeper}.RageQuit(suite.ctx, lock.ID, coins2)
	suite.Require().NoError(err)

	storedLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lock.ID)
	suite.Require().NoError(err)

	suite.Require().Equal(storedLock.Coins, coins2)
}
