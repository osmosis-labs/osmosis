package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
	pooltypes "github.com/osmosis-labs/osmosis/x/pool-incentives/types"
)

func (suite *KeeperTestSuite) TestGRPCGaugeByID() {
	suite.SetupTest()

	// create a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// Ensure that a querying for a gauge with an ID that doesn't exist returns an error
	res, err := suite.app.IncentivesKeeper.GaugeByID(sdk.WrapSDKContext(suite.ctx), &types.GaugeByIDRequest{Id: 1000})
	suite.Require().Error(err)
	suite.Require().Equal(res, (*types.GaugeByIDResponse)(nil))

	// Check that querying a gauge with an ID that exists returns the gauge.
	res, err = suite.app.IncentivesKeeper.GaugeByID(sdk.WrapSDKContext(suite.ctx), &types.GaugeByIDRequest{Id: gaugeID})
	suite.Require().NoError(err)
	suite.Require().NotEqual(res.Gauge, nil)
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
	suite.Require().Equal(res.Gauge.String(), expectedGauge.String())
}

func (suite *KeeperTestSuite) TestGRPCGauges() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.Gauges(sdk.WrapSDKContext(suite.ctx), &types.GaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.app.IncentivesKeeper.Gauges(sdk.WrapSDKContext(suite.ctx), &types.GaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
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
	suite.Require().Equal(res.Data[0].String(), expectedGauge.String())
}

func (suite *KeeperTestSuite) TestGRPCActiveGauges() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ActiveGauges(sdk.WrapSDKContext(suite.ctx), &types.ActiveGaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, gauge, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	suite.ctx = suite.ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// final check
	res, err = suite.app.IncentivesKeeper.ActiveGauges(sdk.WrapSDKContext(suite.ctx), &types.ActiveGaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
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
	suite.Require().Equal(res.Data[0].String(), expectedGauge.String())
}

func (suite *KeeperTestSuite) TestGRPCUpcomingGauges() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.UpcomingGauges(sdk.WrapSDKContext(suite.ctx), &types.UpcomingGaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.app.IncentivesKeeper.UpcomingGauges(sdk.WrapSDKContext(suite.ctx), &types.UpcomingGaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
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
	suite.Require().Equal(res.Data[0].String(), expectedGauge.String())
}

func (suite *KeeperTestSuite) TestGRPCRewardsEst() {
	suite.SetupTest()

	// initial check
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	res, err := suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner: lockOwner.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// setup lock and gauge
	lockOwner, _, coins, _ := suite.SetupLockAndGauge(false)

	res, err = suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner:    lockOwner.String(),
		EndEpoch: 100,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}

func (suite *KeeperTestSuite) TestRewardsEstWithPoolIncentives() {
	suite.SetupTest()

	// initial check
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	res, err := suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner: lockOwner.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// setup lock and gauge
	lockOwner, gaugeID, coins, _ := suite.SetupLockAndGauge(true)
	distrRecord := pooltypes.DistrRecord{
		GaugeId: gaugeID,
		Weight:  sdk.NewInt(100),
	}
	err = suite.app.PoolIncentivesKeeper.ReplaceDistrRecords(suite.ctx, distrRecord)
	suite.Require().NoError(err)

	res, err = suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner:    lockOwner.String(),
		EndEpoch: 10,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	epochIdentifier := suite.app.MintKeeper.GetParams(suite.ctx).EpochIdentifier
	curEpochNumber := suite.app.EpochsKeeper.GetEpochInfo(suite.ctx, epochIdentifier).CurrentEpoch
	suite.app.EpochsKeeper.AfterEpochEnd(suite.ctx, epochIdentifier, curEpochNumber)
	// TODO: Figure out what this number should be
	mintCoins := sdk.NewCoin(coins[0].Denom, sdk.NewInt(1500000))

	res, err = suite.app.IncentivesKeeper.RewardsEst(sdk.WrapSDKContext(suite.ctx), &types.RewardsEstRequest{
		Owner:    lockOwner.String(),
		EndEpoch: 10,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins.Add(mintCoins))
}

// TODO: make this test table driven, or simpler
// I have no idea at a glance what its doing.
func (suite *KeeperTestSuite) TestGRPCToDistributeCoins() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create locks
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	gauges := []types.Gauge{*gauge}

	// check after gauge creation
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, gauges)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check gauge changes after distribution
	gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	suite.Require().Equal(gauge.FilledEpochs, uint64(1))
	suite.Require().Equal(gauge.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})
	gauges = []types.Gauge{*gauge}

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// check after distribution
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins.Sub(distrCoins))

	// distribute second round to stakers
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, gauges)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 6)}, distrCoins)

	// final check
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))
}

func (suite *KeeperTestSuite) TestGRPCDistributedCoins() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create locks
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	gauges := []types.Gauge{*gauge}

	// check after gauge creation
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.Distribute(suite.ctx, gauges)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check gauge changes after distribution
	gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	suite.Require().Equal(gauge.FilledEpochs, uint64(1))
	suite.Require().Equal(gauge.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})
	gauges = []types.Gauge{*gauge}

	// check after distribution
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, distrCoins)

	// distribute second round to stakers
	distrCoins, err = suite.app.IncentivesKeeper.Distribute(suite.ctx, gauges)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 6)}, distrCoins)

	// final check
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}

func (suite *KeeperTestSuite) TestGRPCCurrentReward() {
	suite.SetupTest()

	denom := "stake"
	duration := time.Hour
	currentReward := types.CurrentReward{
		Period:             1,
		LastProcessedEpoch: 1,
		Coin:               sdk.NewInt64Coin("stake", 100),
		Rewards:            sdk.Coins{sdk.NewInt64Coin("reward", 100)},
		Denom:              denom,
		Duration:           duration,
	}

	err := suite.app.IncentivesKeeper.SetCurrentReward(suite.ctx, currentReward, denom, duration)
	suite.Require().NoError(err)

	res, err := suite.app.IncentivesKeeper.CurrentReward(sdk.WrapSDKContext(suite.ctx), &types.CurrentRewardRequest{
		Denom:            denom,
		LockableDuration: duration,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(currentReward.Period, res.Period)
	suite.Require().Equal(currentReward.LastProcessedEpoch, res.LastProcessedEpoch)
	suite.Require().Equal(currentReward.Coin, res.Coin)
	suite.Require().Equal(currentReward.Rewards, res.Reward)
}

func (suite *KeeperTestSuite) TestGRPCHistoricalReward() {
	suite.SetupTest()

	denom := "stake"
	duration := time.Hour
	historicalReward := types.HistoricalReward{
		Period:                1,
		CumulativeRewardRatio: sdk.NewDecCoinsFromCoins(sdk.NewInt64Coin("reward", 100)),
		LastEligibleEpoch:     1,
	}

	err := suite.app.IncentivesKeeper.AddHistoricalReward(suite.ctx, historicalReward, denom, duration, 1, 1)
	suite.Require().NoError(err)

	res, err := suite.app.IncentivesKeeper.HistoricalReward(sdk.WrapSDKContext(suite.ctx), &types.HistoricalRewardRequest{
		Denom:            denom,
		LockableDuration: duration,
		Period:           1,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(historicalReward.CumulativeRewardRatio, res.CumulativeRewardRatio)
	suite.Require().Equal(historicalReward.Period, res.Period)
	suite.Require().Equal(historicalReward.LastEligibleEpoch, res.LastEligibleEpoch)
}

func (suite *KeeperTestSuite) TestGRPCPeriodLockReward() {
	suite.SetupTest()

	periodLockReward := types.PeriodLockReward{
		ID:      1,
		Period:  make(map[string]uint64),
		Rewards: sdk.NewCoins(sdk.NewInt64Coin("reward", 100)),
	}
	periodLockReward.Period["stake/1h"] = 1

	err := suite.app.IncentivesKeeper.SetPeriodLockReward(suite.ctx, periodLockReward)
	suite.Require().NoError(err)

	res, err := suite.app.IncentivesKeeper.PeriodLockReward(sdk.WrapSDKContext(suite.ctx), &types.PeriodLockRewardRequest{
		Id: 1,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(periodLockReward.ID, res.ID)
	suite.Require().Equal(periodLockReward.Period, res.Period)
	suite.Require().Equal(periodLockReward.Rewards, res.Rewards)
}

func (suite *KeeperTestSuite) TestGRPCRewards() {
	suite.SetupTest()

	denom := "stake"
	duration := time.Second
	currentReward := types.CurrentReward{
		Period:             1,
		LastProcessedEpoch: 1,
		Coin:               sdk.NewInt64Coin("lptoken", 100),
		Rewards:            sdk.Coins{sdk.NewInt64Coin("reward", 100)},
		Denom:              denom,
		Duration:           duration,
	}

	err := suite.app.IncentivesKeeper.SetCurrentReward(suite.ctx, currentReward, denom, duration)
	suite.Require().NoError(err)

	periodLockReward := types.PeriodLockReward{
		ID:      1,
		Period:  make(map[string]uint64),
		Rewards: sdk.NewCoins(sdk.NewInt64Coin("reward", 100)),
	}
	periodLockReward.Period["lptoken/1s"] = 1

	err = suite.app.IncentivesKeeper.SetPeriodLockReward(suite.ctx, periodLockReward)
	suite.Require().NoError(err)

	// setup lock and gauge
	lockOwner, _, _, _ := suite.SetupLockAndGauge(false)
	res, err := suite.app.IncentivesKeeper.Rewards(sdk.WrapSDKContext(suite.ctx), &types.RewardsRequest{
		Owner:   lockOwner.String(),
		LockIds: []uint64{1},
	})
	suite.Require().NoError(err)
	suite.Require().Equal(periodLockReward.Rewards, res.Coins)
}
