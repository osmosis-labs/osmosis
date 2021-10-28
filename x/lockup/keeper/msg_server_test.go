package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

// TODO: Make table driven
func (suite *KeeperTestSuite) TestMsgLockTokens() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	err := suite.app.BankKeeper.SetBalances(suite.ctx, addr1, coins)
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.LockTokens(suite.ctx, addr1, coins, time.Second)
	suite.Require().NoError(err)

	// creation of lock via LockTokens
	msgServer := keeper.NewMsgServerImpl(suite.app.LockupKeeper)
	_, err = msgServer.LockTokens(sdk.WrapSDKContext(suite.ctx), types.NewMsgLockTokens(addr1, time.Second, coins))

	// check locks
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)
	// check accumulation store is correctly updated
	accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	// add more tokens to lock via LockTokens
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	err = suite.app.BankKeeper.SetBalances(suite.ctx, addr1, addCoins)
	suite.Require().NoError(err)

	_, err = msgServer.LockTokens(sdk.WrapSDKContext(suite.ctx), types.NewMsgLockTokens(addr1, locks[0].Duration, addCoins))
	suite.Require().NoError(err)

	// check locks after adding tokens to lock
	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins.Add(addCoins...))

	// check accumulation store is correctly updated
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stake",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "20")
}

func (suite *KeeperTestSuite) TestMsgBeginPartialUnlocking() {
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

	// 1st begin partial unlocking
	msgServer := keeper.NewMsgServerImpl(suite.app.LockupKeeper)
	_, err = msgServer.BeginPartialUnlocking(sdk.WrapSDKContext(suite.ctx), types.NewMsgBeginPartialUnlocking(addr1, 1, sdk.Coins{sdk.NewInt64Coin("stake", 1)}))
	suite.Require().NoError(err)

	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)

	// check locks
	suite.Require().Len(locks, 2)
	suite.Require().Equal(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), false)
	suite.Require().Equal(locks[0].Coins.AmountOf("stake"), sdk.NewInt(9))
	suite.Require().NotEqual(locks[1].EndTime, time.Time{})
	suite.Require().Equal(locks[1].IsUnlocking(), true)
	suite.Require().Equal(locks[1].Coins.AmountOf("stake"), sdk.NewInt(1))

	// 2nd begin partial unlocking (all amount)
	_, err = msgServer.BeginPartialUnlocking(sdk.WrapSDKContext(suite.ctx), types.NewMsgBeginPartialUnlocking(addr1, 1, sdk.Coins{sdk.NewInt64Coin("stake", 9)}))
	suite.Require().NoError(err)

	locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)

	// check locks
	suite.Require().Len(locks, 2)
	suite.Require().NotEqual(locks[0].EndTime, time.Time{})
	suite.Require().Equal(locks[0].IsUnlocking(), true)
	suite.Require().Equal(locks[0].Coins.AmountOf("stake"), sdk.NewInt(9))
	suite.Require().NotEqual(locks[1].EndTime, time.Time{})
	suite.Require().Equal(locks[1].IsUnlocking(), true)
	suite.Require().Equal(locks[1].Coins.AmountOf("stake"), sdk.NewInt(1))
}
