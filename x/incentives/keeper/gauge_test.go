package keeper_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v7/x/incentives/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v7/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func (suite *KeeperTestSuite) TestGaugeOperations() {
	testCases := []struct {
		isPerpetual bool
		numLocks    int
	}{
		{
			isPerpetual: true,
			numLocks:    1,
		},
		{
			isPerpetual: false,
			numLocks:    1,
		},
		{
			isPerpetual: true,
			numLocks:    2,
		},
		{
			isPerpetual: false,
			numLocks:    2,
		},
	}
	for _, tc := range testCases {
		// test for module get gauges
		suite.SetupTest()

		// initial module gauges check
		gauges := suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 0)
		gaugeIdsByDenom := suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 0)

		// setup lock and gauge
		lockOwners := suite.SetupManyLocks(tc.numLocks, defaultLiquidTokens, defaultLPTokens, time.Second)
		gaugeID, _, coins, startTime := suite.SetupNewGauge(tc.isPerpetual, sdk.Coins{sdk.NewInt64Coin("stake", 12)})
		// evenly distributed per lock
		expectedCoinsPerLock := sdk.Coins{sdk.NewInt64Coin("stake", 12/int64(tc.numLocks))}
		// set expected epochs
		var expectedNumEpochsPaidOver int
		if tc.isPerpetual {
			expectedNumEpochsPaidOver = 1
		} else {
			expectedNumEpochsPaidOver = 2
		}

		// check gauges
		gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)
		expectedGauge := types.Gauge{
			Id:          gaugeID,
			IsPerpetual: tc.isPerpetual,
			DistributeTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Denom:         "lptoken",
				Duration:      time.Second,
			},
			Coins:             coins,
			NumEpochsPaidOver: uint64(expectedNumEpochsPaidOver),
			FilledEpochs:      0,
			DistributedCoins:  sdk.Coins{},
			StartTime:         startTime,
		}
		suite.Require().Equal(expectedGauge.String(), gauges[0].String())

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)
		suite.Require().Equal(gaugeID, gaugeIdsByDenom[0])

		// check rewards estimation
		rewardsEst := suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
		suite.Require().Equal(expectedCoinsPerLock.String(), rewardsEst.String())

		// check gauges
		gauges = suite.App.IncentivesKeeper.GetNotFinishedGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)
		suite.Require().Equal(expectedGauge.String(), gauges[0].String())

		// check upcoming gauges
		gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// start distribution
		suite.Ctx = suite.Ctx.WithBlockTime(startTime)
		gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
		suite.Require().NoError(err)
		err = suite.App.IncentivesKeeper.MoveUpcomingGaugeToActiveGauge(suite.Ctx, *gauge)
		suite.Require().NoError(err)

		// check active gauges
		gauges = suite.App.IncentivesKeeper.GetActiveGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// check upcoming gauges
		gauges = suite.App.IncentivesKeeper.GetUpcomingGauges(suite.Ctx)
		suite.Require().Len(gauges, 0)

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)

		// check gauge ids by other denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lpt")
		suite.Require().Len(gaugeIdsByDenom, 0)

		// distribute coins to stakers
		distrCoins, err := suite.App.IncentivesKeeper.Distribute(suite.Ctx, []types.Gauge{*gauge})
		suite.Require().NoError(err)
		// We hardcoded 12 "stake" tokens when initializing gauge
		suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", int64(12/expectedNumEpochsPaidOver))}, distrCoins)

		if tc.isPerpetual {
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
			suite.Require().Equal(sdk.Coins{sdk.NewInt64Coin("stake", 200)}, distrCoins)
		} else {
			// add to gauge
			addCoins := sdk.Coins{sdk.NewInt64Coin("stake", 200)}
			suite.AddToGauge(addCoins, gaugeID)
		}

		// check active gauges
		gauges = suite.App.IncentivesKeeper.GetActiveGauges(suite.Ctx)
		suite.Require().Len(gauges, 1)

		// check gauge ids by denom
		gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
		suite.Require().Len(gaugeIdsByDenom, 1)

		// finish distribution for non perpetual gauge
		if !tc.isPerpetual {
			err = suite.App.IncentivesKeeper.MoveActiveGaugeToFinishedGauge(suite.Ctx, *gauge)
			suite.Require().NoError(err)
		}

		// check non-perpetual gauges (finished + rewards estimate empty)
		if !tc.isPerpetual {

			// check finished gauges
			gauges = suite.App.IncentivesKeeper.GetFinishedGauges(suite.Ctx)
			suite.Require().Len(gauges, 1)

			// check gauge by ID
			gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID)
			suite.Require().NoError(err)
			suite.Require().NotNil(gauge)
			suite.Require().Equal(gauges[0], *gauge)

			// check invalid gauge ID
			_, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeID+1000)
			suite.Require().Error(err)
			rewardsEst = suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
			suite.Require().Equal(sdk.Coins{}, rewardsEst)

			// check gauge ids by denom
			gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
			suite.Require().Len(gaugeIdsByDenom, 0)
		} else { // check perpetual gauges (not finished + rewards estimate empty)

			// check finished gauges
			gauges = suite.App.IncentivesKeeper.GetFinishedGauges(suite.Ctx)
			suite.Require().Len(gauges, 0)

			// check rewards estimation
			rewardsEst = suite.App.IncentivesKeeper.GetRewardsEst(suite.Ctx, lockOwners[0], []lockuptypes.PeriodLock{}, 100)
			suite.Require().Equal(sdk.Coins(nil), rewardsEst)

			// check gauge ids by denom
			gaugeIdsByDenom = suite.App.IncentivesKeeper.GetAllGaugeIDsByDenom(suite.Ctx, "lptoken")
			suite.Require().Len(gaugeIdsByDenom, 1)
		}
	}

}
