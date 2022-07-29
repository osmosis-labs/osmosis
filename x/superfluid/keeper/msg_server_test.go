package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	lockupkeeper "github.com/osmosis-labs/osmosis/v10/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/types"
	v8constants "github.com/osmosis-labs/osmosis/v10/app/upgrades/v8/constants"
)

func (suite *KeeperTestSuite) TestMsgSuperfluidDelegate() {
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
			name: "superfluid delegation for not allowed asset",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
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

		lockupMsgServer := lockupkeeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

		valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

		msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)
		_, err = msgServer.SuperfluidDelegate(c, types.NewMsgSuperfluidDelegate(test.param.lockOwner, resp.ID, valAddrs[0]))

		if test.expectPass {
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgSuperfluidUndelegate() {
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
			name: "superfluid undelegation for not superfluid delegated lockup",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
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

		lockupMsgServer := lockupkeeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

		msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)
		_, err = msgServer.SuperfluidUndelegate(c, types.NewMsgSuperfluidUndelegate(test.param.lockOwner, resp.ID))

		if test.expectPass {
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgSuperfluidUnbondLock() {
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
			name: "superfluid unbond lock that is not superfluid lockup",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
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

		lockupMsgServer := lockupkeeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))

		msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)
		_, err = msgServer.SuperfluidUnbondLock(c, types.NewMsgSuperfluidUnbondLock(test.param.lockOwner, resp.ID))

		if test.expectPass {
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

func (suite *KeeperTestSuite) TestMsgLockAndSuperfluidDelegate() {
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
			name: "superfluid lock and superfluid delegate for not allowed asset",
			param: param{
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
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

		c := sdk.WrapSDKContext(suite.Ctx)
		valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

		msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)
		_, err := msgServer.LockAndSuperfluidDelegate(c, types.NewMsgLockAndSuperfluidDelegate(test.param.lockOwner, test.param.coinsToLock, valAddrs[0]))

		if test.expectPass {
			suite.Require().NoError(err)
		} else {
			suite.Require().Error(err)
		}
	}
}

// TestMsgSuperfluidDelegate_Event tests that events are correctly emitted
// when calling SuperfluidDelegate.
func (suite *KeeperTestSuite) TestMsgSuperfluidDelegate_Event() {
	type param struct {
		lockOwner sdk.AccAddress
		duration  time.Duration
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "basic valid",
			param: param{
				lockOwner: sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:  time.Hour * 504,
			},
			expectPass: true,
		},
		{
			name: "invalid",
			param: param{
				lockOwner: sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:  time.Second,
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		suite.SetupTest()

		denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})
		coinsToLock := sdk.NewCoins(sdk.NewCoin(denoms[0], sdk.NewInt(20)))
		suite.FundAcc(test.param.lockOwner, coinsToLock)

		lockupMsgServer := lockupkeeper.NewMsgServerImpl(suite.App.LockupKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)
		resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, coinsToLock))

		valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

		msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)
		_, err = msgServer.SuperfluidDelegate(c, types.NewMsgSuperfluidDelegate(test.param.lockOwner, resp.ID, valAddrs[0]))

		if test.expectPass {
			suite.Require().NoError(err)
			assertEventEmitted(suite, suite.Ctx, types.TypeEvtSuperfluidDelegate, 1)
		} else {
			suite.Require().Error(err)
		}
	}
}

// TestMsgSuperfluidUndelegate_Event tests that events are correctly emitted
// when calling SuperfluidUndelegate.
func (suite *KeeperTestSuite) TestMsgSuperfluidUndelegate_Event() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
		expSuperUnbondingErr  []bool
		// expected amount of delegation to intermediary account
	}{
		{
			"with single validator and single superfluid delegation and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{2},
			[]bool{true},
		},
	}

	for _, test := range testCases {
		suite.SetupTest()
		msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)
		c := sdk.WrapSDKContext(suite.Ctx)

		// setup validators
		valAddrs := suite.SetupValidators(test.validatorStats)

		denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20)})

		// setup superfluid delegations
		suite.setupSuperfluidDelegations(valAddrs, test.superDelegations, denoms)
		for index, lockId := range test.superUnbondingLockIds {
			lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
			if err != nil {
				lock = &lockuptypes.PeriodLock{}
			}

			// superfluid undelegate
			sender, _ := sdk.AccAddressFromBech32(lock.Owner)
			_, err = msgServer.SuperfluidUndelegate(c, types.NewMsgSuperfluidUndelegate(sender, lockId))
			if test.expSuperUnbondingErr[index] {
				suite.Require().Error(err)
				continue
			} else {
				suite.Require().NoError(err)
				assertEventEmitted(suite, suite.Ctx, types.TypeEvtSuperfluidUndelegate, 1)
			}
		}
	}
}

// TestMsgSuperfluidUnbondLock_Event tests that events are correctly emitted
// when calling SuperfluidUnbondLock.
func (suite *KeeperTestSuite) TestMsgSuperfluidUnbondLock_Event() {
	suite.SetupTest()
	msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)

	// setup validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// setup superfluid delegations
	_, _, locks := suite.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)

	for _, lock := range locks {
		startTime := time.Now()
		sender, _ := sdk.AccAddressFromBech32(lock.Owner)

		// first we test that SuperfluidUnbondLock would cause error before undelegating
		_, err := msgServer.SuperfluidUnbondLock(sdk.WrapSDKContext(suite.Ctx), types.NewMsgSuperfluidUnbondLock(sender, lock.ID))
		suite.Require().Error(err)

		// undelegation needs to happen prior to SuperfluidUnbondLock
		err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lock.ID)
		suite.Require().NoError(err)

		// test SuperfluidUnbondLock
		unbondLockStartTime := startTime.Add(time.Hour)
		suite.Ctx = suite.Ctx.WithBlockTime(unbondLockStartTime)
		_, err = msgServer.SuperfluidUnbondLock(sdk.WrapSDKContext(suite.Ctx), types.NewMsgSuperfluidUnbondLock(sender, lock.ID))
		suite.Require().NoError(err)
		assertEventEmitted(suite, suite.Ctx, types.TypeEvtSuperfluidUnbondLock, 1)
	}
}

// TestMsgUnPoolWhitelistedPool_Event tests that events are correctly emitted
// when calling UnPoolWhitelistedPool.
func (suite *KeeperTestSuite) TestMsgUnPoolWhitelistedPool_Event() {
	suite.SetupTest()
	msgServer := keeper.NewMsgServerImpl(suite.App.SuperfluidKeeper)

	// setup validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	denoms, poolIds := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20)})

	// whitelist designated pools
	suite.App.SuperfluidKeeper.SetUnpoolAllowedPools(suite.Ctx, poolIds)

	// setup superfluid delegations
	_, _, locks := suite.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)

	for index, poolId := range poolIds {
		sender, _ := sdk.AccAddressFromBech32(locks[index].Owner)
		suite.Ctx = suite.Ctx.WithBlockHeight(v8constants.UpgradeHeight)
		_, err := msgServer.UnPoolWhitelistedPool(sdk.WrapSDKContext(suite.Ctx), types.NewMsgUnPoolWhitelistedPool(sender, poolId))
		suite.Require().NoError(err)
		assertEventEmitted(suite, suite.Ctx, types.TypeEvtUnpoolId, 1)
	}
}

func assertEventEmitted(suite *KeeperTestSuite, ctx sdk.Context, eventTypeExpected string, numEventsExpected int) {
	allEvents := ctx.EventManager().Events()
	// filter out other events
	actualEvents := make([]sdk.Event, 0, 1)
	for _, event := range allEvents {
		if event.Type == eventTypeExpected {
			actualEvents = append(actualEvents, event)
		}
	}
	suite.Require().Equal(numEventsExpected, len(actualEvents))
}
