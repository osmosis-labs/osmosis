package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v11/x/lockup/types"
)

func (suite *KeeperTestSuite) TestSuperfluidDelegatedValidatorJailed() {
	testCases := []struct {
		name                 string
		validatorStats       []stakingtypes.BondStatus
		delegatorNumber      int
		superDelegations     []superfluidDelegation
		jailedValIndexes     []int64
		expJailedLockIndexes []int64
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]int64{0},
			[]int64{0},
		},
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
			slashFactor := suite.App.SlashingKeeper.SlashFractionDoubleSign(suite.Ctx)

			// setup superfluid delegations
			for _, del := range tc.superDelegations {
				valAddr := valAddrs[del.valIndex]
				delAddr := delAddrs[del.delIndex]
				lock := suite.SetupSuperfluidDelegate(delAddr, valAddr, denoms[del.lpIndex], del.lpAmount)

				// save accounts and locks for future use
				locks = append(locks, lock)
			}

			// slash validator
			for _, valIndex := range tc.jailedValIndexes {
				validator, found := suite.App.StakingKeeper.GetValidator(suite.Ctx, valAddrs[valIndex])
				suite.Require().True(found)
				suite.Ctx = suite.Ctx.WithBlockHeight(100)
				consAddr, err := validator.GetConsAddr()
				suite.Require().NoError(err)
				// slash by slash factor
				power := sdk.TokensToConsensusPower(validator.Tokens, sdk.DefaultPowerReduction)

				// Note: this calls BeforeValidatorSlashed hook
				suite.App.EvidenceKeeper.HandleEquivocationEvidence(suite.Ctx, &evidencetypes.Equivocation{
					Height:           80,
					Time:             time.Time{},
					Power:            power,
					ConsensusAddress: consAddr.String(),
				})
				val, found := suite.App.StakingKeeper.GetValidatorByConsAddr(suite.Ctx, consAddr)
				suite.Require().True(found)
				suite.Require().Equal(val.Jailed, true)
			}

			// check lock changes after validator & lockups slashing
			for _, lockIndex := range tc.expJailedLockIndexes {
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

func (suite *KeeperTestSuite) TestTryUnbondingSuperfluidLockupDirectly() {
	testCases := []struct {
		name               string
		validatorStats     []stakingtypes.BondStatus
		delegatorNumber    int
		superDelegations   []superfluidDelegation
		expInterDelegation []sdk.Dec
	}{
		{
			"with single validator and single superfluid delegation",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
		{
			"with single validator and additional superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(20000000)}, // 50% x 20 x 1000000 x 2
		},
		{
			"with multiple validators and multiple superfluid delegations",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]sdk.Dec{sdk.NewDec(10000000), sdk.NewDec(10000000)}, // 50% x 20 x 1000000
		},
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

			// setup superfluid delegations
			_, _, _ = delAddrs, valAddrs, denoms
			_, locks := suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations, denoms)

			for _, lock := range locks {
				err := suite.App.LockupKeeper.BeginUnlock(suite.Ctx, lock.ID, sdk.Coins{})
				suite.Require().Error(err)
			}
		})
	}
}
