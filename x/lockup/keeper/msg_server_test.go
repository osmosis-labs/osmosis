package keeper_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v7/x/lockup/types"
)

func (suite *KeeperTestSuite) TestMsgLockTokens() {
	type param struct {
		coinsToLock         sdk.Coins
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "creation of lock via lockTokens",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "locking more coins than are in the address",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 20)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		err := simapp.FundAccount(suite.app.BankKeeper, suite.ctx, test.param.lockOwner, test.param.coinsInOwnerAddress)
		suite.Require().NoError(err)

		_, err = suite.app.LockupKeeper.LockTokens(suite.ctx, test.param.lockOwner, test.param.coinsToLock, test.param.duration)

		if test.expectPass {
			// creation of lock via LockTokens
			msgServer := keeper.NewMsgServerImpl(suite.app.LockupKeeper)
			_, err = msgServer.LockTokens(sdk.WrapSDKContext(suite.ctx), types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

			// Check Locks
			locks, err := suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
			suite.Require().NoError(err)
			suite.Require().Len(locks, 1)
			suite.Require().Equal(locks[0].Coins, test.param.coinsToLock)

			// check accumulation store is correctly updated
			accum := suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      test.param.duration,
			})
			suite.Require().Equal(accum.String(), "10")

			// add more tokens to lock via LockTokens
			err = simapp.FundAccount(suite.app.BankKeeper, suite.ctx, test.param.lockOwner, test.param.coinsInOwnerAddress)
			suite.Require().NoError(err)

			_, err = msgServer.LockTokens(sdk.WrapSDKContext(suite.ctx), types.NewMsgLockTokens(test.param.lockOwner, locks[0].Duration, test.param.coinsToLock))
			suite.Require().NoError(err)

			// check locks after adding tokens to lock
			locks, err = suite.app.LockupKeeper.GetPeriodLocks(suite.ctx)
			suite.Require().NoError(err)
			suite.Require().Len(locks, 1)
			suite.Require().Equal(locks[0].Coins, test.param.coinsToLock.Add(test.param.coinsToLock...))

			// check accumulation store is correctly updated
			accum = suite.app.LockupKeeper.GetPeriodLocksAccumulation(suite.ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      test.param.duration,
			})
			suite.Require().Equal(accum.String(), "20")

		} else {
			// Fail simple lock token
			suite.Require().Error(err)
		}
	}
}
