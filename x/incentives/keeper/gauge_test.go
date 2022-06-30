package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Tests error handling for creating a gauge with an invalid duration.
func (suite *KeeperTestSuite) TestInvalidDurationGaugeCreationValidation() {
	suite.SetupTest()

	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration / 2, // 0.5 second, invalid duration
	}
	_, err := suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	distrTo.Duration = defaultLockDuration
	_, err = suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

// Tests error handling for creating a gauge with an invalid denom.
func (suite *KeeperTestSuite) TestNonExistentDenomGaugeCreation() {
	suite.SetupTest()

	addrNoSupply := sdk.AccAddress([]byte("Gauge_Creation_Addr_"))
	addrs := suite.SetupManyLocks(1, defaultLiquidTokens, defaultLPTokens, defaultLockDuration)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         defaultLPDenom,
		Duration:      defaultLockDuration,
	}
	_, err := suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrNoSupply, defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	_, err = suite.App.IncentivesKeeper.CreateGauge(suite.Ctx, false, addrs[0], defaultLiquidTokens, distrTo, time.Time{}, 1)
	suite.Require().NoError(err)
}

// TODO: Make this test table driven
// OR if it needs to be script based,
// remove lots of boilerplate so this can actually be followed
// TODO: Make issue for this test

// Tests non perpetual gauge distribution logic.
func (suite *KeeperTestSuite) TestNonPerpetualGaugeOperations() {
	// test for module get gauges
	suite.SetupTest()

	// initial module gauges check
	gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 0)

	lockOwners := suite.SetupManyLocks(5, defaultLiquidTokens, defaultLPTokens, time.Second)
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	// evenly distributed per lock
	expectedCoinsPerLock := sdk.Coins{sdk.NewInt64Coin("stake", 2)}

	// check gauges
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].Id, gaugeID)
	suite.Require().Equal(gauges[0].Coins, coins)
	suite.Require().Equal(gauges[0].NumEpochsPaidOver, uint64(2))
	suite.Require().Equal(gauges[0].FilledEpochs, uint64(0))
	suite.Require().Equal(gauges[0].DistributedCoins, sdk.Coins(nil))
	suite.Require().Equal(gauges[0].StartTime.Unix(), startTime.Unix())

	// check rewards estimation
	rewardsEst := suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
	suite.Require().Equal(expectedCoinsPerLock.String(), rewardsEst.String())

	// add to gauge
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToGauge(addCoins, gaugeID)

	// check gauges
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
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
	gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// start distribution
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// check upcoming gauges
	gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
	suite.Require().Len(gauges, 0)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check active gauges
	gauges = suite.App.IncentivesKeeper.GetActiveGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// check gauge ids by denom
	gaugeIds := suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
	suite.Require().Len(gaugeIds, 1)

	// finish distribution
	err = suite.App.IncentivesKeeper.MoveActiveGaugeToFinishedGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// check finished gauges
	gauges = suite.App.IncentivesKeeper.GetFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// check gauge by ID
	gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	suite.Require().Equal(*gauge, gauges[0])

	// check invalid gauge ID
	_, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID+1000)
	suite.Require().Error(err)

	rewardsEst = suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
	suite.Require().Equal(sdk.Coins{}, rewardsEst)
}

// Tests perpetual gauge distribution logic.
func (suite *KeeperTestSuite) TestPerpetualGaugeOperations() {
	// test for module get gauges
	suite.SetupTest()

	// initial module gauges check
	gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 0)

	// setup lock and gauge
	lockOwner, gaugeID, coins, startTime := suite.SetupLockAndGauge(true)

	// check gauges
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
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
	rewardsEst := suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwner, []lockuptypes.PeriodLock{}, 100)
	suite.Require().Equal(coins.String(), rewardsEst.String())

	// check gauges
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
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
	gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// start distribution
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// check upcoming gauges
	gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
	suite.Require().Len(gauges, 0)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// distributing twice without adding more for perpetual gauge
	gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	distrCoins, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().True(distrCoins.Empty())

	// add to gauge
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToGauge(addCoins, gaugeID)

	// distributing twice with adding more for perpetual gauge
	gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	distrCoins, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 200)})

	// check active gauges
	gauges = suite.App.IncentivesKeeper.GetActiveGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)

	// check gauge ids by denom
	gaugeIds := suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
	suite.Require().Len(gaugeIds, 1)

	// check finished gauges
	gauges = suite.App.IncentivesKeeper.GetFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 0)

	// check rewards estimation
	rewardsEst = suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwner, []lockuptypes.PeriodLock{}, 100)
	suite.Require().Equal(sdk.Coins(nil), rewardsEst)
}

// Tests gauge distribution using the gauges by denom keeper.
func (suite *KeeperTestSuite) TestGaugesByDenom() {
	// TODO: This is not a good test. We should refactor it to be table driven,
	// specifying a list of gauges to define, and then the expected result of gauges by denom
	// TODO: Make issue for this.
	testGaugeByDenom := func(isPerpetual bool) {
		// test for module get gauges
		suite.SetupTest()

		// initial module gauges check
		gaugeIds := suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIds, 0)

		// setup lock and gauge
		_, gaugeID, _, startTime := suite.SetupLockAndGauge(isPerpetual)

		// check gauges
		gaugeIds = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIds, 1, "perpetual %b", isPerpetual)
		suite.Require().Equal(gaugeIds[0], gaugeID)

		// start distribution
		suite.Ctx = suite.Ctx.WithBlockTime(startTime)
		gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
		suite.Require().NoError(err)
		err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
		suite.Require().NoError(err)

		// check gauge ids by denom
		gaugeIds = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIds, 1)

		// check gauge ids by other denom
		gaugeIds = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lpt")
		suite.Require().Len(gaugeIds, 0)

		// distribute coins to stakers
		_, err = suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
		suite.Require().NoError(err)

		// finish distribution for non perpetual gauge
		if !gauge.IsPerpetual {
			err = suite.App.IncentivesKeeper.MoveActiveGaugeToFinishedGauge(suite.Ctx, *gauge)
			suite.Require().NoError(err)
		}

		expectedNumGauges := 1
		if !isPerpetual {
			expectedNumGauges = 0
		}
		// check gauge ids by denom
		gaugeIds = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIds, expectedNumGauges)
	}

	testGaugeByDenom(true)
	testGaugeByDenom(false)
}
