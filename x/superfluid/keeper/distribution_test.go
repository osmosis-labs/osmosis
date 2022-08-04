package keeper_test

import (
	lockuptypes "github.com/osmosis-labs/osmosis/v11/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v11/x/superfluid/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (suite *KeeperTestSuite) TestMoveSuperfluidDelegationRewardToGauges() {
	type gaugeChecker struct {
		intermediaryAccIndex uint64
		valIndex             int64
		lpIndex              int64
		rewarded             bool
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
			[]superfluidDelegation{{0, 0, 0, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, 0, true}},
		},
		{
			"two LP tokens delegation to a single validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {0, 0, 1, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, 0, true}, {1, 0, 1, true}},
		},
		{
			"one LP token with two locks to a single validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 0, 0, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, 0, true}},
		},
		{
			"add unbonded validator case",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, 0, true}, {1, 1, 0, false}},
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
			unbondingDuration := suite.App.StakingKeeper.GetParams(suite.Ctx).UnbondingTime

			// allocate rewards to first validator
			for _, valIndex := range tc.rewardedVals {
				suite.AllocateRewardsToValidator(valAddrs[valIndex], sdk.NewInt(20000))
			}

			// move intermediary account delegation rewards to gauges
			suite.App.SuperfluidKeeper.MoveSuperfluidDelegationRewardToGauges(suite.Ctx)

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*suite.App.SuperfluidKeeper)(suite.Ctx)
			suite.Require().False(broken, reason)

			// check gauge balance
			for _, gaugeCheck := range tc.gaugeChecks {
				gaugeId := intermediaryAccs[gaugeCheck.intermediaryAccIndex].GaugeId
				gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         keeper.StakingSyntheticDenom(denoms[gaugeCheck.lpIndex], valAddrs[gaugeCheck.valIndex].String()),
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
