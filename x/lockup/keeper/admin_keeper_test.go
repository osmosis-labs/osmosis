package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestRelock() {
	s.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	// lock with balance
	s.FundAcc(addr1, coins)
	lock, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr1, coins, time.Second)
	s.Require().NoError(err)

	// lock with balance with same id
	coins2 := sdk.Coins{sdk.NewInt64Coin("stake2", 10)}
	s.FundAcc(addr1, coins2)
	err = keeper.AdminKeeper{*s.App.LockupKeeper}.Relock(s.Ctx, lock.ID, coins2)
	s.Require().NoError(err)

	storedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
	s.Require().NoError(err)

	s.Require().Equal(storedLock.Coins, coins2)
}

func (s *KeeperTestSuite) BreakLock() {
	s.SetupTest()

	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	// lock with balance
	s.FundAcc(addr1, coins)

	lock, err := s.App.LockupKeeper.CreateLock(s.Ctx, addr1, coins, time.Second)

	s.Require().NoError(err)

	// break lock
	err = keeper.AdminKeeper{*s.App.LockupKeeper}.BreakLock(s.Ctx, lock.ID)
	s.Require().NoError(err)

	_, err = s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
	s.Require().Error(err)
}
