package keeper_test

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type superfluidDelegation struct {
	delIndex int64
	valIndex int64
	lpIndex  int64
	lpAmount int64
}

func (suite *KeeperTestSuite) TestSuperfluidDelegate() {
	testCases := []struct {
		name               string
		validatorStats     []stakingtypes.BondStatus
		superDelegations   []superfluidDelegation
		expInterDelegation []sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"with single validator and additional superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(20000000)}, // 50% x 20 x 1000000 x 2
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000), sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"add unbonding validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonding},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000), sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"add unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000), sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()
			bondDenom := suite.App.StakingKeeper.BondDenom(suite.Ctx)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// get pre-superfluid delgations osmo supply and supplyWithOffset
			presupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
			presupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)

			// setup superfluid delegations
			_, intermediaryAccs, locks := suite.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)

			// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
			postsupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
			postsupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)
			suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
			suite.Require().Equal(postsupplyWithOffset.String(), presupplyWithOffset.String())

			unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

			for index, del := range tc.superDelegations {
				lock := locks[index]
				valAddr := valAddrs[del.valIndex]

				// check synthetic lockup creation
				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
				suite.Require().Equal(synthLock.SynthDenom, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				suite.Require().Equal(synthLock.EndTime, time.Time{})

				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddr.String(), 0)

				// Check lockID connection with intermediary account
				intAcc := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lock.ID)
				suite.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())
			}

			for index, expAcc := range intermediaryAccs {
				// check intermediary account creation
				gotAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, expAcc.GetAccAddress())
				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)
				suite.Require().GreaterOrEqual(gotAcc.GaugeId, uint64(1))

				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				suite.Require().NoError(err)

				// check gauge creation
				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gotAcc.GaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         keeper.StakingSyntheticDenom(expAcc.Denom, valAddr.String()),
					Duration:      unbondingDuration,
				})
				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
				suite.Require().Equal(gauge.StartTime, suite.Ctx.BlockTime())
				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

				// check delegation from intermediary account to validator
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, expAcc.GetAccAddress(), valAddr)
				suite.Require().True(found)
				suite.Require().Equal(tc.expInterDelegation[index], delegation.Shares)
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// try delegating twice with same lockup
			for _, lock := range locks {
				err := suite.App.SuperfluidKeeper.SuperfluidDelegate(suite.Ctx, lock.Owner, lock.ID, valAddrs[0].String())
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestValidateLockForSFDelegate() {
	lockOwner := suite.TestAccs[0]

	tests := []struct {
		name                             string
		lock                             *lockuptypes.PeriodLock
		sender                           string
		skParams                         types.Params
		superfluidAssetToSet             types.SuperfluidAsset
		lockIdAlreadySuperfluidDelegated bool
		expectedErr                      error
	}{
		{
			name: "valid gamm lock",
			lock: &lockuptypes.PeriodLock{
				Owner:    lockOwner.String(),
				Coins:    sdk.NewCoins(sdk.NewCoin(DefaultGammAsset, sdk.NewInt(100))),
				Duration: time.Hour * 24 * 21,
				ID:       1,
			},
			superfluidAssetToSet: types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			expectedErr:          nil,
		},
		{
			name: "valid cl lock",
			lock: &lockuptypes.PeriodLock{
				Owner:    lockOwner.String(),
				Coins:    sdk.NewCoins(sdk.NewCoin(cltypes.GetConcentratedLockupDenomFromPoolId(1), sdk.NewInt(100))),
				Duration: time.Hour * 24 * 21,
				ID:       1,
			},
			superfluidAssetToSet: types.SuperfluidAsset{Denom: cltypes.GetConcentratedLockupDenomFromPoolId(1), AssetType: types.SuperfluidAssetTypeConcentratedShare},
			expectedErr:          nil,
		},
		{
			name: "invalid lock - not superfluid asset",
			lock: &lockuptypes.PeriodLock{
				Owner:    lockOwner.String(),
				Coins:    sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(100))),
				Duration: time.Hour * 24 * 21,
				ID:       1,
			},
			superfluidAssetToSet: types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			expectedErr:          errorsmod.Wrapf(types.ErrNonSuperfluidAsset, "denom: %s", "uosmo"),
		},
		{
			name: "invalid lock - unbonding lockup not supported",
			lock: &lockuptypes.PeriodLock{
				Owner:    lockOwner.String(),
				Coins:    sdk.NewCoins(sdk.NewCoin(DefaultGammAsset, sdk.NewInt(100))),
				Duration: time.Hour * 24 * 21,
				ID:       1,
				EndTime:  time.Now().Add(time.Hour * 24),
			},
			superfluidAssetToSet: types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			expectedErr:          errorsmod.Wrapf(types.ErrUnbondingLockupNotSupported, "lock id : %d", uint64(1)),
		},
		{
			name: "invalid lock - not enough lockup duration",
			lock: &lockuptypes.PeriodLock{
				Owner:    lockOwner.String(),
				Coins:    sdk.NewCoins(sdk.NewCoin(DefaultGammAsset, sdk.NewInt(100))),
				Duration: time.Hour * 24,
				ID:       1,
			},
			superfluidAssetToSet: types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			expectedErr: errorsmod.Wrapf(types.ErrNotEnoughLockupDuration,
				"lock duration (%d) must be greater than unbonding time (%d)",
				time.Hour*24, time.Hour*24*21),
		},
		{
			name: "invalid lock - already used superfluid lockup",
			lock: &lockuptypes.PeriodLock{
				Owner:    lockOwner.String(),
				Coins:    sdk.NewCoins(sdk.NewCoin(DefaultGammAsset, sdk.NewInt(100))),
				Duration: time.Hour * 24 * 21,
				ID:       1,
			},
			superfluidAssetToSet:             types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			lockIdAlreadySuperfluidDelegated: true,
			expectedErr:                      errorsmod.Wrapf(types.ErrAlreadyUsedSuperfluidLockup, "lock id : %d", uint64(1)),
		},
	}

	for _, test := range tests {
		suite.Run(test.name, func() {
			suite.SetupTest()

			suite.App.SuperfluidKeeper.SetSuperfluidAsset(suite.Ctx, test.superfluidAssetToSet)

			if test.lockIdAlreadySuperfluidDelegated {
				intermediateAccount := types.NewSuperfluidIntermediaryAccount(test.lock.Coins[0].Denom, lockOwner.String(), 1)
				suite.App.SuperfluidKeeper.SetLockIdIntermediaryAccountConnection(suite.Ctx, test.lock.ID, intermediateAccount)
			}

			err := suite.App.SuperfluidKeeper.ValidateLockForSFDelegate(suite.Ctx, test.lock, lockOwner.String())
			if test.expectedErr != nil {
				suite.Require().Error(err)
				suite.Require().Equal(test.expectedErr.Error(), err.Error())
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSuperfluidUndelegate() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
		expSuperUnbondingErr  []bool
		// expected amount of delegation to intermediary account
		expInterDelegation []sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		// {
		// 	"with single validator, single superfluid delegation, add more tokens to the lock, and single undelegation",
		// 	[]stakingtypes.BondStatus{stakingtypes.Bonded},
		// 	[]superfluidDelegation{{0, 0, 1000000}},
		// 	[]uint64{1},
		// 	[]uint64{1},
		// 	[]bool{false},
		// 	[]sdk.Dec{sdk.ZeroDec()},
		// },
		{
			"with single validator and additional superfluid delegations and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"with multiple validators and multiple superfluid delegations and multiple undelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"add unbonding validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonding},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"add unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{2},
			[]bool{true},
			[]sdk.Dec{},
		},
		{
			"try undelegating twice for same lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1, 1},
			[]bool{false, true},
			[]sdk.Dec{sdk.ZeroDec()},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			bondDenom := suite.App.StakingKeeper.GetParams(suite.Ctx).BondDenom

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, _ := suite.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for index, lockId := range tc.superUnbondingLockIds {
				// get intermediary account
				accAddr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockId)
				intermediaryAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				if err != nil {
					lock = &lockuptypes.PeriodLock{}
				}

				// get pre-superfluid delgations osmo supply and supplyWithOffset
				presupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
				presupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)

				// superfluid undelegate
				err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lockId)
				if tc.expSuperUnbondingErr[index] {
					suite.Require().Error(err)
					continue
				}
				suite.Require().NoError(err)

				// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
				postsupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
				postsupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)
				suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
				suite.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

				// check lockId and intermediary account connection deletion
				addr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockId)
				suite.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().NoError(err)
				suite.Require().Equal(synthLock.UnderlyingLockId, lockId)
				suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(unbondingDuration))
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// check remaining intermediary account delegation
			for index, expDelegation := range tc.expInterDelegation {
				acc := intermediaryAccs[index]
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, acc.GetAccAddress(), valAddr)
				if expDelegation.IsZero() {
					suite.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
				} else {
					suite.Require().True(found)
					suite.Require().Equal(expDelegation, delegation.Shares)
				}
			}

			// try undelegating twice
			for index, lockId := range tc.superUnbondingLockIds {
				if tc.expSuperUnbondingErr[index] {
					continue
				}

				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)

				err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lockId)
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSuperfluidUndelegateToConcentratedPosition() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
		expSuperUnbondingErr  []bool
		// expected amount of delegation to intermediary account
		expInterDelegation []sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"with single validator and additional superfluid delegations and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"with multiple validators and multiple superfluid delegations and multiple undelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"add unbonding validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonding},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"add unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]sdk.Dec{sdk.ZeroDec()},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{2},
			[]bool{true},
			[]sdk.Dec{},
		},
		{
			"try undelegating twice for same lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1, 1},
			[]bool{false, true},
			[]sdk.Dec{sdk.ZeroDec()},
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			bondDenom := suite.App.StakingKeeper.GetParams(suite.Ctx).BondDenom

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, _ := suite.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for index, lockId := range tc.superUnbondingLockIds {
				// get intermediary account
				accAddr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockId)
				intermediaryAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				if err != nil {
					lock = &lockuptypes.PeriodLock{}
				}

				// get pre-superfluid delgations osmo supply and supplyWithOffset
				presupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
				presupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)

				// superfluid undelegate
				_, err = suite.App.SuperfluidKeeper.SuperfluidUndelegateToConcentratedPosition(suite.Ctx, lock.Owner, lockId)
				if tc.expSuperUnbondingErr[index] {
					suite.Require().Error(err)
					continue
				}
				suite.Require().NoError(err)

				// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
				postsupply := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
				postsupplyWithOffset := suite.App.BankKeeper.GetSupplyWithOffset(suite.Ctx, bondDenom)
				suite.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
				suite.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

				// check lockId and intermediary account connection deletion
				addr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lockId)
				suite.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().Error(err)

				// check unbonding synthetic lockup creation
				// since this is the concentrated liquidity path, no new synthetic lockup should be created
				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				suite.Require().Error(err)
				suite.Require().Nil(synthLock)
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// check remaining intermediary account delegation
			for index, expDelegation := range tc.expInterDelegation {
				acc := intermediaryAccs[index]
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, acc.GetAccAddress(), valAddr)
				if expDelegation.IsZero() {
					suite.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
				} else {
					suite.Require().True(found)
					suite.Require().Equal(expDelegation, delegation.Shares)
				}
			}

			// try undelegating twice
			for index, lockId := range tc.superUnbondingLockIds {
				if tc.expSuperUnbondingErr[index] {
					continue
				}

				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)

				_, err = suite.App.SuperfluidKeeper.SuperfluidUndelegateToConcentratedPosition(suite.Ctx, lock.Owner, lockId)
				suite.Require().Error(err)
			}
		})
	}
}

// TestSuperfluidUnbondLock tests the following.
//  1. test SuperfluidUnbondLock does not work before undelegation
//  2. test SuperfluidUnbondLock makes underlying lock start unlocking
//  3. test that synthetic lockup being finished does not mean underlying lock is finished
//  4. test after SuperfluidUnbondLock + lockup time, the underlying lock is finished
func (suite *KeeperTestSuite) TestSuperfluidUnbondLock() {
	suite.SetupTest()

	// setup validators
	valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// setup superfluid delegations
	_, intermediaryAccs, locks := suite.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)
	suite.checkIntermediaryAccountDelegations(intermediaryAccs)

	for _, lock := range locks {
		startTime := time.Now()
		suite.Ctx = suite.Ctx.WithBlockTime(startTime)
		accAddr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lock.ID)
		intermediaryAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, accAddr)
		valAddr := intermediaryAcc.ValAddr

		// first we test that SuperfluidUnbondLock would cause error before undelegating
		err := suite.App.SuperfluidKeeper.SuperfluidUnbondLock(suite.Ctx, lock.ID, lock.GetOwner())
		suite.Require().Error(err)

		// undelegation needs to happen prior to SuperfluidUnbondLock
		err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lock.ID)
		suite.Require().NoError(err)
		balances := suite.App.BankKeeper.GetAllBalances(suite.Ctx, lock.OwnerAddress())
		suite.Require().Equal(0, balances.Len())

		// check that unbonding synth has been created correctly after undelegation
		unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
		synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().NoError(err)
		suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
		suite.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(unbondingDuration))

		// test SuperfluidUnbondLock
		unbondLockStartTime := startTime.Add(time.Hour)
		suite.Ctx = suite.Ctx.WithBlockTime(unbondLockStartTime)
		err = suite.App.SuperfluidKeeper.SuperfluidUnbondLock(suite.Ctx, lock.ID, lock.GetOwner())
		suite.Require().NoError(err)

		// check that SuperfluidUnbondLock makes underlying lock start unlocking
		// we run WithdrawAllMaturedLocks to ensure that lock isn't getting finished immediately
		suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx)
		updatedLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
		suite.Require().NoError(err)
		suite.Require().True(updatedLock.IsUnlocking())

		// check if finsihed unlocking synth lock did not increase balance
		balances = suite.App.BankKeeper.GetAllBalances(suite.Ctx, lock.OwnerAddress())
		suite.Require().Equal(0, balances.Len())

		// test that synth lock finish does not mean underlying lock is finished
		suite.Ctx = suite.Ctx.WithBlockTime((startTime.Add(unbondingDuration)))
		suite.App.LockupKeeper.DeleteAllMaturedSyntheticLocks(suite.Ctx)
		suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx)
		_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		suite.Require().Error(err)
		updatedLock, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
		suite.Require().NoError(err)
		suite.Require().True(updatedLock.IsUnlocking())

		// test after SuperfluidUnbondLock + lockup unbonding duration, lock is finished and does not exist
		suite.Ctx = suite.Ctx.WithBlockTime(unbondLockStartTime.Add(unbondingDuration))
		suite.App.LockupKeeper.WithdrawAllMaturedLocks(suite.Ctx)
		_, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
		suite.Require().Error(err)

		// check if finished unlocking successfully increased balance
		balances = suite.App.BankKeeper.GetAllBalances(suite.Ctx, lock.OwnerAddress())
		suite.Require().Equal(1, balances.Len())
		suite.Require().Equal(denoms[0], balances[0].Denom)
		suite.Require().Equal(sdk.NewInt(1000000), balances[0].Amount)

		// check invariant is fine
		reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
		suite.Require().False(broken, reason)
	}
}

func (suite *KeeperTestSuite) TestSuperfluidUndelegateAndUnbondLock() {
	var lockAmount int64 = 1000000
	testCases := []struct {
		name            string
		testInvalidLock bool
		unlockAmount    sdk.Int
		expectErr       bool
		splitLockId     bool
		undelegating    bool
		unbond          bool
	}{
		{
			name:            "lock doesn't exist",
			testInvalidLock: true,
			unlockAmount:    sdk.NewInt(0),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "unlock amount = 0",
			testInvalidLock: false,
			unlockAmount:    sdk.NewInt(0),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "unlock amount > locked amount",
			testInvalidLock: false,
			unlockAmount:    sdk.NewInt(lockAmount + 1),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "lock is not split if unlock amount = locked amount",
			testInvalidLock: false,
			unlockAmount:    sdk.NewInt(lockAmount),
			expectErr:       false,
			splitLockId:     false,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "lock is split if unlock amount < locked amount",
			testInvalidLock: false,
			unlockAmount:    sdk.NewInt(lockAmount / 2),
			expectErr:       false,
			splitLockId:     true,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "undelegate and unbond an undelegating lock",
			testInvalidLock: false,
			unlockAmount:    sdk.NewInt(1),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    true,
			unbond:          false,
		},
		{
			name:            "undelegate and unbond an unlocking lock",
			testInvalidLock: false,
			unlockAmount:    sdk.NewInt(1),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    true,
			unbond:          true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// setup validators
			valAddrs := suite.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, locks := suite.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, lockAmount}}, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)
			suite.Require().True(len(locks) > 0)

			// test invalid lock
			if tc.testInvalidLock {
				lock := lockuptypes.PeriodLock{}
				_, err := suite.App.SuperfluidKeeper.SuperfluidUndelegateAndUnbondLock(suite.Ctx, lock.ID, lock.GetOwner(), sdk.NewInt(1))
				suite.Require().Error(err)
				return
			}

			// test undelegated lock
			if tc.undelegating {
				lock := locks[0]
				err := suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.GetOwner(), lock.ID)
				suite.Require().NoError(err)
			}

			// test unbond lock
			if tc.unbond {
				lock := locks[0]
				err := suite.App.SuperfluidKeeper.SuperfluidUnbondLock(suite.Ctx, lock.ID, lock.GetOwner())
				suite.Require().NoError(err)
			}

			for _, lock := range locks {
				startTime := time.Now()
				suite.Ctx = suite.Ctx.WithBlockTime(startTime)
				accAddr := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, lock.ID)
				intermediaryAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				// get OSMO total supply and amount to be burned
				bondDenom := suite.App.StakingKeeper.BondDenom(suite.Ctx)
				supplyBefore := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
				osmoAmount, err := suite.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(suite.Ctx, intermediaryAcc.Denom, tc.unlockAmount)
				suite.Require().NoError(err)

				unbondLockStartTime := startTime.Add(time.Hour)
				suite.Ctx = suite.Ctx.WithBlockTime(unbondLockStartTime)
				lockId, err := suite.App.SuperfluidKeeper.SuperfluidUndelegateAndUnbondLock(suite.Ctx, lock.ID, lock.GetOwner(), tc.unlockAmount)
				if tc.expectErr {
					suite.Require().Error(err)
					continue
				}

				suite.Require().NoError(err)

				// check OSMO total supply and burnt amount
				suite.Require().True(osmoAmount.IsPositive())
				supplyAfter := suite.App.BankKeeper.GetSupply(suite.Ctx, bondDenom)
				suite.Require().Equal(supplyAfter, supplyBefore.Sub(sdk.NewCoin(bondDenom, osmoAmount)))

				if tc.splitLockId {
					suite.Require().Equal(lockId, lock.ID+1)

					// check original underlying lock
					updatedLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lock.ID)
					suite.Require().NoError(err)
					suite.Require().False(updatedLock.IsUnlocking())
					suite.Require().Equal(updatedLock.Coins[0].Amount, lock.Coins[0].Amount.Sub(tc.unlockAmount))

					// check newly created underlying lock
					newLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
					suite.Require().NoError(err)
					suite.Require().True(newLock.IsUnlocking())
					suite.Require().Equal(newLock.Coins[0].Amount, tc.unlockAmount)

					// check original synthetic lock
					stakingDenom := keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr)
					synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, stakingDenom)

					suite.Require().NoError(err)
					suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
					suite.Require().Equal(synthLock.SynthDenom, stakingDenom)
					suite.Require().Equal(synthLock.EndTime, time.Time{})

					// check unstaking synthetic lock is not created for the original synthetic lock
					unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
					unstakingDenom := keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr)
					_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, unstakingDenom)
					suite.Require().Error(err)

					// check newly created unstaking synthetic lock
					newSynthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lockId, unstakingDenom)

					suite.Require().NoError(err)
					suite.Require().Equal(newSynthLock.UnderlyingLockId, lockId)
					suite.Require().Equal(newSynthLock.SynthDenom, unstakingDenom)
					suite.Require().Equal(newSynthLock.EndTime, suite.Ctx.BlockTime().Add(unbondingDuration))
				} else {
					suite.Require().Equal(lockId, lock.ID)

					// check underlying lock
					updatedLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
					suite.Require().NoError(err)
					suite.Require().True(updatedLock.IsUnlocking())
					suite.Require().Equal(updatedLock.Coins[0].Amount, tc.unlockAmount)

					// check synthetic lock
					unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime
					unstakingDenom := keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr)

					synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, unstakingDenom)
					suite.Require().NoError(err)
					suite.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
					suite.Require().Equal(synthLock.SynthDenom, unstakingDenom)
					suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(unbondingDuration))

					// check staking synthetic lock is deleted
					stakingDenom := keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr)
					_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, lock.ID, stakingDenom)
					suite.Require().Error(err)
				}

				// check invariant is fine
				reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
				suite.Require().False(broken, reason)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts() {
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		superDelegations []superfluidDelegation
		// denom of the superfluid asset is the key, multiplier is the value
		multipliersByDenom map[string]sdk.Dec
	}{
		{
			"with single validator and single delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			map[string]sdk.Dec{DefaultGammAsset: sdk.NewDec(10)},
		},
		{
			"with single validator and additional delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			map[string]sdk.Dec{DefaultGammAsset: sdk.NewDec(10)},
		},
		{
			"with multiple validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			map[string]sdk.Dec{DefaultGammAsset: sdk.NewDec(10)},
		},
		{
			"with single validator and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 1, 1000000}},
			map[string]sdk.Dec{DefaultGammAsset: sdk.NewDec(10), "gamm/pool/2": sdk.NewDec(10)},
		},
		{
			"with multiple validators and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 1, 1, 1000000}},
			map[string]sdk.Dec{DefaultGammAsset: sdk.NewDec(10), "gamm/pool/2": sdk.NewDec(10)},
		},
		{
			"zero price multiplier check",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			map[string]sdk.Dec{DefaultGammAsset: sdk.NewDec(0)},
		},
		{
			"dust price multiplier check",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			map[string]sdk.Dec{DefaultGammAsset: sdk.NewDecWithPrec(1, 10)}, // 10^-10
		},
	}

	for _, tc := range testCases {
		tc := tc
		suite.Run(tc.name, func() {
			suite.SetupTest()

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, locks := suite.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			// we make a map of intermediary account to delegation shares to store delegation share
			// before refreshing intermediary account delegations on epoch
			interAccIndexToDenomShare := make(map[int]sdk.Dec)
			for accIndex, intermediaryAcc := range intermediaryAccs {
				valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, intermediaryAcc.GetAccAddress(), valAddr)
				suite.Require().True(found)

				interAccIndexToDenomShare[accIndex] = delegation.Shares
			}

			for denom, multiplier := range tc.multipliersByDenom {
				suite.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(suite.Ctx, 2, denom, multiplier)
			}

			suite.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.Ctx)

			originalMultiplier := sdk.NewDec(20)
			for interAccIndex, intermediaryAcc := range intermediaryAccs {
				multiplier := tc.multipliersByDenom[intermediaryAcc.Denom]
				oldDelegation := interAccIndexToDenomShare[interAccIndex]
				expDelegation := oldDelegation.Mul(multiplier).Quo(originalMultiplier)
				lpTokenAmount := sdk.NewInt(1000000)
				decAmt := multiplier.Mul(lpTokenAmount.ToDec())
				denom := intermediaryAcc.Denom
				_, err := suite.App.SuperfluidKeeper.GetSuperfluidAsset(suite.Ctx, denom)
				suite.Require().NoError(err)
				expAmount := suite.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(suite.Ctx, decAmt.RoundInt())

				// check delegation changes
				valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
				suite.Require().NoError(err)
				delegation, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, intermediaryAcc.GetAccAddress(), valAddr)
				if expAmount.IsPositive() {
					suite.Require().True(found)
					suite.Require().Equal(delegation.Shares, expDelegation)
				} else {
					suite.Require().False(found)
				}
			}

			// unbond all lockups
			for _, lock := range locks {
				// superfluid undelegate
				// handling the case same lockup is used for further delegation
				cacheCtx, write := suite.Ctx.CacheContext()
				err := suite.App.SuperfluidKeeper.SuperfluidUndelegate(cacheCtx, lock.Owner, lock.ID)
				if err == nil {
					write()
				}
			}
			unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

			for _, intermediaryAcc := range intermediaryAccs {
				suite.Ctx = suite.Ctx.WithBlockTime(suite.Ctx.BlockTime().Add(unbondingDuration + time.Second))
				suite.App.EndBlocker(suite.Ctx, abci.RequestEndBlock{Height: suite.Ctx.BlockHeight()})

				unbonded := suite.App.BankKeeper.GetBalance(suite.Ctx, intermediaryAcc.GetAccAddress(), sdk.DefaultBondDenom)
				suite.Require().True(unbonded.IsZero())
			}

			// refresh intermediary account delegations
			suite.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(suite.Ctx)

			for _, intermediaryAcc := range intermediaryAccs {
				// check unbonded amount is removed after refresh operation
				refreshed := suite.App.BankKeeper.GetBalance(suite.Ctx, intermediaryAcc.GetAccAddress(), sdk.DefaultBondDenom)
				suite.Require().True(refreshed.IsZero())
			}
		})
	}
}

// type superfluidRedelegation struct {
// 	lockId      uint64
// 	oldValIndex int64
// 	newValIndex int64
// }

// func (suite *KeeperTestSuite) TestSuperfluidRedelegate() {
// 	testCases := []struct {
// 		name                    string
// 		validatorStats          []stakingtypes.BondStatus
// 		superDelegations        []superfluidDelegation
// 		superRedelegations      []superfluidRedelegation
// 		expSuperRedelegationErr []bool
// 	}{
// 		{
// 			"with single validator and single superfluid delegation with single redelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, DefaultGammAsset, 1000000}},
// 			[]superfluidRedelegation{{1, 0, 1}}, // lock1 => val0 -> val1
// 			[]bool{false},
// 		},
// 		{
// 			"with multiple superfluid delegations with single redelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, DefaultGammAsset, 1000000}, {0, DefaultGammAsset, 1000000}},
// 			[]superfluidRedelegation{{1, 0, 1}}, // lock1 => val0 -> val1
// 			[]bool{false},
// 		},
// 		{
// 			"with multiple superfluid delegations with multiple redelegations",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, DefaultGammAsset, 1000000}, {0, DefaultGammAsset, 1000000}},
// 			[]superfluidRedelegation{{1, 0, 1}, {2, 0, 1}}, // lock1 => val0 -> val1, lock2 => val0 -> val1
// 			[]bool{false, false},
// 		},
// 		{
// 			"try redelegating back from new validator to original validator",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, DefaultGammAsset, 1000000}, {0, DefaultGammAsset, 1000000}},
// 			[]superfluidRedelegation{{1, 0, 1}, {1, 1, 0}}, // lock1 => val0 -> val1, lock1 => val1 -> val0
// 			[]bool{false, true},
// 		},
// 		{
// 			"not available lock id redelegation",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, DefaultGammAsset, 1000000}},
// 			[]superfluidRedelegation{{2, 0, 1}}, // lock1 => val0 -> val1
// 			[]bool{true},
// 		},
// 		{
// 			"redelegation for same validator",
// 			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
// 			[]superfluidDelegation{{0, DefaultGammAsset, 1000000}},
// 			[]superfluidRedelegation{{1, 0, 0}}, // lock1 => val0 -> val0
// 			[]bool{true},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		suite.Run(tc.name, func() {
// 			suite.SetupTest()

// 			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
// 			suite.Require().Equal(poolId, uint64(1))

// 			// setup validators
// 			valAddrs := suite.SetupValidators(tc.validatorStats)

// 			// setup superfluid delegations
// 			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
// 			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

// 			// execute redelegation and check changes on store
// 			for index, srd := range tc.superRedelegations {
// 				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, srd.lockId)
// 				if err != nil {
// 					lock = &lockuptypes.PeriodLock{}
// 				}

// 				// superfluid redelegate
// 				err = suite.App.SuperfluidKeeper.SuperfluidRedelegate(suite.Ctx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
// 				if tc.expSuperRedelegationErr[index] {
// 					suite.Require().Error(err)
// 					continue
// 				}
// 				suite.Require().NoError(err)

// 				// check previous validator bonding synthetic lockup deletion
// 				_, err = suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().Error(err)

// 				// check unbonding synthetic lockup creation
// 				params := suite.App.SuperfluidKeeper.GetParams(suite.Ctx)
// 				synthLock, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, srd.lockId, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(synthLock.UnderlyingLockId, srd.lockId)
// 				suite.Require().Equal(synthLock.Suffix, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				suite.Require().Equal(synthLock.EndTime, suite.Ctx.BlockTime().Add(params.UnbondingDuration))

// 				// check synthetic lockup creation
// 				synthLock2, err := suite.App.LockupKeeper.GetSyntheticLockup(suite.Ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(synthLock2.UnderlyingLockId, srd.lockId)
// 				suite.Require().Equal(synthLock2.Suffix, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
// 				suite.Require().Equal(synthLock2.EndTime, time.Time{})

// 				// check intermediary account creation
// 				lock, err = suite.App.LockupKeeper.GetLockByID(suite.Ctx, srd.lockId)
// 				suite.Require().NoError(err)

// 				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddrs[srd.newValIndex].String(), 1)
// 				gotAcc := suite.App.SuperfluidKeeper.GetIntermediaryAccount(suite.Ctx, expAcc.GetAccAddress())
// 				suite.Require().Equal(gotAcc.Denom, expAcc.Denom)
// 				suite.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)

// 				// check gauge creation
// 				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gotAcc.GaugeId)
// 				suite.Require().NoError(err)
// 				suite.Require().Equal(gauge.Id, gotAcc.GaugeId)
// 				suite.Require().Equal(gauge.IsPerpetual, true)
// 				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
// 					LockQueryType: lockuptypes.ByDuration,
// 					Denom:         expAcc.Denom + keeper.StakingSuffix(valAddrs[srd.newValIndex].String()),
// 					Duration:      params.UnbondingDuration,
// 				})
// 				suite.Require().Equal(gauge.Coins, sdk.Coins(nil))
// 				suite.Require().Equal(gauge.StartTime, suite.Ctx.BlockTime())
// 				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
// 				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
// 				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

// 				// Check lockID connection with intermediary account
// 				intAcc := suite.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(suite.Ctx, srd.lockId)
// 				suite.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())

// 				// check delegation from intermediary account to validator
// 				_, found := suite.App.StakingKeeper.GetDelegation(suite.Ctx, expAcc.GetAccAddress(), valAddrs[srd.newValIndex])
// 				suite.Require().True(found)
// 			}

// 			// try redelegating twice
// 			for index, srd := range tc.superRedelegations {
// 				if tc.expSuperRedelegationErr[index] {
// 					continue
// 				}
// 				cacheCtx, _ := suite.Ctx.CacheContext()
// 				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, srd.lockId)
// 				suite.Require().NoError(err)
// 				err = suite.App.SuperfluidKeeper.SuperfluidRedelegate(cacheCtx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }
