package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	minttypes "github.com/osmosis-labs/osmosis/v7/x/mint/types"
	"github.com/osmosis-labs/osmosis/v7/x/superfluid/keeper"
)

func (suite *KeeperTestSuite) allocateRewardsToValidator(valAddr sdk.ValAddress) {
	validator, found := suite.app.StakingKeeper.GetValidator(suite.ctx, valAddr)
	suite.Require().True(found)

	// allocate reward tokens to distribution module
	coins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20000))}
	suite.app.BankKeeper.MintCoins(suite.ctx, minttypes.ModuleName, coins)
	suite.app.BankKeeper.SendCoinsFromModuleToModule(suite.ctx, minttypes.ModuleName, distrtypes.ModuleName, coins)

	// allocate rewards to validator
	suite.ctx = suite.ctx.WithBlockHeight(suite.ctx.BlockHeight() + 1)
	decTokens := sdk.DecCoins{{Denom: sdk.DefaultBondDenom, Amount: sdk.NewDec(20000)}}
	suite.app.DistrKeeper.AllocateTokensToValidator(suite.ctx, validator, decTokens)
	suite.app.DistrKeeper.IncrementValidatorPeriod(suite.ctx, validator)
}

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
		superDelegations []superfluidDelegation
		rewardedVals     []int64
		gaugeChecks      []gaugeChecker
	}{
		{
			"happy path with single validator and delegator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}},
			[]int64{0},
			[]gaugeChecker{{1, 0, "gamm/pool/1", true}},
		},
		{
			"two LP tokens delegation to a single validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/2"}},
			[]int64{0},
			[]gaugeChecker{{1, 0, "gamm/pool/1", true}, {2, 0, "gamm/pool/2", true}},
		},
		{
			"one LP token with two locks to a single validator",
			[]stakingtypes.BondStatus{stakingtypes.Bonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {0, "gamm/pool/1"}},
			[]int64{0},
			[]gaugeChecker{{1, 0, "gamm/pool/1", true}},
		},
		{
			"add unbonded validator case",
			[]stakingtypes.BondStatus{stakingtypes.Bonded, stakingtypes.Unbonded},
			[]superfluidDelegation{{0, "gamm/pool/1"}, {1, "gamm/pool/1"}},
			[]int64{0},
			[]gaugeChecker{{1, 0, "gamm/pool/1", true}, {2, 1, "gamm/pool/1", false}},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			// setup validators
			valAddrs := suite.SetupValidators(tc.validatorStats)

			// setup superfluid delegations
			suite.SetupSuperfluidDelegations(valAddrs, tc.superDelegations)
			params := suite.app.SuperfluidKeeper.GetParams(suite.ctx)

			// allocate rewards to first validator
			for _, valIndex := range tc.rewardedVals {
				suite.allocateRewardsToValidator(valAddrs[valIndex])
			}

			// move intermediary account delegation rewards to gauges
			suite.app.SuperfluidKeeper.MoveSuperfluidDelegationRewardToGauges(suite.ctx)

			// check gauge balance
			for _, gaugeCheck := range tc.gaugeChecks {
				gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeCheck.gaugeId)
				suite.Require().NoError(err)
				suite.Require().Equal(gauge.Id, gaugeCheck.gaugeId)
				suite.Require().Equal(gauge.IsPerpetual, true)
				suite.Require().Equal(gauge.DistributeTo, lockuptypes.QueryCondition{
					LockQueryType: lockuptypes.ByDuration,
					Denom:         gaugeCheck.lpDenom + keeper.StakingSuffix(valAddrs[gaugeCheck.valIndex].String()),
					Duration:      params.UnbondingDuration,
				})
				if gaugeCheck.rewarded {
					suite.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsPositive())
				} else {
					suite.Require().True(gauge.Coins.AmountOf(sdk.DefaultBondDenom).IsZero())
				}
				suite.Require().Equal(gauge.StartTime, suite.ctx.BlockTime())
				suite.Require().Equal(gauge.NumEpochsPaidOver, uint64(1))
				suite.Require().Equal(gauge.FilledEpochs, uint64(0))
				suite.Require().Equal(gauge.DistributedCoins, sdk.Coins(nil))
			}
		})
	}
}
