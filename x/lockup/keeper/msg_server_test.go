package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v11/x/lockup/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/lockup/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

		suite.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		_, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

		if test.expectPass {
			// creation of lock via LockTokens
			msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
			_, err = msgServer.LockTokens(sdk.WrapSDKContext(suite.Ctx), types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

			// Check Locks
			locks, err := suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
			suite.Require().NoError(err)
			suite.Require().Len(locks, 1)
			suite.Require().Equal(locks[0].Coins, test.param.coinsToLock)

			// check accumulation store is correctly updated
			accum := suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      test.param.duration,
			})
			suite.Require().Equal(accum.String(), "10")

			// add more tokens to lock via LockTokens
			suite.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

			_, err = msgServer.LockTokens(sdk.WrapSDKContext(suite.Ctx), types.NewMsgLockTokens(test.param.lockOwner, locks[0].Duration, test.param.coinsToLock))
			suite.Require().NoError(err)

			// check locks after adding tokens to lock
			locks, err = suite.App.LockupKeeper.GetPeriodLocks(suite.Ctx)
			suite.Require().NoError(err)
			suite.Require().Len(locks, 1)
			suite.Require().Equal(locks[0].Coins, test.param.coinsToLock.Add(test.param.coinsToLock...))

			// check accumulation store is correctly updated
			accum = suite.App.LockupKeeper.GetPeriodLocksAccumulation(suite.Ctx, types.QueryCondition{
				LockQueryType: types.ByDuration,
				Denom:         "stake",
				Duration:      test.param.duration,
			})
			suite.Require().Equal(accum.String(), "20")
			// check to see if events are emitted properly
			suite.AssertEventEmitted(suite.Ctx, types.TypeEvtLockTokens, 1)
		} else {
			// Fail simple lock token
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgBeginUnlocking() {
	type param struct {
		coinsToLock         sdk.Coins
		isSyntheticLockup   bool
		coinsToUnlock       sdk.Coins
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
			name: "unlock full amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   false,
				coinsToUnlock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "unlock partial amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   false,
				coinsToUnlock:       sdk.Coins{sdk.NewInt64Coin("stake", 5)},        // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "unlock zero amount of tokens via begin unlock",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   false,
				coinsToUnlock:       sdk.Coins{},                                    // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "unlock partial amount of tokens via begin unlock for lockup with synthetic versions",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   true,
				coinsToUnlock:       sdk.Coins{sdk.NewInt64Coin("stake", 5)},        // setup wallet
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		suite.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		suite.Require().NoError(err)

		if test.param.isSyntheticLockup {
			err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, resp.ID, "synthetic", time.Second, false)
			suite.Require().NoError(err)
		}

		_, err = msgServer.BeginUnlocking(c, types.NewMsgBeginUnlocking(test.param.lockOwner, resp.ID, test.param.coinsToUnlock))

		if test.expectPass {
			suite.Require().NoError(err)
			suite.AssertEventEmitted(suite.Ctx, types.TypeEvtBeginUnlock, 1)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgBeginUnlockingAll() {
	type param struct {
		coinsToLock         sdk.Coins
		isSyntheticLockup   bool
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
			name: "unlock all lockups",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   false,
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: true,
		},
		{
			name: "unlock all when synthetic versions exists",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup:   true,
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		suite.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		suite.Require().NoError(err)

		if test.param.isSyntheticLockup {
			err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, resp.ID, "synthetic", time.Second, false)
			suite.Require().NoError(err)
		}

		_, err = msgServer.BeginUnlockingAll(c, types.NewMsgBeginUnlockingAll(test.param.lockOwner))

		if test.expectPass {
			suite.Require().NoError(err)
			suite.AssertEventEmitted(suite.Ctx, types.TypeEvtBeginUnlockAll, 1)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgEditLockup() {
	type param struct {
		coinsToLock       sdk.Coins
		isSyntheticLockup bool
		lockOwner         sdk.AccAddress
		duration          time.Duration
		newDuration       time.Duration
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "edit lockups by duration",
			param: param{
				coinsToLock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup: false,
				lockOwner:         sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:          time.Second,
				newDuration:       time.Second * 2,
			},
			expectPass: true,
		},
		{
			name: "edit lockups by lesser duration",
			param: param{
				coinsToLock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup: false,
				lockOwner:         sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:          time.Second,
				newDuration:       time.Second / 2,
			},
			expectPass: false,
		},
		{
			name: "disallow edit when synthetic lockup exists",
			param: param{
				coinsToLock:       sdk.Coins{sdk.NewInt64Coin("stake", 10)}, // setup wallet
				isSyntheticLockup: true,
				lockOwner:         sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:          time.Second,
				newDuration:       time.Second * 2,
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		err := simapp.FundAccount(suite.App.BankKeeper, suite.Ctx, test.param.lockOwner, test.param.coinsToLock)
		suite.Require().NoError(err)

		msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := msgServer.LockTokens(c, types.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		suite.Require().NoError(err)

		if test.param.isSyntheticLockup {
			err = suite.App.LockupKeeper.CreateSyntheticLockup(suite.Ctx, resp.ID, "synthetic", time.Second, false)
			suite.Require().NoError(err)
		}

		_, err = msgServer.ExtendLockup(c, types.NewMsgExtendLockup(test.param.lockOwner, resp.ID, test.param.newDuration))

		if test.expectPass {
			suite.Require().NoError(err, test.name)
			suite.AssertEventEmitted(suite.Ctx, types.TypeEvtLockTokens, 2)
		} else {
			suite.Require().Error(err, test.name)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgLockToken_EmitEvent() {
	type param struct {
		coinsToLock         sdk.Coins
		lockOwner           sdk.AccAddress
		duration            time.Duration
		coinsInOwnerAddress sdk.Coins
	}

	type events struct {
		Id       string
		Owner    string
		Coins    string
		Duration string
		EndTime  string
	}

	defaultLocParams := param{
		coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
		lockOwner:           sdk.AccAddress([]byte("addr1---------------")), // setup wallet
		duration:            time.Second,
		coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
	}

	testCases := map[string]struct {
		events     events
		expectPass bool
	}{
		"events emitted correctly": {
			events: events{
				Id:       "1",
				Owner:    "osmo1v9jxgu3395kj6tfd95kj6tfd95kj6tfd0hptum",
				Coins:    "10stake",
				Duration: "1s",
				EndTime:  "0001-01-01 00:00:00 +0000 UTC",
			},
			expectPass: true,
		},
		"events emitted with incorrect values": {
			events: events{
				Id:       "1",
				Owner:    "osmo1v9jxgu3395kj6tfd95kj6tfd95kj6tfd0hptum",
				Coins:    "11stake",
				Duration: "1s",
				EndTime:  "0001-01-01 00:00:00 +0000 UTC",
			},
			expectPass: false,
		},
		"events emitted with missing values": {
			events: events{
				Id:       "1",
				Coins:    "10stake",
				Duration: "1s",
			},
			expectPass: false,
		},
	}

	for name, tc := range testCases {
		suite.Run(name, func() {
			suite.SetupTest()

			suite.FundAcc(defaultLocParams.lockOwner, defaultLocParams.coinsInOwnerAddress)

			msgServer := keeper.NewMsgServerImpl(suite.App.LockupKeeper)
			c := sdk.WrapSDKContext(suite.Ctx)
			_, err := msgServer.LockTokens(c, types.NewMsgLockTokens(defaultLocParams.lockOwner, defaultLocParams.duration, defaultLocParams.coinsToLock))
			suite.Require().NoError(err)

			expectedEvents := sdk.Events{
				sdk.NewEvent(
					types.TypeEvtLockTokens,
					sdk.NewAttribute(types.AttributePeriodLockID, tc.events.Id),
					sdk.NewAttribute(types.AttributePeriodLockOwner, tc.events.Owner),
					sdk.NewAttribute(types.AttributePeriodLockAmount, tc.events.Coins),
					sdk.NewAttribute(types.AttributePeriodLockDuration, tc.events.Duration),
					sdk.NewAttribute(types.AttributePeriodLockUnlockTime, tc.events.EndTime),
				),
			}

			allEvents := suite.Ctx.EventManager().Events()
			// filter out other events
			actualEvents := make([]sdk.Event, 0)
			for _, event := range allEvents {
				if event.Type == types.TypeEvtLockTokens {
					actualEvents = append(actualEvents, event)
				}
			}

			if tc.expectPass {
				suite.Require().Equal(expectedEvents[0], actualEvents[0])
			} else {
				suite.Require().NotEqual(expectedEvents[0], actualEvents[0])
			}

		})
	}
}
