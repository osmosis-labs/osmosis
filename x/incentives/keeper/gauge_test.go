package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	epochtypes "github.com/osmosis-labs/osmosis/x/epochs/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
)

func (suite *KeeperTestSuite) TestInvalidDurationGaugeCreationValidation() {
	suite.SetupTest()

	addr := sdk.AccAddress([]byte("addr1---------------"))
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	suite.app.BankKeeper.SetBalances(suite.ctx, addr, coins)
	distrTo := lockuptypes.QueryCondition{
		LockQueryType: lockuptypes.ByDuration,
		Denom:         "lptoken",
		Duration:      time.Second / 2, // 0.5 second
	}
	_, err := suite.app.IncentivesKeeper.CreateGauge(suite.ctx, false, addr, coins, distrTo, time.Time{}, 1)
	suite.Require().Error(err)

	distrTo.Duration = time.Second
	_, err = suite.app.IncentivesKeeper.CreateGauge(suite.ctx, false, addr, coins, distrTo, time.Time{}, 1)
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

	// setup lock and gauge
	lockOwner, gaugeID, coins, startTime := suite.SetupLockAndGauge(false)

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
	rewardsEst := suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, 100)
	suite.Require().Equal(coins.String(), rewardsEst.String())

	// add to gauge
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToGauge(addCoins, gaugeID)

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

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

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

	// check rewards estimation
	rewardsEst = suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, 100)
	suite.Require().Equal(sdk.Coins{}, rewardsEst)
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
	rewardsEst := suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, 100)
	suite.Require().Equal(coins.String(), rewardsEst.String())

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

	// distributing twice without adding more for perpetual gauge
	gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)
	suite.Require().True(distrCoins.Empty())

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
	rewardsEst = suite.app.IncentivesKeeper.GetRewardsEst(suite.ctx, lockOwner, []lockuptypes.PeriodLock{}, 100)
	suite.Require().Equal(sdk.Coins(nil), rewardsEst)
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

func (suite *KeeperTestSuite) TestNonPerpetualActiveGaugesByDenom() {
	// test for module get gauges
	suite.SetupTest()

	// initial module gauges check
	gaugeIds := suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
	suite.Require().Len(gaugeIds, 0)

	// setup lock and gauge
	_, gaugeID, _, startTime := suite.SetupLockAndGauge(false)

	// check gauges
	gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
	suite.Require().Len(gaugeIds, 1)
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

	// distribute coins to stakers
	_, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// finish distribution
	err = suite.app.IncentivesKeeper.FinishDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// check gauge ids by denom
	gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
	suite.Require().Len(gaugeIds, 0)
}

func (suite *KeeperTestSuite) TestPerpetualActiveGaugesByDenom() {
	// test for module get gauges
	suite.SetupTest()

	// initial module gauges check
	gaugeIds := suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
	suite.Require().Len(gaugeIds, 0)

	// setup lock and gauge
	_, gaugeID, _, startTime := suite.SetupLockAndGauge(true)

	// check gauges
	gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
	suite.Require().Len(gaugeIds, 1)
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

	// check gauge ids by other denom
	gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "token")
	suite.Require().Len(gaugeIds, 0)

	// distribute coins to stakers
	_, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// check gauge ids by denom
	gaugeIds = suite.app.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.ctx, "lptoken")
	suite.Require().Len(gaugeIds, 1)
}

func (suite *KeeperTestSuite) TestComplexScenarioGauge() {
	suite.SetupTest()

	durations := []time.Duration{
		time.Second * 5,  // 5 secs
		time.Second * 10, // 10 secs
		time.Second * 15, // 15 secs
	}

	createGauge := func(addr sdk.AccAddress, coins sdk.Coins, denom string, duration time.Duration, epochs uint64) (uint64, *types.Gauge) {
		distrTo := lockuptypes.QueryCondition{
			LockQueryType: lockuptypes.ByDuration,
			Denom:         denom,
			Duration:      duration,
		}
		return suite.CreateGauge(epochs == 1, addr, coins, distrTo, time.Time{}, epochs)
	}

	suite.app.IncentivesKeeper.SetLockableDurations(suite.ctx, durations)

	// epoch == 10 secs
	epochDur := time.Second * 10
	suite.app.IncentivesKeeper.SetParams(suite.ctx, types.NewParams("testepoch"))
	epochInfo := epochtypes.EpochInfo{
		Identifier:            "testepoch",
		StartTime:             time.Time{},
		Duration:              epochDur,
		CurrentEpoch:          0,
		CurrentEpochStartTime: time.Time{},
		EpochCountingStarted:  false,
		CurrentEpochEnded:     true,
	}
	suite.app.EpochsKeeper.SetEpochInfo(suite.ctx, epochInfo)

	blockTime := time.Second * 5

	isEpochEnd := func(blockN int) bool {
		return (time.Duration(blockN)*blockTime)%epochDur == 0
	}

	nextBlock := func() {
		suite.ctx = suite.ctx.WithBlockTime(suite.ctx.BlockTime().Add(blockTime))
	}

	creator := sdk.AccAddress([]byte("addr1---------------"))
	lockdenom1 := "denom1"
	lockdenom2 := "denom2"
	rewarddenom := "reward"
	address1 := sdk.AccAddress([]byte("testaddr00----------"))
	address2 := sdk.AccAddress([]byte("testaddr01----------"))

	lock1 := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(lockdenom1, amt) }
	lock2 := func(amt int64) sdk.Coin { return sdk.NewInt64Coin(lockdenom2, amt) }
	reward := func(amt int64) sdk.Coins { return sdk.NewCoins(sdk.NewInt64Coin(rewarddenom, amt)) }

	// denom := func(i int) string { return []string{lockdenom1, lockdenom2}[i%2] }
	// epoch := func(i int) uint64 { return []uint64{10, 1}[i%4/2] }
	gauges := make([]uint64, len(durations)*4)
	for i, duration := range durations {
		gauges[i*4+0], _ = createGauge(creator, reward(100), lockdenom1, duration, 10) // non perpetual, denom 1
		gauges[i*4+1], _ = createGauge(creator, reward(200), lockdenom2, duration, 10) // non perpetual, denom 2
		gauges[i*4+2], _ = createGauge(creator, reward(100), lockdenom1, duration, 1)  // perpetual, denom 1
		gauges[i*4+3], _ = createGauge(creator, reward(200), lockdenom2, duration, 1)  // perpetual, denom 2
	}

	blockN := 0

	// no lockups on first 4 blocks == 2 epochs
	for ; blockN < 4; blockN++ {
		if isEpochEnd(blockN) {
			for _, gaugeID := range gauges {
				gauge, _ := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
				distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
				suite.Require().NoError(err)
				suite.Require().True(distrCoins.Empty())
			}
		}
		for _, gaugeID := range gauges {
			gauge, _ := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
			suite.Require().True(gauge.DistributedCoins.Empty())
		}

		nextBlock()
	}

	// add locksups for account 1, denom 1, duraiton 10 secs
	suite.LockTokens(address1, sdk.NewCoins(lock1(1)), time.Second*10)

	// single lockup on next 4 blocks == 2 epochs
	for ; blockN < 8; blockN++ {
		if isEpochEnd(blockN) {
			for _, gaugeID := range gauges {
				gauge, _ := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
				_, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
				suite.Require().NoError(err)
			}
		}

		nextBlock()
	}

	// address 1, each epoch:
	// reward1(10) by gauge 1
	// reward1(10) by gauge 5
	// address 1, one time
	// reward2(100) by gauge 3
	// reward2(100) by gauge 7
	suite.Require().True(reward(240).IsEqual(
		suite.app.BankKeeper.GetAllBalances(suite.ctx, address1),
	))

	// another lockup for account 2, denom 1, duration 6 secs
	suite.LockTokens(address2, sdk.NewCoins(lock1(2)), time.Second*6)

	// two lockups on next 6 blocks == 3 epochs
	for ; blockN < 14; blockN++ {
		if isEpochEnd(blockN) {
			for _, gaugeID := range gauges {
				gauge, _ := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
				_, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
				suite.Require().NoError(err)
			}
		}

		nextBlock()
	}

	// address 1, each epoch:
	// reward(3) by gauge 1
	// reward(10) by gauge 5
	suite.Require().True(reward(240 + 39).IsEqual(
		suite.app.BankKeeper.GetAllBalances(suite.ctx, address1),
	))

	// address 2, each epoch:
	// reward(7) by gauge 1
	suite.Require().True(reward(0 + 21).IsEqual(
		suite.app.BankKeeper.GetAllBalances(suite.ctx, address2),
	))

	// lockups for address 1, denom 1/2, duration 16 secs
	suite.LockTokens(address1, sdk.NewCoins(lock1(2), lock2(1)), time.Second*16)

	// four lockups on next 4 blocks == 2 epochs
	for ; blockN < 18; blockN++ {
		if isEpochEnd(blockN) {
			for _, gaugeID := range gauges {
				gauge, _ := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
				_, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, *gauge)
				suite.Require().NoError(err)
			}
		}

		nextBlock()
	}

	// address 1, each epoch:
	// reward(6) by gauge 1
	// reward(20) by gauge 2
	// reward(10) by gauge 5
	// reward(20) by gauge 6
	// reward(10) by gauge 9
	// reward(20) by gauge 10
	// addres 1, one time:
	// reward(200) by gauge 4
	// reward(200) by gauge 8
	// reward(100) by gauge 11
	// reward(200) by gauge 12

	suite.Require().True(reward(240 + 39 + 872).IsEqual(
		suite.app.BankKeeper.GetAllBalances(suite.ctx, address1),
	))

	// address 2, each epoch:
	// reward(4) by gauge 1
	suite.Require().True(reward(0 + 21 + 8).IsEqual(
		suite.app.BankKeeper.GetAllBalances(suite.ctx, address2),
	))

}
