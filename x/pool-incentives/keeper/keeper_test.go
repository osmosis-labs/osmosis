package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
	appParams "github.com/osmosis-labs/osmosis/v15/app/params"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v15/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v15/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestCreateBalancerPoolGauges() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(suite.Ctx)
	suite.Equal(3, len(lockableDurations))

	for i := 0; i < 3; i++ {
		poolId := suite.PrepareBalancerPool()
		pool, err := suite.App.GAMMKeeper.GetPoolAndPoke(suite.Ctx, poolId)
		suite.NoError(err)

		poolLpDenom := gammtypes.GetPoolShareDenom(pool.GetId())

		// Same amount of gauges as lockableDurations must be created for every pool created.
		gaugeId, err := keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[0])
		suite.NoError(err)
		gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
		suite.NoError(err)
		suite.Equal(0, len(gauge.Coins))
		suite.Equal(true, gauge.IsPerpetual)
		suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		suite.Equal(lockableDurations[0], gauge.DistributeTo.Duration)

		gaugeId, err = keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[1])
		suite.NoError(err)
		gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
		suite.NoError(err)
		suite.Equal(0, len(gauge.Coins))
		suite.Equal(true, gauge.IsPerpetual)
		suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		suite.Equal(lockableDurations[1], gauge.DistributeTo.Duration)

		gaugeId, err = keeper.GetPoolGaugeId(suite.Ctx, poolId, lockableDurations[2])
		suite.NoError(err)
		gauge, err = suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
		suite.NoError(err)
		suite.Equal(0, len(gauge.Coins))
		suite.Equal(true, gauge.IsPerpetual)
		suite.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		suite.Equal(lockableDurations[2], gauge.DistributeTo.Duration)
	}
}

func (suite *KeeperTestSuite) TestCreateConcentratePoolGauges() {
	suite.SetupTest()

	keeper := suite.App.PoolIncentivesKeeper

	for i := 0; i < 3; i++ {
		clPool := suite.PrepareConcentratedPool()

		incParams := suite.App.IncentivesKeeper.GetParams(suite.Ctx).DistrEpochIdentifier
		currEpoch := suite.App.EpochsKeeper.GetEpochInfo(suite.Ctx, incParams)

		// Same amount of gauges as lockableDurations must be created for every pool created.
		gaugeId, err := keeper.GetPoolGaugeId(suite.Ctx, clPool.GetId(), currEpoch.Duration)
		suite.NoError(err)
		gauge, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, gaugeId)
		suite.NoError(err)
		suite.Equal(0, len(gauge.Coins))
		suite.Equal(true, gauge.IsPerpetual)
		suite.Equal(gaugeId, gauge.Id)
	}
}

func (suite *KeeperTestSuite) TestCreateLockablePoolGauges() {
	durations := suite.App.PoolIncentivesKeeper.GetLockableDurations(suite.Ctx)

	tests := []struct {
		name                   string
		poolId                 uint64
		expectedGaugeDurations []time.Duration
		expectedGaugeIds       []uint64
		expectedErr            bool
	}{
		{
			name:                   "Create Gauge with valid PoolId",
			poolId:                 uint64(1),
			expectedGaugeDurations: durations,
			expectedGaugeIds:       []uint64{4, 5, 6}, //note: it's not 1,2,3 because we create 3 gauges during setup of suite.PrepareBalancerPool()
			expectedErr:            false,
		},
		{
			name:                   "Create Gauge with invalid PoolId",
			poolId:                 uint64(0),
			expectedGaugeDurations: nil,
			expectedGaugeIds:       []uint64{},
			expectedErr:            true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			poolId := suite.PrepareBalancerPool()

			err := suite.App.PoolIncentivesKeeper.CreateLockablePoolGauges(suite.Ctx, tc.poolId)
			if tc.expectedErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotEmpty(tc.expectedGaugeDurations)

				for idx, duration := range tc.expectedGaugeDurations {
					actualGaugeId, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, tc.poolId, duration)
					suite.Require().NoError(err)
					suite.Require().Equal(tc.expectedGaugeIds[idx], actualGaugeId)

					// Get gauge information
					gaugeInfo, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, actualGaugeId)
					suite.Require().NoError(err)

					suite.Require().Equal(actualGaugeId, gaugeInfo.Id)
					suite.Require().True(gaugeInfo.IsPerpetual)
					suite.Require().Empty(gaugeInfo.Coins)
					suite.Require().Equal(duration, gaugeInfo.DistributeTo.Duration)
					suite.Require().Equal(suite.Ctx.BlockTime(), gaugeInfo.StartTime)
					suite.Require().Equal(gammtypes.GetPoolShareDenom(poolId), gaugeInfo.DistributeTo.Denom)
					suite.Require().Equal(uint64(1), gaugeInfo.NumEpochsPaidOver)
				}
			}
		})
	}
}

func (suite *KeeperTestSuite) TestCreateConcentratedLiquidityPoolGauge() {
	tests := []struct {
		name            string
		poolId          uint64
		poolType        poolmanagertypes.PoolType
		expectedGaugeId uint64
		expectedErr     bool
	}{
		{
			name:            "Create Gauge with valid PoolId",
			poolId:          uint64(1),
			poolType:        poolmanagertypes.Concentrated,
			expectedGaugeId: 2, // note: it's not 1 because we create one gauge during setup of suite.PrepareConcentratedPool()
			expectedErr:     false,
		},
		{
			name:            "Create Gauge with balancer poolType",
			poolId:          uint64(1),
			poolType:        poolmanagertypes.Balancer,
			expectedGaugeId: 0,
			expectedErr:     true,
		},
		{
			name:            "Create Gauge with invalid PoolId",
			poolId:          uint64(0),
			expectedGaugeId: 0,
			expectedErr:     true,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			if tc.poolType == poolmanagertypes.Concentrated {
				suite.PrepareConcentratedPool().GetId()
			} else {
				suite.PrepareBalancerPool()
			}

			err := suite.App.PoolIncentivesKeeper.CreateConcentratedLiquidityPoolGauge(suite.Ctx, tc.poolId)
			if tc.expectedErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)

				incParams := suite.App.IncentivesKeeper.GetEpochInfo(suite.Ctx)
				// check that the gauge was created successfully
				actualGaugeId, err := suite.App.PoolIncentivesKeeper.GetPoolGaugeId(suite.Ctx, tc.poolId, incParams.Duration)
				suite.Require().NoError(err)

				suite.Require().Equal(tc.expectedGaugeId, actualGaugeId)

				// Get gauge information
				gaugeInfo, err := suite.App.IncentivesKeeper.GetGaugeByID(suite.Ctx, actualGaugeId)
				suite.Require().NoError(err)

				suite.Require().Equal(actualGaugeId, gaugeInfo.Id)
				suite.Require().True(gaugeInfo.IsPerpetual)
				suite.Require().Empty(gaugeInfo.Coins)
				suite.Require().Equal(suite.Ctx.BlockTime(), gaugeInfo.StartTime)
				suite.Require().Equal(appParams.BaseCoinUnit, gaugeInfo.DistributeTo.Denom)
				suite.Require().Equal(uint64(1), gaugeInfo.NumEpochsPaidOver)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetGaugesForCFMMPool() {
	const validPoolId = 1

	tests := map[string]struct {
		poolId         uint64
		expectedGauges incentivestypes.Gauge
		expectError    error
	}{
		"valid pool id - gauges created": {
			poolId: validPoolId,
		},
		"invalid pool id - error": {
			poolId:      validPoolId + 1,
			expectError: types.NoGaugeAssociatedWithPoolError{PoolId: 2, Duration: time.Hour},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()

			suite.PrepareBalancerPool()

			gauges, err := suite.App.PoolIncentivesKeeper.GetGaugesForCFMMPool(suite.Ctx, tc.poolId)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(err, tc.expectError)
				return
			}

			suite.Require().NoError(err)

			// Validate that  3 gauges for each lockable duration were created.
			suite.Require().Equal(3, len(gauges))
			for i, lockableDuration := range suite.App.PoolIncentivesKeeper.GetLockableDurations(suite.Ctx) {
				suite.Require().Equal(uint64(i+1), gauges[i].Id)
				suite.Require().Equal(lockableDuration, gauges[i].DistributeTo.Duration)
				suite.Require().True(gauges[i].IsActiveGauge(suite.Ctx.BlockTime()))
			}
		})
	}
}

func (suite *KeeperTestSuite) TestGetLongestLockableDuration() {
	testCases := []struct {
		name              string
		lockableDurations []time.Duration
		expectedDuration  time.Duration
		expectError       bool
	}{
		{
			name:              "3 lockable Durations",
			lockableDurations: []time.Duration{time.Hour, time.Minute, time.Second},
			expectedDuration:  time.Hour,
		},

		{
			name:              "2 lockable Durations",
			lockableDurations: []time.Duration{time.Second, time.Minute},
			expectedDuration:  time.Minute,
		},
		{
			name:              "1 lockable Durations",
			lockableDurations: []time.Duration{time.Minute},
			expectedDuration:  time.Minute,
		},
		{
			name:              "0 lockable Durations",
			lockableDurations: []time.Duration{},
			expectedDuration:  0,
			expectError:       true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {

			suite.App.PoolIncentivesKeeper.SetLockableDurations(suite.Ctx, tc.lockableDurations)

			result, err := suite.App.PoolIncentivesKeeper.GetLongestLockableDuration(suite.Ctx)
			if tc.expectError {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}

			suite.Require().Equal(tc.expectedDuration, result)
		})
	}
}
