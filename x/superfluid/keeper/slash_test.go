package keeper_test

import (
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v10/x/superfluid/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestBeforeValidatorSlashed() {
	testCases := []struct {
		name                  string
		validatorStats        []stakingtypes.BondStatus
		delegatorNumber       int
		superDelegations      []superfluidDelegation
		slashedValIndexes     []int64
		expSlashedLockIndexes []int64
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]int64{0},
			[]int64{0},
		},
		// {
		// 	"with single validator and multiple superfluid delegations",
		// 	[]stakingtypes.BondStatus{stakingtypes.Bonded},
		// 	2,
		// 	[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 0, 0, 1000000}},
		// 	[]int64{0},
		// 	[]int64{0, 1},
		// },
		// {
		// 	"with multiple validators and multiple superfluid delegations with single validator slash",
		// 	[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
		// 	2,
		// 	[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
		// 	[]int64{0},
		// 	[]int64{0},
		// },
		// {
		// 	"with multiple validators and multiple superfluid delegations with two validators slash",
		// 	[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
		// 	2,
		// 	[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
		// 	[]int64{0, 1},
		// 	[]int64{0, 1},
		// },
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			// Generate delegator addresses
			delAddrs := CreateRandomAccounts(tc.delegatorNumber)

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			denoms, _ := suite.SetupGammPoolsAndSuperfluidAssets([]sdk.Dec{sdk.NewDec(20), sdk.NewDec(20)})

			locks := []lockuptypes.PeriodLock{}
			slashFactor := sdk.NewDecWithPrec(5, 2)

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				delAddr := delAddrs[del.delIndex]
				lock := suite.setupSuperfluidDelegate(delAddr, valAddr, denoms[del.lpIndex], del.lpAmount)

				// save accounts and locks for future use
				locks = append(locks, lock)
			}

			// slash validator
			for _, valIndex := range tc.slashedValIndexes {
				validator, found := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddrs[valIndex])
				suite.Require().True(found)
				suite.Ctx = suite.Ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				suite.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)
				suite.App.StakingKeeper.Slash(suite.Ctx, consAddr, 80, power, slashFactor)
				// Note: this calls BeforeValidatorSlashed hook
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// check lock changes after validator & lockups slashing
			for _, lockIndex := range tc.expSlashedLockIndexes {
				gotLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, locks[lockIndex].ID)
				suite.Require().NoError(err)
				suite.Require().Equal(
					gotLock.Coins.AmountOf(denoms[0]).String(),
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
		delegatorNumber       int
		superDelegations      []superfluidDelegation
		superUnbondingLockIds []uint64
	}{
		{
			"happy path with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]uint64{1},
		},
		{
			"with single validator and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 0, 0, 1000000}},
			[]uint64{1, 2},
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]uint64{1, 2},
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
			_, intermediaryAccs, _ := suite.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			suite.checkIntermediaryAccountDelegations(intermediaryAccs)

			for _, lockId := range tc.superUnbondingLockIds {
				lock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)
				// superfluid undelegate
				err = suite.App.SuperfluidKeeper.SuperfluidUndelegate(suite.Ctx, lock.Owner, lockId)
				suite.Require().NoError(err)
			}

			// slash unbonding lockups for all intermediary accounts
			slashFactor := sdk.NewDecWithPrec(5, 2)
			for i := 0; i < len(valAddrs); i++ {
				suite.App.SuperfluidKeeper.SlashLockupsForValidatorSlash(
					suite.Ctx,
					valAddrs[i],
					suite.Ctx.BlockHeight(),
					slashFactor)
			}

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// check check unbonding lockup changes
			for _, lockId := range tc.superUnbondingLockIds {
				gotLock, err := suite.App.LockupKeeper.GetLockByID(suite.Ctx, lockId)
				suite.Require().NoError(err)
				suite.Require().Equal(gotLock.Coins[0].Amount.String(), sdk.NewInt(950000).String())
			}
		})
	}
}
