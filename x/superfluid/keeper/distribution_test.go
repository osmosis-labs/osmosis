package keeper_test

import (
	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v16/x/superfluid/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestMoveSuperfluidDelegationRewardToGauges() {
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
		// In this case, allocate reward to validators with different stat.
		// There is no difference between Bonded, Unbonding, Unbonded
		{
			"add unbonded, unbonding validator case",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded, stakingtypes.Unbonding},
			3,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}, {2, 2, 0, 1000000}},
			[]int64{0, 1, 2},
			[]gaugeChecker{{0, 0, 0, true}, {1, 1, 0, true}, {2, 2, 0, true}},
		},
		// Do not allocate rewards to the Unbonded validator. Therefore gauges are not distributed
		{
			"Unallocate to Unbonded validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			2,
			[]superfluidDelegation{{0, 0, 0, 1000000}, {1, 1, 0, 1000000}},
			[]int64{0},
			[]gaugeChecker{{0, 0, 0, true}, {1, 1, 0, false}},
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
			_, intermediaryAccs, _ := s.setupSuperfluidDelegations(valAddrs, tc.superDelegations, denoms)
			unbondingDuration := s.App.StakingKeeper.GetParams(s.Ctx).UnbondingTime

			// allocate rewards to first validator
			for _, valIndex := range tc.rewardedVals {
				s.AllocateRewardsToValidator(valAddrs[valIndex], sdk.NewInt(20000))
			}

			// move intermediary account delegation rewards to gauges
			s.App.SuperfluidKeeper.MoveSuperfluidDelegationRewardToGauges(s.Ctx)

			// check invariant is fine
			reason, broken := keeper.AllInvariants(*s.App.SuperfluidKeeper)(s.Ctx)
			s.Require().False(broken, reason)

			// check gauge balance
			for _, gaugeCheck := range tc.gaugeChecks {
				gaugeId := intermediaryAccs[gaugeCheck.intermediaryAccIndex].GaugeId
				gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
				s.Require().NoError(err)
				s.Require().Equal(gauge.Id, gaugeId)
				s.Require().Equal(gauge.IsPerpetual, true)
				s.Require().Equal(lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         keeper.StakingSyntheticDenom(denoms[gaugeCheck.lpIndex], valAddrs[gaugeCheck.valIndex].String()),
					Duration:      unbondingDuration,
				}, gauge.DistributeTo)
				if gaugeCheck.rewarded {
					s.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsPositive())
				} else {
					s.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsZero())
				}
				s.Require().Equal(gauge.StartTime, s.Ctx.BlockTime())
				s.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				s.Require().Equal(gauge.FilledEpochs, uint64(0))
				s.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))
			}
		})
	}
}
