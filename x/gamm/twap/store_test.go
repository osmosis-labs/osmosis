package twap_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

// TestTrackChangedPool takes a list of poolIds as test cases, and runs one list per block.
// Every simulated block, checks that there no changed pools.
// Then runs k.trackChangedPool on every item in the test case list.
// Then checks that changed pools returns the list, deduplicated.
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
	baseTime := time.Unix(1257894000, 0).UTC()
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
		"settwo records, reverse order": {
			// The last record, independent of time, takes precedence for most recent.
			[]types.TwapRecord{tPlusOneRecord, baseRecord},
			1,
			[]types.TwapRecord{baseRecord},
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
