package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	v8constants "github.com/osmosis-labs/osmosis/v27/app/upgrades/v8/constants"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	gammmigration "github.com/osmosis-labs/osmosis/v27/x/gamm/types/migration"
	lockupkeeper "github.com/osmosis-labs/osmosis/v27/x/lockup/keeper"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

var defaultFunds = sdk.NewCoins(defaultPoolAssets[0].Token, sdk.NewCoin("stake", osmomath.NewInt(5000000000)))

func (s *KeeperTestSuite) TestMsgSuperfluidDelegate() {
	type param struct {
		coinsToLock sdk.Coins
		lockOwner   sdk.AccAddress
		duration    time.Duration
	}

	tests := []struct {
		name       string
		param      param
		expectPass bool
	}{
		{
			name: "superfluid delegation for not allowed asset",
			param: param{
				coinsToLock: sdk.Coins{sdk.NewInt64Coin("stake", 10)},       // setup wallet
				lockOwner:   sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:    time.Hour * 504,
			},
			expectPass: false,
		},
		{
			name: "invalid duration",
			param: param{
				lockOwner: sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:  time.Second,
			},
			expectPass: false,
		},
		{
			name: "happy case",
			param: param{
				lockOwner: sdk.AccAddress([]byte("addr1---------------")), // setup wallet
				duration:  time.Hour * 504,
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			lockupMsgServer := lockupkeeper.NewMsgServerImpl(s.App.LockupKeeper)
			c := s.Ctx

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			// If there is no coinsToLock in the param, use pool denom
			if test.param.coinsToLock.Empty() {
				test.param.coinsToLock = sdk.NewCoins(sdk.NewCoin(denoms[0], osmomath.NewInt(20)))
			}
			s.FundAcc(test.param.lockOwner, test.param.coinsToLock)
			resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
			s.Require().NoError(err)

			valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

			msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
			_, err = msgServer.SuperfluidDelegate(c, types.NewMsgSuperfluidDelegate(test.param.lockOwner, resp.ID, valAddrs[0]))

			if test.expectPass {
				s.Require().NoError(err)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtSuperfluidDelegate, 1)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgSuperfluidUndelegate() {
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
		s.SetupTest()

		s.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		lockupMsgServer := lockupkeeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := s.Ctx
		resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		s.Require().NoError(err)

		msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
		_, err = msgServer.SuperfluidUndelegate(c, types.NewMsgSuperfluidUndelegate(test.param.lockOwner, resp.ID))

		if test.expectPass {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

func (s *KeeperTestSuite) TestMsgCreateFullRangePositionAndSuperfluidDelegate() {
	defaultSender := s.TestAccs[0]
	type param struct {
		coinsToLock sdk.Coins
		poolId      uint64
	}

	tests := []struct {
		name               string
		param              param
		expectPass         bool
		expectedLockId     uint64
		expectedPositionId uint64
	}{
		{
			name:               "happy case",
			param:              param{},
			expectPass:         true,
			expectedLockId:     1,
			expectedPositionId: 2,
		},
		{
			name: "superfluid delegation for not allowed asset",
			param: param{
				coinsToLock: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
		{
			name: "invalid pool id",
			param: param{
				poolId: 3,
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()

			ctx := s.Ctx

			clPool := s.PrepareConcentratedPoolWithCoinsAndFullRangePosition(defaultFunds[0].Denom, defaultFunds[1].Denom)
			clLockupDenom := cltypes.GetConcentratedLockupDenomFromPoolId(clPool.GetId())
			err := s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
				Denom:     clLockupDenom,
				AssetType: types.SuperfluidAssetTypeConcentratedShare,
			})
			s.Require().NoError(err)

			// If there is no coinsToLock in the param, use pool denom
			if test.param.coinsToLock.Empty() {
				test.param.coinsToLock = defaultFunds
			}
			if test.param.poolId == 0 {
				test.param.poolId = clPool.GetId()
			}

			s.FundAcc(defaultSender, test.param.coinsToLock)

			valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

			msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
			resp, err := msgServer.CreateFullRangePositionAndSuperfluidDelegate(ctx, types.NewMsgCreateFullRangePositionAndSuperfluidDelegate(defaultSender, test.param.coinsToLock, valAddrs[0].String(), test.param.poolId))

			if test.expectPass {
				s.Require().NoError(err)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtCreateFullRangePositionAndSFDelegate, 1)
				s.Require().Equal(resp.LockID, test.expectedLockId)
				s.Require().Equal(resp.PositionID, test.expectedPositionId)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestMsgSuperfluidUnbondLock() {
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
		s.SetupTest()

		s.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		lockupMsgServer := lockupkeeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := s.Ctx
		resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		s.Require().NoError(err)

		msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
		_, err = msgServer.SuperfluidUnbondLock(c, types.NewMsgSuperfluidUnbondLock(test.param.lockOwner, resp.ID))

		if test.expectPass {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

func (s *KeeperTestSuite) TestMsgSuperfluidUndelegateAndUnbondLock() {
	type param struct {
		coinsToLock         sdk.Coins
		amountToUnlock      sdk.Coin
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
				coinsToLock:         sdk.Coins{sdk.NewInt64Coin("stake", 10)},
				amountToUnlock:      sdk.NewInt64Coin("stake", 10),
				lockOwner:           sdk.AccAddress([]byte("addr1---------------")),
				duration:            time.Second,
				coinsInOwnerAddress: sdk.Coins{sdk.NewInt64Coin("stake", 10)},
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		s.SetupTest()

		s.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		lockupMsgServer := lockupkeeper.NewMsgServerImpl(s.App.LockupKeeper)
		c := s.Ctx
		resp, err := lockupMsgServer.LockTokens(c, lockuptypes.NewMsgLockTokens(test.param.lockOwner, test.param.duration, test.param.coinsToLock))
		s.Require().NoError(err)

		msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
		_, err = msgServer.SuperfluidUndelegateAndUnbondLock(c, types.NewMsgSuperfluidUndelegateAndUnbondLock(test.param.lockOwner, resp.ID, test.param.amountToUnlock))

		if test.expectPass {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

func (s *KeeperTestSuite) TestMsgLockAndSuperfluidDelegate() {
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
		s.SetupTest()

		s.FundAcc(test.param.lockOwner, test.param.coinsInOwnerAddress)

		c := s.Ctx
		valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

		msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
		_, err := msgServer.LockAndSuperfluidDelegate(c, types.NewMsgLockAndSuperfluidDelegate(test.param.lockOwner, test.param.coinsToLock, valAddrs[0]))

		if test.expectPass {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

// TestMsgSuperfluidUndelegate_Event tests that events are correctly emitted
// when calling SuperfluidUndelegate.
func (s *KeeperTestSuite) TestMsgSuperfluidUndelegate_Event() {
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
		s.SetupTest()
		msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
		c := s.Ctx

		// setup validators
		valAddrs := s.SetupValidators(test.validatorStats)

		denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20)})

		// setup superfluid delegations
		s.setupSuperfluidDelegations(valAddrs, test.superDelegations, denoms)
		for index, lockId := range test.superUnbondingLockIds {
			lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
			if err != nil {
				lock = &lockuptypes.PeriodLock{}
			}

			// superfluid undelegate
			sender, _ := sdk.AccAddressFromBech32(lock.Owner)
			_, err = msgServer.SuperfluidUndelegate(c, types.NewMsgSuperfluidUndelegate(sender, lockId))
			if test.expSuperUnbondingErr[index] {
				s.Require().Error(err)
				continue
			} else {
				s.Require().NoError(err)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtSuperfluidUndelegate, 1)
			}
		}
	}
}

// TestMsgSuperfluidUnbondLock_Event tests that events are correctly emitted
// when calling SuperfluidUnbondLock.
func (s *KeeperTestSuite) TestMsgSuperfluidUnbondLock_Event() {
	s.SetupTest()
	msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)

	// setup validators
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

	// setup superfluid delegations
	_, _, locks := s.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)

	for _, lock := range locks {
		startTime := time.Now()
		sender, _ := sdk.AccAddressFromBech32(lock.Owner)

		// first we test that SuperfluidUnbondLock would cause error before undelegating
		_, err := msgServer.SuperfluidUnbondLock(s.Ctx, types.NewMsgSuperfluidUnbondLock(sender, lock.ID))
		s.Require().Error(err)

		// undelegation needs to happen prior to SuperfluidUnbondLock
		err = s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, lock.Owner, lock.ID)
		s.Require().NoError(err)

		// test SuperfluidUnbondLock
		unbondLockStartTime := startTime.Add(time.Hour)
		s.Ctx = s.Ctx.WithBlockTime(unbondLockStartTime)
		_, err = msgServer.SuperfluidUnbondLock(s.Ctx, types.NewMsgSuperfluidUnbondLock(sender, lock.ID))
		s.Require().NoError(err)
		s.AssertEventEmitted(s.Ctx, types.TypeEvtSuperfluidUnbondLock, 1)
	}
}

// TestMsgUnPoolWhitelistedPool_Event tests that events are correctly emitted
// when calling UnPoolWhitelistedPool.
func (s *KeeperTestSuite) TestMsgUnPoolWhitelistedPool_Event() {
	s.SetupTest()
	msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)

	// setup validators
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	denoms, poolIds := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20)})

	// whitelist designated pools
	s.App.SuperfluidKeeper.SetUnpoolAllowedPools(s.Ctx, poolIds)

	// setup superfluid delegations
	_, _, locks := s.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)

	for index, poolId := range poolIds {
		sender, _ := sdk.AccAddressFromBech32(locks[index].Owner)
		s.Ctx = s.Ctx.WithBlockHeight(v8constants.UpgradeHeight)
		_, err := msgServer.UnPoolWhitelistedPool(s.Ctx, types.NewMsgUnPoolWhitelistedPool(sender, poolId))
		s.Require().NoError(err)
		s.AssertEventEmitted(s.Ctx, types.TypeEvtUnpoolId, 1)
	}
}

func (s *KeeperTestSuite) TestUnlockAndMigrateSharesToFullRangeConcentratedPosition_Event() {
	s.SetupTest()

	msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
	s.FundAcc(s.TestAccs[0], defaultAcctFunds)
	fullRangeCoins := sdk.NewCoins(defaultPoolAssets[0].Token, defaultPoolAssets[1].Token)

	// Set validators
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	// Set balancer pool (foo and stake) and make its respective gamm share an authorized superfluid asset
	msg := balancer.NewMsgCreateBalancerPool(s.TestAccs[0], balancer.PoolParams{
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.NewDec(0),
	}, defaultPoolAssets, defaultFutureGovernor)
	balancerPooId, err := s.App.PoolManagerKeeper.CreatePool(s.Ctx, msg)
	s.Require().NoError(err)
	balancerPool, err := s.App.GAMMKeeper.GetPool(s.Ctx, balancerPooId)
	s.Require().NoError(err)
	poolDenom := gammtypes.GetPoolShareDenom(balancerPool.GetId())
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     poolDenom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})
	s.Require().NoError(err)

	// Set concentrated pool with the same denoms as the balancer pool (foo and stake)
	clPool := s.PrepareCustomConcentratedPool(s.TestAccs[0], defaultPoolAssets[0].Token.Denom, defaultPoolAssets[1].Token.Denom, 1, osmomath.ZeroDec())

	// Set migration link between the balancer and concentrated pool
	migrationRecord := gammmigration.MigrationRecords{BalancerToConcentratedPoolLinks: []gammmigration.BalancerToConcentratedPoolLink{
		{BalancerPoolId: balancerPool.GetId(), ClPoolId: clPool.GetId()},
	}}
	err = s.App.GAMMKeeper.OverwriteMigrationRecords(s.Ctx, migrationRecord)
	s.Require().NoError(err)

	// Superfluid delegate the balancer pool shares
	_, _, locks := s.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 9000000000000000000}}, []string{poolDenom})

	// Create full range concentrated position (needed to give the pool an initial spot price and liquidity value)
	s.CreateFullRangePosition(clPool, fullRangeCoins)

	// Add new superfluid asset
	denom := cltypes.GetConcentratedLockupDenomFromPoolId(clPool.GetId())
	err = s.App.SuperfluidKeeper.AddNewSuperfluidAsset(s.Ctx, types.SuperfluidAsset{
		Denom:     denom,
		AssetType: types.SuperfluidAssetTypeConcentratedShare,
	})
	s.Require().NoError(err)

	// Execute UnlockAndMigrateSharesToFullRangeConcentratedPosition message
	sender, err := sdk.AccAddressFromBech32(locks[0].Owner)
	s.Require().NoError(err)
	_, err = msgServer.UnlockAndMigrateSharesToFullRangeConcentratedPosition(s.Ctx,
		types.NewMsgUnlockAndMigrateSharesToFullRangeConcentratedPosition(sender, int64(locks[0].ID), locks[0].Coins[0]))
	s.Require().NoError(err)

	// Asset event emitted
	s.AssertEventEmitted(s.Ctx, types.TypeEvtUnlockAndMigrateShares, 1)
}

// TestAddToConcentratedLiquiditySuperfluidPosition_Events tests that events are correctly emitted
// when calling addToConcentratedLiquiditySuperfluidPosition.
func (s *KeeperTestSuite) TestAddToConcentratedLiquiditySuperfluidPosition_Events() {
	testcases := map[string]struct {
		isLastPositionInPool         bool
		expectedAddedToPositionEvent int
		expectedMessageEvents        int
		expectedError                error
	}{
		"happy path": {
			isLastPositionInPool:         false,
			expectedAddedToPositionEvent: 1,
		},
		"error: last position in pool": {
			isLastPositionInPool:         true,
			expectedAddedToPositionEvent: 0,
			expectedError:                cltypes.AddToLastPositionInPoolError{PoolId: 1, PositionId: 1},
		},
	}

	for name, tc := range testcases {
		s.Run(name, func() {
			s.SetupTest()

			msgServer := keeper.NewMsgServerImpl(s.App.SuperfluidKeeper)
			concentratedLiquidityKeeper := s.App.ConcentratedLiquidityKeeper
			owner := s.TestAccs[0]

			// Position from current account.
			posId, _, _, _, _, poolJoinAcc := s.SetupSuperfluidConcentratedPosition(s.Ctx, true, false, false, owner)

			if !tc.isLastPositionInPool {
				s.FundAcc(s.TestAccs[1], defaultFunds)
				_, err := concentratedLiquidityKeeper.CreateFullRangePosition(s.Ctx, 1, s.TestAccs[1], defaultFunds)
				s.Require().NoError(err)
			}

			// Reset event counts to 0 by creating a new manager.
			s.Ctx = s.Ctx.WithEventManager(sdk.NewEventManager())
			s.Equal(0, len(s.Ctx.EventManager().Events()))

			s.FundAcc(poolJoinAcc, defaultFunds)
			msg := &types.MsgAddToConcentratedLiquiditySuperfluidPosition{
				PositionId:    posId,
				Sender:        poolJoinAcc.String(),
				TokenDesired0: defaultFunds[0],
				TokenDesired1: defaultFunds[1],
			}

			response, err := msgServer.AddToConcentratedLiquiditySuperfluidPosition(s.Ctx, msg)

			if tc.expectedError == nil {
				s.NoError(err)
				s.NotNil(response)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtAddToConcentratedLiquiditySuperfluidPosition, tc.expectedAddedToPositionEvent)
			} else {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError.Error())
				s.Require().Nil(response)
				s.AssertEventEmitted(s.Ctx, types.TypeEvtAddToConcentratedLiquiditySuperfluidPosition, tc.expectedAddedToPositionEvent)
			}
		})
	}
}
