package keeper_test

import (
	"strings"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v10/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v10/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = suite.TestingSuite(nil)

func DivCoin(coins sdk.Coins, divisor int64) sdk.Coins {
	for id, coin := range coins {
		coins[id].Amount = coin.Amount.QuoRaw(divisor)
	}
	return coins
}

// TestDistribute tests that when the distribute command is executed on a provided gauge
// that the correct amount of rewards is sent to the correct lock owners.
func (suite *KeeperTestSuite) TestDistribute() {
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	doubleLengthGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: 2 * defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	noRewardGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{},
	}
	noRewardCoins := sdk.Coins{}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 2000)}
	fiveKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 5000)}
	tests := []struct {
		name            string
		users           []userLocks
		gauges          []perpGaugeDesc
		expectedRewards []sdk.Coins
	}{
		// gauge 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should go to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one lockup, another user with two lockups, single default gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// gauge 1 gives 3k coins. three locks, all eligible.
		// gauge 2 gives 3k coins. one lock, to twoLockupUser.
		// 1k should to oneLockupUser and 5k to twoLockupUser.
		{
			name:            "One user with one lockup (default gauge), another user with two lockups (double length gauge)",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge, doubleLengthGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
		// gauge 1 gives zero rewards.
		// both oneLockupUser and twoLockupUser should get no rewards.
		{
			name:            "One user with one lockup, another user with two lockups, both with no rewards gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge},
			expectedRewards: []sdk.Coins{noRewardCoins, noRewardCoins},
		},
		// gauge 1 gives no rewards.
		// gauge 2 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one lockup and another user with two lockups. No rewards and a default gauge",
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge, defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
	}
	for _, tc := range tests {
		suite.SetupTest()
		// setup gauges and the locks defined in the above tests, then distribute to them
		gauges := suite.SetupGauges(tc.gauges, defaultLPDenom)
		addrs := suite.SetupUserLocks(tc.users)
		_, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, gauges)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
			suite.Require().Equal(tc.expectedRewards[i].String(), bal.String(), "test %v, person %d", tc.name, i)
		}
	}
}

// TestSyntheticDistribute tests that when the distribute command is executed on a provided gauge
// the correct amount of rewards is sent to the correct synthetic lock owners.
func (suite *KeeperTestSuite) TestSyntheticDistribute() {
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	doubleLengthGauge := perpGaugeDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: 2 * defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	noRewardGauge := perpGaugeDesc{
		lockDenom:    defaultLPSyntheticDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{},
	}
	noRewardCoins := sdk.Coins{}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	twoKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 2000)}
	fiveKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 5000)}
	tests := []struct {
		name            string
		users           []userLocks
		gauges          []perpGaugeDesc
		expectedRewards []sdk.Coins
	}{
		// gauge 1 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should go to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup, another user with two synthetic lockups, both with default gauge",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// gauge 1 gives 3k coins. three locks, all eligible.
		// gauge 2 gives 3k coins. one lock, to twoLockupUser.
		// 1k should to oneLockupUser and 5k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup (default gauge), another user with two synthetic lockups (double length gauge)",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge, doubleLengthGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
		// gauge 1 gives zero rewards.
		// both oneLockupUser and twoLockupUser should get no rewards.
		{
			name:            "One user with one synthetic lockup, another user with two synthetic lockups, both with no rewards gauge",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge},
			expectedRewards: []sdk.Coins{noRewardCoins, noRewardCoins},
		},
		// gauge 1 gives no rewards.
		// gauge 2 gives 3k coins. three locks, all eligible. 1k coins per lock.
		// 1k should to oneLockupUser and 2k to twoLockupUser.
		{
			name:            "One user with one synthetic lockup (no rewards gauge), another user with two synthetic lockups (default gauge)",
			users:           []userLocks{oneSyntheticLockupUser, twoSyntheticLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge, defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
	}
	for _, tc := range tests {
		suite.SetupTest()
		// setup gauges and the synthetic locks defined in the above tests, then distribute to them
		gauges := suite.SetupGauges(tc.gauges, defaultLPSyntheticDenom)
		addrs := suite.SetupUserSyntheticLocks(tc.users)
		_, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, gauges)
		suite.Require().NoError(err)
		// check expected rewards against actual rewards received
		for i, addr := range addrs {
			var rewards string
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
			// extract the superbonding tokens from the rewards distribution
			// TODO: figure out a less hacky way of doing this
			if strings.Contains(bal.String(), "lptoken/superbonding,") {
				rewards = strings.Split(bal.String(), "lptoken/superbonding,")[1]
			}
			suite.Require().Equal(tc.expectedRewards[i].String(), rewards, "test %v, person %d", tc.name, i)
		}
	}
}

// TestGetModuleToDistributeCoins tests the sum of coins yet to be distributed
func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))

	tests := []struct {
		name string
		// each sdk.Coins in initialGaugeCoins will be used to create a new gauge
		initialGaugeCoins []sdk.Coins
		// sdk.Coins to add to gauge with gauge id
		coinsToAddToGauges map[uint64]sdk.Coins
		numEpochsPaidOver  uint64
	}{
		{
			name: "one gauge, even gauge coins, distribute over 2 epochs",
			initialGaugeCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 1089328)),
			},
			coinsToAddToGauges: map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin("stake", 2089231840)),
			},
			numEpochsPaidOver: 2,
		},
		{
			name: "two gauge, odd gauge coins, distribute over 8 epochs",
			initialGaugeCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 4304589034599)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 734523458)),
			},
			coinsToAddToGauges: map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin("stake", 3034553240)),
				2: sdk.NewCoins(sdk.NewInt64Coin("stake", 70349583945837)),
			},
			numEpochsPaidOver: 8,
		},
		{
			name: "two gauge, even gauge coins, distribute over 15 epochs",
			initialGaugeCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 4304589034590)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 734523458)),
			},
			coinsToAddToGauges: map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin("stake", 3034553240)),
				2: sdk.NewCoins(sdk.NewInt64Coin("stake", 70349583945830)),
			},
			numEpochsPaidOver: 15,
		},
		{
			name: "two gauge with 2 types of distributed coins",
			initialGaugeCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 50345983533)),
				sdk.NewCoins(sdk.NewInt64Coin("notstake", 643875873404)),
			},
			coinsToAddToGauges: map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin("notstake", 1000384571)),
				2: sdk.NewCoins(sdk.NewInt64Coin("stake", 500222222)),
			},
			numEpochsPaidOver: 3,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		incentivesKeeper := suite.App.IncentivesKeeper

		// check that the sum of coins yet to be distributed is nil
		moduleToDistributeCoins := incentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
		suite.Require().Equal(moduleToDistributeCoins, sdk.Coins(nil))
		// gauges used in this test
		gauges := []*types.Gauge{}

		expModuleToDistributeCoins := sdk.Coins{}
		// coins that will be distributed after the first epoch for each gauge
		expDistributedCoinsFirstEpoch := map[uint64]sdk.Coins{}

		suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

		// create new gauge using initialGaugeCoins
		for _, coins := range tc.initialGaugeCoins {
			_, gauge, gaugeCoins, _ := suite.SetupNewGauge(false, coins, tc.numEpochsPaidOver)
			suite.Require().Equal(coins, gaugeCoins, "test %v", tc.name)

			gauges = append(gauges, gauge)

			moduleToDistributeCoins = incentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
			expModuleToDistributeCoins = expModuleToDistributeCoins.Add(gaugeCoins...)
			suite.Require().Equal(expModuleToDistributeCoins, moduleToDistributeCoins, "test %v", tc.name)

			expDistributedCoinsFirstEpoch[gauge.Id] = coins
		}

		for gaugeID, coins := range tc.coinsToAddToGauges {
			suite.AddToGauge(coins, gaugeID)

			moduleToDistributeCoins = incentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
			expModuleToDistributeCoins = expModuleToDistributeCoins.Add(coins...)
			suite.Require().Equal(expModuleToDistributeCoins, moduleToDistributeCoins, "test %v", tc.name)

			expDistributedCoinsFirstEpoch[gaugeID] = DivCoin(expDistributedCoinsFirstEpoch[gaugeID].Add(coins...), int64(tc.numEpochsPaidOver))
		}

		// move all created gauges from upcoming to active
		// distribute coins from those gauges to stakers
		for _, gauge := range gauges {
			// move this gauge from upcoming to active
			suite.Ctx = suite.Ctx.WithBlockTime(gauge.StartTime)
			gauge, err := incentivesKeeper.GetGaugeByID(suite.Ctx, gauge.Id)
			suite.Require().NoError(err)
			err = incentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
			suite.Require().NoError(err)
			// distribute coins from this gauge to stakers
			distrCoins, err := incentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)
			suite.Require().Equal(distrCoins, expDistributedCoinsFirstEpoch[gauge.Id], "test %v", tc.name)
			// check gauge changes after distribution
			expModuleToDistributeCoins = expModuleToDistributeCoins.Sub(distrCoins)
			moduleToDistributeCoins = incentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
			suite.Require().Equal(expModuleToDistributeCoins, moduleToDistributeCoins, "test %v", tc.name)
		}
	}
}

// TestGetModuleDistributedCoins tests that the sum of coins that have been distributed so far for all of the module is correct.
func (suite *KeeperTestSuite) TestGetModuleDistributedCoins() {
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))

	tests := []struct {
		name string
		// each sdk.Coins in initialGaugeCoins will be used to create a new gauge
		initialGaugeCoins []sdk.Coins
		// sdk.Coins to add to gauge with gauge id
		coinsToAddToGauges map[uint64]sdk.Coins
		numEpochsPaidOver  uint64
	}{
		{
			name: "one gauge, even gauge coins, distribute over 2 epochs",
			initialGaugeCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 1089328)),
			},
			coinsToAddToGauges: map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin("stake", 2089231840)),
			},
			numEpochsPaidOver: 2,
		},
		{
			name: "two gauge, odd gauge coins, distribute over 8 epochs",
			initialGaugeCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 4304589034599)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 734523458)),
			},
			coinsToAddToGauges: map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin("stake", 3034553240)),
				2: sdk.NewCoins(sdk.NewInt64Coin("stake", 70349583945837)),
			},
			numEpochsPaidOver: 8,
		},
		{
			name: "two gauge, even gauge coins, distribute over 15 epochs",
			initialGaugeCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 4304589034590)),
				sdk.NewCoins(sdk.NewInt64Coin("stake", 734523458)),
			},
			coinsToAddToGauges: map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin("stake", 3034553240)),
				2: sdk.NewCoins(sdk.NewInt64Coin("stake", 70349583945830)),
			},
			numEpochsPaidOver: 15,
		},
		{
			name: "two gauge with 2 types of distributed coins",
			initialGaugeCoins: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("stake", 50345983533)),
				sdk.NewCoins(sdk.NewInt64Coin("notstake", 643875873404)),
			},
			coinsToAddToGauges: map[uint64]sdk.Coins{
				1: sdk.NewCoins(sdk.NewInt64Coin("notstake", 1000384571)),
				2: sdk.NewCoins(sdk.NewInt64Coin("stake", 500222222)),
			},
			numEpochsPaidOver: 3,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()

		incentivesKeeper := suite.App.IncentivesKeeper

		// check that the sum of coins yet to be distributed is nil
		moduleDistributedCoins := incentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
		suite.Require().Equal(moduleDistributedCoins, sdk.Coins(nil))
		// gauges used in this test
		gauges := []*types.Gauge{}

		expModuleDistributedCoins := sdk.Coins{}
		// coins that will be distributed after the first epoch for each gauge
		expDistributedCoinsFirstEpoch := map[uint64]sdk.Coins{}

		suite.LockTokens(lockOwner, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)

		// create new gauge using initialGaugeCoins
		for _, coins := range tc.initialGaugeCoins {
			_, gauge, gaugeCoins, _ := suite.SetupNewGauge(false, coins, tc.numEpochsPaidOver)
			suite.Require().Equal(coins, gaugeCoins)

			gauges = append(gauges, gauge)

			moduleDistributedCoins = incentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
			suite.Require().Equal(sdk.Coins(nil), moduleDistributedCoins, "test %v", tc.name)

			expDistributedCoinsFirstEpoch[gauge.Id] = coins
		}

		for gaugeID, coins := range tc.coinsToAddToGauges {
			suite.AddToGauge(coins, gaugeID)

			moduleDistributedCoins = incentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
			suite.Require().Equal(sdk.Coins(nil), moduleDistributedCoins, "test %v", tc.name)

			expDistributedCoinsFirstEpoch[gaugeID] = DivCoin(expDistributedCoinsFirstEpoch[gaugeID].Add(coins...), int64(tc.numEpochsPaidOver))

		}

		// move all created gauges from upcoming to active
		// distribute coins from those gauges to stakers
		for _, gauge := range gauges {
			// move this gauge from upcoming to active
			suite.Ctx = suite.Ctx.WithBlockTime(gauge.StartTime)
			gauge, err := incentivesKeeper.GetGaugeByID(suite.Ctx, gauge.Id)
			suite.Require().NoError(err)
			err = incentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
			suite.Require().NoError(err)
			// distribute coins from this gauge to stakers
			distrCoins, err := incentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
			suite.Require().NoError(err)
			suite.Require().Equal(distrCoins, expDistributedCoinsFirstEpoch[gauge.Id], "test %v", tc.name)
			// check gauge changes after distribution
			expModuleDistributedCoins = expModuleDistributedCoins.Add(DivCoin(gauge.Coins, int64(tc.numEpochsPaidOver))...)
			moduleDistributedCoins = incentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
			suite.Require().Equal(expModuleDistributedCoins, moduleDistributedCoins, "test %v", tc.name)
		}
	}
}

// TestNoLockPerpetualGaugeDistribution tests that the creation of a perp gauge that has no locks associated does not distribute any tokens.
func (suite *KeeperTestSuite) TestNoLockPerpetualGaugeDistribution() {
	suite.SetupTest()

	incentivesKeeper := suite.App.IncentivesKeeper

	// setup a perpetual gauge with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(true, coins, 2)

	// ensure the created gauge has not completed distribution
	gauges := incentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// ensure the not finished gauge matches the previously created gauge
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: true,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 1,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())

	// move the created gauge from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := incentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = incentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := incentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = incentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}

// TestNoLockNonPerpetualGaugeDistribution tests that the creation of a non perp gauge that has no locks associated does not distribute any tokens.
func (suite *KeeperTestSuite) TestNoLockNonPerpetualGaugeDistribution() {
	suite.SetupTest()

	incentivesKeeper := suite.App.IncentivesKeeper

	// setup non-perpetual gauge with no associated locks
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(false, coins, 2)

	// ensure the created gauge has not completed distribution
	gauges := incentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// ensure the not finished gauge matches the previously created gauge
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins,
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())

	// move the created gauge from upcoming to active
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := incentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = incentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := incentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = incentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}
