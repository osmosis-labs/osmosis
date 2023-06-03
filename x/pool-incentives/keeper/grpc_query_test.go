package keeper_test

import (
	"context"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v16/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v16/x/pool-incentives/types"
)

var (
	isPerpetual  = true
	notPerpetual = false
)

func (s *KeeperTestSuite) TestGaugeIds() {
	for _, tc := range []struct {
		desc    string
		request *types.QueryGaugeIdsRequest
		err     bool
	}{
		{
			desc:    "Empty request",
			request: &types.QueryGaugeIdsRequest{},
			err:     true,
		},
		{
			desc: "Nonexistent pool",
			request: &types.QueryGaugeIdsRequest{
				PoolId: 2,
			},
			err: true,
		},
		{
			desc: "Happy case",
			request: &types.QueryGaugeIdsRequest{
				PoolId: 1,
			},
			err: false,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			s.SetupTest()
			queryClient := s.queryClient
			// Prepare a balancer pool
			s.PrepareBalancerPool()
			// LockableDurations should be 1, 3, 7 hours from the default genesis state.
			lockableDurations := s.App.PoolIncentivesKeeper.GetLockableDurations(s.Ctx)
			s.Require().Equal(3, len(lockableDurations))

			res, err := queryClient.GaugeIds(context.Background(), tc.request)
			if tc.err {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(3, len(res.GaugeIdsWithDuration))

				for i := 0; i < len(res.GaugeIdsWithDuration); i++ {
					s.Require().Equal(lockableDurations[i], res.GaugeIdsWithDuration[i].Duration)
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestDistrInfo() {
	for _, tc := range []struct {
		desc                 string
		poolCreated          bool
		weights              []sdk.Int
		expectedTotalWeight  sdk.Int
		expectedRecordLength int
	}{
		{
			desc:                 "No pool exists",
			poolCreated:          false,
			weights:              []sdk.Int{},
			expectedTotalWeight:  sdk.NewInt(0),
			expectedRecordLength: 0,
		},
		{
			desc:                 "Happy case",
			poolCreated:          true,
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			expectedTotalWeight:  sdk.NewInt(600),
			expectedRecordLength: 3,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			s.SetupTest()
			keeper := s.App.PoolIncentivesKeeper
			queryClient := s.queryClient

			if tc.poolCreated {
				poolId := s.PrepareBalancerPool()

				// LockableDurations should be 1, 3, 7 hours from the default genesis state.
				lockableDurations := keeper.GetLockableDurations(s.Ctx)
				s.Require().Equal(3, len(lockableDurations))

				var distRecord []types.DistrRecord
				for i := 0; i < len(lockableDurations); i++ {
					gaugeId, err := keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[i])
					s.Require().NoError(err)
					distRecord = append(distRecord, types.DistrRecord{
						GaugeId: gaugeId,
						Weight:  tc.weights[i],
					})
				}

				// Create 3 records
				err := keeper.UpdateDistrRecords(s.Ctx, distRecord...)
				s.Require().NoError(err)
			}

			res, err := queryClient.DistrInfo(context.Background(), &types.QueryDistrInfoRequest{})
			s.Require().NoError(err)

			s.Require().Equal(tc.expectedTotalWeight, res.DistrInfo.TotalWeight)
			s.Require().Equal(tc.expectedRecordLength, len(res.DistrInfo.Records))
		})
	}
}

func (s *KeeperTestSuite) TestParams() {
	s.SetupTest()

	queryClient := s.queryClient

	res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)

	// Minted denom set as "stake" from the default genesis state
	s.Require().Equal("stake", res.Params.MintedDenom)
}

func (s *KeeperTestSuite) TestLockableDurations() {
	s.SetupTest()

	queryClient := s.queryClient

	res, err := queryClient.LockableDurations(context.Background(), &types.QueryLockableDurationsRequest{})
	s.Require().NoError(err)

	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	s.Require().Equal(3, len(res.LockableDurations))
	s.Require().Equal(time.Hour, res.LockableDurations[0])
	s.Require().Equal(time.Hour*3, res.LockableDurations[1])
	s.Require().Equal(time.Hour*7, res.LockableDurations[2])
}

func (s *KeeperTestSuite) TestIncentivizedPools() {
	for _, tc := range []struct {
		desc                 string
		poolCreated          bool
		weights              []sdk.Int
		clPoolWithGauge      bool
		clGaugeWeight        sdk.Int
		perpetual            bool
		nonPerpetual         bool
		expectedRecordLength int
	}{
		{
			desc:                 "No pool exist",
			poolCreated:          false,
			weights:              []sdk.Int{},
			expectedRecordLength: 0,
		},
		{
			desc:                 "Normal case",
			poolCreated:          true,
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			expectedRecordLength: 3,
		},
		{
			desc:                 "Perpetual",
			poolCreated:          true,
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			perpetual:            true,
			expectedRecordLength: 3,
		},
		{
			desc:                 "Non Perpetual",
			poolCreated:          true,
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			nonPerpetual:         true,
			expectedRecordLength: 0,
		},
		{
			desc:                 "Concentrated case",
			poolCreated:          true,
			clPoolWithGauge:      true,
			clGaugeWeight:        sdk.NewInt(400),
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			expectedRecordLength: 4,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			s.SetupTest()
			keeper := s.App.PoolIncentivesKeeper
			queryClient := s.queryClient
			var poolId uint64

			var lockableDurations []time.Duration
			if tc.poolCreated {
				// prepare a balancer pool
				poolId = s.PrepareBalancerPool()
				// LockableDurations should be 1, 3, 7 hours from the default genesis state.
				lockableDurations = keeper.GetLockableDurations(s.Ctx)
				s.Require().Equal(3, len(lockableDurations))

				var distRecords []types.DistrRecord

				// If appropriate, create a concentrated pool with a gauge.
				// Recall that concentrated pool gauges are created on epoch duration, which
				// is set in default genesis to be 168 hours (1 week). Since this is not included
				// in the set of lockable durations, creating this gauge serves as a way to ensure
				// CL gauges are captured by the IncentivizedPools query.
				if tc.clPoolWithGauge {
					clPool := s.PrepareConcentratedPool()
					epochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration

					clGaugeId, err := keeper.GetPoolGaugeId(s.Ctx, clPool.GetId(), epochDuration)
					s.Require().NoError(err)
					distRecords = append(distRecords, types.DistrRecord{GaugeId: clGaugeId, Weight: tc.clGaugeWeight})
				}

				for i := 0; i < len(lockableDurations); i++ {
					gaugeId, err := keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[i])
					s.Require().NoError(err)
					distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugeId, Weight: tc.weights[i]})

					if tc.perpetual {
						gaugePerpetualId, err := s.App.IncentivesKeeper.CreateGauge(
							s.Ctx, isPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
								LockQueryType: lockuptypes.ByDuration,
								Denom:         "stake",
								Duration:      time.Hour,
							}, time.Now(), 1)
						s.Require().NoError(err)
						distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugePerpetualId, Weight: sdk.NewInt(300)})
					}
					if tc.nonPerpetual {
						gaugeNonPerpetualId, err := s.App.IncentivesKeeper.CreateGauge(
							s.Ctx, notPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
								LockQueryType: lockuptypes.ByDuration,
								Denom:         "stake",
								Duration:      time.Hour,
							}, time.Now(), 1)
						s.Require().NoError(err)
						distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugeNonPerpetualId, Weight: sdk.NewInt(100)})
					}

					// Sort in ascending order of gaugeId
					sort.Slice(distRecords[:], func(i, j int) bool {
						return distRecords[i].GaugeId < distRecords[j].GaugeId
					})

					// update records and ensure that non-perpetuals pot cannot get rewards.
					_ = keeper.UpdateDistrRecords(s.Ctx, distRecords...)
				}
				res, err := queryClient.IncentivizedPools(context.Background(), &types.QueryIncentivizedPoolsRequest{})
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedRecordLength, len(res.IncentivizedPools))
			}
		})
	}
}

func (s *KeeperTestSuite) TestExternalIncentiveGauges() {
	type externalGauge struct {
		isPerpetual bool
	}

	tests := map[string]struct {
		poolCreated          bool
		internalGaugeWeights []sdk.Int
		externalGauges       []externalGauge
		clPoolWithGauge      bool
		clGaugeWeight        sdk.Int

		expectedNumExternalGauges int
		expectedGaugeIDs          []uint64
	}{
		"No pool exist": {
			poolCreated:          false,
			internalGaugeWeights: []sdk.Int{},

			expectedNumExternalGauges: 0,
		},
		"All gauges are internal (no external gauges)": {
			poolCreated:          true,
			internalGaugeWeights: []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},

			expectedNumExternalGauges: 0,
		},
		"Mixed internal and external gauges": {
			poolCreated:          true,
			internalGaugeWeights: []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			externalGauges:       []externalGauge{{isPerpetual: true}, {isPerpetual: false}},

			// Since there is no concentrated pool, there are only 3 internal gauges, one for each lockup duration
			expectedGaugeIDs:          []uint64{4, 5},
			expectedNumExternalGauges: 2,
		},
		"More external gauges than internal gauges": {
			poolCreated:          true,
			internalGaugeWeights: []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			externalGauges:       []externalGauge{{isPerpetual: true}, {isPerpetual: false}, {isPerpetual: true}, {isPerpetual: true}, {isPerpetual: false}},

			// Since there is no concentrated pool, there are only 3 internal gauges, one for each lockup duration
			expectedGaugeIDs:          []uint64{4, 5, 6, 7, 8},
			expectedNumExternalGauges: 5,
		},
		"Same number of external gauges as internal gauges": {
			poolCreated:          true,
			internalGaugeWeights: []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			externalGauges:       []externalGauge{{isPerpetual: true}, {isPerpetual: false}, {isPerpetual: true}},

			// Since there is no concentrated pool, there are only 3 internal gauges, one for each lockup duration
			expectedGaugeIDs:          []uint64{4, 5, 6},
			expectedNumExternalGauges: 3,
		},
		"Internal gauge for concentrated pool exists": {
			poolCreated:          true,
			clPoolWithGauge:      true,
			clGaugeWeight:        sdk.NewInt(100),
			internalGaugeWeights: []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			externalGauges:       []externalGauge{{isPerpetual: true}, {isPerpetual: false}, {isPerpetual: true}, {isPerpetual: true}, {isPerpetual: false}},

			// Since there is a concentrated pool, there are 4 internal gauges, one for each lockup duration and one for the epoch duration
			// (Note that in our test defaults, the epoch duration does not overlap with any lockup durations)
			expectedGaugeIDs:          []uint64{5, 6, 7, 8, 9},
			expectedNumExternalGauges: 5,
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			keeper := s.App.PoolIncentivesKeeper
			queryClient := s.queryClient
			var poolId uint64

			var lockableDurations []time.Duration
			if tc.poolCreated {
				// prepare a balancer pool
				poolId = s.PrepareBalancerPool()
				// LockableDurations should be 1, 3, 7 hours from the default genesis state.
				lockableDurations = keeper.GetLockableDurations(s.Ctx)
				s.Require().Equal(3, len(lockableDurations))

				// --- Internal Gauge Setup ---

				var distRecords []types.DistrRecord

				// If appropriate, create a concentrated pool with an internal gauge.
				// Recall that concentrated pool gauges are created on epoch duration, which
				// is set in default genesis to be 168 hours (1 week).
				if tc.clPoolWithGauge {
					clPool := s.PrepareConcentratedPool()
					epochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration

					clGaugeId, err := keeper.GetPoolGaugeId(s.Ctx, clPool.GetId(), epochDuration)
					s.Require().NoError(err)
					distRecords = append(distRecords, types.DistrRecord{GaugeId: clGaugeId, Weight: tc.clGaugeWeight})
				}

				for i := 0; i < len(lockableDurations); i++ {
					gaugeId, err := keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[i])
					s.Require().NoError(err)
					distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugeId, Weight: tc.internalGaugeWeights[i]})

					// Sort in ascending order of gaugeId
					sort.Slice(distRecords[:], func(i, j int) bool {
						return distRecords[i].GaugeId < distRecords[j].GaugeId
					})
				}

				// --- External Gauge Setup ---

				// Create external gauges if appropriate.
				// Note that we do not add these gauges to distrRecords, which is how we
				// ensure they are classified as external
				if tc.externalGauges != nil {
					for _, externalBalGauge := range tc.externalGauges {
						_, err := s.App.IncentivesKeeper.CreateGauge(
							s.Ctx, externalBalGauge.isPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
								LockQueryType: lockuptypes.ByDuration,
								Denom:         "stake",
								Duration:      time.Hour,
							}, time.Now(), 1)
						s.Require().NoError(err)
					}
				}

				// update records and ensure that non-perpetuals cannot get rewards.
				_ = keeper.UpdateDistrRecords(s.Ctx, distRecords...)

				// --- System under test ---

				res, err := queryClient.ExternalIncentiveGauges(context.Background(), &types.QueryExternalIncentiveGaugesRequest{})

				// --- Assertions ---

				s.Require().NoError(err)
				s.Require().Equal(tc.expectedNumExternalGauges, len(res.Data))

				// Ensure retrieved gauge IDs are correct
				if tc.expectedGaugeIDs != nil {
					s.Require().Equal(tc.expectedNumExternalGauges, len(tc.expectedGaugeIDs))

					for i, gaugeId := range tc.expectedGaugeIDs {
						s.Require().Equal(gaugeId, res.Data[i].Id)
					}
				}
			}
		})
	}
}

func (s *KeeperTestSuite) TestGaugeIncentivePercentage() {
	s.SetupTest()

	keeper := s.App.PoolIncentivesKeeper
	queryClient := s.queryClient

	poolId := s.PrepareBalancerPool()
	// LockableDurations should be 1, 3, 7 hours from the default genesis state.
	lockableDurations := keeper.GetLockableDurations(s.Ctx)
	s.Require().Equal(3, len(lockableDurations))

	gauge1Id, err := keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[0])
	s.Require().NoError(err)

	gauge2Id, err := keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[1])
	s.Require().NoError(err)

	gauge3Id, err := keeper.GetPoolGaugeId(s.Ctx, poolId, lockableDurations[2])
	s.Require().NoError(err)

	// Create 3 records
	err = keeper.UpdateDistrRecords(s.Ctx, types.DistrRecord{
		GaugeId: gauge1Id,
		Weight:  sdk.NewInt(100),
	}, types.DistrRecord{
		GaugeId: gauge2Id,
		Weight:  sdk.NewInt(200),
	}, types.DistrRecord{
		GaugeId: gauge3Id,
		Weight:  sdk.NewInt(300),
	})
	s.Require().NoError(err)

	res, err := queryClient.GaugeIds(context.Background(), &types.QueryGaugeIdsRequest{
		PoolId: poolId,
	})

	s.Require().NoError(err)
	s.Require().Equal("16.666666666666666700", res.GaugeIdsWithDuration[0].GaugeIncentivePercentage)
	s.Require().Equal("33.333333333333333300", res.GaugeIdsWithDuration[1].GaugeIncentivePercentage)
	s.Require().Equal("50.000000000000000000", res.GaugeIdsWithDuration[2].GaugeIncentivePercentage)
}
