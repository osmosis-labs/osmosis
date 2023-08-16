package twap_test

import (
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v17/x/twap"

	gammtypes "github.com/osmosis-labs/osmosis/v17/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v17/x/twap/types"
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
		"lexicographic fooling pool Ids": {
			recordsToSet: []types.TwapRecord{
				withPoolId(baseRecord, 1),
				withPoolId(baseRecord, 2),
				withPoolId(baseRecord, 10),
				withPoolId(baseRecord, 11),
				withPoolId(baseRecord, 20),
			},
			poolId:          1,
			expectedRecords: []types.TwapRecord{baseRecord},
		},
		"lexicographic fooling pool Ids, with carry": {
			recordsToSet: []types.TwapRecord{
				withPoolId(baseRecord, 9),
				withPoolId(baseRecord, 10),
				withPoolId(baseRecord, 11),
				withPoolId(baseRecord, 19),
				withPoolId(baseRecord, 90),
			},
			poolId:          9,
			expectedRecords: []types.TwapRecord{withPoolId(baseRecord, 9)},
		},
		"set multi-asset pool record": {
			recordsToSet: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
				newEmptyPriceRecord(1, baseTime, denom0, denom2),
				newEmptyPriceRecord(1, baseTime, denom1, denom2),
			},
			poolId: 1,
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
				newEmptyPriceRecord(1, baseTime, denom0, denom2),
				newEmptyPriceRecord(1, baseTime, denom1, denom2),
			},
		},
		"set multi-asset pool record - reverse order": {
			recordsToSet: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom2),
				newEmptyPriceRecord(1, baseTime, denom1, denom2),
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
			},
			poolId: 1,
			expectedRecords: []types.TwapRecord{
				newEmptyPriceRecord(1, baseTime, denom0, denom1),
				newEmptyPriceRecord(1, baseTime, denom0, denom2),
				newEmptyPriceRecord(1, baseTime, denom1, denom2),
			},
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
	baseRecord := withPrice0Set(newEmptyPriceRecord(1, baseTime, denom0, denom1), sdk.OneDec())
	tMin1 := baseTime.Add(-time.Second)
	tMin1Record := withPrice0Set(newEmptyPriceRecord(1, tMin1, denom0, denom1), sdk.OneDec())
	tPlus1 := baseTime.Add(time.Second)
	tPlus1Record := withPrice0Set(newEmptyPriceRecord(1, tPlus1, denom0, denom1), sdk.OneDec())

	tests := map[string]struct {
		recordsToSet   []types.TwapRecord
		input          getRecordInput
		expectedRecord types.TwapRecord
		expErr         error
	}{
		"no entries": {[]types.TwapRecord{}, defaultInputAt(baseTime), baseRecord, fmt.Errorf(
			"getTwapRecord: querying for assets %s %s that are not in pool id %d",
			baseRecord.Asset0Denom, baseRecord.Asset1Denom, 1)},
		"get at latest (exact)": {[]types.TwapRecord{baseRecord}, defaultInputAt(baseTime), baseRecord, nil},
		"rev at latest (exact)": {[]types.TwapRecord{baseRecord}, defaultRevInputAt(baseTime), baseRecord, nil},

		"get latest (exact) w/ past entries": {
			[]types.TwapRecord{tMin1Record, baseRecord}, defaultInputAt(baseTime), baseRecord, nil,
		},
		"get entry (exact) w/ a subsequent entry": {
			[]types.TwapRecord{tMin1Record, baseRecord}, defaultInputAt(tMin1), tMin1Record, nil,
		},
		"get sandwitched entry (exact)": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record}, defaultInputAt(baseTime), baseRecord, nil,
		},
		"rev sandwitched entry (exact)": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record}, defaultRevInputAt(baseTime), baseRecord, nil,
		},

		"get future":                 {[]types.TwapRecord{baseRecord}, defaultInputAt(tPlus1), baseRecord, nil},
		"get future w/ past entries": {[]types.TwapRecord{tMin1Record, baseRecord}, defaultInputAt(tPlus1), baseRecord, nil},

		"get in between entries (2 entry)": {
			[]types.TwapRecord{tMin1Record, baseRecord},
			defaultInputAt(baseTime.Add(-time.Millisecond)), tMin1Record, nil,
		},
		"get in between entries (3 entry)": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
			defaultInputAt(baseTime.Add(-time.Millisecond)), tMin1Record, nil,
		},
		"get in between entries (3 entry) #2": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
			defaultInputAt(baseTime.Add(time.Millisecond)), baseRecord, nil,
		},

		"query too old": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
			defaultInputAt(baseTime.Add(-time.Second * 2)),
			baseRecord,
			twap.TimeTooOldError{Time: baseTime.Add(-time.Second * 2)},
		},

		"non-existent pool ID": {
			[]types.TwapRecord{tMin1Record, baseRecord, tPlus1Record},
			wrongPoolIdInputAt(baseTime), baseRecord, fmt.Errorf(
				"getTwapRecord: querying for assets %s %s that are not in pool id %d",
				baseRecord.Asset0Denom, baseRecord.Asset1Denom, 2),
		},
		"pool2 record get": {
			recordsToSet:   []types.TwapRecord{newEmptyPriceRecord(2, baseTime, denom0, denom1)},
			input:          wrongPoolIdInputAt(baseTime),
			expectedRecord: newEmptyPriceRecord(2, baseTime, denom0, denom1),
			expErr:         nil,
		},
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
			if test.expErr != nil {
				s.Require().Equal(test.expErr, err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expectedRecord, record)
		})
	}
}

// TestPruneRecordsBeforeTime tests that all twap records earlier than
// current block time - given time are pruned from the store while
// the newest record for each pool before the time to keep is preserved.
func (s *TestSuite) TestPruneRecordsBeforeTimeButNewest() {
	// N.B.: the records follow the following naming convention:
	// <pool id><delta from base time in seconds><delta from base time in milliseconds>
	// These are manually created to be able to refer to them by name
	// for convenience.

	// Create 6 records of 4 pools from base time, each in different pool with the difference of 1 second between them. Pool 2 is a 3 asset pool.
	pool1Min2SBaseMs, pool2Min1SBaseMsAB, pool2Min1SBaseMsAC, pool2Min1SBaseMsBC, pool3BaseSecBaseMs, pool4Plus1SBaseMs := s.createTestRecordsFromTime(baseTime)

	// Create 6 records of 4 pools from base time - 1 ms, each in different pool with the difference of 1 second between them. Pool 2 is a 3 asset pool.
	pool1Min2SMin1Ms, pool2Min1SMin1MsAB, pool2Min1SMin1MsAC, pool2Min1SMin1MsBC, pool3BaseSecMin1Ms, pool4Plus1SMin1Ms := s.createTestRecordsFromTime(baseTime.Add(-time.Millisecond))

	// Create 6 records of 4 pools from base time - 2 ms, each in different pool with the difference of 1 second between them. Pool 2 is a 3 asset pool.
	pool1Min2SMin2Ms, pool2Min1SMin2MsAB, pool2Min1SMin2MsAC, pool2Min1SMin2MsBC, pool3BaseSecMin2Ms, pool4Plus1SMin2Ms := s.createTestRecordsFromTime(baseTime.Add(2 * -time.Millisecond))

	// Create 6 records of 4 pools from base time - 3 ms, each in different pool with the difference of 1 second between them. Pool 2 is a 3 asset pool.
	pool1Min2SMin3Ms, pool2Min1SMin3MsAB, pool2Min1SMin3MsAC, pool2Min1SMin3MsBC, pool3BaseSecMin3Ms, pool4Plus1SMin3Ms := s.createTestRecordsFromTime(baseTime.Add(3 * -time.Millisecond))

	// Create 12 records in the same pool from base time , each record with the difference of 1 second between them.
	pool5Min2SBaseMsAB, pool5Min2SBaseMsAC, pool5Min2SBaseMsBC,
		pool5Min1SBaseMsAB, pool5Min1SBaseMsAC, pool5Min1SBaseMsBC,
		pool5BaseSecBaseMsAB, pool5BaseSecBaseMsAC, pool5BaseSecBaseMsBC,
		pool5Plus1SBaseMsAB, pool5Plus1SBaseMsAC, pool5Plus1SBaseMsBC := s.CreateTestRecordsFromTimeInPool(baseTime, 5)

	// Create 12 records in the same pool from base time - 1 ms, each record with the difference of 1 second between them
	pool5Min2SMin1MsAB, pool5Min2SMin1MsAC, pool5Min2SMin1MsBC,
		pool5Min1SMin1MsAB, pool5Min1SMin1MsAC, pool5Min1SMin1MsBC,
		pool5BaseSecMin1MsAB, pool5BaseSecMin1MsAC, pool5BaseSecMin1MsBC,
		pool5Plus1SMin1MsAB, pool5Plus1SMin1MsAC, pool5Plus1SMin1MsBC := s.CreateTestRecordsFromTimeInPool(baseTime.Add(-time.Millisecond), 5)

	tests := map[string]struct {
		// order does not follow any specific pattern
		// across many test cases on purpose.
		recordsToPreSet []types.TwapRecord

		lastKeptTime time.Time

		expectedKeptRecords []types.TwapRecord
	}{
		"base time; across pool 3; 4 records; 3 before lastKeptTime; 2 deleted and newest kept": {
			recordsToPreSet: []types.TwapRecord{
				pool3BaseSecMin1Ms, // base time - 1ms; kept since newest before lastKeptTime
				pool3BaseSecBaseMs, // base time; kept since at lastKeptTime
				pool3BaseSecMin3Ms, // base time - 3ms; deleted
				pool3BaseSecMin2Ms, // base time - 2ms; deleted
			},

			lastKeptTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{pool3BaseSecMin1Ms, pool3BaseSecBaseMs},
		},
		"base time - 1s - 2 ms; across pool 2; 12 records; 3 before lastKeptTime; none pruned since newest kept": {
			recordsToPreSet: []types.TwapRecord{
				pool2Min1SMin2MsAB, pool2Min1SMin2MsAC, pool2Min1SMin2MsBC, // base time - 1s - 2ms; kept since at lastKeptTime
				pool2Min1SMin1MsAB, pool2Min1SMin1MsAC, pool2Min1SMin1MsBC, // base time - 1s - 1ms; kept since older than at lastKeptTime
				pool2Min1SBaseMsAB, pool2Min1SBaseMsAC, pool2Min1SBaseMsBC, // base time - 1s; kept since older than lastKeptTime
				pool2Min1SMin3MsAB, pool2Min1SMin3MsAC, pool2Min1SMin3MsBC, // base time - 1s - 3ms; kept since newest before lastKeptTime
			},

			lastKeptTime: baseTime.Add(-time.Second).Add(2 * -time.Millisecond),

			expectedKeptRecords: []types.TwapRecord{
				pool2Min1SMin3MsAB, pool2Min1SMin3MsAC, pool2Min1SMin3MsBC,
				pool2Min1SMin2MsAB, pool2Min1SMin2MsAC, pool2Min1SMin2MsBC,
				pool2Min1SMin1MsAB, pool2Min1SMin1MsAC, pool2Min1SMin1MsBC,
				pool2Min1SBaseMsAB, pool2Min1SBaseMsAC, pool2Min1SBaseMsBC,
			},
		},
		"base time - 2s - 3 ms; across pool 1; 4 records; none before lastKeptTime; none pruned": {
			recordsToPreSet: []types.TwapRecord{
				pool1Min2SMin3Ms, // base time - 2s - 3ms; kept since older than lastKeptTime
				pool1Min2SMin1Ms, // base time - 2s - 1ms; kept since older than lastKeptTime
				pool1Min2SMin2Ms, // base time - 2s - 2ms; kept since older than lastKeptTime
				pool1Min2SBaseMs, // base time - 2s; kept since older than lastKeptTime
			},

			lastKeptTime: baseTime.Add(2 * -time.Second).Add(3 * -time.Millisecond),

			expectedKeptRecords: []types.TwapRecord{pool1Min2SMin3Ms, pool1Min2SMin2Ms, pool1Min2SMin1Ms, pool1Min2SBaseMs},
		},
		"base time + 1s + 1ms; across pool 4; 4 records; all before lastKeptTime; 3 deleted and newest kept": {
			recordsToPreSet: []types.TwapRecord{
				pool4Plus1SBaseMs, // base time + 1s; kept since newest before lastKeptTime
				pool4Plus1SMin3Ms, // base time + 1s - 3ms; deleted
				pool4Plus1SMin1Ms, // base time + 1s -1ms; deleted
				pool4Plus1SMin2Ms, // base time + 1s - 2ms; deleted
			},

			lastKeptTime: baseTime.Add(time.Second).Add(time.Millisecond),

			expectedKeptRecords: []types.TwapRecord{pool4Plus1SBaseMs},
		},
		"base time; across pool 3 and pool 5; pool 3: 4 total records; 3 before lastKeptTime; 2 deleted and newest 2 kept. pool 5: 24 total records; 12 before lastKeptTime; 12 deleted and 12 kept": {
			recordsToPreSet: []types.TwapRecord{
				pool3BaseSecMin3Ms, // base time - 3ms; deleted
				pool3BaseSecMin2Ms, // base time - 2ms; deleted
				pool3BaseSecMin1Ms, // base time - 1ms; kept since newest before lastKeptTime
				pool3BaseSecBaseMs, // base time; kept since at lastKeptTime

				pool5Min2SBaseMsAB, pool5Min2SBaseMsAC, pool5Min2SBaseMsBC, // base time - 2s; deleted
				pool5Min1SBaseMsAB, pool5Min1SBaseMsAC, pool5Min1SBaseMsBC, // base time - 1s; ; deleted
				pool5BaseSecBaseMsAB, pool5BaseSecBaseMsAC, pool5BaseSecBaseMsBC, // base time; kept since at lastKeptTime
				pool5Plus1SBaseMsAB, pool5Plus1SBaseMsAC, pool5Plus1SBaseMsBC, // base time + 1s; kept since older than lastKeptTime

				pool5Min2SMin1MsAB, pool5Min2SMin1MsAC, pool5Min2SMin1MsBC, // base time - 2s - 1ms; deleted
				pool5Min1SMin1MsAB, pool5Min1SMin1MsAC, pool5Min1SMin1MsBC, // base time - 1s - 1ms; deleted
				pool5BaseSecMin1MsAB, pool5BaseSecMin1MsAC, pool5BaseSecMin1MsBC, // base time - 1ms; kept since newest before lastKeptTime
				pool5Plus1SMin1MsAB, pool5Plus1SMin1MsAC, pool5Plus1SMin1MsBC, // base time + 1s - 1ms; kept since older than lastKeptTime
			},

			lastKeptTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{
				pool3BaseSecMin1Ms,
				pool5BaseSecMin1MsAB, pool5BaseSecMin1MsAC, pool5BaseSecMin1MsBC,
				pool3BaseSecBaseMs,
				pool5BaseSecBaseMsAB, pool5BaseSecBaseMsAC, pool5BaseSecBaseMsBC,
				pool5Plus1SMin1MsAB, pool5Plus1SMin1MsAC, pool5Plus1SMin1MsBC,
				pool5Plus1SBaseMsAB, pool5Plus1SBaseMsAC, pool5Plus1SBaseMsBC,
			},
		},
		"base time - 1s - 2 ms; all pools; all test records": {
			recordsToPreSet: []types.TwapRecord{
				pool3BaseSecMin3Ms, // base time - 3ms; kept since older
				pool3BaseSecMin2Ms, // base time - 2ms; kept since older
				pool3BaseSecMin1Ms, // base time - 1ms; kept since older
				pool3BaseSecBaseMs, // base time; kept since older

				pool2Min1SMin3MsAB, pool2Min1SMin3MsAC, pool2Min1SMin3MsBC, // base time - 1s - 3ms; kept since newest before lastKeptTime
				pool2Min1SMin2MsAB, pool2Min1SMin2MsAC, pool2Min1SMin2MsBC, // base time - 1s - 2ms; kept since at lastKeptTime
				pool2Min1SMin1MsAB, pool2Min1SMin1MsAC, pool2Min1SMin1MsBC, // base time - 1s - 1ms; kept since older
				pool2Min1SBaseMsAB, pool2Min1SBaseMsAC, pool2Min1SBaseMsBC, // base time - 1s; kept since older

				pool1Min2SMin3Ms, // base time - 2s - 3ms; deleted
				pool1Min2SMin2Ms, // base time - 2s - 2ms; deleted
				pool1Min2SMin1Ms, // base time - 2s - 1ms; deleted
				pool1Min2SBaseMs, // base time - 2s; kept since newest before lastKeptTime

				pool4Plus1SMin3Ms, // base time + 1s - 3ms; kept since older
				pool4Plus1SMin2Ms, // base time + 1s - 2ms; kept since older
				pool4Plus1SMin1Ms, // base time + 1s -1ms; kept since older
				pool4Plus1SBaseMs, // base time + 1s; kept since older

				pool5Min2SBaseMsAB, pool5Min2SBaseMsAC, pool5Min2SBaseMsBC, // base time - 2s; kept since newest before lastKeptTime
				pool5Min1SBaseMsAB, pool5Min1SBaseMsAC, pool5Min1SBaseMsBC, // base time - 1s; kept since older
				pool5BaseSecBaseMsAB, pool5BaseSecBaseMsAC, pool5BaseSecBaseMsBC, // base time; kept since older
				pool5Plus1SBaseMsAB, pool5Plus1SBaseMsAC, pool5Plus1SBaseMsBC, // base time + 1s; kept since older

				pool5Min2SMin1MsAB, pool5Min2SMin1MsAC, pool5Min2SMin1MsBC, // base time - 2s - 1ms; deleted
				pool5Min1SMin1MsAB, pool5Min1SMin1MsAC, pool5Min1SMin1MsBC, // base time - 1s - 1ms; kept since older
				pool5BaseSecMin1MsAB, pool5BaseSecMin1MsAC, pool5BaseSecMin1MsBC, // base time - 1ms; kept since older
				pool5Plus1SMin1MsAB, pool5Plus1SMin1MsAC, pool5Plus1SMin1MsBC, // base time + 1s - 1ms; kept since older
			},

			lastKeptTime: baseTime.Add(-time.Second).Add(2 * -time.Millisecond),

			expectedKeptRecords: []types.TwapRecord{
				pool1Min2SBaseMs,                                           // base time - 2s; kept since newest before lastKeptTime
				pool5Min2SBaseMsAB, pool5Min2SBaseMsAC, pool5Min2SBaseMsBC, // base time - 2s; kept since newest before lastKeptTime
				pool2Min1SMin3MsAB, pool2Min1SMin3MsAC, pool2Min1SMin3MsBC, // base time - 1s - 3ms; kept since newest before lastKeptTime
				pool2Min1SMin2MsAB, pool2Min1SMin2MsAC, pool2Min1SMin2MsBC, // base time - 1s - 2ms; kept since at lastKeptTime
				pool2Min1SMin1MsAB, pool2Min1SMin1MsAC, pool2Min1SMin1MsBC, // base time - 1s - 1ms; kept since older
				pool5Min1SMin1MsAB, pool5Min1SMin1MsAC, pool5Min1SMin1MsBC, // base time - 1s - 1ms; kept since older
				pool2Min1SBaseMsAB, pool2Min1SBaseMsAC, pool2Min1SBaseMsBC, // base time - 1s; kept since older
				pool5Min1SBaseMsAB, pool5Min1SBaseMsAC, pool5Min1SBaseMsBC, // base time - 1s; kept since older
				pool3BaseSecMin3Ms,                                               // base time - 3ms; kept since older
				pool3BaseSecMin2Ms,                                               // base time - 2ms; kept since older
				pool3BaseSecMin1Ms,                                               // base time - 1ms; kept since older
				pool5BaseSecMin1MsAB, pool5BaseSecMin1MsAC, pool5BaseSecMin1MsBC, // base time - 1ms; kept since older
				pool3BaseSecBaseMs,                                               // base time; kept since older
				pool5BaseSecBaseMsAB, pool5BaseSecBaseMsAC, pool5BaseSecBaseMsBC, // base time; kept since older
				pool4Plus1SMin3Ms,                                             // base time + 1s - 3ms; kept since older
				pool4Plus1SMin2Ms,                                             // base time + 1s - 2ms; kept since older
				pool4Plus1SMin1Ms,                                             // base time + 1s -1ms; kept since older
				pool5Plus1SMin1MsAB, pool5Plus1SMin1MsAC, pool5Plus1SMin1MsBC, // base time + 1s - 1ms; kept since older
				pool4Plus1SBaseMs,                                             // base time + 1s; kept since older
				pool5Plus1SBaseMsAB, pool5Plus1SBaseMsAC, pool5Plus1SBaseMsBC, // base time + 1s; kept since older
			},
		},
		"no pre-set records - no error": {
			recordsToPreSet: []types.TwapRecord{},

			lastKeptTime: baseTime,

			expectedKeptRecords: []types.TwapRecord{},
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(tc.recordsToPreSet)

			ctx := s.Ctx
			twapKeeper := s.twapkeeper

			err := twapKeeper.PruneRecordsBeforeTimeButNewest(ctx, tc.lastKeptTime)
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

func (s *TestSuite) TestAccumulatorOverflow() {
	maxSpotPrice := gammtypes.MaxSpotPrice
	tests := map[string]struct {
		// timeDelta is duration in nano seconds.
		// we use sdk.Dec here because time.Duration would automatically cap to
		// time.duration.maxDuration without erroring.
		timeDelta sdk.Dec
		panics    bool
	}{
		"no overflow": {
			// 2562047h47m16.854775807s in duration, this is over 292 years.
			timeDelta: sdk.NewDec(2).Power(128),
			panics:    false,
		},
		"overflow": {
			timeDelta: sdk.NewDec(2).Power(129),
			panics:    true,
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()

			var accumulatorVal sdk.Dec

			fmt.Println(time.Duration(math.Pow(2, 128)))
			if test.panics {
				s.Require().Panics(func() {
					// accumulator value is calculated via spot price * time delta
					accumulatorVal = maxSpotPrice.Mul(test.timeDelta)
				})
			} else {
				twapRecordToStore := types.TwapRecord{
					PoolId:                      basePoolId,
					Asset0Denom:                 denom0,
					Asset1Denom:                 denom1,
					P0ArithmeticTwapAccumulator: accumulatorVal,
				}

				s.twapkeeper.StoreNewRecord(s.Ctx, twapRecordToStore)
			}
		})
	}
}

func (s *TestSuite) TestGetAllHistoricalPoolIndexedTWAPsForPooId() {
	baseRecord := newEmptyPriceRecord(1, baseTime, denom0, denom1)
	tPlusOneRecord := newEmptyPriceRecord(1, tPlusOne, denom0, denom1)
	tests := map[string]struct {
		recordsToSet    []types.TwapRecord
		poolId          uint64
		expectedRecords []types.TwapRecord
	}{
		"set single record": {
			poolId:          1,
			expectedRecords: []types.TwapRecord{baseRecord},
		},
		"query non-existent pool": {
			poolId:          2,
			expectedRecords: []types.TwapRecord{},
		},
		"set single record, different pool ID": {
			poolId:          2,
			expectedRecords: []types.TwapRecord{newEmptyPriceRecord(2, baseTime, denom0, denom1)},
		},
		"set two records": {
			poolId:          1,
			expectedRecords: []types.TwapRecord{baseRecord, tPlusOneRecord},
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			twapKeeper := s.twapkeeper
			s.preSetRecords(test.expectedRecords)

			// System under test.
			actualRecords, err := twapKeeper.GetAllHistoricalPoolIndexedTWAPsForPoolId(s.Ctx, test.poolId)
			s.NoError(err)

			// Assertions.
			s.Equal(test.expectedRecords, actualRecords)

		})
	}

}
