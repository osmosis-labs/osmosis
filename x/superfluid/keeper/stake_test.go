package keeper_test

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	cltypes "github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v16/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v16/x/superfluid/types"

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

func (s *KeeperTestSuite) TestSuperfluidDelegate() {
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
		s.Run(tc.name, func() {
			s.SetupTest()
			bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// get pre-superfluid delgations osmo supply and supplyWithOffset
			presupply := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
			presupplyWithOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, bondDenom)

			// setup superfluid delegations
			_, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)

			// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
			postsupply := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
			postsupplyWithOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, bondDenom)
			s.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
			s.Require().Equal(postsupplyWithOffset.String(), presupplyWithOffset.String())

			unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime

			for index, del := range tc.superDelegations {
				lock := locks[index]
				valAddr := valAddrs[del.valIndex]

				// check synthetic lockup creation
				synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				s.Require().NoError(err)
				s.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
				s.Require().Equal(synthLock.SynthDenom, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr.String()))
				s.Require().Equal(synthLock.EndTime, time.Time{})

				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddr.String(), 0)

				// Check lockID connection with intermediary account
				intAcc := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lock.ID)
				s.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())
			}

			for index, expAcc := range intermediaryAccs {
				// check intermediary account creation
				gotAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, expAcc.GetAccAddress())
				s.Require().Equal(gotAcc.Denom, expAcc.Denom)
				s.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)
				s.Require().GreaterOrEqual(gotAcc.GaugeId, uint64(1))

				valAddr, err := sdk.ValAddressFromBech32(expAcc.ValAddr)
				s.Require().NoError(err)

				// check gauge creation
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gotAcc.GaugeId)
				s.Require().NoError(err)
				s.Require().Equal(gauge.Id, gotAcc.GaugeId)
				s.Require().Equal(gauge.IsPerpetual, true)
				s.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         keeper.StakingSyntheticDenom(expAcc.Denom, valAddr.String()),
					Duration:      unbondingDuration,
				})
				s.Require().Equal(gauge.Coins, sdk.Coins(nil))
				s.Require().Equal(gauge.StartTime, s.Ctx.BlockTime())
				s.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				s.Require().Equal(gauge.FilledEpochs, uint64(0))
				s.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

				// check delegation from intermediary account to validator
				delegation, found := s.App.StakingKeeper.GetDelegation(s.Ctx, expAcc.GetAccAddress(), valAddr)
				s.Require().True(found)
				s.Require().Equal(tc.expInterDelegation[index], delegation.Shares)
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
			s.Require().False(broken, reason)

			// try delegating twice with same lockup
			for _, lock := range locks {
				err := s.App.SuperfluidKeeper.SuperfluidDelegate(s.Ctx, lock.Owner, lock.ID, valAddrs[0].String())
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestValidateLockForSFDelegate() {
	lockOwner := s.TestAccs[0]

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
		s.Run(test.name, func() {
			s.SetupTest()

			s.App.SuperfluidKeeper.SetSuperfluidAsset(s.Ctx, test.superfluidAssetToSet)

			if test.lockIdAlreadySuperfluidDelegated {
				intermediateAccount := types.NewSuperfluidIntermediaryAccount(test.lock.Coins[0].Denom, lockOwner.String(), 1)
				s.App.SuperfluidKeeper.SetLockIdIntermediaryAccountConnection(s.Ctx, test.lock.ID, intermediateAccount)
			}

			err := s.App.SuperfluidKeeper.ValidateLockForSFDelegate(s.Ctx, test.lock, lockOwner.String())
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().Equal(test.expectedErr.Error(), err.Error())
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSuperfluidUndelegate() {
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
		s.Run(tc.name, func() {
			s.SetupTest()

			bondDenom := s.App.StakingKeeper.GetParams(s.Ctx).BondDenom

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, _ := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			s.checkIntermediaryAccountDelegations(intermediaryAccs)

			for index, lockId := range tc.superUnbondingLockIds {
				// get intermediary account
				accAddr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lockId)
				intermediaryAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
				if err != nil {
					lock = &lockuptypes.PeriodLock{}
				}

				// get pre-superfluid delgations osmo supply and supplyWithOffset
				presupply := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
				presupplyWithOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, bondDenom)

				// superfluid undelegate
				err = s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, lock.Owner, lockId)
				if tc.expSuperUnbondingErr[index] {
					s.Require().Error(err)
					continue
				}
				s.Require().NoError(err)

				// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
				postsupply := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
				postsupplyWithOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, bondDenom)
				s.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
				s.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

				// check lockId and intermediary account connection deletion
				addr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lockId)
				s.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				s.Require().Error(err)

				// check unbonding synthetic lockup creation
				unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime
				synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				s.Require().NoError(err)
				s.Require().Equal(synthLock.UnderlyingLockId, lockId)
				s.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				s.Require().Equal(synthLock.EndTime, s.Ctx.BlockTime().Add(unbondingDuration))
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
			s.Require().False(broken, reason)

			// check remaining intermediary account delegation
			for index, expDelegation := range tc.expInterDelegation {
				acc := intermediaryAccs[index]
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				s.Require().NoError(err)
				delegation, found := s.App.StakingKeeper.GetDelegation(s.Ctx, acc.GetAccAddress(), valAddr)
				if expDelegation.IsZero() {
					s.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
				} else {
					s.Require().True(found)
					s.Require().Equal(expDelegation, delegation.Shares)
				}
			}

			// try undelegating twice
			for index, lockId := range tc.superUnbondingLockIds {
				if tc.expSuperUnbondingErr[index] {
					continue
				}

				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
				s.Require().NoError(err)

				err = s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, lock.Owner, lockId)
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSuperfluidUndelegateToConcentratedPosition() {
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
		s.Run(tc.name, func() {
			s.SetupTest()

			bondDenom := s.App.StakingKeeper.GetParams(s.Ctx).BondDenom

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, _ := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			s.checkIntermediaryAccountDelegations(intermediaryAccs)

			for index, lockId := range tc.superUnbondingLockIds {
				// get intermediary account
				accAddr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lockId)
				intermediaryAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
				if err != nil {
					lock = &lockuptypes.PeriodLock{}
				}

				// get pre-superfluid delgations osmo supply and supplyWithOffset
				presupply := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
				presupplyWithOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, bondDenom)

				// superfluid undelegate
				_, err = s.App.SuperfluidKeeper.SuperfluidUndelegateToConcentratedPosition(s.Ctx, lock.Owner, lockId)
				if tc.expSuperUnbondingErr[index] {
					s.Require().Error(err)
					continue
				}
				s.Require().NoError(err)

				// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
				postsupply := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
				postsupplyWithOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, bondDenom)
				s.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
				s.Require().True(postsupplyWithOffset.IsEqual(presupplyWithOffset))

				// check lockId and intermediary account connection deletion
				addr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lockId)
				s.Require().Equal(addr.String(), "")

				// check bonding synthetic lockup deletion
				_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lockId, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				s.Require().Error(err)

				// check unbonding synthetic lockup creation
				// since this is the concentrated liquidity path, no new synthetic lockup should be created
				synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lockId, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
				s.Require().Error(err)
				s.Require().Nil(synthLock)
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
			s.Require().False(broken, reason)

			// check remaining intermediary account delegation
			for index, expDelegation := range tc.expInterDelegation {
				acc := intermediaryAccs[index]
				valAddr, err := sdk.ValAddressFromBech32(acc.ValAddr)
				s.Require().NoError(err)
				delegation, found := s.App.StakingKeeper.GetDelegation(s.Ctx, acc.GetAccAddress(), valAddr)
				if expDelegation.IsZero() {
					s.Require().False(found, "expected no delegation, found delegation w/ %d shares", delegation.Shares)
				} else {
					s.Require().True(found)
					s.Require().Equal(expDelegation, delegation.Shares)
				}
			}

			// try undelegating twice
			for index, lockId := range tc.superUnbondingLockIds {
				if tc.expSuperUnbondingErr[index] {
					continue
				}

				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
				s.Require().NoError(err)

				_, err = s.App.SuperfluidKeeper.SuperfluidUndelegateToConcentratedPosition(s.Ctx, lock.Owner, lockId)
				s.Require().Error(err)
			}
		})
	}
}

// TestSuperfluidUnbondLock tests the following.
//  1. test SuperfluidUnbondLock does not work before undelegation
//  2. test SuperfluidUnbondLock makes underlying lock start unlocking
//  3. test that synthetic lockup being finished does not mean underlying lock is finished
//  4. test after SuperfluidUnbondLock + lockup time, the underlying lock is finished
func (s *KeeperTestSuite) TestSuperfluidUnbondLock() {
	s.SetupTest()

	// setup validators
	valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

	// setup superfluid delegations
	_, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, 1000000}}, denoms)
	s.checkIntermediaryAccountDelegations(intermediaryAccs)

	for _, lock := range locks {
		startTime := time.Now()
		s.Ctx = s.Ctx.WithBlockTime(startTime)
		accAddr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lock.ID)
		intermediaryAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, accAddr)
		valAddr := intermediaryAcc.ValAddr

		// first we test that SuperfluidUnbondLock would cause error before undelegating
		err := s.App.SuperfluidKeeper.SuperfluidUnbondLock(s.Ctx, lock.ID, lock.GetOwner())
		s.Require().Error(err)

		// undelegation needs to happen prior to SuperfluidUnbondLock
		err = s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, lock.Owner, lock.ID)
		s.Require().NoError(err)
		balances := s.App.BankKeeper.GetAllBalances(s.Ctx, lock.OwnerAddress())
		s.Require().Equal(0, balances.Len())

		// check that unbonding synth has been created correctly after undelegation
		unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime
		synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		s.Require().NoError(err)
		s.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
		s.Require().Equal(synthLock.SynthDenom, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		s.Require().Equal(synthLock.EndTime, s.Ctx.BlockTime().Add(unbondingDuration))

		// test SuperfluidUnbondLock
		unbondLockStartTime := startTime.Add(time.Hour)
		s.Ctx = s.Ctx.WithBlockTime(unbondLockStartTime)
		err = s.App.SuperfluidKeeper.SuperfluidUnbondLock(s.Ctx, lock.ID, lock.GetOwner())
		s.Require().NoError(err)

		// check that SuperfluidUnbondLock makes underlying lock start unlocking
		// we run WithdrawAllMaturedLocks to ensure that lock isn't getting finished immediately
		s.App.LockupKeeper.WithdrawAllMaturedLocks(s.Ctx)
		updatedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
		s.Require().NoError(err)
		s.Require().True(updatedLock.IsUnlocking())

		// check if finsihed unlocking synth lock did not increase balance
		balances = s.App.BankKeeper.GetAllBalances(s.Ctx, lock.OwnerAddress())
		s.Require().Equal(0, balances.Len())

		// test that synth lock finish does not mean underlying lock is finished
		s.Ctx = s.Ctx.WithBlockTime((startTime.Add(unbondingDuration)))
		s.App.LockupKeeper.DeleteAllMaturedSyntheticLocks(s.Ctx)
		s.App.LockupKeeper.WithdrawAllMaturedLocks(s.Ctx)
		_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
		s.Require().Error(err)
		updatedLock, err = s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
		s.Require().NoError(err)
		s.Require().True(updatedLock.IsUnlocking())

		// test after SuperfluidUnbondLock + lockup unbonding duration, lock is finished and does not exist
		s.Ctx = s.Ctx.WithBlockTime(unbondLockStartTime.Add(unbondingDuration))
		s.App.LockupKeeper.WithdrawAllMaturedLocks(s.Ctx)
		_, err = s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
		s.Require().Error(err)

		// check if finished unlocking successfully increased balance
		balances = s.App.BankKeeper.GetAllBalances(s.Ctx, lock.OwnerAddress())
		s.Require().Equal(1, balances.Len())
		s.Require().Equal(denoms[0], balances[0].Denom)
		s.Require().Equal(sdk.NewInt(1000000), balances[0].Amount)

		// check invariant is fine
		reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
		s.Require().False(broken, reason)
	}
}

func (s *KeeperTestSuite) TestSuperfluidUndelegateAndUnbondLock() {
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
		s.Run(tc.name, func() {
			s.SetupTest()

			// setup validators
			valAddrs := s.SetupValidators([]stakingtypes.BondStatus{stakingtypes.Bonded})

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, lockAmount}}, denoms)
			s.checkIntermediaryAccountDelegations(intermediaryAccs)
			s.Require().True(len(locks) > 0)

			// test invalid lock
			if tc.testInvalidLock {
				lock := lockuptypes.PeriodLock{}
				_, err := s.App.SuperfluidKeeper.SuperfluidUndelegateAndUnbondLock(s.Ctx, lock.ID, lock.GetOwner(), sdk.NewInt(1))
				s.Require().Error(err)
				return
			}

			// test undelegated lock
			if tc.undelegating {
				lock := locks[0]
				err := s.App.SuperfluidKeeper.SuperfluidUndelegate(s.Ctx, lock.GetOwner(), lock.ID)
				s.Require().NoError(err)
			}

			// test unbond lock
			if tc.unbond {
				lock := locks[0]
				err := s.App.SuperfluidKeeper.SuperfluidUnbondLock(s.Ctx, lock.ID, lock.GetOwner())
				s.Require().NoError(err)
			}

			for _, lock := range locks {
				startTime := time.Now()
				s.Ctx = s.Ctx.WithBlockTime(startTime)
				accAddr := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, lock.ID)
				intermediaryAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, accAddr)
				valAddr := intermediaryAcc.ValAddr

				// get OSMO total supply and amount to be burned
				bondDenom := s.App.StakingKeeper.BondDenom(s.Ctx)
				supplyBefore := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
				osmoAmount, err := s.App.SuperfluidKeeper.GetSuperfluidOSMOTokens(s.Ctx, intermediaryAcc.Denom, tc.unlockAmount)
				s.Require().NoError(err)

				unbondLockStartTime := startTime.Add(time.Hour)
				s.Ctx = s.Ctx.WithBlockTime(unbondLockStartTime)
				lockId, err := s.App.SuperfluidKeeper.SuperfluidUndelegateAndUnbondLock(s.Ctx, lock.ID, lock.GetOwner(), tc.unlockAmount)
				if tc.expectErr {
					s.Require().Error(err)
					continue
				}

				s.Require().NoError(err)

				// check OSMO total supply and burnt amount
				s.Require().True(osmoAmount.IsPositive())
				supplyAfter := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
				s.Require().Equal(supplyAfter, supplyBefore.Sub(sdk.NewCoin(bondDenom, osmoAmount)))

				if tc.splitLockId {
					s.Require().Equal(lockId, lock.ID+1)

					// check original underlying lock
					updatedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
					s.Require().NoError(err)
					s.Require().False(updatedLock.IsUnlocking())
					s.Require().Equal(updatedLock.Coins[0].Amount, lock.Coins[0].Amount.Sub(tc.unlockAmount))

					// check newly created underlying lock
					newLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
					s.Require().NoError(err)
					s.Require().True(newLock.IsUnlocking())
					s.Require().Equal(newLock.Coins[0].Amount, tc.unlockAmount)

					// check original synthetic lock
					stakingDenom := keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr)
					synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, stakingDenom)

					s.Require().NoError(err)
					s.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
					s.Require().Equal(synthLock.SynthDenom, stakingDenom)
					s.Require().Equal(synthLock.EndTime, time.Time{})

					// check unstaking synthetic lock is not created for the original synthetic lock
					unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime
					unstakingDenom := keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr)
					_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, unstakingDenom)
					s.Require().Error(err)

					// check newly created unstaking synthetic lock
					newSynthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lockId, unstakingDenom)

					s.Require().NoError(err)
					s.Require().Equal(newSynthLock.UnderlyingLockId, lockId)
					s.Require().Equal(newSynthLock.SynthDenom, unstakingDenom)
					s.Require().Equal(newSynthLock.EndTime, s.Ctx.BlockTime().Add(unbondingDuration))
				} else {
					s.Require().Equal(lockId, lock.ID)

					// check underlying lock
					updatedLock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, lockId)
					s.Require().NoError(err)
					s.Require().True(updatedLock.IsUnlocking())
					s.Require().Equal(updatedLock.Coins[0].Amount, tc.unlockAmount)

					// check synthetic lock
					unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime
					unstakingDenom := keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr)

					synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, unstakingDenom)
					s.Require().NoError(err)
					s.Require().Equal(synthLock.UnderlyingLockId, lock.ID)
					s.Require().Equal(synthLock.SynthDenom, unstakingDenom)
					s.Require().Equal(synthLock.EndTime, s.Ctx.BlockTime().Add(unbondingDuration))

					// check staking synthetic lock is deleted
					stakingDenom := keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr)
					_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, stakingDenom)
					s.Require().Error(err)
				}

				// check invariant is fine
				reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
				s.Require().False(broken, reason)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRefreshIntermediaryDelegationAmounts() {
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
		s.Run(tc.name, func() {
			s.SetupTest()

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			s.checkIntermediaryAccountDelegations(intermediaryAccs)

			// we make a map of intermediary account to delegation shares to store delegation share
			// before refreshing intermediary account delegations on epoch
			interAccIndexToDenomShare := make(map[int]sdk.Dec)
			for accIndex, intermediaryAcc := range intermediaryAccs {
				valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
				s.Require().NoError(err)
				delegation, found := s.App.StakingKeeper.GetDelegation(s.Ctx, intermediaryAcc.GetAccAddress(), valAddr)
				s.Require().True(found)

				interAccIndexToDenomShare[accIndex] = delegation.Shares
			}

			for denom, multiplier := range tc.multipliersByDenom {
				s.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(s.Ctx, 2, denom, multiplier)
			}

			s.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(s.Ctx)

			originalMultiplier := sdk.NewDec(20)
			for interAccIndex, intermediaryAcc := range intermediaryAccs {
				multiplier := tc.multipliersByDenom[intermediaryAcc.Denom]
				oldDelegation := interAccIndexToDenomShare[interAccIndex]
				expDelegation := oldDelegation.Mul(multiplier).Quo(originalMultiplier)
				lpTokenAmount := sdk.NewInt(1000000)
				decAmt := multiplier.Mul(lpTokenAmount.ToDec())
				denom := intermediaryAcc.Denom
				_, err := s.App.SuperfluidKeeper.GetSuperfluidAsset(s.Ctx, denom)
				s.Require().NoError(err)
				expAmount := s.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(s.Ctx, decAmt.RoundInt())

				// check delegation changes
				valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
				s.Require().NoError(err)
				delegation, found := s.App.StakingKeeper.GetDelegation(s.Ctx, intermediaryAcc.GetAccAddress(), valAddr)
				if expAmount.IsPositive() {
					s.Require().True(found)
					s.Require().Equal(delegation.Shares, expDelegation)
				} else {
					s.Require().False(found)
				}
			}

			// unbond all lockups
			for _, lock := range locks {
				// superfluid undelegate
				// handling the case same lockup is used for further delegation
				cacheCtx, write := s.Ctx.CacheContext()
				err := s.App.SuperfluidKeeper.SuperfluidUndelegate(cacheCtx, lock.Owner, lock.ID)
				if err == nil {
					write()
				}
			}
			unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime

			for _, intermediaryAcc := range intermediaryAccs {
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(unbondingDuration + time.Second))
				s.App.EndBlocker(s.Ctx, abci.RequestEndBlock{Height: s.Ctx.BlockHeight()})

				unbonded := s.App.BankKeeper.GetBalance(s.Ctx, intermediaryAcc.GetAccAddress(), sdk.DefaultBondDenom)
				s.Require().True(unbonded.IsZero())
			}

			// refresh intermediary account delegations
			s.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(s.Ctx)

			for _, intermediaryAcc := range intermediaryAccs {
				// check unbonded amount is removed after refresh operation
				refreshed := s.App.BankKeeper.GetBalance(s.Ctx, intermediaryAcc.GetAccAddress(), sdk.DefaultBondDenom)
				s.Require().True(refreshed.IsZero())
			}
		})
	}
}

// type superfluidRedelegation struct {
// 	lockId      uint64
// 	oldValIndex int64
// 	newValIndex int64
// }

// func (s *KeeperTestSuite) TestSuperfluidRedelegate() {
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
// 		s.Run(tc.name, func() {
// 			s.SetupTest()

// 			poolId := s.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
// 			s.Require().Equal(poolId, uint64(1))

// 			// setup validators
// 			valAddrs := s.SetupValidators(tc.validatorStats)

// 			// setup superfluid delegations
// 			intermediaryAccs, _ := s.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
// 			s.checkIntermediaryAccountDelegations(intermediaryAccs)

// 			// execute redelegation and check changes on store
// 			for index, srd := range tc.superRedelegations {
// 				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, srd.lockId)
// 				if err != nil {
// 					lock = &lockuptypes.PeriodLock{}
// 				}

// 				// superfluid redelegate
// 				err = s.App.SuperfluidKeeper.SuperfluidRedelegate(s.Ctx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
// 				if tc.expSuperRedelegationErr[index] {
// 					s.Require().Error(err)
// 					continue
// 				}
// 				s.Require().NoError(err)

// 				// check previous validator bonding synthetic lockup deletion
// 				_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				s.Require().Error(err)

// 				// check unbonding synthetic lockup creation
// 				params := s.App.SuperfluidKeeper.GetParams(s.Ctx)
// 				synthLock, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, srd.lockId, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				s.Require().NoError(err)
// 				s.Require().Equal(synthLock.UnderlyingLockId, srd.lockId)
// 				s.Require().Equal(synthLock.Suffix, keeper.UnstakingSuffix(valAddrs[srd.oldValIndex].String()))
// 				s.Require().Equal(synthLock.EndTime, s.Ctx.BlockTime().Add(params.UnbondingDuration))

// 				// check synthetic lockup creation
// 				synthLock2, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, srd.lockId, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
// 				s.Require().NoError(err)
// 				s.Require().Equal(synthLock2.UnderlyingLockId, srd.lockId)
// 				s.Require().Equal(synthLock2.Suffix, keeper.StakingSuffix(valAddrs[srd.newValIndex].String()))
// 				s.Require().Equal(synthLock2.EndTime, time.Time{})

// 				// check intermediary account creation
// 				lock, err = s.App.LockupKeeper.GetLockByID(s.Ctx, srd.lockId)
// 				s.Require().NoError(err)

// 				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddrs[srd.newValIndex].String(), 1)
// 				gotAcc := s.App.SuperfluidKeeper.GetIntermediaryAccount(s.Ctx, expAcc.GetAccAddress())
// 				s.Require().Equal(gotAcc.Denom, expAcc.Denom)
// 				s.Require().Equal(gotAcc.ValAddr, expAcc.ValAddr)

// 				// check gauge creation
// 				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gotAcc.GaugeId)
// 				s.Require().NoError(err)
// 				s.Require().Equal(gauge.Id, gotAcc.GaugeId)
// 				s.Require().Equal(gauge.IsPerpetual, true)
// 				s.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
// 					LockQueryType: lockuptypes.ByDuration,
// 					Denom:         expAcc.Denom + keeper.StakingSuffix(valAddrs[srd.newValIndex].String()),
// 					Duration:      params.UnbondingDuration,
// 				})
// 				s.Require().Equal(gauge.Coins, sdk.Coins(nil))
// 				s.Require().Equal(gauge.StartTime, s.Ctx.BlockTime())
// 				s.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
// 				s.Require().Equal(gauge.FilledEpochs, uint64(0))
// 				s.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))

// 				// Check lockID connection with intermediary account
// 				intAcc := s.App.SuperfluidKeeper.GetLockIdIntermediaryAccountConnection(s.Ctx, srd.lockId)
// 				s.Require().Equal(intAcc.String(), expAcc.GetAccAddress().String())

// 				// check delegation from intermediary account to validator
// 				_, found := s.App.StakingKeeper.GetDelegation(s.Ctx, expAcc.GetAccAddress(), valAddrs[srd.newValIndex])
// 				s.Require().True(found)
// 			}

// 			// try redelegating twice
// 			for index, srd := range tc.superRedelegations {
// 				if tc.expSuperRedelegationErr[index] {
// 					continue
// 				}
// 				cacheCtx, _ := s.Ctx.CacheContext()
// 				lock, err := s.App.LockupKeeper.GetLockByID(s.Ctx, srd.lockId)
// 				s.Require().NoError(err)
// 				err = s.App.SuperfluidKeeper.SuperfluidRedelegate(cacheCtx, lock.Owner, srd.lockId, valAddrs[srd.newValIndex].String())
// 				s.Require().Error(err)
// 			}
// 		})
// 	}
// }
