package keeper_test

import (
	// "fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"
	pooltypes "github.com/osmosis-labs/osmosis/v7/x/pool-incentives/types"
	query "github.com/cosmos/cosmos-sdk/types/query"
)

func (suite *KeeperTestSuite) TestGRPCGaugeByID() {
	suite.SetupTest()

	// create a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// Ensure that a querying for a gauge with an ID that doesn't exist returns an error
	res, err := suite.querier.GaugeByID(sdk.WrapSDKContext(suite.Ctx), &types.GaugeByIDRequest{Id: 1000})
	suite.Require().Error(err)
	suite.Require().Equal(res, (*types.GaugeByIDResponse)(nil))

	// Check that querying a gauge with an ID that exists returns the gauge.
	res, err = suite.querier.GaugeByID(sdk.WrapSDKContext(suite.Ctx), &types.GaugeByIDRequest{Id: gaugeID})
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
	res, err := suite.querier.Gauges(sdk.WrapSDKContext(suite.Ctx), &types.GaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.querier.Gauges(sdk.WrapSDKContext(suite.Ctx), &types.GaugesRequest{})
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

	// filtering check
	for i := 0; i < 10; i++ {
		suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))
	}

	filter := query.PageRequest{Limit: 10}
	res, err = suite.querier.Gauges(sdk.WrapSDKContext(suite.Ctx), &types.GaugesRequest{Pagination: &filter})
	suite.Require().Len(res.Data, 10)
}

func (suite *KeeperTestSuite) TestGRPCActiveGauges() {
	suite.SetupTest()

	// initial check
	res, err := suite.querier.ActiveGauges(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, gauge, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.querier.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// final check
	res, err = suite.querier.ActiveGauges(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesRequest{})
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

	// filtering check 
	for i := 0; i < 20; i++ {
		_, gauge, _, _ := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// set up more 9 active gauges => 10
		if i < 9 {
			suite.querier.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
		}
	}

	res, err = suite.querier.ActiveGauges(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesRequest{Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.Data, 5)

	res, err = suite.querier.ActiveGauges(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesRequest{Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().Len(res.Data, 10)
}

func (suite *KeeperTestSuite) TestGRPCActiveGaugesPerDenom() {
	suite.SetupTest()

	// initial check
	res, err := suite.querier.ActiveGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesPerDenomRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, gauge, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	// final check
	res, err = suite.querier.ActiveGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesPerDenomRequest{Denom: "lptoken", Pagination: nil})
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

	// filtering check 
	for i := 0; i < 20; i++ {
		_, gauge, _, _ := suite.SetupNewGaugeWithDenom(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)}, "pool")
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// set up 10 active gauges with "pool" denom
		if i < 10 {
			suite.querier.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
		}
	}

	res, err = suite.querier.ActiveGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesPerDenomRequest{Denom: "lptoken", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.Data, 1)

	res, err = suite.querier.ActiveGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.Data, 5)

	res, err = suite.querier.ActiveGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.ActiveGaugesPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().Len(res.Data, 10)
}

func (suite *KeeperTestSuite) TestGRPCUpcomingGauges() {
	suite.SetupTest()

	// initial check
	res, err := suite.querier.UpcomingGauges(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingGaugesRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})

	// final check
	res, err = suite.querier.UpcomingGauges(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingGaugesRequest{})
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

	// filtering check 
	for i := 0; i < 20; i++ {
		_, gauge, _, _ := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)})
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// set up more 9 active gauges => Upcoming = 1 + (20 -9) = 12
		if i < 9 {
			suite.querier.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
		}
	}

	res, err = suite.querier.UpcomingGauges(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingGaugesRequest{Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.Data, 5)

	res, err = suite.querier.UpcomingGauges(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingGaugesRequest{Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().Len(res.Data, 12)
}

func (suite *KeeperTestSuite) TestGRPCUpcomingGaugesPerDenom() {
	suite.SetupTest()

	upcomingGaugeRequest := types.UpcomingGaugesPerDenomRequest{Denom: "lptoken", Pagination: nil}
	// initial check, no gauges when none exist
	res, err := suite.querier.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &upcomingGaugeRequest)
	suite.Require().NoError(err)
	suite.Require().Len(res.UpcomingGauges, 0)

	// create a gauge, and check upcoming gauge is working
	gaugeID, gauge, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	res, err = suite.querier.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &upcomingGaugeRequest)
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
	suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	res, err = suite.querier.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &upcomingGaugeRequest)
	suite.Require().NoError(err)
	suite.Require().Len(res.UpcomingGauges, 0)

	// filtering check 
	for i := 0; i < 20; i++ {
		_, gauge, _, _ := suite.SetupNewGaugeWithDenom(false, sdk.Coins{sdk.NewInt64Coin("stake", 3)}, "pool")
		suite.Ctx = suite.Ctx.WithBlockTime(startTime.Add(time.Second))

		// set up 10 active gauges with "pool" denom => 10 upcoming
		if i < 10 {
			suite.querier.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
		}
	}

	res, err = suite.querier.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingGaugesPerDenomRequest{Denom: "lptoken", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.UpcomingGauges, 0)

	res, err = suite.querier.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingGaugesPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 5}})
	suite.Require().Len(res.UpcomingGauges, 5)

	res, err = suite.querier.UpcomingGaugesPerDenom(sdk.WrapSDKContext(suite.Ctx), &types.UpcomingGaugesPerDenomRequest{Denom: "pool", Pagination: &query.PageRequest{Limit: 15}})
	suite.Require().Len(res.UpcomingGauges, 10)
}

func (suite *KeeperTestSuite) TestGRPCRewardsEst() {
	suite.SetupTest()

	// initial check
	lockOwner := sdk.AccAddress([]byte("addr1---------------"))
	res, err := suite.querier.RewardsEst(sdk.WrapSDKContext(suite.Ctx), &types.RewardsEstRequest{
		Owner: lockOwner.String(),
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins{})

	// setup lock and gauge
	lockOwner, _, coins, _ := suite.SetupLockAndGauge(false)

	res, err = suite.querier.RewardsEst(sdk.WrapSDKContext(suite.Ctx), &types.RewardsEstRequest{
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
	res, err := suite.querier.RewardsEst(sdk.WrapSDKContext(suite.Ctx), &types.RewardsEstRequest{
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
	err = suite.App.PoolIncentivesKeeper.ReplaceDistrRecords(suite.Ctx, distrRecord)
	suite.Require().NoError(err)

	res, err = suite.querier.RewardsEst(sdk.WrapSDKContext(suite.Ctx), &types.RewardsEstRequest{
		Owner:    lockOwner.String(),
		EndEpoch: 10,
	})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	epochIdentifier := suite.App.MintKeeper.GetParams(suite.Ctx).EpochIdentifier
	curEpochNumber := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, epochIdentifier).CurrentEpoch
	suite.App.EpochsKeeper.AfterEpochEnd(suite.Ctx, epochIdentifier, curEpochNumber)
	// TODO: Figure out what this number should be
	mintCoins := sdk.NewCoin(coins[0].Denom, sdk.NewInt(1500000))

	res, err = suite.querier.RewardsEst(sdk.WrapSDKContext(suite.Ctx), &types.RewardsEstRequest{
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
	res, err := suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create locks
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	gauge, err := suite.querier.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	gauges := []types.Gauge{*gauge}

	// check after gauge creation
	res, err = suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)

	// distribute coins to stakers
	distrCoins, err := suite.querier.Distribute(suite.Ctx, gauges)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check gauge changes after distribution
	gauge, err = suite.querier.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	suite.Require().Equal(gauge.FilledEpochs, uint64(1))
	suite.Require().Equal(gauge.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})
	gauges = []types.Gauge{*gauge}

	// start distribution
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	err = suite.querier.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// check after distribution
	res, err = suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins.Sub(distrCoins))

	// distribute second round to stakers
	distrCoins, err = suite.querier.Distribute(suite.Ctx, gauges)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 6)}, distrCoins)

	// final check
	res, err = suite.querier.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))
}

func (suite *KeeperTestSuite) TestGRPCDistributedCoins() {
	suite.SetupTest()

	// initial check
	res, err := suite.querier.ModuleDistributedCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// create locks
	addr1 := sdk.AccAddress([]byte("addr1---------------"))
	addr2 := sdk.AccAddress([]byte("addr2---------------"))
	suite.LockTokens(addr1, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, time.Second)
	suite.LockTokens(addr2, sdk.Coins{sdk.NewInt64Coin("lptoken", 10)}, 2*time.Second)

	// setup a gauge
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 10)})
	gauge, err := suite.querier.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	gauges := []types.Gauge{*gauge}

	// check after gauge creation
	res, err = suite.querier.ModuleDistributedCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// start distribution
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	err = suite.querier.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.querier.Distribute(suite.Ctx, gauges)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})

	// check gauge changes after distribution
	gauge, err = suite.querier.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	suite.Require().Equal(gauge.FilledEpochs, uint64(1))
	suite.Require().Equal(gauge.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 4)})
	gauges = []types.Gauge{*gauge}

	// check after distribution
	res, err = suite.querier.ModuleDistributedCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, distrCoins)

	// distribute second round to stakers
	distrCoins, err = suite.querier.Distribute(suite.Ctx, gauges)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 6)}, distrCoins)

	// final check
	res, err = suite.querier.ModuleDistributedCoins(sdk.WrapSDKContext(suite.Ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}
