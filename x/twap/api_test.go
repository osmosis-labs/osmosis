package twap_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v11/x/twap"
	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

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
	initStartRecord.Asset0Denom, initStartRecord.Asset1Denom = denomB, denomA

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
		expError   error
	}{
		"no record (wrong pool ID)":                         {initStartRecord, initStartRecord, baseTime, 4, denomA, denomB, fmt.Errorf("twap not found")},
		"default record":                                    {initStartRecord, initStartRecord, baseTime, 1, denomA, denomB, nil},
		"default record but same denom":                     {initStartRecord, initStartRecord, baseTime, 1, denomA, denomA, fmt.Errorf("both assets cannot be of the same denom: assetA: %s, assetB: %s", denomA, denomA)},
		"default record wrong order (should get reordered)": {initStartRecord, initStartRecord, baseTime, 1, denomB, denomA, nil},
		"one second later record":                           {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOne, 1, denomA, denomB, nil},
		"idempotent overwrite":                              {initStartRecord, initStartRecord, baseTime, 1, denomA, denomB, nil},
		"idempotent overwrite2":                             {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOne, 1, denomA, denomB, nil},
		"diff spot price": {zeroAccumTenPoint1Record,
			recordWithUpdatedAccum(zeroAccumTenPoint1Record, OneSec.MulInt64(10), OneSec.QuoInt64(10)),
			tPlusOne, 1, denomA, denomB, nil},
		// TODO: Overflow
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// setup time
			s.Ctx = s.Ctx.WithBlockTime(tc.time)
			tc.expRecord.Time = tc.time

			s.twapkeeper.StoreNewRecord(s.Ctx, tc.startRecord)

			actualRecord, err := s.twapkeeper.GetBeginBlockAccumulatorRecord(s.Ctx, tc.poolId, tc.baseDenom, tc.quoteDenom)

			if tc.expError != nil {
				s.Require().Equal(tc.expError, err)
				return
			}

			// ensure denom order was corrected
			s.Require().True(actualRecord.Asset0Denom < actualRecord.Asset1Denom)

			s.Require().NoError(err)
			s.Require().Equal(tc.expRecord, actualRecord)
		})
	}
}

type getTwapInput struct {
	poolId          uint64
	quoteAssetDenom string
	baseAssetDenom  string
	startTime       time.Time
	endTime         time.Time
}

func makeSimpleTwapInput(startTime time.Time, endTime time.Time, isQuoteTokenA bool) getTwapInput {
	quoteAssetDenom, baseAssetDenom := denom0, denom1
	if isQuoteTokenA {
		baseAssetDenom, quoteAssetDenom = quoteAssetDenom, baseAssetDenom
	}
	return getTwapInput{1, quoteAssetDenom, baseAssetDenom, startTime, endTime}
}

// TestGetArithmeticTwap tests if we get the expected twap value from `GetArithmeticTwap`.
// We test the method directly by updating the accumulator and storing the twap records
// manually in this test.
func (s *TestSuite) TestGetArithmeticTwap() {
	quoteAssetA := true
	quoteAssetB := false

	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        getTwapInput
		expTwap      sdk.Dec
		expectError  error
	}{
		"(1 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record) start and end point to same record, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(1, 1),
		},
		"(1 record) start and end point to same record, end time = now": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOneMin, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(2 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(2 record) start and end exact, different records": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(10*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(2 record) start exact, end after second record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(20*time.Second), quoteAssetA),
			expTwap:      sdk.NewDecWithPrec(75, 1), // 10 for 10s, 5 for 10s
		},
		"(2 record) start exact, end after second record, sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(20*time.Second), quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(15, 2), // .1 for 10s, .2 for 10s
		},
		// start at 5 second after first twap record, end at 5 second after second twap record
		"(2 record) start and end interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(5*time.Second), baseTime.Add(20*time.Second), quoteAssetA),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: ThreePlusOneThird.MulInt64(2),
		},

		"(3 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(10*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(5),
		},
		"(3 record) start and end exactly at record times, different records": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(20*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(5),
		},
		"(3 record) start at second record, end after third record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), quoteAssetA),
			expTwap:      sdk.NewDecWithPrec(35, 1), // 5 for 10s, 2 for 10s
		},
		"(3 record) start at second record, end after third record, sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(35, 2), // 0.2 for 10s, 0.5 for 10s
		},
		// start in middle of first and second record, end in middle of second and third record
		"(3 record) interpolate: in between second and third record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(15*time.Second), baseTime.Add(25*time.Second), quoteAssetA),
			expTwap:      sdk.NewDecWithPrec(35, 1), // 5 for 5s, 2 for 5 = 35 / 10 = 3.5
		},
		// interpolate in time closer to second record
		"(3 record) interpolate: get twap closer to second record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(15*time.Second), baseTime.Add(30*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(3), // 5 for 5s, 2 for 10s = 45 / 15 = 3
		},

		// error catching
		"end time in future": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, quoteAssetA),
			expectError:  types.EndTimeInFutureError{BlockTime: baseTime, EndTime: tPlusOne},
		},
		"start time after end time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(tPlusOne, baseTime, quoteAssetA),
			expectError:  types.StartTimeAfterEndTimeError{StartTime: tPlusOne, EndTime: baseTime},
		},
		"start time too old (end time = now)": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Hour), baseTime, quoteAssetA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
		"start time too old": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Second),
			input:        makeSimpleTwapInput(baseTime.Add(-time.Hour), baseTime, quoteAssetA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
		// TODO: overflow tests, multi-asset pool handling
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			twap, err := s.twapkeeper.GetArithmeticTwap(s.Ctx, test.input.poolId,
				test.input.quoteAssetDenom, test.input.baseAssetDenom,
				test.input.startTime, test.input.endTime)

			if test.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectError)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expTwap, twap)
		})
	}
}

// TestGetArithmeticTwap_PruningRecordKeepPeriod is similar to TestGetArithmeticTwap.
// It specifically focuses on testing edge cases related to the
// pruning record keep period when interacting with GetArithmeticTwap.
// The goal of this test is to make sure that we are able to calculate the twap correctly
// when they are at or below the (current block time - default record history keep period).
// This is conditional on the records being present in the store earlier than startTime.
// If there is no such record, we expect an error.
func (s *TestSuite) TestGetArithmeticTwap_PruningRecordKeepPeriod() {
	const quoteAssetA = true

	var (
		defaultRecordHistoryKeepPeriod = types.DefaultParams().RecordHistoryKeepPeriod

		// baseTimePlusKeepPeriod = baseTime + defaultRecordHistoryKeepPeriod
		baseTimePlusKeepPeriod = baseTime.Add(defaultRecordHistoryKeepPeriod)

		// oneHourBeforeKeepThreshold =  baseKeepThreshold - 1 hour
		oneHourBeforeKeepThreshold = baseTimePlusKeepPeriod.Add(-time.Hour)

		// oneHourAfterKeepThreshold = baseKeepThreshold + 1 hour
		oneHourAfterKeepThreshold = baseTimePlusKeepPeriod.Add(time.Hour)

		periodBetweenBaseAndOneHourBeforeThreshold           = (defaultRecordHistoryKeepPeriod.Milliseconds() - time.Hour.Milliseconds())
		accumBeforeKeepThreshold0, accumBeforeKeepThreshold1 = sdk.NewDec(periodBetweenBaseAndOneHourBeforeThreshold * 10), sdk.NewDec(periodBetweenBaseAndOneHourBeforeThreshold * 10)
		// recordBeforeKeepThreshold is a record with t=baseTime+keepPeriod-1h, sp0=30(sp1=0.3) accumulators set relative to baseRecord
		recordBeforeKeepThreshold types.TwapRecord = newTwapRecordWithDefaults(oneHourBeforeKeepThreshold, sdk.NewDec(30), accumBeforeKeepThreshold0, accumBeforeKeepThreshold1)
	)

	// N.B.: when ctxTime = end time, we trigger the "TWAP to now path".
	// As a result, we duplicate the test cases by triggering both "to now" and "with end time" paths
	// To trigger "with end time" path, we make end time less than ctxTime.
	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        getTwapInput
		expTwap      sdk.Dec
		expectError  error
	}{
		"(1 record at keep threshold); to now; ctxTime = at keep threshold; start time = end time = base keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTimePlusKeepPeriod, baseTimePlusKeepPeriod, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record at keep threshold); with end time; ctxTime = at keep threshold; start time = end time = base keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTimePlusKeepPeriod.Add(-time.Millisecond), baseTimePlusKeepPeriod.Add(-time.Millisecond), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record younger than keep threshold); to now; ctxTime = start time = end time = after keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourAfterKeepThreshold, oneHourAfterKeepThreshold, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record younger than keep threshold); with end time; ctxTime = start time = end time = after keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourAfterKeepThreshold.Add(-time.Millisecond), oneHourAfterKeepThreshold.Add(-time.Millisecond), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record older than keep threshold); to now; ctxTime = baseTime, start time = end time = before keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourBeforeKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourBeforeKeepThreshold, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record older than keep threshold); with end time; ctxTime = baseTime, start time = end time = before keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourBeforeKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold.Add(-time.Millisecond), oneHourBeforeKeepThreshold.Add(-time.Millisecond), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record older than keep threshold); to now; ctxTime = after keep threshold, start time = before keep threshold; end time = after keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourAfterKeepThreshold, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record older than keep threshold); with end time; ctxTime = after keep threshold, start time = before keep threshold; end time = after keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourAfterKeepThreshold.Add(-time.Millisecond), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record at keep threshold); to now; ctxTime = base keep threshold, start time = base time - 1ms (source of error); end time = base keep threshold; error": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Millisecond), baseTimePlusKeepPeriod, quoteAssetA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Millisecond)},
		},
		"(1 record at keep threshold); with end time; ctxTime = base keep threshold, start time = base time - 1ms (source of error); end time = base keep threshold - ms; error": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Millisecond), baseTimePlusKeepPeriod.Add(-time.Millisecond), quoteAssetA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Millisecond)},
		},
		"(2 records); to now; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime, baseTimePlusKeepPeriod, quoteAssetA),
			// expTwap: = (10 * (172800s - 3600s) + 30 * 3600s) / 172800s = 10.416666666666666666
			expTwap: sdk.MustNewDecFromStr("10.416666666666666666"),
		},
		"(2 records); with end time; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      baseTimePlusKeepPeriod.Add(time.Millisecond),
			input:        makeSimpleTwapInput(baseTime, baseTimePlusKeepPeriod.Add(-time.Millisecond), quoteAssetA),
			// expTwap: = (10 * (172800000ms - 3600000ms) + 30 * 3599999ms) / 172799999ms approx = 10.41666655333719
			expTwap: sdk.MustNewDecFromStr("10.416666553337190702"),
		},
		"(2 records); to now; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(baseTime, oneHourAfterKeepThreshold, quoteAssetA),
			// expTwap: = (10 * (172800s - 3600s) + 30 * 3600s * 2) / (172800s + 3600s) approx = 10.816326530612244
			expTwap: sdk.MustNewDecFromStr("10.816326530612244897"),
		},
		"(2 records); with end time; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(baseTime, oneHourAfterKeepThreshold.Add(-time.Millisecond), quoteAssetA),
			// expTwap: = (10 * (172800000ms - 3600000ms) + 30 * (3600000ms + 3599999ms)) / (172800000ms + 3599999ms) approx = 10.81632642186126
			expTwap: sdk.MustNewDecFromStr("10.816326421861260894"),
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			var twap sdk.Dec
			var err error

			twap, err = s.twapkeeper.GetArithmeticTwap(s.Ctx, test.input.poolId,
				test.input.quoteAssetDenom, test.input.baseAssetDenom,
				test.input.startTime, test.input.endTime)

			if test.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectError)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expTwap, twap)
		})
	}
}

// TODO: implement
// func (s *TestSuite) TestGetArithmeticTwapWithErrorRecords() {
// }

// TestGetArithmeticTwapToNow tests if we get the expected twap value from `GetArithmeticTwapToNow`.
// We test the method directly by updating the accumulator and storing the twap records
// manually in this test.
func (s *TestSuite) TestGetArithmeticTwapToNow() {
	makeSimpleTwapToNowInput := func(startTime time.Time, isQuoteTokenA bool) getTwapInput {
		return makeSimpleTwapInput(startTime, startTime, isQuoteTokenA)
	}

	quoteAssetA := true
	quoteAssetB := false

	tests := map[string]struct {
		recordsToSet  []types.TwapRecord
		ctxTime       time.Time
		input         getTwapInput
		expTwap       sdk.Dec
		expectedError error
	}{
		"(1 record) start time = record time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime, quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(1 record) start time = record time, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime, quoteAssetB),
			expTwap:      sdk.NewDecWithPrec(1, 1),
		},
		"(1 record) to_now: start time > record time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(10),
		},
		"(2 record) to now: start time = second record time": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), quoteAssetA),
			expTwap:      sdk.NewDec(5), // 10 for 0s, 5 for 10s
		},
		"(2 record) to now: start time = second record time, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
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
			recordsToSet:  []types.TwapRecord{baseRecord},
			ctxTime:       tPlusOne,
			input:         makeSimpleTwapToNowInput(baseTime.Add(-time.Hour), quoteAssetA),
			expectedError: twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
		// TODO: overflow tests
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			var twap sdk.Dec
			var err error

			// test the values of `GetArithmeticTwapToNow` if bool in test field is true
			twap, err = s.twapkeeper.GetArithmeticTwapToNow(s.Ctx, test.input.poolId,
				test.input.quoteAssetDenom, test.input.baseAssetDenom,
				test.input.startTime)

			if test.expectedError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectedError)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expTwap, twap)
		})
	}
}
