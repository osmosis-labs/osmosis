package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v27/app/apptesting"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	incentivestypes "github.com/osmosis-labs/osmosis/v27/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolincentivestypes "github.com/osmosis-labs/osmosis/v27/x/pool-incentives/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()

	s.queryClient = types.NewQueryClient(s.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestCreateBalancerPoolGauges() {
	s.SetupTest()

	keeper := s.App.PoolIncentivesKeeper

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(s.Ctx)
	s.Equal(3, len(lockableDurations))

	for i := 0; i < 3; i++ {
		poolId := s.PrepareBalancerPool()
		pool, err := s.App.GAMMKeeper.GetPoolAndPoke(s.Ctx, poolId)
		s.NoError(err)

		poolLpDenom := gammtypes.GetPoolShareDenom(pool.GetId())

		// Same amount of gauges as lockableDurations must be created for every pool created.
		gaugeId, err := keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[0])
		s.NoError(err)
		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
		s.NoError(err)
		s.Equal(0, len(gauge.Coins))
		s.Equal(true, gauge.IsPerpetual)
		s.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		s.Equal(lockableDurations[0], gauge.DistributeTo.Duration)

		gaugeId, err = keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[1])
		s.NoError(err)
		gauge, err = s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
		s.NoError(err)
		s.Equal(0, len(gauge.Coins))
		s.Equal(true, gauge.IsPerpetual)
		s.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		s.Equal(lockableDurations[1], gauge.DistributeTo.Duration)

		gaugeId, err = keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[2])
		s.NoError(err)
		gauge, err = s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
		s.NoError(err)
		s.Equal(0, len(gauge.Coins))
		s.Equal(true, gauge.IsPerpetual)
		s.Equal(poolLpDenom, gauge.DistributeTo.Denom)
		s.Equal(lockableDurations[2], gauge.DistributeTo.Duration)
	}
}

func (s *KeeperTestSuite) TestCreateConcentratePoolGauges() {
	s.SetupTest()

	keeper := s.App.PoolIncentivesKeeper

	for i := 0; i < 3; i++ {
		clPool := s.PrepareConcentratedPool()

		incParams := s.App.IncentivesKeeper.GetParams(s.Ctx).DistrEpochIdentifier
		currEpoch := s.App.EpochsKeeper.GetEpochInfo(s.Ctx, incParams)

		// Same amount of gauges as lockableDurations must be created for every pool created.
		gaugeId, err := keeper.GetPoolGaugeId(s.Ctx, clPool.GetId(), currEpoch.Duration)
		s.NoError(err)

		// Same amount of NoLock gauges as lockableDurations must be created for every pool created.
		gaugeIds, err := keeper.GetNoLockGaugeIdsFromPool(s.Ctx, clPool.GetId())
		s.NoError(err)

		s.Equal(1, len(gaugeIds))

		s.Equal(gaugeId, gaugeIds[0])

		gauge, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, gaugeId)
		s.NoError(err)
		s.Equal(0, len(gauge.Coins))
		s.Equal(true, gauge.IsPerpetual)
		s.Equal(gaugeId, gauge.Id)
	}
}

func (s *KeeperTestSuite) TestCreateLockablePoolGauges() {
	durations := s.App.PoolIncentivesKeeper.GetLockableDurations(s.Ctx)

	tests := []struct {
		name                      string
		poolId                    uint64
		isInvalidLockableDuration bool
		expectedGaugeDurations    []time.Duration
		expectedGaugeIds          []uint64
		expectedErr               bool
	}{
		{
			name:                   "Create Gauge with valid PoolId",
			poolId:                 uint64(1),
			expectedGaugeDurations: durations,
			expectedGaugeIds:       []uint64{4, 5, 6}, // note: it's not 1,2,3 because we create 3 gauges during setup of s.PrepareBalancerPool()
			expectedErr:            false,
		},
		{
			name:                   "Create Gauge with invalid PoolId",
			poolId:                 uint64(0),
			expectedGaugeDurations: nil,
			expectedGaugeIds:       []uint64{},
			expectedErr:            true,
		},
		{
			name:                      "error: invalid lockable duration",
			poolId:                    uint64(1),
			isInvalidLockableDuration: true,

			expectedErr: true,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.SetupTest()

			poolId := s.PrepareBalancerPool()

			// This should trigger error when creating a pool id <> gauge id internal incentive link.
			if tc.isInvalidLockableDuration {
				durations = []time.Duration{time.Duration(0)}
				s.App.PoolIncentivesKeeper.SetLockableDurations(s.Ctx, durations)
			}

			err := s.App.PoolIncentivesKeeper.CreateLockablePoolGauges(s.Ctx, tc.poolId)
			if tc.expectedErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NotEmpty(tc.expectedGaugeDurations)

				for idx, duration := range tc.expectedGaugeDurations {
					actualGaugeId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, tc.poolId, duration)
					s.Require().NoError(err)
					s.Require().Equal(tc.expectedGaugeIds[idx], actualGaugeId)

					// Get gauge information
					gaugeInfo, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, actualGaugeId)
					s.Require().NoError(err)

					s.Require().Equal(actualGaugeId, gaugeInfo.Id)
					s.Require().True(gaugeInfo.IsPerpetual)
					s.Require().Empty(gaugeInfo.Coins)
					s.Require().Equal(duration, gaugeInfo.DistributeTo.Duration)
					s.Require().Equal(s.Ctx.BlockTime(), gaugeInfo.StartTime)
					s.Require().Equal(gammtypes.GetPoolShareDenom(poolId), gaugeInfo.DistributeTo.Denom)
					s.Require().Equal(uint64(1), gaugeInfo.NumEpochsPaidOver)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestCreateConcentratedLiquidityPoolGauge() {
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
			expectedGaugeId: 2, // note: it's not 1 because we create one gauge during setup of s.PrepareConcentratedPool()
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
		s.Run(tc.name, func() {
			s.SetupTest()
			if tc.poolType == poolmanagertypes.Concentrated {
				s.PrepareConcentratedPool().GetId()
			} else {
				s.PrepareBalancerPool()
			}

			err := s.App.PoolIncentivesKeeper.CreateConcentratedLiquidityPoolGauge(s.Ctx, tc.poolId)
			if tc.expectedErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)

				incParams := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx)
				// check that the gauge was created successfully
				actualGaugeId, err := s.App.PoolIncentivesKeeper.GetPoolGaugeId(s.Ctx, tc.poolId, incParams.Duration)
				s.Require().NoError(err)

				s.Require().Equal(tc.expectedGaugeId, actualGaugeId)

				// Get gauge information
				gaugeInfo, err := s.App.IncentivesKeeper.GetGaugeByID(s.Ctx, actualGaugeId)
				s.Require().NoError(err)

				s.Require().Equal(actualGaugeId, gaugeInfo.Id)
				s.Require().True(gaugeInfo.IsPerpetual)
				s.Require().Empty(gaugeInfo.Coins)
				s.Require().Equal(s.Ctx.BlockTime(), gaugeInfo.StartTime)
				s.Require().Equal(incentivestypes.NoLockInternalGaugeDenom(tc.poolId), gaugeInfo.DistributeTo.Denom)
				s.Require().Equal(uint64(1), gaugeInfo.NumEpochsPaidOver)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetGaugesForCFMMPool() {
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
		s.Run(name, func() {
			s.SetupTest()

			s.PrepareBalancerPool()

			gauges, err := s.App.PoolIncentivesKeeper.GetGaugesForCFMMPool(s.Ctx, tc.poolId)

			if tc.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, tc.expectError)
				return
			}

			s.Require().NoError(err)

			// Validate that  3 gauges for each lockable duration were created.
			s.Require().Equal(3, len(gauges))
			for i, lockableDuration := range s.App.PoolIncentivesKeeper.GetLockableDurations(s.Ctx) {
				s.Require().Equal(uint64(i+1), gauges[i].Id)
				s.Require().Equal(lockableDuration, gauges[i].DistributeTo.Duration)
				s.Require().True(gauges[i].IsActiveGauge(s.Ctx.BlockTime()))
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetLongestLockableDuration() {
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
		s.Run(tc.name, func() {
			s.App.PoolIncentivesKeeper.SetLockableDurations(s.Ctx, tc.lockableDurations)

			result, err := s.App.PoolIncentivesKeeper.GetLongestLockableDuration(s.Ctx)
			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}

			s.Require().Equal(tc.expectedDuration, result)
		})
	}
}

func (s *KeeperTestSuite) TestIsPoolIncentivized() {
	testCases := []struct {
		name                   string
		poolIdToQuery          uint64
		expectedIsIncentivized bool
	}{
		{
			name:                   "Incentivized Pool",
			poolIdToQuery:          1,
			expectedIsIncentivized: true,
		},
		{
			name:                   "Unincentivized Pool",
			poolIdToQuery:          2,
			expectedIsIncentivized: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()
			s.PrepareAllSupportedPools()

			s.App.PoolIncentivesKeeper.SetDistrInfo(s.Ctx, poolincentivestypes.DistrInfo{
				TotalWeight: osmomath.NewInt(100),
				Records: []poolincentivestypes.DistrRecord{
					{
						GaugeId: 1,
						Weight:  osmomath.NewInt(50),
					},
				},
			})

			actualIsIncentivized, err := s.App.PoolIncentivesKeeper.IsPoolIncentivized(s.Ctx, tc.poolIdToQuery)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedIsIncentivized, actualIsIncentivized)
		})
	}
}

// Tests that for every supported internally incentivized pool,
// the appropriate gauge ID is returned.
// For balancer and stableswap, returns the longest duration gauge ID.
// For CL, returns the gauge ID for the current epoch incentive duration.
// For cosmwasm pool, returns an error.
// For non-existent pool ID, returns an error.
func (suite *KeeperTestSuite) TestGetInternalGaugeIDForPool() {

	// Note that we initialize the same state for all pools.
	suite.SetupTest()

	// Prepare pools and their IDs
	poolInfo := suite.PrepareAllSupportedPools()

	tests := map[string]struct {
		poolID          uint64
		expectedGaugeID uint64
		expectError     error
	}{
		"concentrated pool": {
			poolID:          poolInfo.ConcentratedPoolID,
			expectedGaugeID: poolInfo.ConcentratedGaugeID,
		},
		"balancer pool": {
			poolID:          poolInfo.BalancerPoolID,
			expectedGaugeID: poolInfo.BalancerGaugeID,
		},
		"stableswap pool": {
			poolID:          poolInfo.StableSwapPoolID,
			expectedGaugeID: poolInfo.StableSwapGaugeID,
		},
		"cosmwasm pool": {
			poolID:      poolInfo.CosmWasmPoolID,
			expectError: types.UnsupportedPoolTypeError{PoolID: poolInfo.CosmWasmPoolID, PoolType: poolmanagertypes.CosmWasm},
		},
		"pool with given ID does not exist": {
			poolID:      poolInfo.AlloyedPoolID + 1,
			expectError: poolmanagertypes.FailedToFindRouteError{PoolId: poolInfo.AlloyedPoolID + 1},
		},
	}

	for name, tc := range tests {
		suite.Run(name, func() {

			poolIncentivesKeeper := suite.App.PoolIncentivesKeeper

			gaugeID, err := poolIncentivesKeeper.GetInternalGaugeIDForPool(suite.Ctx, tc.poolID)

			if tc.expectError != nil {
				suite.Require().Error(err)
				suite.Require().ErrorIs(tc.expectError, err)
				return
			}
			suite.Require().NoError(err)
			suite.Require().Equal(tc.expectedGaugeID, gaugeID)
		})
	}
}
