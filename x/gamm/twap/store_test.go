package twap_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
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
	tPlusOne := baseTime.Add(time.Second)
	baseRecord := newEmptyPriceRecord(1, baseTime, "tokenB", "tokenA")
	tPlusOneRecord := newEmptyPriceRecord(1, tPlusOne, "tokenB", "tokenA")
	tests := map[string]struct {
		recordsToSet    []types.TwapRecord
		poolId          uint64
		expectedRecords []types.TwapRecord
	}{
		"set single record": {
			[]types.TwapRecord{baseRecord},
			1,
			[]types.TwapRecord{baseRecord},
		},
		"query non-existent pool": {
			[]types.TwapRecord{baseRecord},
			2,
			[]types.TwapRecord{},
		},
		"set two records": {
			[]types.TwapRecord{baseRecord, tPlusOneRecord},
			1,
			[]types.TwapRecord{tPlusOneRecord},
		},
		"set two records, reverse order": {
			// The last record, independent of time, takes precedence for most recent.
			[]types.TwapRecord{tPlusOneRecord, baseRecord},
			1,
			[]types.TwapRecord{baseRecord},
		},
		"set multi-asset pool record": {
			[]types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, "tokenB", "tokenA"),
				newEmptyPriceRecord(1, baseTime, "tokenC", "tokenB"),
				newEmptyPriceRecord(1, baseTime, "tokenC", "tokenA")},
			1,
			[]types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, "tokenB", "tokenA"),
				newEmptyPriceRecord(1, baseTime, "tokenC", "tokenA"),
				newEmptyPriceRecord(1, baseTime, "tokenC", "tokenB")},
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
	defaultInputAt := func(t time.Time) getRecordInput { return getRecordInput{1, t, "tokenB", "tokenA"} }
	wrongPoolIdInputAt := func(t time.Time) getRecordInput { return getRecordInput{2, t, "tokenB", "tokenA"} }
	defaultRevInputAt := func(t time.Time) getRecordInput { return getRecordInput{1, t, "tokenA", "tokenB"} }
	baseRecord := newEmptyPriceRecord(1, baseTime, "tokenB", "tokenA")
	tMin1 := baseTime.Add(-time.Second)
	tMin1Record := newEmptyPriceRecord(1, tMin1, "tokenB", "tokenA")
	tPlus1 := baseTime.Add(time.Second)
	tPlus1Record := newEmptyPriceRecord(1, tPlus1, "tokenB", "tokenA")

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
