package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/x/lockup/types"
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
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 1000)})
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)

	// check after gauge creation
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(coins, res.Coins)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.DistributeAllGauges(suite.ctx)
	suite.Require().NoError(err)
	// 1000 stake over 2 epochs = 500 distributed
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 500)}, distrCoins)

	// check gauge changes after distribution
	gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	suite.Require().Equal(gauge.FilledEpochs, uint64(1))
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 500)}, gauge.DistributedCoins)

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// check after distribution
	res, err = suite.app.IncentivesKeeper.ModuleToDistributeCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleToDistributeCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins.Sub(distrCoins))

	// distribute second round to stakers
	distrCoins, err = suite.app.IncentivesKeeper.DistributeAllGauges(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 500)}, distrCoins)

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
	gaugeID, _, coins, startTime := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 1000)})
	gauge, err := suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)

	// check after gauge creation
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, sdk.Coins(nil))

	// start distribution
	suite.ctx = suite.ctx.WithBlockTime(startTime)
	err = suite.app.IncentivesKeeper.BeginDistribution(suite.ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.app.IncentivesKeeper.DistributeAllGauges(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 500)})

	// check gauge changes after distribution
	gauge, err = suite.app.IncentivesKeeper.GetGaugeByID(suite.ctx, gaugeID)
	suite.Require().NoError(err)
	suite.Require().NotNil(gauge)
	suite.Require().Equal(gauge.FilledEpochs, uint64(1))
	fmt.Println("Entering test")
	suite.Require().Equal(gauge.DistributedCoins, sdk.Coins{sdk.NewInt64Coin("stake", 500)})
	fmt.Println("Leaving test")

	// check after distribution
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, distrCoins)

	// distribute second round to stakers
	distrCoins, err = suite.app.IncentivesKeeper.DistributeAllGauges(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 500)})

	// final check
	res, err = suite.app.IncentivesKeeper.ModuleDistributedCoins(sdk.WrapSDKContext(suite.ctx), &types.ModuleDistributedCoinsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(res.Coins, coins)
}
