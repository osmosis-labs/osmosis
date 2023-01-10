package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

func (s *TestSuite) TestMigrateExistingPools() {
	// create two pools before migration
	s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins[0], defaultTwoAssetCoins[1])
	s.PrepareBalancerPool()

	// suppose upgrade happened and increment block height and block time
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second * 10))

	// run migration logic
	latestPoolId := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx) - 1
	err := s.twapkeeper.MigrateExistingPools(s.Ctx, latestPoolId)
	s.Require().NoError(err)

	// check transient store
	changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
	s.Require().Equal(int(latestPoolId), len(changedPools))

	upgradeTime := s.Ctx.BlockTime()

	// iterate through all pools, check that all state entries have been correctly updated
	for poolId := 1; poolId <= int(latestPoolId); poolId++ {
		recentTwapRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, uint64(poolId))
		poolDenoms, err := s.App.GAMMKeeper.GetPoolDenoms(s.Ctx, uint64(poolId))
		s.Require().NoError(err)
		denomPairs := types.GetAllUniqueDenomPairs(poolDenoms)
		s.Require().NoError(err)
		s.Require().Equal(len(denomPairs), len(recentTwapRecords))

		// ensure that the migrate logic has been triggered by checking that
		// the twap record time has been updated to the current ctx block time
		s.Require().Equal(upgradeTime, recentTwapRecords[0].Time)

		twapRecord, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, uint64(poolId), recentTwapRecords[0].Asset0Denom, recentTwapRecords[0].Asset1Denom)
		s.Require().NoError(err)
		s.Require().Equal(upgradeTime, twapRecord.Time)

		twapRecordBeforeTime, err := s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, uint64(poolId), s.Ctx.BlockTime(), twapRecord.Asset0Denom, twapRecord.Asset1Denom)
		s.Require().NoError(err)
		s.Require().Equal(upgradeTime, twapRecordBeforeTime.Time)
	}
}

func (s *TestSuite) TestMigrateExistingPoolsError() {
	// create two pools before migration
	s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins[0], defaultTwoAssetCoins[1])
	s.PrepareBalancerPool()

	// suppose upgrade happened and increment block height and block time
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
	s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Second * 10))

	// run migration logic
	// should error when we try to migrate with pool ID that does not exist
	latestPoolIdPlusOne := s.App.PoolManagerKeeper.GetNextPoolId(s.Ctx)
	err := s.twapkeeper.MigrateExistingPools(s.Ctx, latestPoolIdPlusOne)
	s.Require().Error(err)
}

// TestTwapRecord_GeometricTwap_MarshalUnmarshal this test proves that migrations
// to initialize geometric twap accumulators are not required.
// This is because proto marshalling will initialize the field to the zero value.
// Zero value is the expected initialization value for the geometric twap accumulator.
func (suite *TestSuite) TestTwapRecord_GeometricTwap_MarshalUnmarshal() {
	originalRecord := types.TwapRecord{
		Asset0Denom: "uatom",
		Asset1Denom: "uusd",
	}

	suite.Require().True(originalRecord.GeometricTwapAccumulator.IsNil())

	bz, err := proto.Marshal(&originalRecord)
	suite.Require().NoError(err)

	var deserialized types.TwapRecord
	err = proto.Unmarshal(bz, &deserialized)
	suite.Require().NoError(err)

	suite.Require().Equal(originalRecord, deserialized)
	suite.Require().Equal(originalRecord.String(), deserialized.String())

	suite.Require().False(originalRecord.GeometricTwapAccumulator.IsNil())
	suite.Require().Equal(sdk.ZeroDec(), originalRecord.GeometricTwapAccumulator)
}

func (suite *TestSuite) TestInitializeGeometricTwap() {
	var (
		expectedGeometricAccumUpdate = sdk.OneDec()
	)

	// sets the geometric accumulator of every given record to expectedGeometricAccumUpdate
	// and returns the result.
	updateGeometricAccum := func(records []types.TwapRecord) []types.TwapRecord {
		for i := range records {
			records[i].GeometricTwapAccumulator = expectedGeometricAccumUpdate
		}
		return records
	}

	// concatenates all element slices into a single slice and
	// returns the result.
	merge := func(records [][]types.TwapRecord) []types.TwapRecord {
		var result []types.TwapRecord
		for _, r := range records {
			r := r
			result = append(result, r...)
		}
		return result
	}

	// returns three asset record with the given pool id and time, everything other parameter set to 1.
	threeAssetRecordWithPoolIdAndTime := func(poolId uint64, time time.Time) []types.TwapRecord {
		return newThreeAssetRecord(poolId, time, sdk.OneDec(), sdk.OneDec(), sdk.OneDec(), sdk.OneDec(), sdk.OneDec(), sdk.OneDec(), sdk.OneDec())
	}

	tests := map[string]struct {
		expectError bool

		preExistingRecords []types.TwapRecord

		// pool id -> most recent records
		expectedMostRecent map[uint64][]types.TwapRecord
	}{
		"one record, one pool": {
			preExistingRecords: []types.TwapRecord{
				baseRecord,
			},

			expectedMostRecent: map[uint64][]types.TwapRecord{
				baseRecord.PoolId: updateGeometricAccum([]types.TwapRecord{baseRecord}),
			},
		},
		"three records, one pool": {
			preExistingRecords: threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime),

			expectedMostRecent: map[uint64][]types.TwapRecord{
				basePoolId: updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime)),
			},
		},
		"two pools, one record each": {
			preExistingRecords: []types.TwapRecord{
				baseRecord,
				withPoolId(baseRecord, baseRecord.PoolId+1),
			},

			expectedMostRecent: map[uint64][]types.TwapRecord{
				baseRecord.PoolId:     updateGeometricAccum([]types.TwapRecord{baseRecord}),
				baseRecord.PoolId + 1: updateGeometricAccum([]types.TwapRecord{withPoolId(baseRecord, baseRecord.PoolId+1)}),
			},
		},
		"two pools, multiple records each at the same time": {
			preExistingRecords: append(
				threeAssetRecordWithPoolIdAndTime(baseRecord.PoolId, baseTime),
				threeAssetRecordWithPoolIdAndTime(baseRecord.PoolId+1, baseTime)...),

			expectedMostRecent: map[uint64][]types.TwapRecord{
				baseRecord.PoolId:     updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime)),
				baseRecord.PoolId + 1: updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId+1, baseTime)),
			},
		},
		"two pools, multiple records each at different times": {
			preExistingRecords: append(
				threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime),
				threeAssetRecordWithPoolIdAndTime(basePoolId+1, baseTime.Add(time.Second))...),

			expectedMostRecent: map[uint64][]types.TwapRecord{
				baseRecord.PoolId:     updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime)),
				baseRecord.PoolId + 1: updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId+1, baseTime.Add(time.Second))),
			},
		},
		"three pools, 2 at the same time, 1 is not, entries and different time for each": {
			preExistingRecords: merge([][]types.TwapRecord{
				threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime),
				threeAssetRecordWithPoolIdAndTime(basePoolId+1, baseTime.Add(time.Second)),
				threeAssetRecordWithPoolIdAndTime(basePoolId+2, baseTime.Add(time.Second)),
			}),

			expectedMostRecent: map[uint64][]types.TwapRecord{
				baseRecord.PoolId:     updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime)),
				baseRecord.PoolId + 1: updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId+1, baseTime.Add(time.Second))),
				baseRecord.PoolId + 2: updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId+2, baseTime.Add(time.Second))),
			},
		},
		"three pools, 2 at the same time, 1 is not, entries and different time for each. entries at various times per pool": {
			preExistingRecords: merge([][]types.TwapRecord{
				// pool 1
				threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime),
				threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime.Add(time.Second)),
				threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime.Add(time.Second*2)),
				// pool 2
				threeAssetRecordWithPoolIdAndTime(basePoolId+1, baseTime.Add(time.Second)),
				threeAssetRecordWithPoolIdAndTime(basePoolId+1, baseTime.Add(time.Second*3)),
				// pool 3
				threeAssetRecordWithPoolIdAndTime(basePoolId+2, baseTime.Add(time.Second)),
				threeAssetRecordWithPoolIdAndTime(basePoolId+2, baseTime.Add(time.Second*2)),
				threeAssetRecordWithPoolIdAndTime(basePoolId+2, baseTime.Add(time.Second*3)),
				threeAssetRecordWithPoolIdAndTime(basePoolId+2, baseTime.Add(time.Second*4)),
			}),

			expectedMostRecent: map[uint64][]types.TwapRecord{
				baseRecord.PoolId: merge([][]types.TwapRecord{
					updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId, baseTime.Add(time.Second*2))),
				},
				),
				baseRecord.PoolId + 1: merge([][]types.TwapRecord{
					updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId+1, baseTime.Add(time.Second*3))),
				},
				),
				baseRecord.PoolId + 2: merge([][]types.TwapRecord{
					updateGeometricAccum(threeAssetRecordWithPoolIdAndTime(basePoolId+2, baseTime.Add(time.Second*4))),
				},
				),
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		suite.Run(name, func() {
			suite.SetupTest()
			k := suite.App.TwapKeeper
			ctx := suite.Ctx

			suite.preSetRecords(tc.preExistingRecords)

			// Grab all historical pool indexed twap records prior to initializing the
			// geometric accumulator.
			oldHistoricalPoolIndexed, err := k.GetAllHistoricalPoolIndexedTWAPs(ctx)
			suite.Require().NoError(err)
			// Initialize the geometric accumulator to get the expected values.
			expectedHistoricalPoolIndexed := updateGeometricAccum(oldHistoricalPoolIndexed)

			// Repeat for historical time indexed records.
			oldHistoricalTimeIndexed, err := k.GetAllHistoricalTimeIndexedTWAPs(ctx)
			suite.Require().NoError(err)
			expectedHistoricalTimeIndexed := updateGeometricAccum(oldHistoricalTimeIndexed)

			err = k.InitializeGeometricTwap(ctx, sdk.OneDec())

			if tc.expectError {
				suite.Require().Error(err)
				return
			}

			suite.Require().NoError(err)

			// Validate pool indexed
			actualHistoricalPoolIndexed, err := k.GetAllHistoricalPoolIndexedTWAPs(ctx)
			suite.Require().NoError(err)
			suite.Require().Equal(expectedHistoricalPoolIndexed, actualHistoricalPoolIndexed)

			// Validate time indexed
			actualHistoricalTimeIndexed, err := k.GetAllHistoricalTimeIndexedTWAPs(ctx)
			suite.Require().NoError(err)
			suite.Require().Equal(expectedHistoricalTimeIndexed, actualHistoricalTimeIndexed)

			// Validate most recent
			for poolId, expectedMostRecentRecord := range tc.expectedMostRecent {
				actualMostRecentRecord, err := k.GetAllMostRecentRecordsForPool(ctx, poolId)
				suite.Require().NoError(err)

				suite.Require().Equal(len(tc.expectedMostRecent[poolId]), len(actualMostRecentRecord))

				for i, expectedRecord := range expectedMostRecentRecord {
					suite.Require().Equal(expectedRecord, actualMostRecentRecord[i])
				}
			}
		})
	}
}
