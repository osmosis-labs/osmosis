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
				newEmptyPriceRecord(1, baseTime, denom2, denom0),
				newEmptyPriceRecord(1, baseTime, denom2, denom1)},
			poolId: 1,
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
				newEmptyPriceRecord(1, baseTime, denom2, denom1),
				newEmptyPriceRecord(1, baseTime, denom2, denom0)},
		},
		"set multi-asset pool record - reverse order": {
			recordsToSet: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom2, denom0),
				newEmptyPriceRecord(1, baseTime, denom2, denom1),
				newEmptyPriceRecord(1, baseTime, denom0, denom1)},
			poolId: 1,
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
				newEmptyPriceRecord(1, baseTime, denom2, denom1),
				newEmptyPriceRecord(1, baseTime, denom2, denom0)},
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
		"rev at latest (exact)": {[]types.TwapRecord{baseRecord}, defaultRevInputAt(baseTime), baseRecord, true},

		"get latest (exact) w/ past entries": {
			[]types.TwapRecord{tMin1Record, baseRecord}, defaultInputAt(baseTime), baseRecord, false},
		"get entry (exact) w/ a subsequent entry": {
			[]types.TwapRecord{tMin1Record, baseRecord}, defaultInputAt(tMin1), tMin1Record, false},
		"get sandwitched entry (exact)": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record}, defaultInputAt(baseTime), baseRecord, false},
		"rev sandwitched entry (exact)": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record}, defaultRevInputAt(baseTime), baseRecord, true},

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
// current block time - given time are pruned from the store.
func (s *TestSuite) TestPruneRecordsBeforeTime() {
	tMin2Record, tMin1Record, baseRecord, tPlus1Record := s.createTestRecordsFromTime(baseTime)

	// non-ascending insertion order.
	allTestRecords := []types.TwapRecord{tPlus1Record, tMin1Record, baseRecord, tMin2Record}

	tests := map[string]struct {
		recordsToPreSet []types.TwapRecord

		beforeTime time.Time

		expectedKeptRecords []types.TwapRecord

		expErr bool
	}{
		"base time, 1 record before base time (deleted)": {
			recordsToPreSet: []types.TwapRecord{tMin1Record},

			beforeTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{},
		},
		"base time, 2 records before base time (both deleted)": {
			recordsToPreSet: []types.TwapRecord{tMin1Record, tMin2Record},

			beforeTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{},
		},
		"base time, 1 record at base time (not deleted)": {
			recordsToPreSet: []types.TwapRecord{baseRecord},

			beforeTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{baseRecord},
		},
		"base time, 1 record after base time (not deleted)": {
			recordsToPreSet: []types.TwapRecord{tPlus1Record},

			beforeTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{tPlus1Record},
		},
		"base time minus 1, 2 records before (deleted), 1 records at (deleted), 1 records after (not deleted)": {
			recordsToPreSet: allTestRecords,

			beforeTime: tMin1Record.Time,

			expectedKeptRecords: []types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
		},
		"base time minus 2 - 0 records before - all kept": {
			recordsToPreSet: allTestRecords,

			beforeTime: tMin2Record.Time,

			expectedKeptRecords: []types.TwapRecord{tMin2Record, tMin1Record, baseRecord, tPlus1Record},
		},
		"base time plus 2 - all records before - all deleted": {
			recordsToPreSet: allTestRecords,

			beforeTime: tPlus1Record.Time.Add(time.Second),

			expectedKeptRecords: []types.TwapRecord{},
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

			err := twapKeeper.PruneRecordsBeforeTime(ctx, tc.beforeTime)
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
