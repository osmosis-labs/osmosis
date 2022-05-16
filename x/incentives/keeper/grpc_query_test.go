package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v8/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v8/x/lockup/types"
	pooltypes "github.com/osmosis-labs/osmosis/v8/x/pool-incentives/types"
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

func (suite *KeeperTestSuite) TestGRPCActiveGaugesPerDenom() {
	suite.SetupTest()

	// initial check
	res, err := suite.app.IncentivesKeeper.ActiveGaugesPerDenom(sdk.WrapSDKContext(suite.ctx), &types.ActiveGaugesPerDenomRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, gauge, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	suite.ctx = suite.ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	// final check
	res, err = suite.app.IncentivesKeeper.ActiveGaugesPerDenom(sdk.WrapSDKContext(suite.ctx), &types.ActiveGaugesPerDenomRequest{Denom: "lptoken", Pagination: nil})
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

func (suite *KeeperTestSuite) TestGRPCUpcomingGaugesPerDenom() {
	suite.SetupTest()

	upcomingGaugeRequest := types.UpcomingGaugesPerDenomRequest{Denom: "lptoken", Pagination: nil}
	// initial check, no gauges when none exist
	res, err := suite.app.IncentivesKeeper.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.ctx), &upcomingGaugeRequest)
	suite.Require().NoError(err)
	suite.Require().Len(res.UpcomingGauges, 0)

	// create a gauge, and check upcoming gauge is working
	gaugeID, gauge, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	res, err = suite.app.IncentivesKeeper.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.ctx), &upcomingGaugeRequest)
	suite.Require().NoError(err)
	suite.Require().Len(res.UpcomingGauges, 1)
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
	suite.Require().Equal(res.UpcomingGauges[0].String(), expectedGauge.String())

	// final check when gauge is moved from upcoming to active
	suite.ctx = suite.ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	res, err = suite.app.IncentivesKeeper.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.ctx), &upcomingGaugeRequest)
	suite.Require().NoError(err)
	suite.Require().Len(res.UpcomingGauges, 0)
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
