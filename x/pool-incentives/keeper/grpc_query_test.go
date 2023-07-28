package keeper_test

import (
	"context"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
	gammmigration "github.com/osmosis-labs/osmosis/v17/x/gamm/types/migration"
	lockuptypes "github.com/osmosis-labs/osmosis/v17/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v17/x/pool-incentives/types"
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
		desc                   string
		poolCreated            bool
		weights                []sdk.Int
		clPoolWithGauge        bool
		clGaugeWeight          sdk.Int
		perpetual              bool
		nonPerpetual           bool
		setupPoolMigrationLink bool
		expectedRecordLength   int
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
			desc:                 "Concentrated case, no pool migration link",
			poolCreated:          true,
			clPoolWithGauge:      true,
			clGaugeWeight:        sdk.NewInt(400),
			weights:              []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			expectedRecordLength: 3, // Note we expect 3 instead of 4 here, since the pool migration link is not setup.
		},
		{
			desc:                   "Concentrated case, setup pool migration link",
			poolCreated:            true,
			clPoolWithGauge:        true,
			clGaugeWeight:          sdk.NewInt(400),
			weights:                []sdk.Int{sdk.NewInt(100), sdk.NewInt(200), sdk.NewInt(300)},
			setupPoolMigrationLink: true,
			expectedRecordLength:   4,
		},
	} {
		tc := tc
		s.Run(tc.desc, func() {
			s.SetupTest()
			keeper := s.App.PoolIncentivesKeeper
			queryClient := s.queryClient
			var balancerPoolId uint64

			// Replace the longest lockable durations with the epoch duration to match the record that gets auto created when making a cl pool.
			epochDuration := s.App.IncentivesKeeper.GetEpochInfo(s.Ctx).Duration
			lockableDurations := keeper.GetLockableDurations(s.Ctx)
			lockableDurations[len(lockableDurations)-1] = epochDuration
			keeper.SetLockableDurations(s.Ctx, lockableDurations)
			lockableDurations = keeper.GetLockableDurations(s.Ctx)
			s.Require().Equal(3, len(lockableDurations))
			s.App.IncentivesKeeper.SetLockableDurations(s.Ctx, lockableDurations)
			lockableDurations = s.App.IncentivesKeeper.GetLockableDurations(s.Ctx)
			s.Require().Equal(3, len(lockableDurations))

			if tc.poolCreated {
				balancerPoolId = s.PrepareBalancerPoolWithCoins(sdk.NewCoin("eth", sdk.NewInt(100000000000)), sdk.NewCoin("usdc", sdk.NewInt(100000000000)))

				var distRecords []types.DistrRecord

				// Note we set up gauges prior to creating the cl pool, otherwise the cl pool will have no gauge to overwrite, and we cannot verify that it works as expected.
				for i := 0; i < len(lockableDurations); i++ {
					gaugeId, err := keeper.GetPoolGaugeId(s.Ctx, balancerPoolId, lockableDurations[i])
					s.Require().NoError(err)
					distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugeId, Weight: tc.weights[i]})

					if tc.perpetual {
						gaugePerpetualId, err := s.App.IncentivesKeeper.CreateGauge(
							s.Ctx, isPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
								LockQueryType: lockuptypes.ByDuration,
								Denom:         "stake",
								Duration:      time.Hour,
							}, time.Now(), 1, 0)
						s.Require().NoError(err)
						distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugePerpetualId, Weight: sdk.NewInt(300)})
					}
					if tc.nonPerpetual {
						gaugeNonPerpetualId, err := s.App.IncentivesKeeper.CreateGauge(
							s.Ctx, notPerpetual, sdk.AccAddress{}, sdk.Coins{}, lockuptypes.QueryCondition{
								LockQueryType: lockuptypes.ByDuration,
								Denom:         "stake",
								Duration:      time.Hour,
							}, time.Now(), 1, 0)
						s.Require().NoError(err)
						distRecords = append(distRecords, types.DistrRecord{GaugeId: gaugeNonPerpetualId, Weight: sdk.NewInt(100)})
					}

					// Sort in ascending order of gaugeId
					sort.Slice(distRecords[:], func(i, j int) bool {
						return distRecords[i].GaugeId < distRecords[j].GaugeId
					})

					// Update records and ensure that non-perpetuals pot cannot get rewards.
					_ = keeper.UpdateDistrRecords(s.Ctx, distRecords...)
				}

				// Create a concentrated pool with a gauge if required.
				var clPool cltypes.ConcentratedPoolExtension
				if tc.clPoolWithGauge {
					clPool = s.PrepareConcentratedPool()
				}

				// Create a link between the balancer pool and the cl pool if required.
				if tc.setupPoolMigrationLink {
					record := gammmigration.BalancerToConcentratedPoolLink{BalancerPoolId: balancerPoolId, ClPoolId: clPool.GetId()}
					err := s.App.GAMMKeeper.ReplaceMigrationRecords(s.Ctx, []gammmigration.BalancerToConcentratedPoolLink{record})
					s.Require().NoError(err)
				}
			}
			// System under test.
			res, err := queryClient.IncentivizedPools(context.Background(), &types.QueryIncentivizedPoolsRequest{})
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedRecordLength, len(res.IncentivizedPools))
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
							}, time.Now(), 1, 0)
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

// TestExternalIncentiveGauges_NoLock tests the ExternalIncentiveGauges
// around gauges of type NoLock.
// For every test case, this test sets up a balancer pool and a concentrated pool.
// Balancer pool creates 3 internal gauges with ids 1, 2, 3
// Concentrated pool creates 1 internal gauge with id 4
// Next, each test case creates external gauges with different distributeTo conditions
// based on the test case parameters.
// Finally, the test calls ExternalIncentiveGauges and ensures that the correct
// gauges are returned.
func (s *KeeperTestSuite) TestExternalIncentiveGauges_NoLock() {
	type gaugeConfig struct {
		distributeTo lockuptypes.QueryCondition
		poolId       uint64
	}

	const (
		defaultIsPerpetual       = true
		defaultNumEpochsPaidOver = 1

		balancerPoolId     = 1
		concentratedPoolId = 2

		// Balancer pool creates 3 internal gauges with ids 1, 2, 3
		// Concentrated pool creates 1 internal gauge with id 4
		firstExpectedExternalGaugeId = 5

		defaultDenom = "stake"
	)

	var (
		defaultCoins = sdk.NewCoins(sdk.NewCoin(defaultDenom, sdk.NewInt(10000000000)))

		defaultLockableDuration = s.App.IncentivesKeeper.GetLockableDurations(s.Ctx)[0]

		defaultNoLockGaugeConfig = gaugeConfig{
			distributeTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.NoLock,
			},
			poolId: concentratedPoolId,
		}

		defaultByDurationGaugeConfig = gaugeConfig{
			distributeTo: lockuptypes.QueryCondition{
				LockQueryType: lockuptypes.ByDuration,
				Duration:      defaultLockableDuration,
				Denom:         defaultDenom,
			},
		}
	)

	tests := map[string]struct {
		gaugesToCreate []gaugeConfig

		expectedGaugeIds []uint64

		expectError bool
	}{
		"1 no lock external gauge": {
			gaugesToCreate: []gaugeConfig{
				defaultNoLockGaugeConfig,
			},
			expectedGaugeIds: []uint64{firstExpectedExternalGaugeId},
		},
		"1 by duration external gauge": {
			gaugesToCreate: []gaugeConfig{
				defaultByDurationGaugeConfig,
			},
			expectedGaugeIds: []uint64{firstExpectedExternalGaugeId},
		},
		"5 gauges no lock and by duration mixed": {
			gaugesToCreate: []gaugeConfig{
				defaultByDurationGaugeConfig,
				defaultNoLockGaugeConfig,
				defaultByDurationGaugeConfig,
				defaultNoLockGaugeConfig,
				defaultByDurationGaugeConfig,
			},
			expectedGaugeIds: []uint64{
				firstExpectedExternalGaugeId,
				firstExpectedExternalGaugeId + 1,
				firstExpectedExternalGaugeId + 2,
				firstExpectedExternalGaugeId + 3,
				firstExpectedExternalGaugeId + 4,
			},
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()

			defaultStartTime := s.Ctx.BlockTime()

			queryClient := s.queryClient

			// Prepare a balancer and a CL pool
			// Creates 3 NoLock internal gauges with ids 1, 2, 3
			s.PrepareBalancerPool()
			// Creates 1 NoLock internal gauge with id 4
			s.PrepareConcentratedPool()

			// Pre-create external gauges
			for _, gauge := range tc.gaugesToCreate {

				// Fund creator
				s.FundAcc(s.TestAccs[0], defaultCoins)

				// Note that some parameters are defaults as they are not relevant to this test
				_, err := s.App.IncentivesKeeper.CreateGauge(s.Ctx, defaultIsPerpetual, s.TestAccs[0], defaultCoins, gauge.distributeTo, defaultStartTime, defaultNumEpochsPaidOver, gauge.poolId)
				s.Require().NoError(err)
			}

			res, err := queryClient.ExternalIncentiveGauges(context.Background(), &types.QueryExternalIncentiveGaugesRequest{})

			if tc.expectError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(len(tc.expectedGaugeIds), len(res.Data))
				for i, gaugeId := range tc.expectedGaugeIds {
					s.Require().Equal(gaugeId, res.Data[i].Id)
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
