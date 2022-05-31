package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v9/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v9/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TestDistribute tests that when the distribute command is executed on
// a provided gauge,
func (suite *KeeperTestSuite) TestDistribute() {
	twoLockupUser := userLocks{
		lockDurations: []time.Duration{defaultLockDuration, 2 * defaultLockDuration},
		lockAmounts:   []sdk.Coins{defaultLPTokens, defaultLPTokens},
	}
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
		users           []userLocks
		gauges          []perpGaugeDesc
		expectedRewards []sdk.Coins
	}{
		// gauge 1 gives 3k coins. Three locks, all eligible. 1k coins per lock
		// so 1k to oneLockupUser, 2k to twoLockupUser
		{
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
		// gauge 1 gives 3k coins. Three locks, all eligible.
		// gauge 2 gives 3k coins to one lock, in twoLockupUser
		// so 1k to oneLockupUser, 5k to twoLockupUser
		{
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{defaultGauge, doubleLengthGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, fiveKRewardCoins},
		},
		{
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge},
			expectedRewards: []sdk.Coins{noRewardCoins, noRewardCoins},
		},
		{
			users:           []userLocks{oneLockupUser, twoLockupUser},
			gauges:          []perpGaugeDesc{noRewardGauge, defaultGauge},
			expectedRewards: []sdk.Coins{oneKRewardCoins, twoKRewardCoins},
		},
	}
	for tcIndex, tc := range tests {
		suite.SetupTest()
		gauges := suite.SetupGauges(tc.gauges)
		addrs := suite.SetupUserLocks(tc.users)
		_, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, gauges)
		suite.Require().NoError(err)
		// Check expected rewards
		for i, addr := range addrs {
			bal := suite.App.BankKeeper.GetAllBalances(suite.Ctx, addr)
			suite.Require().Equal(tc.expectedRewards[i].String(), bal.String(), "tcnum %d, person %d", tcIndex, i)
		}
	}

	// TODO: test distribution for synthetic lockup as well
}

// TODO: Make this test table driven, or move whatever it tests into
// the much simpler TestDistribute
func (suite *KeeperTestSuite) TestGetModuleToDistributeCoins() {
	// test for module get gauges
	suite.SetupTest()

	// initial check
	coins := suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup lock and gauge
	_, gaugeID, gaugeCoins, startTime := suite.SetupLockAndGauge(false)

	// check after gauge creation
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins)

	// add to gauge and check
	addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
	suite.AddToGauge(addCoins, gaugeID)
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...))

	// check after creating another gauge from another address
	_, _, gaugeCoins2, _ := suite.SetupNewGauge(false, sdk.Coins{sdk.NewInt64Coin("stake", 1000)})

	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...))

	// start distribution
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 105)})

	// check gauge changes after distribution
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, gaugeCoins.Add(addCoins...).Add(gaugeCoins2...).Sub(distrCoins))
}

// TODO: Make this test table driven, or move whatever it tests into
// the much simpler TestDistribute
func (suite *KeeperTestSuite) TestGetModuleDistributedCoins() {
	suite.SetupTest()

	// initial check
	coins := suite.App.IncentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// setup lock and gauge
	_, gaugeID, _, startTime := suite.SetupLockAndGauge(false)

	// check after gauge creation
	coins = suite.App.IncentivesKeeper.GetModuleDistributedCoins(suite.Ctx)
	suite.Require().Equal(coins, sdk.Coins(nil))

	// start distribution
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins{sdk.NewInt64Coin("stake", 5)})

	// check after distribution
	coins = suite.App.IncentivesKeeper.GetModuleToDistributeCoins(suite.Ctx)
	suite.Require().Equal(coins, distrCoins)
}

func (suite *KeeperTestSuite) TestNoLockPerpetualGaugeDistribution() {
	// test for module get gauges
	suite.SetupTest()

	// setup no lock perpetual gauge
	coins := sdk.Coins{sdk.NewInt64Coin("stake", 10)}
	gaugeID, _, _, startTime := suite.SetupNewGauge(true, coins)

	// check gauges
	gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
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
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
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
	gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
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
	suite.Ctx = suite.Ctx.WithBlockTime(startTime)
	gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
	suite.Require().NoError(err)
	err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
	suite.Require().NoError(err)

	// distribute coins to stakers, since it's perpetual distribute everything on single distribution
	distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
	suite.Require().NoError(err)
	suite.Require().Equal(distrCoins, sdk.Coins(nil))

	// check state is same after distribution
	gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
	suite.Require().Len(gauges, 1)
	suite.Require().Equal(gauges[0].String(), expectedGauge.String())
}
