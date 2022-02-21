package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	appparams "github.com/osmosis-labs/osmosis/v7/app/params"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (suite *KeeperTestSuite) TestBeforeValidatorSlashed() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		superDelegations      []superfluidDelegation
		slashedValIndexes     []int64
		expSlashedLockIndexes []int64
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]int64{0},
			[]int64{0},
		},
		{
			"with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]int64{0},
			[]int64{0, 1},
		},
		{
			"with multiple validators and multiple superfluid delegations with single validator slash",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/1"}},
			[]int64{0},
			[]int64{0},
		},
		{
			"with multiple validators and multiple superfluid delegations with two validators slash",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/1"}},
			[]int64{0, 1},
			[]int64{0, 1},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			intermediaryAccs := []types.SuperfluidIntermediaryAccount{}
			locks := []lockuptypes.PeriodLock{}
			slashFactor := sdk.NewDecWithPrec(5, 2)

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				lock := suite.SetupSuperfluidDelegate(valAddr, del.lpDenom)
				expAcc := types.NewSuperfluidIntermediaryAccount(lock.Coins[0].Denom, valAddr.String(), 0)

				// save accounts and locks for future use
				intermediaryAccs = append(intermediaryAccs, expAcc)
				locks = append(locks, lock)
			}

			// slash validator
			for _, valIndex := range tc.slashedValIndexes {
				validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddrs[valIndex])
				suite.Require().True(found)
				suite.ctx = suite.ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				suite.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)
				suite.app.StakingKeeper.Slash(suite.ctx, consAddr, 80, power, slashFactor)
				// Note: this calls BeforeValidatorSlashed hook
			}

			// check lock changes after validator & lockups slashing
			for _, lockIndex := range tc.expSlashedLockIndexes {
				gotLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, locks[lockIndex].ID)
				suite.Require().NoError(err)
				suite.Require().Equal(
					gotLock.Coins.AmountOf("gamm/pool/1").String(),
					sdk.NewDec(1000000).Mul(sdk.OneDec().Sub(slashFactor)).TruncateInt().String(),
				)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSlashLockupsForUnbondingDelegationSlash() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
	}{
		{
			"happy path with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]uint64{1},
		},
		{
			"with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]uint64{1, 2},
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/1"}},
			[]uint64{1, 2},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			poolId := suite.createGammPool([]string{appparams.BaseCoinUnit, "foo"})
			suite.Require().Equal(poolId, uint64(1))

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)
			// setup superfluid delegations
			intermediaryAccs, _ := suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.superUnbondingLockIds {
				lock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)
				// superfluid undelegate
				err = suite.app.SuperfluidKeeper.SuperfluidUndelegate(suite.ctx, lock.Owner, lockId)
				suite.Require().NoError(err)
			}

			// slash unbonding lockups for all intermediary accounts
			slashFactor := sdk.NewDecWithPrec(5, 2)
			for i := 0; i < len(valAddrs); i++ {
				suite.app.SuperfluidKeeper.SlashLockupsForValidatorSlash(
					suite.ctx,
					valAddrs[i],
					suite.ctx.BlockHeight(),
					slashFactor)
			}

			// check check unbonding lockup changes
			for _, lockId := range tc.superUnbondingLockIds {
				gotLock, err := suite.app.LockupKeeper.GetLockByID(suite.ctx, lockId)
				suite.Require().NoError(err)
				suite.Require().Equal(gotLock.Coins.AmountOf("gamm/pool/1").String(), sdk.NewInt(950000).String())
			}
		})
	}
}
