package twap_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

// TestTrackChangedPool takes a list of poolIds as test cases, and runs one list per block.
// Every simulated block, checks that there no changed pools.
// Then runs k.trackChangedPool on every item in the test case list.
// Then, checks that changed pools return the list, deduplicated.
//
// This achieves testing the functionality that we depend on, that this clears every end block.
func (s *TestSuite) TestTrackChangedPool() {
	tests := map[string][]uint64{
		"single":         {1},
		"duplicated":     {1, 1},
		"four":           {1, 2, 3, 4},
		"many with dups": {1, 2, 3, 4, 3, 2, 1},
	}
	for name, test := range tests {
		s.Run(name, func() {
			// Test that no cumulative ID registers as tracked
			changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
			s.Require().Empty(changedPools)

			// Track every pool in list
			cumulativeIds := map[uint64]bool{}
			for _, v := range test {
				cumulativeIds[v] = true
				s.twapkeeper.TrackChangedPool(s.Ctx, v)
			}

			changedPools = s.twapkeeper.GetChangedPools(s.Ctx)
			s.Require().Equal(len(cumulativeIds), len(changedPools))
			for _, v := range changedPools {
				s.Require().True(cumulativeIds[v])
			}
			s.Commit()
		})
	}
}

// TestGetAllMostRecentRecordsForPool takes a list of records as test cases,
// and runs storeNewRecord for everything in sequence.
// Then it runs GetAllMostRecentRecordsForPool, and sees if its equal to expected
func (s *TestSuite) TestGetAllMostRecentRecordsForPool() {
	baseRecord := newEmptyPriceRecord(1, baseTime, denom0, denom1)
	tPlusOneRecord := newEmptyPriceRecord(1, tPlusOne, denom0, denom1)
	tests := map[string]struct {
		recordsToSet    []types.TwapRecord
		poolId          uint64
		expectedRecords []types.TwapRecord
	}{
		"set single record": {
			recordsToSet:    []types.TwapRecord{baseRecord},
			poolId:          1,
			expectedRecords: []types.TwapRecord{baseRecord},
		},
		"query non-existent pool": {
			recordsToSet:    []types.TwapRecord{baseRecord},
			poolId:          2,
			expectedRecords: []types.TwapRecord{},
		},
		"set single record, different pool ID": {
			recordsToSet:    []types.TwapRecord{newEmptyPriceRecord(2, baseTime, denom0, denom1)},
			poolId:          2,
			expectedRecords: []types.TwapRecord{newEmptyPriceRecord(2, baseTime, denom0, denom1)},
		},
		"set two records": {
			recordsToSet:    []types.TwapRecord{baseRecord, tPlusOneRecord},
			poolId:          1,
			expectedRecords: []types.TwapRecord{tPlusOneRecord},
		},
		"set two records, reverse order": {
			// The last record, independent of time, takes precedence for most recent.
			recordsToSet:    []types.TwapRecord{tPlusOneRecord, baseRecord},
			poolId:          1,
			expectedRecords: []types.TwapRecord{baseRecord},
		},
		"set multi-asset pool record": {
			recordsToSet: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
				newEmptyPriceRecord(1, baseTime, denom0, denom2),
				newEmptyPriceRecord(1, baseTime, denom1, denom2)},
			poolId: 1,
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
				newEmptyPriceRecord(1, baseTime, denom0, denom2),
				newEmptyPriceRecord(1, baseTime, denom1, denom2)},
		},
		"set multi-asset pool record - reverse order": {
			recordsToSet: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom2),
				newEmptyPriceRecord(1, baseTime, denom1, denom2),
				newEmptyPriceRecord(1, baseTime, denom0, denom1)},
			poolId: 1,
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
				newEmptyPriceRecord(1, baseTime, denom0, denom2),
				newEmptyPriceRecord(1, baseTime, denom1, denom2)},
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			for _, record := range test.recordsToSet {
				s.twapkeeper.StoreNewRecord(s.Ctx, record)
			}
			actualRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, test.poolId)
			s.Require().NoError(err)
			s.Require().Equal(test.expectedRecords, actualRecords)
		})
	}
}

// TestGetRecordAtOrBeforeTime takes a list of records as test cases,
// and runs storeNewRecord for everything in sequence.
// Then it runs GetRecordAtOrBeforeTime, and sees if its equal to expected
func (s *TestSuite) TestGetRecordAtOrBeforeTime() {
	type getRecordInput struct {
		poolId      uint64
		t           time.Time
		asset0Denom string
		asset1Denom string
	}
	defaultInputAt := func(t time.Time) getRecordInput { return getRecordInput{1, t, denom0, denom1} }
	wrongPoolIdInputAt := func(t time.Time) getRecordInput { return getRecordInput{2, t, denom0, denom1} }
	defaultRevInputAt := func(t time.Time) getRecordInput { return getRecordInput{1, t, denom1, denom0} }
	baseRecord := newEmptyPriceRecord(1, baseTime, denom0, denom1)
	tMin1 := baseTime.Add(-time.Second)
	tMin1Record := newEmptyPriceRecord(1, tMin1, denom0, denom1)
	tPlus1 := baseTime.Add(time.Second)
	tPlus1Record := newEmptyPriceRecord(1, tPlus1, denom0, denom1)

	tests := map[string]struct {
		recordsToSet   []types.TwapRecord
		input          getRecordInput
		expectedRecord types.TwapRecord
		expErr         bool
	}{
		"no entries":            {[]types.TwapRecord{}, defaultInputAt(baseTime), baseRecord, true},
		"get at latest (exact)": {[]types.TwapRecord{baseRecord}, defaultInputAt(baseTime), baseRecord, false},
		"rev at latest (exact)": {[]types.TwapRecord{baseRecord}, defaultRevInputAt(baseTime), baseRecord, false},

		"get latest (exact) w/ past entries": {
			[]types.TwapRecord{tMin1Record, baseRecord}, defaultInputAt(baseTime), baseRecord, false},
		"get entry (exact) w/ a subsequent entry": {
			[]types.TwapRecord{tMin1Record, baseRecord}, defaultInputAt(tMin1), tMin1Record, false},
		"get sandwitched entry (exact)": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record}, defaultInputAt(baseTime), baseRecord, false},
		"rev sandwitched entry (exact)": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record}, defaultRevInputAt(baseTime), baseRecord, false},

		"get future":                 {[]types.TwapRecord{baseRecord}, defaultInputAt(tPlus1), baseRecord, false},
		"get future w/ past entries": {[]types.TwapRecord{tMin1Record, baseRecord}, defaultInputAt(tPlus1), baseRecord, false},

		"get in between entries (2 entry)": {
			[]types.TwapRecord{tMin1Record, baseRecord},
			defaultInputAt(baseTime.Add(-time.Millisecond)), tMin1Record, false},
		"get in between entries (3 entry)": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
			defaultInputAt(baseTime.Add(-time.Millisecond)), tMin1Record, false},
		"get in between entries (3 entry) #2": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
			defaultInputAt(baseTime.Add(time.Millisecond)), baseRecord, false},

		"query too old": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
			defaultInputAt(baseTime.Add(-time.Second * 2)),
			baseRecord, true},

		"non-existent pool ID": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
			wrongPoolIdInputAt(baseTime), baseRecord, true},
		"pool2 record get": {
			recordsToSet:   []types.TwapRecord{newEmptyPriceRecord(2, baseTime, denom0, denom1)},
			input:          wrongPoolIdInputAt(baseTime),
			expectedRecord: newEmptyPriceRecord(2, baseTime, denom0, denom1),
			expErr:         false},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			for _, record := range test.recordsToSet {
				s.twapkeeper.StoreNewRecord(s.Ctx, record)
			}
			record, err := s.twapkeeper.GetRecordAtOrBeforeTime(
				s.Ctx,
				test.input.poolId, test.input.t, test.input.asset0Denom, test.input.asset1Denom)
			if test.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expectedRecord, record)
		})
	}
}

// TestPruneRecordsBeforeTime tests that all twap records earlier than
// current block time - given time are pruned from the store while
// the newest record before the time to keep is preserved.
func (s *TestSuite) TestPruneRecordsBeforeTimeButNewest() {
	// These are manually created to be able to refer to them by name
	// for convenience.
	pool1BaseRecordMin2S, pool2BaseRecordMin1S, pool3BaseRecordBase, pool4BaseRecordPlus1S := s.createTestRecordsFromTime(baseTime)
	pool1Min1MsMin2S, pool2Min1MsMin1S, pool3Min1MsBase, pool4Min1MsPlus1S := s.createTestRecordsFromTime(baseTime.Add(-time.Millisecond))
	pool1Min2MsMin2S, pool2Min2MsMin1S, pool3Min2MsBase, pool4Min2MsPlus1S := s.createTestRecordsFromTime(baseTime.Add(2 * -time.Millisecond))
	pool1Min3MsMin2S, pool2Min3MsMin1S, pool3Min3MsBase, pool4Min3MsPlus1S := s.createTestRecordsFromTime(baseTime.Add(3 * -time.Millisecond))

	// Create records that match times 1:1 with other pools but have a different pool ID.
	pool5BaseRecordMin2S, pool5BaseRecordMin1S, pool5BaseRecordBase, pool5BaseRecordPlus1S := s.createTestRecordsFromTime(baseTime)
	pool5BaseRecordMin2S.PoolId = 5
	pool5BaseRecordMin1S.PoolId = 5
	pool5BaseRecordBase.PoolId = 5
	pool5BaseRecordPlus1S.PoolId = 5

	pool5Min1MsMin2S, pool5Min1MsMin1S, pool5Min1MsBase, pool5Min1MsPlus1S := s.createTestRecordsFromTime(baseTime.Add(-time.Millisecond))
	pool5Min1MsMin2S.PoolId = 5
	pool5Min1MsMin1S.PoolId = 5
	pool5Min1MsBase.PoolId = 5
	pool5Min1MsPlus1S.PoolId = 5

	tests := map[string]struct {
		// order does not follow any specific pattern
		// across many test cases on purpose.
		recordsToPreSet []types.TwapRecord

		beforeTime time.Time

		expectedKeptRecords []types.TwapRecord

		expErr bool
	}{
		"base time; across pool 3; 4 records; 3 before prune time; 2 deleted and newest kept": {
			recordsToPreSet: []types.TwapRecord{
				pool3Min1MsBase,     // base time - 1ms; kept since newest
				pool3BaseRecordBase, // base time; kept since at prune time
				pool3Min3MsBase,     // base time - 3ms; deleted
				pool3Min2MsBase,     // base time - 2ms; deleted
			},

			beforeTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{pool3Min1MsBase, pool3BaseRecordBase},
		},
		"base time - 1s - 2 ms; across pool 2; 4 records; 1 before prune time; none pruned since newest kept": {
			recordsToPreSet: []types.TwapRecord{
				pool2Min2MsMin1S,     // base time - 1s - 2ms; kept since at prune time
				pool2Min1MsMin1S,     // base time - 1s - 1ms; kept since older than at prune time
				pool2BaseRecordMin1S, // base time - 1s; kept since older than prune time
				pool2Min3MsMin1S,     // base time - 1s - 3ms; kept since newest
			},

			beforeTime: baseTime.Add(-time.Second).Add(2 * -time.Millisecond),

			expectedKeptRecords: []types.TwapRecord{
				pool2Min3MsMin1S,
				pool2Min2MsMin1S,
				pool2Min1MsMin1S,
				pool2BaseRecordMin1S,
			},
		},
		"base time - 2s - 3 ms; across pool 1; 4 records; none before prune time; none pruned": {
			recordsToPreSet: []types.TwapRecord{
				pool1Min3MsMin2S,     // base time - 2s - 3ms; kept since older than prune time
				pool1Min1MsMin2S,     // base time - 2s - 1ms; kept since older than prune time
				pool1Min2MsMin2S,     // base time - 2s - 2ms; kept since older than prune time
				pool1BaseRecordMin2S, // base time - 2s; kept since older than prune time
			},

			beforeTime: baseTime.Add(2 * -time.Second).Add(3 * -time.Millisecond),

			expectedKeptRecords: []types.TwapRecord{pool1Min3MsMin2S, pool1Min2MsMin2S, pool1Min1MsMin2S, pool1BaseRecordMin2S},
		},
		"base time + 1s + 1ms; across pool 4; 4 records; all before prune time; 3 deleted and newest kept": {
			recordsToPreSet: []types.TwapRecord{
				pool4BaseRecordPlus1S, // base time + 1s; kept since newest
				pool4Min3MsPlus1S,     // base time + 1s - 3ms; deleted
				pool4Min1MsPlus1S,     // base time + 1s -1ms; deleted
				pool4Min2MsPlus1S,     // base time + 1s - 2ms; deleted
			},

			beforeTime: baseTime.Add(time.Second).Add(time.Millisecond),

			expectedKeptRecords: []types.TwapRecord{pool4BaseRecordPlus1S},
		},
		"base time; across pool 3 and pool 5; pool 3: 4 total records; 3 before prune time; 2 deleted and newest kept. pool 5: 8 total records; 5 before prune time; 4 deleted and 1 kept": {
			recordsToPreSet: []types.TwapRecord{
				pool3Min3MsBase,     // base time - 3ms; deleted
				pool3Min2MsBase,     // base time - 2ms; deleted
				pool3Min1MsBase,     // base time - 1ms; kept since newest
				pool3BaseRecordBase, // base time; kept since at prune time

				pool5BaseRecordMin2S,  // base time - 2s; deleted
				pool5BaseRecordMin1S,  // base time - 1s; ; deleted
				pool5BaseRecordBase,   // base time; kept since at prune time
				pool5BaseRecordPlus1S, // base time + 1s; kept since older than prune time

				pool5Min1MsMin2S,  // base time - 2s - 1ms; deleted
				pool5Min1MsMin1S,  // base time - 1s - 1ms; deleted
				pool5Min1MsBase,   // base time - 1ms; kept since newest
				pool5Min1MsPlus1S, // base time + 1s - 1ms; kept since older than prune time
			},

			beforeTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{
				pool3Min1MsBase,
				pool5Min1MsBase,
				pool3BaseRecordBase,
				pool5BaseRecordBase,
				pool5Min1MsPlus1S,
				pool5BaseRecordPlus1S,
			},
		},
		"base time - 1s - 2 ms; all pools; all test records": {
			recordsToPreSet: []types.TwapRecord{
				pool3Min3MsBase,     // base time - 3ms; kept since older
				pool3Min2MsBase,     // base time - 2ms; kept since older
				pool3Min1MsBase,     // base time - 1ms; kept since older
				pool3BaseRecordBase, // base time; kept since older

				pool2Min3MsMin1S,     // base time - 1s - 3ms; kept since newest
				pool2Min2MsMin1S,     // base time - 1s - 2ms; kept since at prune time
				pool2Min1MsMin1S,     // base time - 1s - 1ms; kept since older
				pool2BaseRecordMin1S, // base time - 1s; kept since older

				pool1Min3MsMin2S,     // base time - 2s - 3ms; deleted
				pool1Min2MsMin2S,     // base time - 2s - 2ms; deleted
				pool1Min1MsMin2S,     // base time - 2s - 1ms; deleted
				pool1BaseRecordMin2S, // base time - 2s; kept since newest

				pool4Min3MsPlus1S,     // base time + 1s - 3ms; kept since older
				pool4Min2MsPlus1S,     // base time + 1s - 2ms; kept since older
				pool4Min1MsPlus1S,     // base time + 1s -1ms; kept since older
				pool4BaseRecordPlus1S, // base time + 1s; kept since older

				pool5BaseRecordMin2S,  // base time - 2s; kept since newest
				pool5BaseRecordMin1S,  // base time - 1s; kept since older
				pool5BaseRecordBase,   // base time; kept since older
				pool5BaseRecordPlus1S, // base time + 1s; kept since older

				pool5Min1MsMin2S,  // base time - 2s - 1ms; deleted
				pool5Min1MsMin1S,  // base time - 1s - 1ms; kept since older
				pool5Min1MsBase,   // base time - 1ms; kept since older
				pool5Min1MsPlus1S, // base time + 1s - 1ms; kept since older
			},

			beforeTime: baseTime.Add(-time.Second).Add(2 * -time.Millisecond),

			expectedKeptRecords: []types.TwapRecord{
				pool1BaseRecordMin2S,  // base time - 2s; kept since newest
				pool5BaseRecordMin2S,  // base time - 2s; kept since newest
				pool2Min3MsMin1S,      // base time - 1s - 3ms; kept since newest
				pool2Min2MsMin1S,      // base time - 1s - 2ms; kept since at prune time
				pool2Min1MsMin1S,      // base time - 1s - 1ms; kept since older
				pool5Min1MsMin1S,      // base time - 1s - 1ms; kept since older
				pool2BaseRecordMin1S,  // base time - 1s; kept since older
				pool5BaseRecordMin1S,  // base time - 1s; kept since older
				pool3Min3MsBase,       // base time - 3ms; kept since older
				pool3Min2MsBase,       // base time - 2ms; kept since older
				pool3Min1MsBase,       // base time - 1ms; kept since older
				pool5Min1MsBase,       // base time - 1ms; kept since older
				pool3BaseRecordBase,   // base time; kept since older
				pool5BaseRecordBase,   // base time; kept since older
				pool4Min3MsPlus1S,     // base time + 1s - 3ms; kept since older
				pool4Min2MsPlus1S,     // base time + 1s - 2ms; kept since older
				pool4Min1MsPlus1S,     // base time + 1s -1ms; kept since older
				pool5Min1MsPlus1S,     // base time + 1s - 1ms; kept since older
				pool4BaseRecordPlus1S, // base time + 1s; kept since older
				pool5BaseRecordPlus1S, // base time + 1s; kept since older
			},
		},
		"no pre-set records - no error": {
			recordsToPreSet: []types.TwapRecord{},

			beforeTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{},
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(tc.recordsToPreSet)

			ctx := s.Ctx
			twapKeeper := s.twapkeeper

			err := twapKeeper.PruneRecordsBeforeTimeButNewest(ctx, tc.beforeTime)
			if tc.expErr {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)

			s.validateExpectedRecords(tc.expectedKeptRecords)
		})
	}
}

func (s *TestSuite) TestGetAllHistoricalTimeIndexedTWAPs() {
	tests := map[string]struct {
		expectedRecords []types.TwapRecord
	}{
		"no records": {
			expectedRecords: []types.TwapRecord{},
		},
		"one record": {
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, "tokenB", "tokenA"),
			},
		},
		"multiple records": {
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, "tokenB", "tokenA"),
				newEmptyPriceRecord(2, baseTime, "tokenA", "tokenC"),
				newEmptyPriceRecord(3, baseTime, "tokenB", "tokenC"),
			},
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// Setup.
			s.SetupTest()
			ctx := s.Ctx
			twapKeeper := s.twapkeeper
			s.preSetRecords(tc.expectedRecords)

			// System under test.
			actualRecords, err := twapKeeper.GetAllHistoricalPoolIndexedTWAPs(ctx)
			s.NoError(err)

			// Assertions.
			s.Equal(tc.expectedRecords, actualRecords)
		})
	}
}

func (s *TestSuite) TestGetAllHistoricalPoolIndexedTWAPs() {
	tests := map[string]struct {
		expectedRecords []types.TwapRecord
	}{
		"no records": {
			expectedRecords: []types.TwapRecord{},
		},
		"one record": {
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, "tokenB", "tokenA"),
			},
		},
		"multiple records": {
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, "tokenB", "tokenA"),
				newEmptyPriceRecord(2, baseTime, "tokenA", "tokenC"),
				newEmptyPriceRecord(3, baseTime, "tokenB", "tokenC"),
			},
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// Setup.
			s.SetupTest()
			ctx := s.Ctx
			twapKeeper := s.twapkeeper
			s.preSetRecords(tc.expectedRecords)

			// System under test.
			actualRecords, err := twapKeeper.GetAllHistoricalPoolIndexedTWAPs(ctx)
			s.NoError(err)

			// Assertions.
			s.Equal(tc.expectedRecords, actualRecords)
		})
	}
}
