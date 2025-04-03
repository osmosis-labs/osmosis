package twap_test

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	sdkrand "github.com/osmosis-labs/osmosis/v27/simulation/simtypes/random"
	"github.com/osmosis-labs/osmosis/v27/x/gamm/pool-models/balancer"
	"github.com/osmosis-labs/osmosis/v27/x/twap"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

var (
	// when creating twap inputs, we use a function that, when set to true, returns the
	// base asset as the lexicographically smaller denom and the quote as the larger. When
	// set to false, this order is switched. These constants are provided to understand the
	// base/quote asset for every test at a glance rather than a raw boolean value.
	baseQuoteAB = true
	baseQuoteBA = false

	ThreePlusOneThird osmomath.Dec = osmomath.MustNewDecFromStr("3.333333333333333333")

	// base record is a record with t=baseTime, sp0=10(sp1=0.1) accumulators set to 0
	baseRecord types.TwapRecord = newTwoAssetPoolTwapRecordWithDefaults(baseTime, osmomath.NewDec(10), osmomath.ZeroDec(), osmomath.ZeroDec(), osmomath.ZeroDec())

	threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC types.TwapRecord = newThreeAssetPoolTwapRecordWithDefaults(
		baseTime,
		osmomath.NewDec(10), // spot price 0
		osmomath.ZeroDec(),  // accum A
		osmomath.ZeroDec(),  // accum B
		osmomath.ZeroDec(),  // accum C
		osmomath.ZeroDec(),  // geomAccum AB
		osmomath.ZeroDec(),  // geomAccum AC
		osmomath.ZeroDec(),  // geomAccum BC
	)

	// accumA = 10 seconds * (spot price = 10) = OneSec * 10 * 10
	// accumB = 10 seconds * (spot price = 0.1) = OneSec
	// accumC = 10 seconds * (spot price = 20) = OneSec * 10 * 20
	accumA, accumB, accumC osmomath.Dec = OneSec.MulInt64(10 * 10), OneSec, OneSec.MulInt64(10 * 20)

	// geomAccumAB = 10 seconds * (log_{1.0001}{spot price = 10})
	geomAccumAB = geometricTenSecAccum.MulInt64(10)
	geomAccumAC = geomAccumAB
	// geomAccumBC = 10 seconds * (log_{1.0001}{spot price = 0.1})
	geomAccumBC = OneSec.Mul(logOneOverTen).MulInt64(10)

	// accumulators updated from baseRecord with
	// t = baseTime + 10
	// spA = 5, spB = 0.2, spC = 10
	tPlus10sp5Record = newTwoAssetPoolTwapRecordWithDefaults(
		baseTime.Add(10*time.Second), osmomath.NewDec(5), accumA, accumB, geomAccumAB)

	tPlus10sp5ThreeAssetRecordAB, tPlus10sp5ThreeAssetRecordAC, tPlus10sp5ThreeAssetRecordBC = newThreeAssetPoolTwapRecordWithDefaults(
		baseTime.Add(10*time.Second), osmomath.NewDec(5), accumA, accumB, accumC, geomAccumAB, geomAccumAC, geomAccumBC)

	// accumulators updated from tPlus10sp5Record with
	// t = baseTime + 20
	// spA = 2, spB = 0.5, spC = 4
	tPlus20sp2Record = newTwoAssetPoolTwapRecordWithDefaults(
		baseTime.Add(20*time.Second),
		osmomath.NewDec(2),          // spot price 0
		OneSec.MulInt64(10*10+5*10), // accum A
		OneSec.MulInt64(3),          // accum B
		osmomath.ZeroDec(),          // TODO: choose correct
	)

	tPlus20sp2ThreeAssetRecordAB, tPlus20sp2ThreeAssetRecordAC, tPlus20sp2ThreeAssetRecordBC = newThreeAssetPoolTwapRecordWithDefaults(
		baseTime.Add(20*time.Second),
		osmomath.NewDec(2),           // spot price 0
		OneSec.MulInt64(10*10+5*10),  // accum A
		OneSec.MulInt64(3),           // accum B
		OneSec.MulInt64(20*10+10*10), // accum C
		osmomath.ZeroDec(),           // TODO: choose correct
		osmomath.ZeroDec(),           // TODO: choose correct
		osmomath.ZeroDec(),           // TODO: choose correct
	)

	errSpotPrice = errors.New("twap: error in pool spot price occurred between start and end time, twap result may be faulty")
)

func (s *TestSuite) TestGetBeginBlockAccumulatorRecord() {
	poolId, denomA, denomB := s.setupDefaultPool()
	initStartRecord := newRecord(poolId, s.Ctx.BlockTime(), osmomath.OneDec(), osmomath.ZeroDec(), osmomath.ZeroDec(), osmomath.ZeroDec())
	initStartRecord.PoolId, initStartRecord.Height = poolId, s.Ctx.BlockHeight()
	initStartRecord.Asset0Denom, initStartRecord.Asset1Denom = denomB, denomA

	zeroAccumTenPoint1Record := recordWithUpdatedSpotPrice(initStartRecord, osmomath.NewDec(10), osmomath.NewDecWithPrec(1, 1))

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
		"one second later record":                           {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec, osmomath.ZeroDec()), tPlusOne, 1, denomA, denomB, nil},
		"idempotent overwrite":                              {initStartRecord, initStartRecord, baseTime, 1, denomA, denomB, nil},
		"idempotent overwrite2":                             {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec, osmomath.ZeroDec()), tPlusOne, 1, denomA, denomB, nil},
		"diff spot price": {
			zeroAccumTenPoint1Record,
			recordWithUpdatedAccum(zeroAccumTenPoint1Record, OneSec.MulInt64(10), OneSec.QuoInt64(10), geometricTenSecAccum),
			tPlusOne, 1, denomA, denomB, nil,
		},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// setup time
			s.Ctx = s.Ctx.WithBlockTime(tc.time)
			tc.expRecord.Time = tc.time

			s.twapkeeper.StoreNewRecord(s.Ctx, tc.startRecord)

			actualRecord, err := s.twapkeeper.GetBeginBlockAccumulatorRecord(s.Ctx, tc.poolId, tc.baseDenom, tc.quoteDenom)

			if tc.expError != nil {
				s.Require().ErrorContains(err, fmt.Sprintf("%s", tc.expError))
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

func makeSimpleThreeAssetTwapInput(startTime time.Time, endTime time.Time, baseQuoteAB bool) []getTwapInput {
	var twapInput []getTwapInput
	twapInput = formatSimpleTwapInput(twapInput, startTime, endTime, baseQuoteAB, denom0, denom1, 2)
	twapInput = formatSimpleTwapInput(twapInput, startTime, endTime, baseQuoteAB, denom2, denom0, 2)
	twapInput = formatSimpleTwapInput(twapInput, startTime, endTime, baseQuoteAB, denom1, denom2, 2)
	return twapInput
}

func formatSimpleTwapInput(twapInput []getTwapInput, startTime time.Time, endTime time.Time, baseQuoteAB bool, denomA, denomB string, poolID uint64) []getTwapInput {
	quoteAssetDenom, baseAssetDenom := denomA, denomB
	if baseQuoteAB {
		baseAssetDenom, quoteAssetDenom = quoteAssetDenom, baseAssetDenom
	}
	twapInput = append(twapInput, getTwapInput{poolID, quoteAssetDenom, baseAssetDenom, startTime, endTime})
	return twapInput
}

// TestGetArithmeticTwap tests if we get the expected twap value from `GetArithmeticTwap`.
// We test the method directly by updating the accumulator and storing the twap records
// manually in this test.
func (s *TestSuite) TestGetArithmeticTwap() {
	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        getTwapInput
		expTwap      osmomath.Dec
		expectError  error
		expectSpErr  time.Time
	}{
		"(1 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record) start and end point to same record, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseQuoteAB),
			expTwap:      osmomath.NewDecWithPrec(1, 1),
		},
		"(1 record) start and end point to same record, end time = now": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOneMin, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record) spot price error before start time": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, tMinOne)},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(2 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(2 record) start and end exact, different records": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(10*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(2 record) start exact, end after second record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(20*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDecWithPrec(75, 1), // 10 for 10s, 5 for 10s
		},
		"(2 record) start exact, end after second record, sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(20*time.Second), baseQuoteAB),
			expTwap:      osmomath.NewDecWithPrec(15, 2), // .1 for 10s, .2 for 10s
		},
		// start at 5 second after first twap record, end at 5 second after second twap record
		"(2 record) start and end interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(5*time.Second), baseTime.Add(20*time.Second), baseQuoteBA),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: ThreePlusOneThird.MulInt64(2),
		},

		"(3 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(10*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDec(5),
		},
		"(3 record) start and end exactly at record times, different records": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(20*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDec(5),
		},
		"(3 record) start at second record, end after third record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDecWithPrec(35, 1), // 5 for 10s, 2 for 10s
		},
		"(3 record) start at second record, end after third record, sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), baseQuoteAB),
			expTwap:      osmomath.NewDecWithPrec(35, 2), // 0.2 for 10s, 0.5 for 10s
		},
		// start in middle of first and second record, end in middle of second and third record
		"(3 record) interpolate: in between second and third record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(15*time.Second), baseTime.Add(25*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDecWithPrec(35, 1), // 5 for 5s, 2 for 5 = 35 / 10 = 3.5
		},
		// interpolate in time closer to second record
		"(3 record) interpolate: get twap closer to second record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(15*time.Second), baseTime.Add(30*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDec(3), // 5 for 5s, 2 for 10s = 45 / 15 = 3
		},

		// error catching
		"end time in future": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseQuoteBA),
			expectError:  types.EndTimeInFutureError{BlockTime: baseTime, EndTime: tPlusOne},
		},
		"start time after end time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(tPlusOne, baseTime, baseQuoteBA),
			expectError:  types.StartTimeAfterEndTimeError{StartTime: tPlusOne, EndTime: baseTime},
		},
		"start time too old (end time = now)": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Hour), baseTime, baseQuoteBA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
		"start time too old": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Second),
			input:        makeSimpleTwapInput(baseTime.Add(-time.Hour), baseTime, baseQuoteBA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
		"spot price error in record at record time (start time = record time)": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, baseTime)},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOneMin, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
			expectError:  errSpotPrice,
			expectSpErr:  baseTime,
		},
		"spot price error in record at record time (start time > record time)": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, baseTime)},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(tPlusOne, tPlusOneMin, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
			expectError:  errSpotPrice,
			expectSpErr:  baseTime,
		},
		"spot price error in record after record time": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, tPlusOne)},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOneMin, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
			expectError:  errSpotPrice,
			expectSpErr:  baseTime,
		},
		// should error, since start time may have been used to interpolate this value
		"spot price error exactly at start time": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, baseTime)},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
			expectError:  errSpotPrice,
			expectSpErr:  baseTime,
		},
		"spot price error exactly at end time": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, tPlusOne)},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
			expectError:  errSpotPrice,
			expectSpErr:  baseTime,
		},
		// should not happen, but if it did would error
		"spot price error after end time": {
			recordsToSet: []types.TwapRecord{withLastErrTime(baseRecord, tPlusOneMin)},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
			expectError:  errSpotPrice,
			expectSpErr:  baseTime,
		},
	}
	counter := uint64(0)
	for name, test := range tests {
		curPoolId := counter
		s.Run(name, func() {
			s.preSetRecordsWithPoolId(curPoolId, test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			twap, err := s.twapkeeper.GetArithmeticTwap(s.Ctx, curPoolId,
				test.input.baseAssetDenom, test.input.quoteAssetDenom,
				test.input.startTime, test.input.endTime)

			if test.expectError != nil || !test.expectSpErr.IsZero() {
				s.Require().Error(err)
				s.Require().Equal(test.expectError, err)
				s.Require().Equal(test.expTwap, twap)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expTwap, twap)
		})
		counter++
	}
}

func (s *TestSuite) TestGetArithmeticTwap_ThreeAsset() {
	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        []getTwapInput
		expTwap      []osmomath.Dec
		expectError  error
	}{
		"(2 pairs of 3 records) start exact, end after second record": {
			recordsToSet: []types.TwapRecord{
				threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC,
				tPlus10sp5ThreeAssetRecordAB, tPlus10sp5ThreeAssetRecordAC, tPlus10sp5ThreeAssetRecordBC,
			},
			ctxTime: tPlusOneMin,
			input:   makeSimpleThreeAssetTwapInput(baseTime, baseTime.Add(20*time.Second), baseQuoteBA),
			// A 10 for 10s, 5 for 10s = 150/20 = 7.5
			// C 20 for 10s, 10 for 10s = 300/20 = 15
			// B .1 for 10s, .2 for 10s = 3/20 = 0.15
			expTwap: []osmomath.Dec{osmomath.NewDecWithPrec(75, 1), osmomath.NewDec(15), osmomath.NewDecWithPrec(15, 2)},
		},
		"(3 pairs of 3 record) start at second record, end after third record": {
			recordsToSet: []types.TwapRecord{
				threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC,
				tPlus10sp5ThreeAssetRecordAB, tPlus10sp5ThreeAssetRecordAC, tPlus10sp5ThreeAssetRecordBC,
				tPlus20sp2ThreeAssetRecordAB, tPlus20sp2ThreeAssetRecordAC, tPlus20sp2ThreeAssetRecordBC,
			},
			ctxTime: tPlusOneMin,
			input:   makeSimpleThreeAssetTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), baseQuoteBA),
			// A 5 for 10s, 2 for 10s = 70/20 = 3.5
			// C 10 for 10s, 4 for 10s = 140/20 = 7
			// B .2 for 10s, .5 for 10s = 7/20 = 0.35
			expTwap: []osmomath.Dec{osmomath.NewDecWithPrec(35, 1), osmomath.NewDec(7), osmomath.NewDecWithPrec(35, 2)},
		},
		// interpolate in time closer to second record
		"(3 record) interpolate: get twap closer to second record": {
			recordsToSet: []types.TwapRecord{
				threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC,
				tPlus10sp5ThreeAssetRecordAB, tPlus10sp5ThreeAssetRecordAC, tPlus10sp5ThreeAssetRecordBC,
				tPlus20sp2ThreeAssetRecordAB, tPlus20sp2ThreeAssetRecordAC, tPlus20sp2ThreeAssetRecordBC,
			},
			ctxTime: tPlusOneMin,
			input:   makeSimpleThreeAssetTwapInput(baseTime.Add(15*time.Second), baseTime.Add(30*time.Second), baseQuoteBA),
			// A 5 for 5s, 2 for 10s = 45/15 = 3
			// C 10 for 5s, 4 for 10s = 140/15 = 6
			// B .2 for 5s, .5 for 10s = 7/15 = .4
			expTwap: []osmomath.Dec{osmomath.NewDec(3), osmomath.NewDec(6), osmomath.NewDecWithPrec(4, 1)},
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			for i, record := range test.input {
				twap, err := s.twapkeeper.GetArithmeticTwap(s.Ctx, record.poolId,
					record.baseAssetDenom, record.quoteAssetDenom,
					record.startTime, record.endTime)

				if test.expectError != nil {
					s.Require().Error(err)
					s.Require().ErrorIs(err, test.expectError)
					return
				}
				s.Require().NoError(err)
				s.Require().Equal(test.expTwap[i], twap)
			}
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
	var (
		defaultRecordHistoryKeepPeriod = types.DefaultParams().RecordHistoryKeepPeriod

		// baseTimePlusKeepPeriod = baseTime + defaultRecordHistoryKeepPeriod
		baseTimePlusKeepPeriod = baseTime.Add(defaultRecordHistoryKeepPeriod)

		// oneHourBeforeKeepThreshold =  baseKeepThreshold - 1 hour
		oneHourBeforeKeepThreshold = baseTimePlusKeepPeriod.Add(-time.Hour)

		// oneHourAfterKeepThreshold = baseKeepThreshold + 1 hour
		oneHourAfterKeepThreshold = baseTimePlusKeepPeriod.Add(time.Hour)

		periodBetweenBaseAndOneHourBeforeThreshold           = (defaultRecordHistoryKeepPeriod.Milliseconds() - time.Hour.Milliseconds())
		accumBeforeKeepThreshold0, accumBeforeKeepThreshold1 = osmomath.NewDec(periodBetweenBaseAndOneHourBeforeThreshold * 10), osmomath.NewDec(periodBetweenBaseAndOneHourBeforeThreshold * 10)
		geomAccumBeforeKeepThreshold                         = osmomath.NewDec(periodBetweenBaseAndOneHourBeforeThreshold).Mul(logTen)
		// recordBeforeKeepThreshold is a record with t=baseTime+keepPeriod-1h, sp0=30(sp1=0.3) accumulators set relative to baseRecord
		recordBeforeKeepThreshold = newTwoAssetPoolTwapRecordWithDefaults(oneHourBeforeKeepThreshold, osmomath.NewDec(30), accumBeforeKeepThreshold0, accumBeforeKeepThreshold1, geomAccumBeforeKeepThreshold)
	)

	// N.B.: when ctxTime = end time, we trigger the "TWAP to now path".
	// As a result, we duplicate the test cases by triggering both "to now" and "with end time" paths
	// To trigger "with end time" path, we make end time less than ctxTime.
	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        getTwapInput
		expTwap      osmomath.Dec
		expectError  error
	}{
		"(1 record at keep threshold); to now; ctxTime = at keep threshold; start time = end time = base keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTimePlusKeepPeriod, baseTimePlusKeepPeriod, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record at keep threshold); with end time; ctxTime = at keep threshold; start time = end time = base keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTimePlusKeepPeriod.Add(-time.Millisecond), baseTimePlusKeepPeriod.Add(-time.Millisecond), baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record younger than keep threshold); to now; ctxTime = start time = end time = after keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourAfterKeepThreshold, oneHourAfterKeepThreshold, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record younger than keep threshold); with end time; ctxTime = start time = end time = after keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourAfterKeepThreshold.Add(-time.Millisecond), oneHourAfterKeepThreshold.Add(-time.Millisecond), baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record older than keep threshold); to now; ctxTime = baseTime, start time = end time = before keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourBeforeKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourBeforeKeepThreshold, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record older than keep threshold); with end time; ctxTime = baseTime, start time = end time = before keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourBeforeKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold.Add(-time.Millisecond), oneHourBeforeKeepThreshold.Add(-time.Millisecond), baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record older than keep threshold); to now; ctxTime = after keep threshold, start time = before keep threshold; end time = after keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourAfterKeepThreshold, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record older than keep threshold); with end time; ctxTime = after keep threshold, start time = before keep threshold; end time = after keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourAfterKeepThreshold.Add(-time.Millisecond), baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record at keep threshold); to now; ctxTime = base keep threshold, start time = base time - 1ms (source of error); end time = base keep threshold; error": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Millisecond), baseTimePlusKeepPeriod, baseQuoteBA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Millisecond)},
		},
		"(1 record at keep threshold); with end time; ctxTime = base keep threshold, start time = base time - 1ms (source of error); end time = base keep threshold - ms; error": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Millisecond), baseTimePlusKeepPeriod.Add(-time.Millisecond), baseQuoteBA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Millisecond)},
		},
		"(2 records); to now; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime, baseTimePlusKeepPeriod, baseQuoteBA),
			// expTwap: = (10 * (172800s - 3600s) + 30 * 3600s) / 172800s = 10.416666666666666666
			expTwap: osmomath.MustNewDecFromStr("10.416666666666666666"),
		},
		"(2 records); with end time; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      baseTimePlusKeepPeriod.Add(time.Millisecond),
			input:        makeSimpleTwapInput(baseTime, baseTimePlusKeepPeriod.Add(-time.Millisecond), baseQuoteBA),
			// expTwap: = (10 * (172800000ms - 3600000ms) + 30 * 3599999ms) / 172799999ms approx = 10.41666655333719
			expTwap: osmomath.MustNewDecFromStr("10.416666553337190702"),
		},
		"(2 records); to now; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(baseTime, oneHourAfterKeepThreshold, baseQuoteBA),
			// expTwap: = (10 * (172800s - 3600s) + 30 * 3600s * 2) / (172800s + 3600s) approx = 10.816326530612244
			expTwap: osmomath.MustNewDecFromStr("10.816326530612244897"),
		},
		"(2 records); with end time; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(baseTime, oneHourAfterKeepThreshold.Add(-time.Millisecond), baseQuoteBA),
			// expTwap: = (10 * (172800000ms - 3600000ms) + 30 * (3600000ms + 3599999ms)) / (172800000ms + 3599999ms) approx = 10.81632642186126
			expTwap: osmomath.MustNewDecFromStr("10.816326421861260894"),
		},
	}

	counter := uint64(0)
	for name, test := range tests {
		curPoolId := counter // Capture the current value of the counter for use within the goroutine
		s.Run(name, func() {
			s.preSetRecordsWithPoolId(curPoolId, test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			var twap osmomath.Dec
			var err error

			twap, err = s.twapkeeper.GetArithmeticTwap(s.Ctx, curPoolId,
				test.input.baseAssetDenom, test.input.quoteAssetDenom,
				test.input.startTime, test.input.endTime)

			if test.expectError != nil {
				s.Require().Error(err)
				s.Require().ErrorIs(err, test.expectError)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expTwap, twap)
		})
		counter++
	}
}

func (s *TestSuite) TestGetArithmeticTwap_PruningRecordKeepPeriod_ThreeAsset() {
	var (
		defaultRecordHistoryKeepPeriod = types.DefaultParams().RecordHistoryKeepPeriod
		baseTimePlusKeepPeriod         = baseTime.Add(defaultRecordHistoryKeepPeriod)
		oneHourBeforeKeepThreshold     = baseTimePlusKeepPeriod.Add(-time.Hour)
		oneHourAfterKeepThreshold      = baseTimePlusKeepPeriod.Add(time.Hour)

		periodBetweenBaseAndOneHourBeforeThreshold                                      = (defaultRecordHistoryKeepPeriod.Milliseconds() - time.Hour.Milliseconds())
		accumBeforeKeepThresholdA, accumBeforeKeepThresholdB, accumBeforeKeepThresholdC = osmomath.NewDec(periodBetweenBaseAndOneHourBeforeThreshold * 10), osmomath.NewDecWithPrec(1, 1).MulInt64(periodBetweenBaseAndOneHourBeforeThreshold), osmomath.NewDec(periodBetweenBaseAndOneHourBeforeThreshold * 20)
		// recordBeforeKeepThreshold is a record with t=baseTime+keepPeriod-1h, spA=30(spB=0.3)(spC=60) accumulators set relative to baseRecord
		recordBeforeKeepThresholdAB, recordBeforeKeepThresholdAC, recordBeforeKeepThresholdBC = newThreeAssetPoolTwapRecordWithDefaults(
			oneHourBeforeKeepThreshold,
			osmomath.NewDec(30),
			accumBeforeKeepThresholdA,
			accumBeforeKeepThresholdB,
			accumBeforeKeepThresholdC,
			osmomath.ZeroDec(), // TODO: choose correct
			osmomath.ZeroDec(), // TODO: choose correct
			osmomath.ZeroDec(), // TODO: choose correct
		)
	)

	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        []getTwapInput
		expTwap      []osmomath.Dec
		expectError  error
	}{
		"(2 sets of 3 records); to now; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{
				threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC,
				recordBeforeKeepThresholdAB, recordBeforeKeepThresholdAC, recordBeforeKeepThresholdBC,
			},
			ctxTime: baseTimePlusKeepPeriod,
			input:   makeSimpleThreeAssetTwapInput(baseTime, baseTimePlusKeepPeriod, baseQuoteBA),
			// A 10 for 169200s, 30 for 3600s = 1800000/172800 = 10.416666
			// C 20 for 169200s, 60 for 3600s = 100/172800 = 20.83333333
			// B .1 for 169200s, .033 for 3600s = 17040/172800 = 0.0986111
			expTwap: []osmomath.Dec{osmomath.MustNewDecFromStr("10.416666666666666666"), osmomath.MustNewDecFromStr("20.833333333333333333"), osmomath.MustNewDecFromStr("0.098611111111111111")},
		},
		"(2 sets of 3 records); with end time; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{
				threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC,
				recordBeforeKeepThresholdAB, recordBeforeKeepThresholdAC, recordBeforeKeepThresholdBC,
			},
			ctxTime: oneHourAfterKeepThreshold,
			input:   makeSimpleThreeAssetTwapInput(baseTime, oneHourAfterKeepThreshold.Add(-time.Millisecond), baseQuoteBA),
			// A 10 for 169200000ms, 30 for 7199999ms = 1907999970/176399999 = 10.81632642
			// C 20 for 169200000ms, 60 for 7199999ms = 3815999940/176399999 = 21.6326528
			// B .1 for 169200000ms, .033 for 7199999ms = 17159999/176399999 = 0.09727891
			expTwap: []osmomath.Dec{osmomath.MustNewDecFromStr("10.816326421861260894"), osmomath.MustNewDecFromStr("21.632652843722521789"), osmomath.MustNewDecFromStr("0.097278911927129130")},
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			for i, record := range test.input {
				twap, err := s.twapkeeper.GetArithmeticTwap(s.Ctx, record.poolId,
					record.baseAssetDenom, record.quoteAssetDenom,
					record.startTime, record.endTime)

				if test.expectError != nil {
					s.Require().Error(err)
					s.Require().ErrorIs(err, test.expectError)
					return
				}
				s.Require().NoError(err)
				s.Require().Equal(test.expTwap[i], twap)
			}
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

	tests := map[string]struct {
		recordsToSet  []types.TwapRecord
		ctxTime       time.Time
		input         getTwapInput
		expTwap       osmomath.Dec
		expectedError error
	}{
		"(1 record) start time = record time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime, baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(1 record) start time = record time, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime, baseQuoteAB),
			expTwap:      osmomath.NewDecWithPrec(1, 1),
		},
		"(1 record) to_now: start time > record time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDec(10),
		},
		"(2 record) to now: start time = second record time": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), baseQuoteBA),
			expTwap:      osmomath.NewDec(5), // 10 for 0s, 5 for 10s
		},
		"(2 record) to now: start time = second record time, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), baseQuoteAB),
			expTwap:      osmomath.NewDecWithPrec(2, 1),
		},
		"(2 record) first record time < start time < second record time": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(20 * time.Second),
			input:        makeSimpleTwapToNowInput(baseTime.Add(5*time.Second), baseQuoteBA),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: ThreePlusOneThird.MulInt64(2),
		},

		// error catching
		"start time too old": {
			recordsToSet:  []types.TwapRecord{baseRecord},
			ctxTime:       tPlusOne,
			input:         makeSimpleTwapToNowInput(baseTime.Add(-time.Hour), baseQuoteBA),
			expectedError: twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
		"start time too new": {
			recordsToSet:  []types.TwapRecord{baseRecord},
			ctxTime:       tPlusOne,
			input:         makeSimpleTwapToNowInput(baseTime.Add(time.Hour), baseQuoteBA),
			expectedError: types.StartTimeAfterEndTimeError{StartTime: baseTime.Add(time.Hour), EndTime: tPlusOne},
		},
		"spot price error in record at record time (start time > record time)": {
			recordsToSet:  []types.TwapRecord{withLastErrTime(baseRecord, baseTime)},
			ctxTime:       tPlusOneMin,
			input:         makeSimpleTwapInput(tPlusOne, tPlusOneMin, baseQuoteBA),
			expTwap:       osmomath.NewDec(10),
			expectedError: errSpotPrice,
		},
	}
	counter := uint64(0)
	for name, test := range tests {
		curPoolId := counter
		s.Run(name, func() {
			s.preSetRecordsWithPoolId(curPoolId, test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			var twap osmomath.Dec
			var err error

			// test the values of `GetArithmeticTwapToNow` if bool in test field is true
			twap, err = s.twapkeeper.GetArithmeticTwapToNow(s.Ctx, curPoolId,
				test.input.baseAssetDenom, test.input.quoteAssetDenom,
				test.input.startTime)

			if test.expectedError != nil {
				s.Require().Error(err)
				s.Require().Equal(test.expectedError, err)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(test.expTwap, twap)
		})
		counter++
	}
}

func (s *TestSuite) TestGetArithmeticTwapToNow_ThreeAsset() {
	makeSimpleThreeAssetTwapToNowInput := func(startTime time.Time, baseQuoteAB bool) []getTwapInput {
		return makeSimpleThreeAssetTwapInput(startTime, startTime, baseQuoteAB)
	}

	tests := map[string]struct {
		recordsToSet  []types.TwapRecord
		ctxTime       time.Time
		input         []getTwapInput
		expTwap       []osmomath.Dec
		expectedError error
	}{
		"(2 pairs of 3 records) to now: start time = second record time": {
			recordsToSet: []types.TwapRecord{
				threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC,
				tPlus10sp5ThreeAssetRecordAB, tPlus10sp5ThreeAssetRecordAC, tPlus10sp5ThreeAssetRecordBC,
			},
			ctxTime: tPlusOneMin,
			input:   makeSimpleThreeAssetTwapToNowInput(baseTime.Add(10*time.Second), baseQuoteBA),
			// A 10 for 0s, 5 for 10s = 50/10 = 5
			// C 20 for 0s, 10 for 10s = 100/10 = 10
			// B .1 for 0s, .2 for 10s = 2/10 = 0.2
			expTwap: []osmomath.Dec{osmomath.NewDec(5), osmomath.NewDec(10), osmomath.NewDecWithPrec(2, 1)},
		},
		"(2 pairs of 3 records) first record time < start time < second record time": {
			recordsToSet: []types.TwapRecord{
				threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC,
				tPlus10sp5ThreeAssetRecordAB, tPlus10sp5ThreeAssetRecordAC, tPlus10sp5ThreeAssetRecordBC,
			},
			ctxTime: baseTime.Add(20 * time.Second),
			input:   makeSimpleThreeAssetTwapToNowInput(baseTime.Add(5*time.Second), baseQuoteBA),
			// A 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			// C 20 for 5s, 10 for 10s = 200/15 = 13 + 1/3 = 13.333333
			// B .1 for 5s, .2 for 10s = 2.5/15 = 0.1666666
			expTwap: []osmomath.Dec{ThreePlusOneThird.MulInt64(2), osmomath.MustNewDecFromStr("13.333333333333333333"), osmomath.MustNewDecFromStr("0.166666666666666666")},
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			for i, record := range test.input {
				twap, err := s.twapkeeper.GetArithmeticTwapToNow(s.Ctx, record.poolId,
					record.baseAssetDenom, record.quoteAssetDenom,
					record.startTime)

				if test.expectedError != nil {
					s.Require().Error(err)
					s.Require().ErrorIs(err, test.expectedError)
					return
				}
				s.Require().NoError(err)
				s.Require().Equal(test.expTwap[i], twap)
			}
		})
	}
}

// TestGeometricTwapToNow_BalancerPool_Randomized the goal of this test case is to validate
// that no internal panics occur when computing geometric twap. It also sanity checks
// that geometric twap is roughly close to spot price.
func (s *TestSuite) TestGeometricTwapToNow_BalancerPool_Randomized() {
	seed := int64(1)
	r := rand.New(rand.NewSource(seed))
	retries := 200

	maxUint64 := ^uint(0)

	for i := 0; i < retries; i++ {
		elapsedTimeMs := sdkrand.RandIntBetween(r, 1, int(maxUint64>>1))
		weightA := osmomath.NewInt(int64(sdkrand.RandIntBetween(r, 1, 1000)))
		tokenASupply := osmomath.NewInt(int64(sdkrand.RandIntBetween(r, 10_000, 1_000_000_000_000_000_000)))

		tokenBSupply := osmomath.NewInt(int64(sdkrand.RandIntBetween(r, 10_000, 1_000_000_000_000_000_000)))
		weightB := osmomath.NewInt(int64(sdkrand.RandIntBetween(r, 1, 1000)))

		s.Run(fmt.Sprintf("elapsedTimeMs=%d, weightA=%d, tokenASupply=%d, weightB=%d, tokenBSupply=%d", elapsedTimeMs, weightA, tokenASupply, weightB, tokenBSupply), func() {
			ctx := s.Ctx
			app := s.App

			assets := []balancer.PoolAsset{
				{
					Token:  sdk.NewCoin(denom0, tokenASupply),
					Weight: weightA,
				},
				{
					Token:  sdk.NewCoin(denom1, tokenBSupply),
					Weight: weightB,
				},
			}

			poolId := s.PrepareCustomBalancerPool(assets, balancer.PoolParams{
				SwapFee: osmomath.ZeroDec(),
				ExitFee: osmomath.ZeroDec(),
			})

			// We add 1ms to avoid always landing on the same block time
			// In that case, the most recent spot price would be used
			// instead of interpolation.
			oldTime := ctx.BlockTime().Add(1 * time.Millisecond)
			newTime := oldTime.Add(time.Duration(elapsedTimeMs))

			ctx = ctx.WithBlockTime(newTime)

			spotPrice, err := app.GAMMKeeper.CalculateSpotPrice(ctx, poolId, denom1, denom0)
			s.Require().NoError(err)

			twap, err := app.TwapKeeper.GetGeometricTwapToNow(ctx, poolId, denom0, denom1, oldTime)
			s.Require().NoError(err)

			osmomath.ErrTolerance{
				MultiplicativeTolerance: osmomath.SmallestDec(),
			}.CompareBigDec(
				spotPrice,
				osmomath.BigDecFromDec(twap),
			)
		})
	}
}
