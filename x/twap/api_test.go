package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v10/x/twap/types"
)

type getTwapInput struct {
	poolId          uint64
	quoteAssetDenom string
	baseAssetDenom  string
	startTime       time.Time
	endTime         time.Time
}

var (
	ThreePlusOneThird sdk.Dec = sdk.MustNewDecFromStr("3.333333333333333333")

	// base record is a record with t=baseTime, sp0=10(sp1=0.1) accumulators set to 0
	baseRecord types.TwapRecord = newTwapRecordWithDefaults(baseTime, sdk.NewDec(10), sdk.ZeroDec(), sdk.ZeroDec())

	// accum0 = 10 seconds * (spot price = 10) = OneSec * 10 * 10
	// accum1 = 10 seconds * (spot price = 0.1) = OneSec
	accum0, accum1 sdk.Dec = OneSec.MulInt64(10 * 10), OneSec

	// accumulators updated from baseRecord with
	// t = baseTime + 10
	// sp0 = 5, sp1 = 0.2
	tPlus10sp5Record = newTwapRecordWithDefaults(
		baseTime.Add(10*time.Second), sdk.NewDec(5), accum0, accum1)

	// accumulators updated from tPlus10sp5Record with
	// t = baseTime + 20
	// sp0 = 2, sp1 = 0.5
	tPlus20sp2Record = newTwapRecordWithDefaults(
		baseTime.Add(20*time.Second), sdk.NewDec(2), OneSec.MulInt64(10*10+5*10), OneSec.MulInt64(3))
)

func (s *TestSuite) TestGetBeginBlockAccumulatorRecord() {
	poolId, denomA, denomB := s.setupDefaultPool()
	initStartRecord := newRecord(s.Ctx.BlockTime(), sdk.OneDec(), sdk.ZeroDec(), sdk.ZeroDec())
	initStartRecord.PoolId, initStartRecord.Height = poolId, s.Ctx.BlockHeight()
	initStartRecord.Asset0Denom, initStartRecord.Asset1Denom = denomA, denomB

	zeroAccumTenPoint1Record := recordWithUpdatedSpotPrice(initStartRecord, sdk.NewDec(10), sdk.NewDecWithPrec(1, 1))

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
		"no record (wrong pool ID)": {initStartRecord, initStartRecord, baseTime, 4, denomA, denomB, true},
		"default record":            {initStartRecord, initStartRecord, baseTime, 1, denomA, denomB, false},
		"one second later record":   {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOne, 1, denomA, denomB, false},
		"idempotent overwrite":      {initStartRecord, initStartRecord, baseTime, 1, denomA, denomB, false},
		"idempotent overwrite2":     {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOne, 1, denomA, denomB, false},
		"diff spot price": {zeroAccumTenPoint1Record,
			recordWithUpdatedAccum(zeroAccumTenPoint1Record, OneSec.MulInt64(10), OneSec.QuoInt64(10)),
			tPlusOne, 1, denomA, denomB, false},
		// TODO: Overflow
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// setup time
			s.Ctx = s.Ctx.WithBlockTime(tc.time)
			tc.expRecord.Time = tc.time

			s.twapkeeper.StoreNewRecord(s.Ctx, tc.startRecord)

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

// TestGetArithmeticTwap tests if we get the expected twap value from `GetArithmeticTwap`.
// We test the method directly by updating the accumulator and storing the twap records
// manually in this test.
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

	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        getTwapInput
		expTwap      sdk.Dec
		expErrorStr  string
	}{
		"(1 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record) start and end point to same record, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(1, 1),
		},
		"(1 record) start and end point to same record, end time = now": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(time.Minute), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},

		"(2 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(2 record) start and end exact, different records": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(10*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(2 record) start exact, end after second record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(20*time.Second), quoteAssetA),
			expTwap:      sdk.NewDecWithPrec(75, 1), // 10 for 10s, 5 for 10s
		},
		"(2 record) start exact, end after second record, sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(20*time.Second), quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(15, 2), // .1 for 10s, .2 for 10s
		},
		// start at 5 second after first twap record, end at 5 second after second twap record
		"(2 record) start and end interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime.Add(5*time.Second), baseTime.Add(20*time.Second), quoteAssetA),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: ThreePlusOneThird.MulInt64(2),
		},

		"(3 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(10*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(5),
		},
		"(3 record) start and end exactly at record times, different records": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(20*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(5),
		},
		"(3 record) start at second record, end after third record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), quoteAssetA),
			expTwap:      sdk.NewDecWithPrec(35, 1), // 5 for 10s, 2 for 10s
		},
		"(3 record) start at second record, end after third record, sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(35, 2), // 0.2 for 10s, 0.5 for 10s
		},
		// start in middle of first and second record, end in middle of second and third record
		"(3 record) interpolate: in between second and third record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime.Add(15*time.Second), baseTime.Add(25*time.Second), quoteAssetA),
			expTwap:      sdk.NewDecWithPrec(35, 1), // 5 for 5s, 2 for 5 = 35 / 10 = 3.5
		},
		// interpolate in time closer to second record
		"(3 record) interpolate: get twap closer to second record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapInput(baseTime.Add(15*time.Second), baseTime.Add(30*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(3), // 5 for 5s, 2 for 10s = 45 / 15 = 3
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
		// TODO: overflow tests
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			for _, record := range test.recordsToSet {
				s.twapkeeper.StoreNewRecord(s.Ctx, record)
			}
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			var twap sdk.Dec
			var err error

			twap, err = s.twapkeeper.GetArithmeticTwap(s.Ctx, test.input.poolId,
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

// TestGetArithmeticTwapToNow tests if we get the expected twap value from `GetArithmeticTwapToNow`.
// We test the method directly by updating the accumulator and storing the twap records
// manually in this test.
func (s *TestSuite) TestGetArithmeticTwapToNow() {

	makeSimpleTwapToNowInput := func(startTime time.Time, isQuoteTokenA bool) getTwapInput {
		quoteAssetDenom, baseAssetDenom := denom0, denom1
		if isQuoteTokenA {
			baseAssetDenom, quoteAssetDenom = quoteAssetDenom, baseAssetDenom
		}
		return getTwapInput{1, quoteAssetDenom, baseAssetDenom, startTime, startTime}
	}

	quoteAssetA := true
	quoteAssetB := false

	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        getTwapInput
		expTwap      sdk.Dec
		expErrorStr  string
	}{
		"(1 record) start time = record time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapToNowInput(baseTime, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record) start time = record time, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapToNowInput(baseTime, quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(1, 1),
		},
		"(1 record) to_now: start time > record time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(2 record) to now: start time = second record time": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(5), // .1 for 10s, .2 for 10s
		},
		"(2 record) to now: start time = second record time, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(time.Minute),
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(2, 1),
		},
		"(2 record) first record time < start time < second record time": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(20 * time.Second),
			input:        makeSimpleTwapToNowInput(baseTime.Add(5*time.Second), quoteAssetA),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: ThreePlusOneThird.MulInt64(2),
		},

		// error catching
		"start time too old": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Second),
			input:        makeSimpleTwapToNowInput(baseTime.Add(-time.Hour), quoteAssetA),
			expErrorStr:  "too old",
		},
		// TODO: overflow tests
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			for _, record := range test.recordsToSet {
				s.twapkeeper.StoreNewRecord(s.Ctx, record)
			}
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			var twap sdk.Dec
			var err error

			// test the values of `GetArithmeticTwapToNow` if bool in test field is true
			twap, err = s.twapkeeper.GetArithmeticTwapToNow(s.Ctx, test.input.poolId,
				test.input.quoteAssetDenom, test.input.baseAssetDenom,
				test.input.startTime)

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
