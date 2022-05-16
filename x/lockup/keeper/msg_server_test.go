package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v8/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v8/x/lockup/types"
)

// TODO: Make table driven
func (suite *KeeperTestSuite) TestMsgLockTokens() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}

	err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, coins)
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
	err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, addCoins)
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
