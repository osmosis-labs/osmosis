package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils"
	appparams "github.com/osmosis-labs/osmosis/v27/app/params"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/keeper"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
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
		expInterDelegation []osmomath.Dec
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]osmomath.Dec{osmomath.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"with single validator and additional superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]osmomath.Dec{osmomath.NewDec(20000000)}, // 50% x 20 x 1000000 x 2
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]osmomath.Dec{osmomath.NewDec(10000000), osmomath.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"add unbonding validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonding},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]osmomath.Dec{osmomath.NewDec(10000000), osmomath.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"add unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]osmomath.Dec{osmomath.NewDec(10000000), osmomath.NewDec(10000000)}, // 50% x 20 x 1000000
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()
			bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
			s.Require().NoError(err)

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			// get pre-superfluid delegations osmo supply and supplyWithOffset
			presupply := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
			presupplyWithOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, bondDenom)

			// setup superfluid delegations
			_, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)

			// ensure post-superfluid delegations osmo supplywithoffset is the same while supply is not
			postsupply := s.App.BankKeeper.GetSupply(s.Ctx, bondDenom)
			postsupplyWithOffset := s.App.BankKeeper.GetSupplyWithOffset(s.Ctx, bondDenom)
			s.Require().False(postsupply.IsEqual(presupply), "presupply: %s   postsupply: %s", presupply, postsupply)
			s.Require().Equal(postsupplyWithOffset.String(), presupplyWithOffset.String())

			stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
			unbondingDuration := stakingParams.UnbondingTime

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
				delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, expAcc.GetAccAddress(), valAddr)
				s.Require().NoError(err)
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
				Coins:    sdk.NewCoins(sdk.NewCoin(DefaultGammAsset, osmomath.NewInt(100))),
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
				Coins:    sdk.NewCoins(sdk.NewCoin(cltypes.GetConcentratedLockupDenomFromPoolId(1), osmomath.NewInt(100))),
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
				Coins:    sdk.NewCoins(sdk.NewCoin(appparams.BaseCoinUnit, osmomath.NewInt(100))),
				Duration: time.Hour * 24 * 21,
				ID:       1,
			},
			superfluidAssetToSet: types.SuperfluidAsset{Denom: DefaultGammAsset, AssetType: types.SuperfluidAssetTypeLPShare},
			expectedErr:          errorsmod.Wrapf(types.ErrNonSuperfluidAsset, "denom: %s", appparams.BaseCoinUnit),
		},
		{
			name: "invalid lock - unbonding lockup not supported",
			lock: &lockuptypes.PeriodLock{
				Owner:    lockOwner.String(),
				Coins:    sdk.NewCoins(sdk.NewCoin(DefaultGammAsset, osmomath.NewInt(100))),
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
				Coins:    sdk.NewCoins(sdk.NewCoin(DefaultGammAsset, osmomath.NewInt(100))),
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
				Coins:    sdk.NewCoins(sdk.NewCoin(DefaultGammAsset, osmomath.NewInt(100))),
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
		expInterDelegation []osmomath.Dec
	}{
		{
			"with single validator and single superfluid delegation and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		// {
		// 	"with single validator, single superfluid delegation, add more tokens to the lock, and single undelegation",
		// 	[]stakingtypes.BondStatus{stakingtypes.Bonded},
		// 	[]superfluidDelegation{{0, 0, 1000000}},
		// 	[]uint64{1},
		// 	[]uint64{1},
		// 	[]bool{false},
		// 	[]osmomath.Dec{osmomath.ZeroDec()},
		// },
		{
			"with single validator and additional superfluid delegations and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"with multiple validators and multiple superfluid delegations and multiple undelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"add unbonding validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonding},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"add unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{2},
			[]bool{true},
			[]osmomath.Dec{},
		},
		{
			"try undelegating twice for same lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1, 1},
			[]bool{false, true},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
			s.Require().NoError(err)
			bondDenom := stakingParams.BondDenom

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

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

				// get pre-superfluid delegations osmo supply and supplyWithOffset
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
				unbondingDuration := stakingParams.UnbondingTime
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
				delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, acc.GetAccAddress(), valAddr)
				if expDelegation.IsZero() {
					s.Require().Error(err, "expected error, found delegation w/ %d shares", delegation.Shares)
				} else {
					s.Require().NoError(err)
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
		expInterDelegation []osmomath.Dec
	}{
		{
			"with single validator and single superfluid delegation and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"with single validator and additional superfluid delegations and single undelegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]uint64{1},
			[]bool{false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"with multiple validators and multiple superfluid delegations and multiple undelegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"add unbonding validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonding},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"add unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
			[]bool{false, false},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
		{
			"undelegating not available lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{2},
			[]bool{true},
			[]osmomath.Dec{},
		},
		{
			"try undelegating twice for same lock id",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1, 1},
			[]bool{false, true},
			[]osmomath.Dec{osmomath.ZeroDec()},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
			s.Require().NoError(err)
			bondDenom := stakingParams.BondDenom

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

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

				// get pre-superfluid delegations osmo supply and supplyWithOffset
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
				delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, acc.GetAccAddress(), valAddr)
				if expDelegation.IsZero() {
					s.Require().Error(err, "expected error, found delegation w/ %d shares", delegation.Shares)
				} else {
					s.Require().NoError(err)
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

	denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

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
		stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
		s.Require().NoError(err)
		unbondingDuration := stakingParams.UnbondingTime
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

		// check if finished unlocking synth lock did not increase balance
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
		s.Require().Equal(osmomath.NewInt(1000000), balances[0].Amount)

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
		unlockAmount    osmomath.Int
		expectErr       bool
		splitLockId     bool
		undelegating    bool
		unbond          bool
	}{
		{
			name:            "lock doesn't exist",
			testInvalidLock: true,
			unlockAmount:    osmomath.NewInt(0),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "unlock amount = 0",
			testInvalidLock: false,
			unlockAmount:    osmomath.NewInt(0),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "unlock amount > locked amount",
			testInvalidLock: false,
			unlockAmount:    osmomath.NewInt(lockAmount + 1),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "lock is not split if unlock amount = locked amount",
			testInvalidLock: false,
			unlockAmount:    osmomath.NewInt(lockAmount),
			expectErr:       false,
			splitLockId:     false,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "lock is split if unlock amount < locked amount",
			testInvalidLock: false,
			unlockAmount:    osmomath.NewInt(lockAmount / 2),
			expectErr:       false,
			splitLockId:     true,
			undelegating:    false,
			unbond:          false,
		},
		{
			name:            "undelegate and unbond an undelegating lock",
			testInvalidLock: false,
			unlockAmount:    osmomath.NewInt(1),
			expectErr:       true,
			splitLockId:     false,
			undelegating:    true,
			unbond:          false,
		},
		{
			name:            "undelegate and unbond an unlocking lock",
			testInvalidLock: false,
			unlockAmount:    osmomath.NewInt(1),
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

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, []superfluidDelegation{{0, 0, 0, lockAmount}}, denoms)
			s.checkIntermediaryAccountDelegations(intermediaryAccs)
			s.Require().True(len(locks) > 0)

			// test invalid lock
			if tc.testInvalidLock {
				lock := lockuptypes.PeriodLock{}
				_, err := s.App.SuperfluidKeeper.SuperfluidUndelegateAndUnbondLock(s.Ctx, lock.ID, lock.GetOwner(), osmomath.NewInt(1))
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
				bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
				s.Require().NoError(err)
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
					stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
					s.Require().NoError(err)
					unbondingDuration := stakingParams.UnbondingTime
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
					stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
					s.Require().NoError(err)
					unbondingDuration := stakingParams.UnbondingTime
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
		multipliersByDenom map[string]osmomath.Dec
	}{
		{
			"with single validator and single delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			map[string]osmomath.Dec{DefaultGammAsset: osmomath.NewDec(10)},
		},
		{
			"with single validator and additional delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			map[string]osmomath.Dec{DefaultGammAsset: osmomath.NewDec(10)},
		},
		{
			"with multiple validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			map[string]osmomath.Dec{DefaultGammAsset: osmomath.NewDec(10)},
		},
		{
			"with single validator and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 1, 1000000}},
			map[string]osmomath.Dec{DefaultGammAsset: osmomath.NewDec(10), "gamm/pool/2": osmomath.NewDec(10)},
		},
		{
			"with multiple validators and multiple denom superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 1, 1, 1000000}},
			map[string]osmomath.Dec{DefaultGammAsset: osmomath.NewDec(10), "gamm/pool/2": osmomath.NewDec(10)},
		},
		{
			"zero price multiplier check",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			map[string]osmomath.Dec{DefaultGammAsset: osmomath.NewDec(0)},
		},
		{
			"dust price multiplier check",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			map[string]osmomath.Dec{DefaultGammAsset: osmomath.NewDecWithPrec(1, 10)}, // 10^-10
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			s.SetupTest()

			// setup validators
			valAddrs := s.SetupValidators(tc.validatorStats)

			denoms, _ := s.SetupGammPoolsAndSuperfluidAssets([]osmomath.Dec{osmomath.NewDec(20), osmomath.NewDec(20)})

			// setup superfluid delegations
			_, intermediaryAccs, locks := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			s.checkIntermediaryAccountDelegations(intermediaryAccs)

			// we make a map of intermediary account to delegation shares to store delegation share
			// before refreshing intermediary account delegations on epoch
			interAccIndexToDenomShare := make(map[int]osmomath.Dec)
			for accIndex, intermediaryAcc := range intermediaryAccs {
				valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
				s.Require().NoError(err)
				delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, intermediaryAcc.GetAccAddress(), valAddr)
				s.Require().NoError(err)

				interAccIndexToDenomShare[accIndex] = delegation.Shares
			}

			for denom, multiplier := range tc.multipliersByDenom {
				s.App.SuperfluidKeeper.SetOsmoEquivalentMultiplier(s.Ctx, 2, denom, multiplier)
			}

			accs := s.App.SuperfluidKeeper.GetAllIntermediaryAccounts(s.Ctx)
			s.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(s.Ctx, accs)

			originalMultiplier := osmomath.NewDec(20)
			for interAccIndex, intermediaryAcc := range intermediaryAccs {
				multiplier := tc.multipliersByDenom[intermediaryAcc.Denom]
				oldDelegation := interAccIndexToDenomShare[interAccIndex]
				expDelegation := oldDelegation.Mul(multiplier).Quo(originalMultiplier)
				lpTokenAmount := osmomath.NewInt(1000000)
				decAmt := multiplier.Mul(lpTokenAmount.ToLegacyDec())
				denom := intermediaryAcc.Denom
				_, err := s.App.SuperfluidKeeper.GetSuperfluidAsset(s.Ctx, denom)
				s.Require().NoError(err)
				expAmount := s.App.SuperfluidKeeper.GetRiskAdjustedOsmoValue(s.Ctx, decAmt.RoundInt())

				// check delegation changes
				valAddr, err := sdk.ValAddressFromBech32(intermediaryAcc.ValAddr)
				s.Require().NoError(err)
				delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, intermediaryAcc.GetAccAddress(), valAddr)
				if expAmount.IsPositive() {
					s.Require().NoError(err)
					s.Require().Equal(delegation.Shares, expDelegation)
				} else {
					s.Require().Error(err)
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
			stakingParams, err := s.App.StakingKeeper.GetParams(s.Ctx)
			s.Require().NoError(err)
			unbondingDuration := stakingParams.UnbondingTime

			for _, intermediaryAcc := range intermediaryAccs {
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(unbondingDuration + time.Second))
				s.App.EndBlocker(s.Ctx)

				unbonded := s.App.BankKeeper.GetBalance(s.Ctx, intermediaryAcc.GetAccAddress(), sdk.DefaultBondDenom)
				s.Require().True(unbonded.IsZero())
			}

			// refresh intermediary account delegations
			accs = s.App.SuperfluidKeeper.GetAllIntermediaryAccounts(s.Ctx)
			s.App.SuperfluidKeeper.RefreshIntermediaryDelegationAmounts(s.Ctx, accs)

			for _, intermediaryAcc := range intermediaryAccs {
				// check unbonded amount is removed after refresh operation
				refreshed := s.App.BankKeeper.GetBalance(s.Ctx, intermediaryAcc.GetAccAddress(), sdk.DefaultBondDenom)
				s.Require().True(refreshed.IsZero())
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnbondConvertAndStake() {
	defaultJoinTime := s.Ctx.BlockTime()
	type tc struct {
		notSuperfluidDelegated bool
		superfluidUndelegating bool
		unlocking              bool
		unlocked               bool
		testCLLock             bool
		expectedError          error
	}
	testCases := map[string]tc{
		"lock that is superfluid delegated": {},
		"lock that is superfluid undelegating": {
			superfluidUndelegating: true,
		},
		"bonded lock, not superfluid delegated": {
			notSuperfluidDelegated: true,
		},
		"lock that is unlocking": {
			unlocking:              true,
			superfluidUndelegating: true,
		},
		"unlocked gamm shares": {
			notSuperfluidDelegated: true,
			unlocked:               true,
		},
		"error: concentrated lock should fail": {
			testCLLock: true,
			expectedError: types.SharesToMigrateDenomPrefixError{
				Denom:               "cl/pool/2",
				ExpectedDenomPrefix: "gamm/pool/",
			},
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)

			var (
				lock             *lockuptypes.PeriodLock
				lockId           uint64
				joinPoolAcc      sdk.AccAddress
				originalValAddr  sdk.ValAddress
				balancerShareOut sdk.Coin
			)
			// we use migration setup for testing with cl lock
			if tc.testCLLock {
				_, _, lock, _, joinPoolAcc, _, _, balancerShareOut, originalValAddr = s.SetupMigrationTest(s.Ctx, !tc.notSuperfluidDelegated, tc.superfluidUndelegating, tc.unlocking, tc.unlocked, osmomath.MustNewDecFromStr("1"))
				synthLockBeforeMigration, _, err := s.App.SuperfluidKeeper.GetMigrationType(s.Ctx, int64(lock.ID))
				s.Require().NoError(err)
				_, lockId, _, err = s.App.SuperfluidKeeper.MigrateSuperfluidBondedBalancerToConcentrated(s.Ctx, joinPoolAcc, lock.ID, lock.Coins[0], synthLockBeforeMigration.SynthDenom, sdk.NewCoins())
				s.Require().NoError(err)
			} else {
				// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
				_, _, lock, _, joinPoolAcc, _, balancerShareOut, originalValAddr = s.SetupUnbondConvertAndStakeTest(s.Ctx, !tc.notSuperfluidDelegated, tc.superfluidUndelegating, tc.unlocking, tc.unlocked)
				lockId = lock.ID
			}

			sender := sdk.MustAccAddressFromBech32(joinPoolAcc.String())
			valAddr := s.SetupValidator(stakingtypes.Bonded)
			minAmountToStake := osmomath.ZeroInt()
			sharesToConvert := sdk.NewInt64Coin("foo", 0)
			if tc.unlocked {
				sharesToConvert = balancerShareOut
			}

			// only test with test related denoms
			balanceBeforeConvertLockToStake := osmoutils.FilterDenoms(s.App.BankKeeper.GetAllBalances(s.Ctx, sender), []string{"foo", "stake", appparams.BaseCoinUnit})

			// system under test
			totalAmtConverted, err := s.App.SuperfluidKeeper.UnbondConvertAndStake(s.Ctx, lockId, sender.String(), valAddr.String(), minAmountToStake, sharesToConvert)
			if tc.expectedError != nil {
				s.Require().Equal(err.Error(), tc.expectedError.Error())
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// Staking & Delegation check
			s.delegationCheck(sender, originalValAddr, valAddr, totalAmtConverted)

			// Bank check
			balanceAfterConvertLockToStake := osmoutils.FilterDenoms(s.App.BankKeeper.GetAllBalances(s.Ctx, sender), []string{"foo", "stake", appparams.BaseCoinUnit})
			s.Require().True(balanceBeforeConvertLockToStake.Equal(balanceAfterConvertLockToStake))

			// if unlocked, no need to check locks since there is no lock existing
			if tc.unlocked {
				return
			}

			// lock check
			s.lockCheck(*lock, valAddr.String())
		})
	}
}

func (s *KeeperTestSuite) TestConvertLockToStake() {
	defaultJoinTime := s.Ctx.BlockTime()
	type tc struct {
		superfluidUndelegating bool
		unlocking              bool
		notSuperfluidDelegated bool

		useMinAmountToStake    bool
		senderIsNotOwnerOfLock bool
		useNonBalancerLock     bool

		expectedError error
	}
	testCases := map[string]tc{
		"lock that is superfluid delegated": {},
		"lock that is superfluid undelegating": {
			unlocking:              true,
			superfluidUndelegating: true,
		},
		"lock that is unlocking": {
			unlocking:              true,
			superfluidUndelegating: false,
		},
		"bonded lock, not superfluid delegated": {
			notSuperfluidDelegated: true,
		},
		// error cases
		"error: min amount to stake greater than actual amount": {
			useMinAmountToStake: true,
			expectedError: types.TokenConvertedLessThenDesiredStakeError{
				ActualTotalAmtToStake:   osmomath.NewInt(8309),
				ExpectedTotalAmtToStake: osmomath.NewInt(999999999),
			},
		},
		"error: use non balancer lock": {
			useNonBalancerLock: true,
			expectedError: types.SharesToMigrateDenomPrefixError{
				Denom:               "foo",
				ExpectedDenomPrefix: "gamm/pool/",
			},
		},
		"error: sender is not owner of lock ": {
			senderIsNotOwnerOfLock: true,
			expectedError: types.LockOwnerMismatchError{
				LockId:        1,
				LockOwner:     s.TestAccs[0].String(),
				ProvidedOwner: s.TestAccs[1].String(),
			},
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)
			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			_, _, lock, _, _, _, _, originalValAddr := s.SetupUnbondConvertAndStakeTest(s.Ctx, !tc.notSuperfluidDelegated, tc.superfluidUndelegating, false, false)

			// testing params
			sender := sdk.MustAccAddressFromBech32(lock.Owner)
			if tc.senderIsNotOwnerOfLock {
				sender = s.TestAccs[1]
			}

			if tc.useNonBalancerLock {
				nonBalancerShareDenomCoins := sdk.NewCoins(sdk.NewCoin("foo", osmomath.NewInt(100)))
				s.FundAcc(sender, nonBalancerShareDenomCoins)
				newLock, err := s.App.LockupKeeper.CreateLock(s.Ctx, sender, nonBalancerShareDenomCoins, time.Second)
				s.Require().NoError(err)
				lock = &newLock
			}

			valAddr := s.SetupValidator(stakingtypes.Bonded)
			minAmountToStake := osmomath.ZeroInt()
			if tc.useMinAmountToStake {
				minAmountToStake = osmomath.NewInt(999999999)
			}

			balanceBeforeConvertLockToStake := s.App.BankKeeper.GetAllBalances(s.Ctx, sender)

			// system under test
			totalAmtConverted, err := s.App.SuperfluidKeeper.ConvertLockToStake(s.Ctx, sender, valAddr.String(), lock.ID, minAmountToStake)
			if tc.expectedError != nil {
				s.Require().Error(err)
				// TODO: come back to this specific err case
				// err check for LockOwnerMismatchError needs further refactoring for all these test cases
				// since lock owner is not know-able at the time of test creation
				if !tc.senderIsNotOwnerOfLock {
					s.Require().Equal(err.Error(), tc.expectedError.Error())
				}
				return
			}
			s.Require().NoError(err)

			// Staking & Delegation check
			s.delegationCheck(sender, originalValAddr, valAddr, totalAmtConverted)

			// Lock check
			s.lockCheck(*lock, valAddr.String())

			// Bank check
			balanceAfterConvertLockToStake := s.App.BankKeeper.GetAllBalances(s.Ctx, sender)
			s.Require().True(balanceBeforeConvertLockToStake.Equal(balanceAfterConvertLockToStake))
		})
	}
}

func (s *KeeperTestSuite) TestConvertUnlockedToStake() {
	defaultJoinTime := s.Ctx.BlockTime()
	type tc struct {
		usePartialShares    bool
		useMinAmountToStake bool
		useNonGammPrefix    bool
		expectedError       error
	}
	testCases := map[string]tc{
		"convert unlocked gamm shares": {},
		"convert partial shares": {
			usePartialShares: true,
		},
		"min amount to stake exceeds exit pool amount": {
			useMinAmountToStake: true,
			expectedError: types.TokenConvertedLessThenDesiredStakeError{
				ActualTotalAmtToStake:   osmomath.NewInt(8309),
				ExpectedTotalAmtToStake: osmomath.NewInt(999999999),
			},
		},
		"error: use non gamm prefix": {
			useNonGammPrefix: true,
			expectedError: types.TokenConvertedLessThenDesiredStakeError{
				ActualTotalAmtToStake:   osmomath.NewInt(8309),
				ExpectedTotalAmtToStake: osmomath.NewInt(999999999),
			},
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			s.Ctx = s.Ctx.WithBlockTime(defaultJoinTime)

			// We bundle all migration setup into a single function to avoid repeating the same code for each test case.
			_, _, _, _, sender, poolId, shareOut, _ := s.SetupUnbondConvertAndStakeTest(s.Ctx, false, false, false, true)

			// testing params
			valAddr := s.SetupValidator(stakingtypes.Bonded)
			minAmtToStake := osmomath.ZeroInt()
			if tc.useMinAmountToStake {
				minAmtToStake = osmomath.NewInt(9999999999)
			}
			sharesToStake := shareOut
			if tc.usePartialShares {
				sharesToStake.Amount = sharesToStake.Amount.Quo(osmomath.NewInt(2))
			}
			if tc.useNonGammPrefix {
				sharesToStake = sdk.NewInt64Coin("foo", 10)
			}

			balanceBeforeConvert := s.App.BankKeeper.GetBalance(s.Ctx, sender, shareOut.Denom)
			s.Require().True(!balanceBeforeConvert.Amount.IsZero())

			bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
			s.Require().NoError(err)
			totalPoolLiquidityBeforeConvert, err := s.App.GAMMKeeper.GetTotalPoolLiquidity(s.Ctx, poolId)
			s.Require().NoError(err)
			bondDenomPoolAmtBeforeConvert := totalPoolLiquidityBeforeConvert.AmountOf(bondDenom)

			var expectedBondDenomAmt osmomath.Int
			// check expected bond denom pool liquidity amount after conversion(only for non error cases)
			if tc.expectedError == nil {
				expectedBondDenomAmt = s.getExpectedBondDenomPoolAmtAfterConvert(sender, poolId, sharesToStake)
			}

			// system under test
			totalAmtConverted, err := s.App.SuperfluidKeeper.ConvertUnlockedToStake(s.Ctx, sender, valAddr.String(), sharesToStake, minAmtToStake)
			if tc.expectedError != nil {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// gamm check
			totalPoolLiquidityAfterConvert, err := s.App.GAMMKeeper.GetTotalPoolLiquidity(s.Ctx, poolId)
			s.Require().NoError(err)
			// check that pool liquidity have reduced
			bondDenomPoolAmtAfterConvert := totalPoolLiquidityAfterConvert.AmountOf(bondDenom)
			s.Require().True(bondDenomPoolAmtAfterConvert.LT(bondDenomPoolAmtBeforeConvert))
			s.Require().True(expectedBondDenomAmt.Equal(bondDenomPoolAmtAfterConvert))

			// Staking & Delegation check
			s.delegationCheck(sender, sdk.ValAddress{}, valAddr, totalAmtConverted)

			// Bank check
			balanceAfterConvertLockToStake := s.App.BankKeeper.GetBalance(s.Ctx, sender, shareOut.Denom)
			if tc.usePartialShares {
				s.Require().True(balanceAfterConvertLockToStake.Amount.Equal(sharesToStake.Amount))
			} else {
				s.Require().True(balanceAfterConvertLockToStake.IsZero())
			}
		})
	}
}

func (s *KeeperTestSuite) TestConvertGammSharesToOsmoAndStake() {
	type tc struct {
		useInvalidValAddr        bool
		useMinAmtToStake         bool
		useValSetPrefSingleVal   bool
		useValSetPrefMultipleVal bool
		useSuperfluid            bool

		expectedError string
	}
	testCases := map[string]tc{
		"superfluid staked, provide validator address": {},
		"use val set preference (single validator)": {
			useValSetPrefSingleVal: true,
		},
		"multiple validator returned from valset pref": {
			useValSetPrefMultipleVal: true,
		},
		"No validator returned from valset, fall back to superfluid delegation": {
			useSuperfluid: true,
		},
		"error: invalid val address": {
			useInvalidValAddr: true,
			expectedError:     "invalid Bech32 prefix; expected osmovaloper, got osmo",
		},
		"error: min amount to stake exceeds actual amount staking": {
			useMinAmtToStake: true,
			expectedError:    "actual amount converted to stake (8309) is less then minimum amount expected to be staked (999999999)",
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.SetupTest()
			bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
			s.Require().NoError(err)

			// use setup helper function to setup pool, fund account with gamm shares
			// note that we're not creating any locks here.
			_, _, _, _, sender, poolId, shareOut, _ := s.SetupUnbondConvertAndStakeTest(s.Ctx, false, false, false, true)
			// exit pool
			exitCoins, err := s.App.GAMMKeeper.ExitPool(s.Ctx, sender, poolId, shareOut.Amount, sdk.NewCoins())
			s.Require().NoError(err)

			// test params
			originalSuperfluidValAddr := ""
			valAddr := s.SetupValidator(stakingtypes.Bonded)
			valAddrString := valAddr.String()
			if tc.useInvalidValAddr {
				valAddrString = s.TestAccs[0].String()
			}

			stakeCoin := sdk.NewInt64Coin(bondDenom, 100000)
			if tc.useValSetPrefSingleVal || tc.useValSetPrefMultipleVal {
				valAddrString = ""

				s.FundAcc(sender, sdk.NewCoins(stakeCoin))
				validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr)
				s.Require().NoError(err)

				_, err = s.App.StakingKeeper.Delegate(s.Ctx, sender, stakeCoin.Amount, stakingtypes.Unbonded, validator, true)
				s.Require().NoError(err)
			}
			if tc.useSuperfluid {
				originalSuperfluidValAddr = valAddrString
				valAddrString = ""
			}

			// if test case is setting multiple validator, stake one more time to a different validator
			if tc.useValSetPrefMultipleVal {
				valAddr2 := s.SetupValidator(stakingtypes.Bonded)
				stakeCoin := sdk.NewInt64Coin(bondDenom, 100000)
				s.FundAcc(sender, sdk.NewCoins(stakeCoin))
				validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, valAddr2)
				s.Require().NoError(err)
				_, err = s.App.StakingKeeper.Delegate(s.Ctx, sender, stakeCoin.Amount, stakingtypes.Unbonded, validator, true)
				s.Require().NoError(err)
			}

			minAmtToStake := osmomath.ZeroInt()
			if tc.useMinAmtToStake {
				minAmtToStake = osmomath.NewInt(999999999)
			}

			// mark expected shares before swap
			nonStakeDenomCoin := osmoutils.FilterDenoms(exitCoins, []string{"foo"})[0]
			stakeDenomCoin := exitCoins.AmountOf(bondDenom)
			// use cache context to get expected amount after swap without changing test state
			cc, _ := s.Ctx.CacheContext()
			tokenOutAmt, _, err := s.App.PoolManagerKeeper.SwapExactAmountIn(cc, sender, poolId, nonStakeDenomCoin, bondDenom, osmomath.ZeroInt())
			s.Require().NoError(err)
			expectedTotalAmtStaked := tokenOutAmt.Add(stakeDenomCoin)

			// mark pool liquidity
			pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
			s.Require().NoError(err)
			poolLiquidityBeforeSwap := pool.GetTotalPoolLiquidity(s.Ctx)
			poolBeforeBondDenomAmt := poolLiquidityBeforeSwap.AmountOf("stake")
			poolBeforeNonBondDenomAmt := poolLiquidityBeforeSwap.AmountOf("foo")

			// system under test.
			totalAmtConverted, err := s.App.SuperfluidKeeper.ConvertGammSharesToOsmoAndStake(s.Ctx, sender, valAddrString, poolId, exitCoins, minAmtToStake, originalSuperfluidValAddr)
			if tc.expectedError != "" {
				s.Require().Equal(err.Error(), tc.expectedError)
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			// check that total Amount converted is equal to (swap result + original stake denom amount)
			s.Require().True(expectedTotalAmtStaked.Equal(totalAmtConverted))

			// check staking
			if tc.useValSetPrefMultipleVal {
				delegations, err := s.App.StakingKeeper.GetAllDelegatorDelegations(s.Ctx, sender)
				s.Require().NoError(err)
				// we used two validators
				s.Require().True(len(delegations) == 2)

				delegation0Shares := delegations[0].Shares
				delegation1Shares := delegations[1].Shares

				shareDiff := delegation0Shares.Sub(delegation1Shares).Abs()

				// in practice, the share amount between two validators should be equal,
				// but due to how we handle truncation and rounding in valset pref, we expect the diff to be under one dec.
				s.Require().True(shareDiff.LTE(osmomath.OneDec()))
			} else {
				_, err := s.App.StakingKeeper.GetDelegation(s.Ctx, sender, valAddr)
				s.Require().NoError(err)
			}

			// check pool
			pool, err = s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
			s.Require().NoError(err)
			poolLiquidityAfterSwap := pool.GetTotalPoolLiquidity(s.Ctx)
			poolAfterBondDenomAmt := poolLiquidityAfterSwap.AmountOf("stake")
			poolAfterNonBondDenomAmt := poolLiquidityAfterSwap.AmountOf("foo")
			// we swapped from non-bond denom to bond denom,
			// thus bond denom token in pool should have decreased, non bond denom token should have increased
			s.Require().True(poolBeforeBondDenomAmt.GT(poolAfterBondDenomAmt))
			s.Require().True(poolBeforeNonBondDenomAmt.LT(poolAfterNonBondDenomAmt))
		})
	}
}

func (s *KeeperTestSuite) TestDelegateBaseOnValsetPref() {
	type tc struct {
		useValAddr                   bool
		haveExistingDelegation       bool
		useOriginalSuperfluidValAddr bool

		useInvalidValAddr bool

		expectedError string
	}
	testCases := map[string]tc{
		"provide val address": {
			useValAddr: true,
		},
		"use valset pref delegation": {
			haveExistingDelegation: true,
		},
		"using valset pref fail, fallback to using provided original superfluid address": {
			useOriginalSuperfluidValAddr: true,
		},
		"error: using valset pref fail, no superfluid address provided": {
			expectedError: "empty address string is not allowed",
		},
		"error: invalid val address provided": {
			useInvalidValAddr: true,
			expectedError:     "ecoding bech32 failed: invalid character not part of charset",
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			s.Setup()
			bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
			s.Require().NoError(err)
			stakeAmount := osmomath.NewInt(100)

			sender := s.TestAccs[0]
			s.FundAcc(sender, sdk.NewCoins(sdk.NewCoin(bondDenom, stakeAmount)))

			var valAddr string
			if tc.useValAddr {
				valAddr = s.SetupValidator(stakingtypes.Bonded).String()
			}
			if tc.useInvalidValAddr {
				valAddr = s.SetupValidator(stakingtypes.Bonded).String() + "invalid"
			}

			var originalSuperfluidValAddr string
			if tc.useOriginalSuperfluidValAddr {
				originalSuperfluidValAddr = s.SetupValidator(stakingtypes.Bonded).String()
			}

			// by having existing delegation, we can test val set pref based delegation
			var superfluidStakedValAddr sdk.ValAddress
			if tc.haveExistingDelegation {
				superfluidStakedValAddr = s.SetupValidator(stakingtypes.Bonded)

				stakeCoin := sdk.NewInt64Coin(bondDenom, 100)
				s.FundAcc(sender, sdk.NewCoins(stakeCoin))
				validator, err := s.App.StakingKeeper.GetValidator(s.Ctx, superfluidStakedValAddr)
				s.Require().NoError(err)
				_, err = s.App.StakingKeeper.Delegate(s.Ctx, sender, stakeCoin.Amount, stakingtypes.Unbonded, validator, true)
				s.Require().NoError(err)
			}

			// system under test
			err = s.App.SuperfluidKeeper.DelegateBaseOnValsetPref(s.Ctx, sender, valAddr, originalSuperfluidValAddr, stakeAmount)
			if tc.expectedError != "" {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expectedError)
				return
			}

			s.Require().NoError(err)

			// check delegation
			if valAddr != "" || originalSuperfluidValAddr != "" {
				// we want to check which ever param that was passed in with value
				var delegatedAddr string
				if valAddr == "" {
					delegatedAddr = originalSuperfluidValAddr
				} else {
					delegatedAddr = valAddr
				}

				val, err := sdk.ValAddressFromBech32(delegatedAddr)
				s.Require().NoError(err)
				del, err := s.App.StakingKeeper.GetDelegation(s.Ctx, sender, val)
				s.Require().NoError(err)
				s.Require().True(del.Shares.RoundInt().Equal(stakeAmount))
				return
			}

			// if we are testing valset-pref case(already deleated), check existing delegation address to see if delegation increased
			if tc.haveExistingDelegation {
				del, err := s.App.StakingKeeper.GetDelegation(s.Ctx, sender, superfluidStakedValAddr)
				s.Require().NoError(err)
				// should be 200(original delegated amount + newly staked amount)
				s.Require().True(del.Shares.RoundInt().Equal(stakeAmount.Mul(osmomath.NewInt(2))))
				return
			}
		})
	}
}

func (s *KeeperTestSuite) SetupUnbondConvertAndStakeTest(ctx sdk.Context, superfluidDelegated, superfluidUndelegating, unlocking, noLock bool) (joinPoolAmt sdk.Coins, balancerIntermediaryAcc types.SuperfluidIntermediaryAccount, balancerLock *lockuptypes.PeriodLock, poolCreateAcc, poolJoinAcc sdk.AccAddress, balancerPooId uint64, balancerPoolShareOut sdk.Coin, valAddr sdk.ValAddress) { //nolint:revive // TODO: refactor this function
	bankKeeper := s.App.BankKeeper
	gammKeeper := s.App.GAMMKeeper
	superfluidKeeper := s.App.SuperfluidKeeper
	lockupKeeper := s.App.LockupKeeper
	stakingKeeper := s.App.StakingKeeper
	poolmanagerKeeper := s.App.PoolManagerKeeper

	// Generate and fund two accounts.
	// Account 1 will be the account that creates the pool.
	// Account 2 will be the account that joins the pool.
	delAddrs := CreateRandomAccounts(2)
	poolCreateAcc = delAddrs[0]
	poolJoinAcc = delAddrs[1]
	for _, acc := range delAddrs {
		err := testutil.FundAccount(ctx, bankKeeper, acc, defaultAcctFunds)
		s.Require().NoError(err)
	}

	// Set up a single validator.
	valAddr = s.SetupValidator(stakingtypes.Bonded)

	// Create a balancer pool of "stake" and "foo".
	msg := balancer.NewMsgCreateBalancerPool(poolCreateAcc, balancer.PoolParams{
		SwapFee: osmomath.NewDecWithPrec(1, 2),
		ExitFee: osmomath.NewDec(0),
	}, defaultPoolAssets, defaultFutureGovernor)
	balancerPooId, err := poolmanagerKeeper.CreatePool(ctx, msg)
	s.Require().NoError(err)

	// Join the balancer pool.
	// Note the account balance before and after joining the pool.
	balanceBeforeJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)
	_, _, err = gammKeeper.JoinPoolNoSwap(ctx, poolJoinAcc, balancerPooId, gammtypes.OneShare.MulRaw(50), sdk.Coins{})
	s.Require().NoError(err)
	balanceAfterJoin := bankKeeper.GetAllBalances(ctx, poolJoinAcc)

	// The balancer join pool amount is the difference between the account balance before and after joining the pool.
	joinPoolAmt, _ = balanceBeforeJoin.SafeSub(balanceAfterJoin...)

	// Determine the balancer pool's LP token denomination.
	balancerPoolDenom := gammtypes.GetPoolShareDenom(balancerPooId)

	// Register the balancer pool's LP token as a superfluid asset
	err = superfluidKeeper.AddNewSuperfluidAsset(ctx, types.SuperfluidAsset{
		Denom:     balancerPoolDenom,
		AssetType: types.SuperfluidAssetTypeLPShare,
	})
	s.Require().NoError(err)

	// Note how much of the balancer pool's LP token the account that joined the pool has.
	balancerPoolShareOut = bankKeeper.GetBalance(ctx, poolJoinAcc, balancerPoolDenom)

	// The unbonding duration is the same as the staking module's unbonding duration.
	stakingParams, err := stakingKeeper.GetParams(ctx)
	unbondingDuration := stakingParams.UnbondingTime

	// Lock the LP tokens for the duration of the unbonding period.
	originalGammLockId := uint64(0)
	if !noLock {
		originalGammLockId = s.LockTokens(poolJoinAcc, sdk.NewCoins(balancerPoolShareOut), unbondingDuration)
	}

	// Superfluid delegate the balancer lock if the test case requires it.
	// Note the intermediary account that was created.
	if superfluidDelegated {
		err = superfluidKeeper.SuperfluidDelegate(ctx, poolJoinAcc.String(), originalGammLockId, valAddr.String())
		s.Require().NoError(err)
		intermediaryAccConnection := superfluidKeeper.GetLockIdIntermediaryAccountConnection(ctx, originalGammLockId)
		balancerIntermediaryAcc = superfluidKeeper.GetIntermediaryAccount(ctx, intermediaryAccConnection)
	}

	// Superfluid undelegate the lock if the test case requires it.
	if superfluidUndelegating {
		err = superfluidKeeper.SuperfluidUndelegate(ctx, poolJoinAcc.String(), originalGammLockId)
		s.Require().NoError(err)
	}

	// Unlock the balancer lock if the test case requires it.
	if unlocking {
		// If lock was superfluid staked, we can't unlock via `BeginUnlock`,
		// we need to unlock lock via `SuperfluidUnbondLock`
		if superfluidUndelegating {
			err = superfluidKeeper.SuperfluidUnbondLock(ctx, originalGammLockId, poolJoinAcc.String())
			s.Require().NoError(err)
		} else {
			lock, err := lockupKeeper.GetLockByID(ctx, originalGammLockId)
			s.Require().NoError(err)
			_, err = lockupKeeper.BeginUnlock(ctx, originalGammLockId, lock.Coins)
			s.Require().NoError(err)
		}
	}

	balancerLock = &lockuptypes.PeriodLock{}
	if !noLock {
		balancerLock, err = lockupKeeper.GetLockByID(ctx, originalGammLockId)
		s.Require().NoError(err)
	}

	s.Require().NoError(err)
	return joinPoolAmt, balancerIntermediaryAcc, balancerLock, poolCreateAcc, poolJoinAcc, balancerPooId, balancerPoolShareOut, valAddr
}

// delegationCheck checks staking related invariants of the test.
// We check the following in this method:
// - if superfluid staked previously, check if the original validator's delegation has been deleted.
// - Check if the delegation of the new validator matches what's expected.
func (s *KeeperTestSuite) delegationCheck(sender sdk.AccAddress, originalValAddr, newValAddr sdk.ValAddress, totalAmtConverted osmomath.Int) {
	if !originalValAddr.Empty() {
		// check if original superfluid staked lock's delegation is successfully deleted
		_, err := s.App.StakingKeeper.GetDelegation(s.Ctx, sender, originalValAddr)
		s.Require().Error(err)
	}
	// check if delegation amount matches
	delegation, err := s.App.StakingKeeper.GetDelegation(s.Ctx, sender, newValAddr)
	s.Require().NoError(err)
	s.Require().True(totalAmtConverted.ToLegacyDec().Equal(delegation.Shares))
	s.Require().True(delegation.Shares.Equal(totalAmtConverted.ToLegacyDec()))
}

// lockCheck checks lock related invariants of the test.
// We check the following in this method:
// - check if old synth lock has been deleted (both staking & unstaking)
// - check if old lock has been successfully deleted.
func (s *KeeperTestSuite) lockCheck(lock lockuptypes.PeriodLock, valAddr string) {
	// The synthetic lockup should be deleted.
	_, err := s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, keeper.StakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
	s.Require().Error(err)

	// intermediary account should have been deleted
	_, err = s.App.LockupKeeper.GetSyntheticLockup(s.Ctx, lock.ID, keeper.UnstakingSyntheticDenom(lock.Coins[0].Denom, valAddr))
	s.Require().Error(err)

	// Lock check
	_, err = s.App.LockupKeeper.GetLockByID(s.Ctx, lock.ID)
	s.Require().Error(err)
}

func (s *KeeperTestSuite) getExpectedBondDenomPoolAmtAfterConvert(sender sdk.AccAddress, poolId uint64, sharesToStake sdk.Coin) osmomath.Int {
	bondDenom, err := s.App.StakingKeeper.BondDenom(s.Ctx)
	s.Require().NoError(err)
	cc, _ := s.Ctx.CacheContext()
	exitCoins, err := s.App.GAMMKeeper.ExitPool(cc, sender, poolId, sharesToStake.Amount, sdk.NewCoins())
	s.Require().NoError(err)

	var nonOsmoCoin sdk.Coin
	for _, exitCoin := range exitCoins {
		// if coin is not uosmo, add it to non-osmo Coins
		if exitCoin.Denom != bondDenom {
			nonOsmoCoin = exitCoin
		}
	}
	_, _, err = s.App.PoolManagerKeeper.SwapExactAmountIn(cc, sender, poolId, nonOsmoCoin, bondDenom, osmomath.ZeroInt())
	s.Require().NoError(err)
	expectedLiquidity, err := s.App.GAMMKeeper.GetTotalPoolLiquidity(cc, poolId)
	s.Require().NoError(err)

	return expectedLiquidity.AmountOf(bondDenom)
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
// 				_, err := s.App.StakingKeeper.GetDelegation(s.Ctx, expAcc.GetAccAddress(), valAddrs[srd.newValIndex])
// 				s.Require().NoError(err)
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
