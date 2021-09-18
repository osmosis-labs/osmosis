package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

// TestDistribute tests that when the distribute command is executed on
// a provided gauge,
func (suite *KeeperTestSuite) TestDistribute() {
	twoLockupUser := userLocks{
		lockDurations: []time.Duration{time.Second, 2 * time.Second},
		lockAmounts:   []sdk.Coins{defaultLPTokens, defaultLPTokens},
	}
	defaultGauge := perpGaugeDesc{
		lockDenom:    defaultLPDenom,
		lockDuration: defaultLockDuration,
		rewardAmount: sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 3000)},
	}
	oneKRewardCoins := sdk.Coins{sdk.NewInt64Coin(defaultRewardDenom, 1000)}
	twoKRewardCoins := oneKRewardCoins.Add(oneKRewardCoins...)
	tests := []struct {
		users           []userLocks
		gauges          []perpGaugeDesc
		expectedRewards []sdk.Coins
	}{
		{
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
	}
	for _, tc := range tests {
		suite.SetupTest()
		gauges := suite.SetupGauges(tc.gauges)
		addrs := suite.SetupUserLocks(tc.users)
		for _, g := range gauges {
			suite.app.IncentivesKeeper.Distribute(suite.ctx, g)
		}
		// Check expected rewards
		for i, addr := range addrs {
			bal := suite.app.BankKeeper.GetAllBalances(suite.ctx, addr)
			suite.Require().Equal(bal.String(), tc.expectedRewards[i].String())
		}
	}
}

func (suite *KeeperTestSuite) TestInvalidDurationGaugeCreationValidation() {
	suite.SetupTest()

	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration / 2, // 0.5 second, invalid duration
	}
	_, err := suite.app.IncentivesKeeper.CreateGauge(suite.ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	distrTo.Duration = defaultLockDuration
	_, err = suite.app.IncentivesKeeper.CreateGauge(suite.ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	// test for module get gauges
	suite.SetupTest()

	// initial check
	coins := suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup lock and gauge
	_, gaugeID, gaugeCoins, startTime := suite.SetupLockAndGauge(false)

	// check after gauge creation
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, gaugeCoins)

	// add to gauge and check
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToGauge(addCoins, gaugeID)
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...))

	// check after creating another gauge from another address
	_, _, gaugeCoins2, _ := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 1000)})

	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...))

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check gauge changes after distribution
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...).Sub(distrCoins))
}

func (suite *KeeperTestSuite) TestGetModuleDistributedCoins() {
	suite.SetupTest()

	// initial check
	coins := suite.app.IncentivesKeeper.GetModuleDistributedCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup lock and gauge
	_, gaugeID, _, startTime := suite.SetupLockAndGauge(false)

	// check after gauge creation
	coins = suite.app.IncentivesKeeper.GetModuleDistributedCoins(suite.ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 5)})

	// check after distribution
	coins = suite.app.IncentivesKeeper.GetModuleToDistributeCoins(suite.ctx)
	suite.Require().Equal(coins, distrCoins)
}

func (suite *KeeperTestSuite) TestNonPerpetualGaugeOperations() {
	// test for module get gauges
	suite.SetupTest()

	// initial module gauges check
	gauges := suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 0)

	lockOwners := suite.SetupManyLocks(5, defaultLiquidTokens, defaultLPTokens, time.Second)
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// check gauges
	gauges = suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].Id, gaugeID)
	suite.Require().Equal(gauges[0].Coins, coins)
	suite.Require().Equal(gauges[0].NumEpochsPaidOver, uint64(2))
	suite.Require().Equal(gauges[0].FilledEpochs, uint64(0))
	suite.Require().Equal(gauges[0].DistributedCoins, sdk.Coins(nil))
	suite.Require().Equal(gauges[0].StartTime.Unix(), startTime.Unix())

	// check rewards estimation
	// rewardsEst := suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
	// suite.Require().Equal(expectedCoinsPerLock.String(), rewardsEst.String())

	// add to gauge
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToGauge(addCoins, gaugeID)
	// 210 coins over 2 epochs to 5 people = 21 coins per person per epoch
	expectedCoinsPerLock := sdk.NewInt64Coin("stake", 21)

	// check gauges
	gauges = suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)
	expectedGauge := types.Gauge{
		Id:          gaugeID,
		IsPerpetual: false,
		DistributeTo: lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         "lptoken",
			Duration:      time.Second,
		},
		Coins:             coins.Add(addCoins...),
		NumEpochsPaidOver: 2,
		FilledEpochs:      0,
		DistributedCoins:  sdk.Coins{},
		StartTime:         startTime,
	}
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())

	// check upcoming gauges
	gauges = suite.app.IncentivesKeeper.GetUpcomingGauges(suite.ctx)
	suite.Require().Len(gauges, 1)

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// check upcoming gauges
	gauges = suite.app.IncentivesKeeper.GetUpcomingGauges(suite.ctx)
	suite.Require().Len(gauges, 0)

	// distribute coins to LPers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})
	suite.Require().Equal(
		expectedCoinsPerLock,
		suite.app.BankKeeper.GetBalance(suite.ctx, lockOwners[0], "stake"))

	// check active gauges
	gauges = suite.app.IncentivesKeeper.GetActiveGauges(suite.ctx)
	suite.Require().Len(gauges, 1)

	// check gauge ids by denom
	gaugeIds := suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
	suite.Require().Len(gaugeIds, 1)

	// finish distribution
	err = suite.app.IncentivesKeeper.FinishDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// check finished gauges
	gauges = suite.app.IncentivesKeeper.GetFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)

	// check gauge by ID
	gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	suite.Require().Equal(*gauge, gauges[0])

	// check invalid gauge ID
	_, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID+1000)
	suite.Require().Error(err)

	// rewardsEst = suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
	// suite.Require().Equal(sdk.Coins{}, rewardsEst)
}

func (suite *KeeperTestSuite) TestPerpetualGaugeOperations() {
	// test for module get gauges
	suite.SetupTest()

	// initial module gauges check
	gauges := suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 0)

	// setup lock and gauge
	lockOwner, gaugeID, coins, startTime := suite.SetupLockAndGauge(true)

	// check gauges
	gauges = suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)
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

	// check rewards estimation
	// rewardsEst := suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, 100)
	// suite.Require().Equal(coins.String(), rewardsEst.String())

	// check gauges
	gauges = suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)
	expectedGauge = types.Gauge{
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

	// check upcoming gauges
	gauges = suite.app.IncentivesKeeper.GetUpcomingGauges(suite.ctx)
	suite.Require().Len(gauges, 1)

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// check upcoming gauges
	gauges = suite.app.IncentivesKeeper.GetUpcomingGauges(suite.ctx)
	suite.Require().Len(gauges, 0)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	suite.Require().Equal(coins.String(),
		suite.app.BankKeeper.GetBalance(suite.ctx, lockOwner, "stake").String())

	// distributing twice without adding more for perpetual gauge
	gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{})

	suite.Require().Equal(coins.String(),
		suite.app.BankKeeper.GetBalance(suite.ctx, lockOwner, "stake").String())

	// add to gauge
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToGauge(addCoins, gaugeID)

	// distributing twice with adding more for perpetual gauge
	gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 200)})

	// check active gauges
	gauges = suite.app.IncentivesKeeper.GetActiveGauges(suite.ctx)
	suite.Require().Len(gauges, 1)

	// check gauge ids by denom
	gaugeIds := suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
	suite.Require().Len(gaugeIds, 1)

	// check finished gauges
	gauges = suite.app.IncentivesKeeper.GetFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 0)

	// check rewards estimation
	// rewardsEst = suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, 100)
	// suite.Require().Equal(sdk.Coins(nil), rewardsEst)
}

func (suite *KeeperTestSuite) TestNoLockPerpetualGaugeDistribution() {
	// test for module get gauges
	suite.SetupTest()

	// setup no lock perpetual gauge
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(true, coins)

	// check gauges
	gauges := suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)
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

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}

func (suite *KeeperTestSuite) TestNoLockNonPerpetualGaugeDistribution() {
	// test for module get gauges
	suite.SetupTest()

	// setup no lock non-perpetual gauge
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(false, coins)

	// check gauges
	gauges := suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)
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

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = suite.app.IncentivesKeeper.GetNotFinishedGauges(suite.ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}

func (suite *KeeperTestSuite) TestGaugesByDenom() {
	// TODO: This is not a good test. We should refactor it to be table driven,
	// specifying a list of gauges to define, and then the expected result of gauges by denom
	testGaugeByDenom := func(isPerpetual bool) {
		// test for module get gauges
		suite.SetupTest()

		// initial module gauges check
		gaugeIds := suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
		suite.Require().Len(gaugeIds, 0)

		// setup lock and gauge
		_, gaugeID, _, startTime := suite.SetupLockAndGauge(isPerpetual)

		// check gauges
		gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
		suite.Require().Len(gaugeIds, 1, "perpetual %b", isPerpetual)
		suite.Require().Equal(gaugeIds[0], gaugeID)

		// start distribution
		suite.ctx = suite.ctx.WithBlockTime(startTime)
		gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
		suite.Require().NoError(err)
		err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
		suite.Require().NoError(err)

		// check gauge ids by denom
		gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
		suite.Require().Len(gaugeIds, 1)

		// check gauge ids by other denom
		gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lpt")
		suite.Require().Len(gaugeIds, 0)

		// distribute coins to stakers
		_, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
		suite.Require().NoError(err)

		// finish distribution for non perpetual gauge
		if !gauge.IsPerpetual {
			err = suite.app.IncentivesKeeper.FinishDistribution(suite.ctx, *gauge)
			suite.Require().NoError(err)
		}

		expectedNumGauges := 1
		if !isPerpetual {
			expectedNumGauges = 0
		}
		// check gauge ids by denom
		gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
		suite.Require().Len(gaugeIds, expectedNumGauges)
	}

	testGaugeByDenom(true)
	testGaugeByDenom(false)
}
