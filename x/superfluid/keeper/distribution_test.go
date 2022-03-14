package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
)

func (suite *KeeperTestSuite) TestMoveSuperfluidDelegationRewardToGauges() {
	type gaugeChecker struct {
		gaugeId  uint64
		valIndex int64
		lpDenom  string
		rewarded bool
	}
	testCases := []struct {
		name             string
		validatorStats   []stakingtypes.BondStatus
		delegatorNumber  int
		superDelegations []superfluidDelegation
		rewardedVals     []int64
		gaugeChecks      []gaugeChecker
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			1,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}},
			[]int64{0},
			[]gaugeChecker{{4, 0, "gamm/pool/1", true}},
		},
		{
			"two LP tokens delegation to a single validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {0, 0, "gamm/pool/2", 1000000}},
			[]int64{0},
			[]gaugeChecker{{4, 0, "gamm/pool/1", true}, {5, 0, "gamm/pool/2", true}},
		},
		{
			"one LP token with two locks to a single validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {1, 0, "gamm/pool/1", 1000000}},
			[]int64{0},
			[]gaugeChecker{{4, 0, "gamm/pool/1", true}},
		},
		{
			"add unbonded validator case",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			2,
			[]superfluidDelegation{{0, 0, "gamm/pool/1", 1000000}, {1, 1, "gamm/pool/1", 1000000}},
			[]int64{0},
			[]gaugeChecker{{4, 0, "gamm/pool/1", true}, {5, 1, "gamm/pool/1", false}},
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

			// setup superfluid delegations
			suite.SetupSuperfluidDelegations(delAddrs, valAddrs, tc.superDelegations)
			unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

			// allocate rewards to first validator
			for _, valIndex := range tc.rewardedVals {
				suite.AllocateRewardsToValidator(valAddrs[valIndex], sdk.NewInt(20000))
			}

			// move intermediary account delegation rewards to gauges
			suite.App.SuperfluidKeeper.MoveSuperfluidDelegationRewardToGauges(suite.Ctx)

			// check gauge balance
			for _, gaugeCheck := range tc.gaugeChecks {
				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeCheck.gaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gaugeCheck.gaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         keeper.StakingSyntheticDenom(gaugeCheck.lpDenom, valAddrs[gaugeCheck.valIndex].String()),
					Duration:      unbondingDuration,
				}, gauge.DistributeTo)
				if gaugeCheck.rewarded {
					suite.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsPositive())
				} else {
					suite.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsZero())
				}
				suite.Require().Equal(gauge.StartTime, suite.Ctx.BlockTime())
				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))
			}
		})
	}
}
