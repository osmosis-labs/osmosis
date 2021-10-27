package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/lockup/types"
)

// TODO: add test for ResetAllShadowLocks
// TODO: add test for querying shadow denoms
// TODO: add test for querying non-shadow queries after adding shadows

func (suite *KeeperTestSuite) TestShadowLockupCreateGetDeleteAccumulation() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

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

	err = suite.app.LockupKeeper.CreateShadowLockup(suite.ctx, 1, "stakedtovalidator1", false)
	suite.Require().NoError(err)

	shadowLock, err := suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator1",
		EndTime: time.Time{},
	})

	shadowLocks := suite.app.LockupKeeper.GetAllShadowsByLockup(suite.ctx, 1)
	suite.Require().Len(shadowLocks, 1)
	suite.Require().Equal(*shadowLock, shadowLocks[0])

	shadowLocks = suite.app.LockupKeeper.GetAllShadows(suite.ctx)
	suite.Require().Len(shadowLocks, 1)
	suite.Require().Equal(*shadowLock, shadowLocks[0])

	// check accumulation store is correctly updated for shadow lock
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "10")

	err = suite.app.LockupKeeper.DeleteShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)

	shadowLock, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().Error(err)
	suite.Require().Nil(shadowLock)

	// check accumulation store is correctly updated for shadow lock
	accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
		LockQueryType: types.ByDuration,
		Denom:         "stakestakedtovalidator1",
		Duration:      time.Second,
	})
	suite.Require().Equal(accum.String(), "0")
}

func (suite *KeeperTestSuite) TestShadowLockupDeleteAllShadowsByLockup() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	err = suite.app.LockupKeeper.CreateShadowLockup(suite.ctx, 1, "stakedtovalidator1", false)
	suite.Require().NoError(err)

	shadowLock, err := suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator1",
		EndTime: time.Time{},
	})

	err = suite.app.LockupKeeper.DeleteAllShadowsByLockup(suite.ctx, 1)
	suite.Require().NoError(err)

	shadowLock, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().Error(err)
	suite.Require().Nil(shadowLock)
}

func (suite *KeeperTestSuite) TestShadowLockupDeleteAllMaturedShadowLocks() {
	suite.SetupTest()

	// lock coins
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.LockTokens(addr1, coins, time.Second)

	// check locks
	locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(locks, 1)
	suite.Require().Equal(locks[0].Coins, coins)

	err = suite.app.LockupKeeper.CreateShadowLockup(suite.ctx, 1, "stakedtovalidator1", false)
	suite.Require().NoError(err)

	err = suite.app.LockupKeeper.CreateShadowLockup(suite.ctx, 1, "stakedtovalidator2", true)
	suite.Require().NoError(err)

	shadowLock, err := suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator1",
		EndTime: time.Time{},
	})
	shadowLock, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator2")
	suite.Require().NoError(err)
	suite.Require().Equal(*shadowLock, types.ShadowLock{
		LockId:  1,
		Shadow:  "stakedtovalidator2",
		EndTime: suite.ctx.BlockTime().Add(time.Second),
	})

	suite.app.LockupKeeper.DeleteAllMaturedShadowLocks(suite.ctx)

	_, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator2")
	suite.Require().NoError(err)

	suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(time.Second * 2))
	suite.app.LockupKeeper.DeleteAllMaturedShadowLocks(suite.ctx)

	_, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator1")
	suite.Require().NoError(err)
	_, err = suite.app.LockupKeeper.GetShadowLockup(suite.ctx, 1, "stakedtovalidator2")
	suite.Require().Error(err)
}
