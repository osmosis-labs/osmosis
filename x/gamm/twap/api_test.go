package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/gamm/twap/types"
)

func (s *TestSuite) TestGetBeginBlockAccumulatorRecord() {
	poolId, denomA, denomB := s.setupDefaultPool()
	initStartRecord := newRecord(s.Ctx.BlockTime(), sdk.OneDec(), sdk.ZeroDec(), sdk.ZeroDec())
	initStartRecord.PoolId, initStartRecord.Height = poolId, s.Ctx.BlockHeight()
	initStartRecord.Asset0Denom, initStartRecord.Asset1Denom = denomA, denomB

	zeroAccumTenPoint1Record := recordWithUpdatedSpotPrice(initStartRecord, sdk.NewDec(10), sdk.NewDecWithPrec(1, 1))

	blankRecord := types.TwapRecord{}
	defaultTime := s.Ctx.BlockTime()

	tPlusOneSec := defaultTime.Add(time.Second)

	tests := map[string]struct {
		// if start record is blank, don't do any sets
		startRecord types.TwapRecord
		// We set it to have the updated time
		expRecord  types.TwapRecord
		time       time.Time
		poolId     uint64
		quoteDenom string
		baseDenom  string
		expError   bool
	}{
		"no record (wrong pool ID)": {blankRecord, blankRecord, defaultTime, 4, denomA, denomB, true},
		"default record":            {blankRecord, initStartRecord, defaultTime, 1, denomA, denomB, false},
		"one second later record":   {blankRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOneSec, 1, denomA, denomB, false},
		"idempotent overwrite":      {initStartRecord, initStartRecord, defaultTime, 1, denomA, denomB, false},
		"idempotent overwrite2":     {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOneSec, 1, denomA, denomB, false},
		"diff spot price": {zeroAccumTenPoint1Record,
			recordWithUpdatedAccum(zeroAccumTenPoint1Record, OneSec.MulInt64(10), OneSec.QuoInt64(10)),
			tPlusOneSec, 1, denomA, denomB, false},
		// TODO: Overflow
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// setup time
			s.Ctx = s.Ctx.WithBlockTime(tc.time)
			tc.expRecord.Time = tc.time

			// setup record
			initSetRecord := tc.startRecord
			if (tc.startRecord == types.TwapRecord{}) {
				initSetRecord = initStartRecord
			}
			s.twapkeeper.StoreNewRecord(s.Ctx, initSetRecord)

			actualRecord, err := s.twapkeeper.GetBeginBlockAccumulatorRecord(s.Ctx, tc.poolId, tc.baseDenom, tc.quoteDenom)

			if tc.expError {
				s.Require().Error(err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(tc.expRecord, actualRecord)
		})
	}
}

func (s *TestSuite) TestGetArithmeticTwap() {
	type getTwapInput struct {
		poolId          uint64
		quoteAssetDenom string
		baseAssetDenom  string
		startTime       time.Time
		endTime         time.Time
	}

	makeSimpleTwapInput := func(startTime time.Time, endTime time.Time, isQuoteTokenA bool) getTwapInput {
		quoteAssetDenom, baseAssetDenom := denom0, denom1
		if isQuoteTokenA {
			baseAssetDenom, quoteAssetDenom = quoteAssetDenom, baseAssetDenom
		}
		return getTwapInput{1, quoteAssetDenom, baseAssetDenom, startTime, endTime}
	}

	quoteAssetA := true
	quoteAssetB := false
	tPlusOne := baseTime.Add(time.Second)
	// base record is a record with t=baseTime, sp0=10, sp1=.1, accumulators set to 0
	baseRecord := newTwapRecordWithDefaults(baseTime, sdk.NewDec(10), sdk.ZeroDec(), sdk.ZeroDec())

	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        getTwapInput
		expTwap      sdk.Dec
		expErrorStr  string
	}{
		"(single record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(single record) start and end point to same record, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(1, 1),
		},

		// error catching
		"end time in future": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetA),
			expErrorStr:  "future",
		},
		"start time after end time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(tPlusOne, baseTime, quoteAssetA),
			expErrorStr:  "after",
		},
		"start time too old (end time = now)": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Hour), baseTime, quoteAssetA),
			expErrorStr:  "too old",
		},
		"start time too old": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Second),
			input:        makeSimpleTwapInput(baseTime.Add(-time.Hour), baseTime, quoteAssetA),
			expErrorStr:  "too old",
		},
		// TODO: overflow tests, multi-asset pool handling, make more record interpolation cases
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			for _, record := range test.recordsToSet {
				s.twapkeeper.StoreNewRecord(s.Ctx, record)
			}
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)
			twap, err := s.twapkeeper.GetArithmeticTwap(s.Ctx, test.input.poolId,
				test.input.quoteAssetDenom, test.input.baseAssetDenom,
				test.input.startTime, test.input.endTime)
			if test.expErrorStr != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), test.expErrorStr)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expTwap, twap)
		})
	}
}

// func (s *TestSuite) TestGetArithmeticTwapToNow() {
// 	tests := map[string]struct {
// 		// if start record is blank, don't do any sets
// 		setupRecords []types.TwapRecord
// 		latestRecord types.TwapRecord
// 		// We set it to have the updated time
// 		expRecord  types.TwapRecord
// 		time       time.Time
// 		poolId     uint64
// 		quoteDenom string
// 		baseDenom  string
// 		expError   bool
// 	}{}
// 	for name, tc := range tests {
// 		s.Run(name, func() {

// 		})
// 	}
// }
