package twap_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/x/twap"
	"github.com/osmosis-labs/osmosis/v12/x/twap/types"
)

const (
	baseAssetA  bool = true
	baseAssetB  bool = false
	baseQuoteAB bool = true
	baseQuoteBA bool = false
	baseQuoteAC bool = true
	baseQuoteCA bool = false
	baseQuoteBC bool = true
	baseQuoteCB bool = false
)

var (
	ThreePlusOneThird sdk.Dec = sdk.MustNewDecFromStr("3.333333333333333333")

	// base record is a record with t=baseTime, sp0=10(sp1=0.1) accumulators set to 0
	baseRecord types.TwapRecord = newTwapRecordWithDefaults(baseTime, sdk.NewDec(10), sdk.ZeroDec(), sdk.ZeroDec())

	// tapRecord are records that would be created from a three asset pool, each with unique denom pairs
	tapRecordAB, tapRecordAC, tapRecordBC = newThreeAssetPoolTwapRecordWithDefaults(
		baseTime, sdk.NewDec(10), sdk.ZeroDec(), sdk.ZeroDec())

	// accum0 = 10 seconds * (spot price = 10) = OneSec * 10 * 10
	// accum1 = 10 seconds * (spot price = 0.1) = OneSec
	accum0, accum1 sdk.Dec = OneSec.MulInt64(10 * 10), OneSec

	// accumulators updated from baseRecord with
	// t = baseTime + 10
	// sp0 = 5, sp1 = 0.2
	tPlus10sp5Record = newTwapRecordWithDefaults(
		baseTime.Add(10*time.Second), sdk.NewDec(5), accum0, accum1)

	tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC = newThreeAssetPoolTwapRecordWithDefaults(
		baseTime.Add(10*time.Second), sdk.NewDec(5), accum0, accum1)

	// accumulators updated from tPlus10sp5Record with
	// t = baseTime + 20
	// sp0 = 2, sp1 = 0.5
	tPlus20sp2Record = newTwapRecordWithDefaults(
		baseTime.Add(20*time.Second), sdk.NewDec(2), OneSec.MulInt64(10*10+5*10), OneSec.MulInt64(3))

	tPlus20sp2TapRecordAB, tPlus20sp2TapRecordAC, tPlus20sp2TapRecordBC = newThreeAssetPoolTwapRecordWithDefaults(
		baseTime.Add(20*time.Second), sdk.NewDec(2), OneSec.MulInt64(10*10+5*10), OneSec.MulInt64(3))
)

func (s *TestSuite) TestGetBeginBlockAccumulatorRecord() {
	poolId, denomA, denomB := s.setupDefaultPool()
	initStartRecord := newRecord(poolId, s.Ctx.BlockTime(), sdk.OneDec(), sdk.ZeroDec(), sdk.ZeroDec())
	initStartRecord[0].Height = s.Ctx.BlockHeight()
	zeroAccumTenPoint1Record := recordWithUpdatedSpotPrice(initStartRecord, sdk.NewDec(10), sdk.NewDecWithPrec(1, 1))

	tapPoolId, tapDenomA, tapDenomB, tapDenomC := s.setupDefaultThreeAssetPool()
	tapInitStartRecord := newTapRecord(tapPoolId, s.Ctx.BlockTime(), sdk.OneDec(), sdk.ZeroDec(), sdk.ZeroDec())
	for i := range tapInitStartRecord {
		tapInitStartRecord[i].Height = s.Ctx.BlockHeight()
	}
	tapZeroAccumTenPoint1Record := recordWithUpdatedSpotPrice(tapInitStartRecord, sdk.NewDec(10), sdk.NewDecWithPrec(1, 1))

	denomAB := [][]string{{denomA}, {denomB}}
	denomBA := [][]string{{denomB}, {denomA}}
	denomAA := [][]string{{denomA}, {denomA}}
	denomABdenomACdenomBC := [][]string{{tapDenomA, tapDenomA, tapDenomB}, {tapDenomB, tapDenomC, tapDenomC}}

	tests := map[string]struct {
		// if start record is blank, don't do any sets
		startRecord []types.TwapRecord
		// We set it to have the updated time
		expRecord []types.TwapRecord
		time      time.Time
		poolId    uint64
		denoms    [][]string
		expError  error
	}{
		"no record (wrong pool ID)":                         {initStartRecord, initStartRecord, baseTime, 4, denomAB, fmt.Errorf("twap not found")},
		"default record":                                    {initStartRecord, initStartRecord, baseTime, poolId, denomAB, nil},
		"default record, three asset pool":                  {tapInitStartRecord, tapInitStartRecord, baseTime, tapPoolId, denomABdenomACdenomBC, nil},
		"default record but same denom":                     {initStartRecord, initStartRecord, baseTime, poolId, denomAA, fmt.Errorf("both assets cannot be of the same denom: assetA: %s, assetB: %s", denomA, denomA)},
		"default record wrong order (should get reordered)": {initStartRecord, initStartRecord, baseTime, poolId, denomBA, nil},
		"one second later record":                           {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOne, poolId, denomAB, nil},
		"one second later record, three asset pool":         {tapInitStartRecord, recordWithUpdatedAccum(tapInitStartRecord, OneSec, OneSec), tPlusOne, tapPoolId, denomABdenomACdenomBC, nil},
		"idempotent overwrite":                              {initStartRecord, initStartRecord, baseTime, poolId, denomAB, nil},
		"idempotent overwrite, three asset pool":            {tapInitStartRecord, tapInitStartRecord, baseTime, tapPoolId, denomABdenomACdenomBC, nil},
		"idempotent overwrite2":                             {initStartRecord, recordWithUpdatedAccum(initStartRecord, OneSec, OneSec), tPlusOne, poolId, denomAB, nil},
		"idempotent overwrite2, three asset pool":           {tapInitStartRecord, recordWithUpdatedAccum(tapInitStartRecord, OneSec, OneSec), tPlusOne, tapPoolId, denomABdenomACdenomBC, nil},
		"diff spot price": {zeroAccumTenPoint1Record,
			recordWithUpdatedAccum(zeroAccumTenPoint1Record, OneSec.MulInt64(10), OneSec.QuoInt64(10)),
			tPlusOne, poolId, denomAB, nil},
		"diff spot price, three asset pool": {tapZeroAccumTenPoint1Record,
			recordWithUpdatedAccum(tapZeroAccumTenPoint1Record, OneSec.MulInt64(10), OneSec.QuoInt64(10)),
			tPlusOne, tapPoolId, denomABdenomACdenomBC, nil},
	}
	for name, tc := range tests {
		s.Run(name, func() {
			// setup time
			s.Ctx = s.Ctx.WithBlockTime(tc.time)
			for n := range tc.startRecord {
				tc.expRecord[n].Time = tc.time

				s.twapkeeper.StoreNewRecord(s.Ctx, tc.startRecord[n])

				actualRecord, err := s.twapkeeper.GetBeginBlockAccumulatorRecord(s.Ctx, tc.poolId, tc.denoms[0][n], tc.denoms[1][n])

				if tc.expError != nil {
					s.Require().Equal(tc.expError, err)
					return
				}

				// ensure denom order was corrected
				s.Require().True(actualRecord.Asset0Denom < actualRecord.Asset1Denom)

				s.Require().NoError(err)
				s.Require().Equal(tc.expRecord[n], actualRecord)
			}
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

func makeSimpleTwapInput(startTime time.Time, endTime time.Time, isBaseTokenA bool) []getTwapInput {
	var twapInput []getTwapInput
	quoteAssetDenom, baseAssetDenom := denom0, denom1
	if isBaseTokenA {
		baseAssetDenom, quoteAssetDenom = quoteAssetDenom, baseAssetDenom
	}
	twapInput = append(twapInput, getTwapInput{1, quoteAssetDenom, baseAssetDenom, startTime, endTime})
	return twapInput
}

// makeSimpleTapTwapInput creates twap outputs that would result from three asset pool
// if baseQuoteXY is false, the baseQuote pair switches to YX
func makeSimpleTapTwapInput(startTime time.Time, endTime time.Time, baseQuoteAB, baseQuoteAC, baseQuoteBC bool) []getTwapInput {
	var twapInput []getTwapInput
	quoteAssetDenom, baseAssetDenom := denom0, denom1
	if baseQuoteAB {
		baseAssetDenom, quoteAssetDenom = quoteAssetDenom, baseAssetDenom
	}
	twapInput = append(twapInput, getTwapInput{2, quoteAssetDenom, baseAssetDenom, startTime, endTime})
	quoteAssetDenom, baseAssetDenom = denom0, denom2
	if baseQuoteAC {
		baseAssetDenom, quoteAssetDenom = quoteAssetDenom, baseAssetDenom
	}
	twapInput = append(twapInput, getTwapInput{2, quoteAssetDenom, baseAssetDenom, startTime, endTime})
	quoteAssetDenom, baseAssetDenom = denom1, denom2
	if baseQuoteBC {
		baseAssetDenom, quoteAssetDenom = quoteAssetDenom, baseAssetDenom
	}
	twapInput = append(twapInput, getTwapInput{2, quoteAssetDenom, baseAssetDenom, startTime, endTime})
	return twapInput
}

// TestGetArithmeticTwap tests if we get the expected twap value from `GetArithmeticTwap`.
// We test the method directly by updating the accumulator and storing the twap records
// manually in this test.
func (s *TestSuite) TestGetArithmeticTwap() {
	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        []getTwapInput
		expTwap      []sdk.Dec
		expectError  error
	}{
		"(1 record) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 pair of 3 records, three asset pool) start and end point to same record": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapInput(baseTime, tPlusOne, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record) start and end point to same record, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseAssetB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(1, 1)},
		},
		"(1 pair of 3 records, three asset pool) start and end point to same record, use sp1": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapInput(baseTime, tPlusOne, baseQuoteBA, baseQuoteCA, baseQuoteCB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1)},
		},
		"(1 record) start and end point to same record, end time = now": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOneMin, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 pair of 3 records, three asset pool) start and end point to same record, end time = now": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapInput(baseTime, tPlusOneMin, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(2 records) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(2 pairs of 3 records, three asset pool) start and end point to same record": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapInput(baseTime, tPlusOne, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(2 record) start and end exact, different records": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(10*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(2 pairs of 3 records, three asset pool) start and end exact, different records": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapInput(baseTime, baseTime.Add(10*time.Second), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(2 records) start exact, end after second record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(20*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(75, 1)}, // 10 for 10s, 5 for 10s
		},
		"(2 pairs of 3 records, three asset pool) start exact, end after second record": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapInput(baseTime, baseTime.Add(20*time.Second), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(75, 1), sdk.NewDecWithPrec(75, 1), sdk.NewDecWithPrec(75, 1)}, // 10 for 10s, 5 for 10s
		},
		"(2 records) start exact, end after second record, sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime, baseTime.Add(20*time.Second), baseAssetB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(15, 2)}, // .1 for 10s, .2 for 10s
		},
		"(2 pairs of 3 records, three asset pool) start exact, end after second record, sp1": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapInput(baseTime, baseTime.Add(20*time.Second), baseQuoteBA, baseQuoteCA, baseQuoteCB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(15, 2), sdk.NewDecWithPrec(15, 2), sdk.NewDecWithPrec(15, 2)}, // .1 for 10s, .2 for 10s
		},
		// start at 5 second after first twap record, end at 5 second after second twap record
		"(2 records) start and end interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(5*time.Second), baseTime.Add(20*time.Second), baseAssetA),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: []sdk.Dec{ThreePlusOneThird.MulInt64(2)},
		},
		"(2 pairs of 3 records, three asset pool) start and end interpolated": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapInput(baseTime.Add(5*time.Second), baseTime.Add(20*time.Second), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: []sdk.Dec{ThreePlusOneThird.MulInt64(2), ThreePlusOneThird.MulInt64(2), ThreePlusOneThird.MulInt64(2)},
		},
		"(3 records) start and end point to same record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(10*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(5)},
		},
		"(3 pairs of 3 records, three asset pool) start and end point to same record": {
			recordsToSet: []types.TwapRecord{
				tapRecordAB, tapRecordAC, tapRecordBC,
				tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC,
				tPlus20sp2TapRecordAB, tPlus20sp2TapRecordAC, tPlus20sp2TapRecordBC},
			ctxTime: tPlusOneMin,
			input:   makeSimpleTapTwapInput(baseTime.Add(10*time.Second), baseTime.Add(10*time.Second), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap: []sdk.Dec{sdk.NewDec(5), sdk.NewDec(5), sdk.NewDec(5)},
		},
		"(3 records) start and end exactly at record times, different records": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(20*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(5)},
		},
		"(3 pairs of 3 records, three asset pool) start and end exactly at record times, different records": {
			recordsToSet: []types.TwapRecord{
				tapRecordAB, tapRecordAC, tapRecordBC,
				tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC,
				tPlus20sp2TapRecordAB, tPlus20sp2TapRecordAC, tPlus20sp2TapRecordBC},
			ctxTime: tPlusOneMin,
			input:   makeSimpleTapTwapInput(baseTime.Add(10*time.Second), baseTime.Add(20*time.Second), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap: []sdk.Dec{sdk.NewDec(5), sdk.NewDec(5), sdk.NewDec(5)},
		},
		"(3 records) start at second record, end after third record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(35, 1)}, // 5 for 10s, 2 for 10s
		},
		"(3 pairs of 3 records, three asset pool) start at second record, end after third record": {
			recordsToSet: []types.TwapRecord{
				tapRecordAB, tapRecordAC, tapRecordBC,
				tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC,
				tPlus20sp2TapRecordAB, tPlus20sp2TapRecordAC, tPlus20sp2TapRecordBC},
			ctxTime: tPlusOneMin,
			input:   makeSimpleTapTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap: []sdk.Dec{sdk.NewDecWithPrec(35, 1), sdk.NewDecWithPrec(35, 1), sdk.NewDecWithPrec(35, 1)}, // 5 for 10s, 2 for 10s
		},
		"(3 records) start at second record, end after third record, sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), baseAssetB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(35, 2)}, // 0.2 for 10s, 0.5 for 10s
		},
		"(3 pairs of 3 records, three asset pool) start at second record, end after third record, sp1": {
			recordsToSet: []types.TwapRecord{
				tapRecordAB, tapRecordAC, tapRecordBC,
				tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC,
				tPlus20sp2TapRecordAB, tPlus20sp2TapRecordAC, tPlus20sp2TapRecordBC},
			ctxTime: tPlusOneMin,
			input:   makeSimpleTapTwapInput(baseTime.Add(10*time.Second), baseTime.Add(30*time.Second), baseQuoteBA, baseQuoteCA, baseQuoteCB),
			expTwap: []sdk.Dec{sdk.NewDecWithPrec(35, 2), sdk.NewDecWithPrec(35, 2), sdk.NewDecWithPrec(35, 2)}, // 0.2 for 10s, 0.5 for 10s
		},
		// start in middle of first and second record, end in middle of second and third record
		"(3 records) interpolate: in between second and third record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(15*time.Second), baseTime.Add(25*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(35, 1)}, // 5 for 5s, 2 for 5 = 35 / 10 = 3.5
		},
		"(3 pairs of 3 records, three asset pool) interpolate: in between second and third record": {
			recordsToSet: []types.TwapRecord{
				tapRecordAB, tapRecordAC, tapRecordBC,
				tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC,
				tPlus20sp2TapRecordAB, tPlus20sp2TapRecordAC, tPlus20sp2TapRecordBC},
			ctxTime: tPlusOneMin,
			input:   makeSimpleTapTwapInput(baseTime.Add(15*time.Second), baseTime.Add(25*time.Second), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap: []sdk.Dec{sdk.NewDecWithPrec(35, 1), sdk.NewDecWithPrec(35, 1), sdk.NewDecWithPrec(35, 1)}, // 5 for 5s, 2 for 5 = 35 / 10 = 3.5
		},
		// interpolate in time closer to second record
		"(3 records) interpolate: get twap closer to second record": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record, tPlus20sp2Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapInput(baseTime.Add(15*time.Second), baseTime.Add(30*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(3)}, // 5 for 5s, 2 for 10s = 45 / 15 = 3
		},
		"(3 pairs of 3 records, three asset pool) interpolate: get twap closer to second record": {
			recordsToSet: []types.TwapRecord{
				tapRecordAB, tapRecordAC, tapRecordBC,
				tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC,
				tPlus20sp2TapRecordAB, tPlus20sp2TapRecordAC, tPlus20sp2TapRecordBC},
			ctxTime: tPlusOneMin,
			input:   makeSimpleTapTwapInput(baseTime.Add(15*time.Second), baseTime.Add(30*time.Second), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap: []sdk.Dec{sdk.NewDec(3), sdk.NewDec(3), sdk.NewDec(3)}, // 5 for 5s, 2 for 10s = 45 / 15 = 3
		},
		// error catching
		"end time in future": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(baseTime, tPlusOne, baseAssetA),
			expectError:  types.EndTimeInFutureError{BlockTime: baseTime, EndTime: tPlusOne},
		},
		"start time after end time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(tPlusOne, baseTime, baseAssetA),
			expectError:  types.StartTimeAfterEndTimeError{StartTime: tPlusOne, EndTime: baseTime},
		},
		"start time too old (end time = now)": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Hour), baseTime, baseAssetA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
		"start time too old": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTime.Add(time.Second),
			input:        makeSimpleTwapInput(baseTime.Add(-time.Hour), baseTime, baseAssetA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			for i, record := range test.input {
				twap, err := s.twapkeeper.GetArithmeticTwap(s.Ctx, record.poolId,
					record.quoteAssetDenom, record.baseAssetDenom,
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
		accumBeforeKeepThreshold0, accumBeforeKeepThreshold1 = sdk.NewDec(periodBetweenBaseAndOneHourBeforeThreshold * 10), sdk.NewDec(periodBetweenBaseAndOneHourBeforeThreshold * 10)
		// recordBeforeKeepThreshold is a record with t=baseTime+keepPeriod-1h, sp0=30(sp1=0.3) accumulators set relative to baseRecord
		recordBeforeKeepThreshold                                                             types.TwapRecord = newTwapRecordWithDefaults(oneHourBeforeKeepThreshold, sdk.NewDec(30), accumBeforeKeepThreshold0, accumBeforeKeepThreshold1)
		recordBeforeKeepThresholdAB, recordBeforeKeepThresholdAC, recordBeforeKeepThresholdBC                  = newThreeAssetPoolTwapRecordWithDefaults(oneHourBeforeKeepThreshold, sdk.NewDec(30), accumBeforeKeepThreshold0, accumBeforeKeepThreshold1)
	)

	// N.B.: when ctxTime = end time, we trigger the "TWAP to now path".
	// As a result, we duplicate the test cases by triggering both "to now" and "with end time" paths
	// To trigger "with end time" path, we make end time less than ctxTime.
	tests := map[string]struct {
		recordsToSet []types.TwapRecord
		ctxTime      time.Time
		input        []getTwapInput
		expTwap      []sdk.Dec
		expectError  error
	}{
		"(1 record at keep threshold); to now; ctxTime = at keep threshold; start time = end time = base keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTimePlusKeepPeriod, baseTimePlusKeepPeriod, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record at keep threshold, three asset pool); to now; ctxTime = at keep threshold; start time = end time = base keep threshold": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTapTwapInput(baseTimePlusKeepPeriod, baseTimePlusKeepPeriod, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record at keep threshold); with end time; ctxTime = at keep threshold; start time = end time = base keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTimePlusKeepPeriod.Add(-time.Millisecond), baseTimePlusKeepPeriod.Add(-time.Millisecond), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record at keep threshold, three asset pool); with end time; ctxTime = at keep threshold; start time = end time = base keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTapTwapInput(baseTimePlusKeepPeriod.Add(-time.Millisecond), baseTimePlusKeepPeriod.Add(-time.Millisecond), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record younger than keep threshold); to now; ctxTime = start time = end time = after keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourAfterKeepThreshold, oneHourAfterKeepThreshold, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record younger than keep threshold, three asset pool); to now; ctxTime = start time = end time = after keep threshold": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTapTwapInput(oneHourAfterKeepThreshold, oneHourAfterKeepThreshold, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record younger than keep threshold); with end time; ctxTime = start time = end time = after keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourAfterKeepThreshold.Add(-time.Millisecond), oneHourAfterKeepThreshold.Add(-time.Millisecond), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record younger than keep threshold, three asset pool); with end time; ctxTime = start time = end time = after keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTapTwapInput(oneHourAfterKeepThreshold.Add(-time.Millisecond), oneHourAfterKeepThreshold.Add(-time.Millisecond), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record older than keep threshold); to now; ctxTime = baseTime, start time = end time = before keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourBeforeKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourBeforeKeepThreshold, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record older than keep threshold, three asset pool); to now; ctxTime = baseTime, start time = end time = before keep threshold": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      oneHourBeforeKeepThreshold,
			input:        makeSimpleTapTwapInput(oneHourBeforeKeepThreshold, oneHourBeforeKeepThreshold, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record older than keep threshold); with end time; ctxTime = baseTime, start time = end time = before keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourBeforeKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold.Add(-time.Millisecond), oneHourBeforeKeepThreshold.Add(-time.Millisecond), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record older than keep threshold, three asset pool); with end time; ctxTime = baseTime, start time = end time = before keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      oneHourBeforeKeepThreshold,
			input:        makeSimpleTapTwapInput(oneHourBeforeKeepThreshold.Add(-time.Millisecond), oneHourBeforeKeepThreshold.Add(-time.Millisecond), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record older than keep threshold); to now; ctxTime = after keep threshold, start time = before keep threshold; end time = after keep threshold": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourAfterKeepThreshold, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record older than keep threshold, three asset pool); to now; ctxTime = after keep threshold, start time = before keep threshold; end time = after keep threshold": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTapTwapInput(oneHourBeforeKeepThreshold, oneHourAfterKeepThreshold, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record older than keep threshold); with end time; ctxTime = after keep threshold, start time = before keep threshold; end time = after keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(oneHourBeforeKeepThreshold, oneHourAfterKeepThreshold.Add(-time.Millisecond), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record older than keep threshold, three asset pool); with end time; ctxTime = after keep threshold, start time = before keep threshold; end time = after keep threshold - 1ms": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTapTwapInput(oneHourBeforeKeepThreshold, oneHourAfterKeepThreshold.Add(-time.Millisecond), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record at keep threshold); to now; ctxTime = base keep threshold, start time = base time - 1ms (source of error); end time = base keep threshold; error": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Millisecond), baseTimePlusKeepPeriod, baseAssetA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Millisecond)},
		},
		"(1 record at keep threshold, three asset pool); to now; ctxTime = base keep threshold, start time = base time - 1ms (source of error); end time = base keep threshold; error": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTapTwapInput(baseTime.Add(-time.Millisecond), baseTimePlusKeepPeriod, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Millisecond)},
		},
		"(1 record at keep threshold); with end time; ctxTime = base keep threshold, start time = base time - 1ms (source of error); end time = base keep threshold - ms; error": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime.Add(-time.Millisecond), baseTimePlusKeepPeriod.Add(-time.Millisecond), baseAssetA),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Millisecond)},
		},
		"(1 record at keep threshold, three asset pool); with end time; ctxTime = base keep threshold, start time = base time - 1ms (source of error); end time = base keep threshold - ms; error": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTapTwapInput(baseTime.Add(-time.Millisecond), baseTimePlusKeepPeriod.Add(-time.Millisecond), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			expectError:  twap.TimeTooOldError{Time: baseTime.Add(-time.Millisecond)},
		},
		"(2 records); to now; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTwapInput(baseTime, baseTimePlusKeepPeriod, baseAssetA),
			// expTwap: = (10 * (172800s - 3600s) + 30 * 3600s) / 172800s = 10.416666666666666666
			expTwap: []sdk.Dec{sdk.MustNewDecFromStr("10.416666666666666666")},
		},
		"(2 records, three asset pool); to now; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, recordBeforeKeepThresholdAB, recordBeforeKeepThresholdAC, recordBeforeKeepThresholdBC},
			ctxTime:      baseTimePlusKeepPeriod,
			input:        makeSimpleTapTwapInput(baseTime, baseTimePlusKeepPeriod, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			// expTwap: = (10 * (172800s - 3600s) + 30 * 3600s) / 172800s = 10.416666666666666666
			expTwap: []sdk.Dec{sdk.MustNewDecFromStr("10.416666666666666666"), sdk.MustNewDecFromStr("10.416666666666666666"), sdk.MustNewDecFromStr("10.416666666666666666")},
		},
		"(2 records); with end time; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      baseTimePlusKeepPeriod.Add(time.Millisecond),
			input:        makeSimpleTwapInput(baseTime, baseTimePlusKeepPeriod.Add(-time.Millisecond), baseAssetA),
			// expTwap: = (10 * (172800000ms - 3600000ms) + 30 * 3599999ms) / 172799999ms approx = 10.41666655333719
			expTwap: []sdk.Dec{sdk.MustNewDecFromStr("10.416666553337190702")},
		},
		"(2 records, three asset pool); with end time; with one directly at threshold, interpolated": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, recordBeforeKeepThresholdAB, recordBeforeKeepThresholdAC, recordBeforeKeepThresholdBC},
			ctxTime:      baseTimePlusKeepPeriod.Add(time.Millisecond),
			input:        makeSimpleTapTwapInput(baseTime, baseTimePlusKeepPeriod.Add(-time.Millisecond), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			// expTwap: = (10 * (172800000ms - 3600000ms) + 30 * 3599999ms) / 172799999ms approx = 10.41666655333719
			expTwap: []sdk.Dec{sdk.MustNewDecFromStr("10.416666553337190702"), sdk.MustNewDecFromStr("10.416666553337190702"), sdk.MustNewDecFromStr("10.416666553337190702")},
		},
		"(2 records); to now; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(baseTime, oneHourAfterKeepThreshold, baseAssetA),
			// expTwap: = (10 * (172800s - 3600s) + 30 * 3600s * 2) / (172800s + 3600s) approx = 10.816326530612244
			expTwap: []sdk.Dec{sdk.MustNewDecFromStr("10.816326530612244897")},
		},
		"(2 records, three asset pool); to now; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, recordBeforeKeepThresholdAB, recordBeforeKeepThresholdAC, recordBeforeKeepThresholdBC},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTapTwapInput(baseTime, oneHourAfterKeepThreshold, baseQuoteAB, baseQuoteAC, baseQuoteBC),
			// expTwap: = (10 * (172800s - 3600s) + 30 * 3600s * 2) / (172800s + 3600s) approx = 10.816326530612244
			expTwap: []sdk.Dec{sdk.MustNewDecFromStr("10.816326530612244897"), sdk.MustNewDecFromStr("10.816326530612244897"), sdk.MustNewDecFromStr("10.816326530612244897")},
		},
		"(2 records); with end time; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{baseRecord, recordBeforeKeepThreshold},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTwapInput(baseTime, oneHourAfterKeepThreshold.Add(-time.Millisecond), baseAssetA),
			// expTwap: = (10 * (172800000ms - 3600000ms) + 30 * (3600000ms + 3599999ms)) / (172800000ms + 3599999ms) approx = 10.81632642186126
			expTwap: []sdk.Dec{sdk.MustNewDecFromStr("10.816326421861260894")},
		},
		"(2 records, three asset pool); with end time; with one before keep threshold, interpolated": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, recordBeforeKeepThresholdAB, recordBeforeKeepThresholdAC, recordBeforeKeepThresholdBC},
			ctxTime:      oneHourAfterKeepThreshold,
			input:        makeSimpleTapTwapInput(baseTime, oneHourAfterKeepThreshold.Add(-time.Millisecond), baseQuoteAB, baseQuoteAC, baseQuoteBC),
			// expTwap: = (10 * (172800000ms - 3600000ms) + 30 * (3600000ms + 3599999ms)) / (172800000ms + 3599999ms) approx = 10.81632642186126
			expTwap: []sdk.Dec{sdk.MustNewDecFromStr("10.816326421861260894"), sdk.MustNewDecFromStr("10.816326421861260894"), sdk.MustNewDecFromStr("10.816326421861260894")},
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			for i, record := range test.input {
				twap, err := s.twapkeeper.GetArithmeticTwap(s.Ctx, record.poolId,
					record.quoteAssetDenom, record.baseAssetDenom,
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
	makeSimpleTwapToNowInput := func(startTime time.Time, isQuoteTokenA bool) []getTwapInput {
		return makeSimpleTwapInput(startTime, startTime, isQuoteTokenA)
	}

	makeSimpleTapTwapToNowInput := func(startTime time.Time, isQuoteTokenA bool) []getTwapInput {
		return makeSimpleTapTwapInput(startTime, startTime, baseQuoteAB, baseQuoteAC, baseQuoteBC)
	}

	baseAssetA := true
	baseAssetB := false

	tests := map[string]struct {
		recordsToSet  []types.TwapRecord
		ctxTime       time.Time
		input         []getTwapInput
		expTwap       []sdk.Dec
		expectedError error
	}{
		"(1 record) start time = record time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record, three asset pool) start time = record time": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapToNowInput(baseTime, baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(1 record) start time = record time, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime, baseAssetB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(1, 1)},
		},
		"(1 record, three asset pool) start time = record time, use sp1": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapToNowInput(baseTime, baseAssetB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(1, 1)},
		},
		"(1 record) to_now: start time > record time": {
			recordsToSet: []types.TwapRecord{baseRecord},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10)},
		},
		"(1 record, three asset pool) to_now: start time > record time": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapToNowInput(baseTime.Add(10*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"(2 record) to now: start time = second record time": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(5)}, // 10 for 0s, 5 for 10s
		},
		"(2 record, three asset pool) to now: start time = second record time": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapToNowInput(baseTime.Add(10*time.Second), baseAssetA),
			expTwap:      []sdk.Dec{sdk.NewDec(5), sdk.NewDec(5), sdk.NewDec(5)}, // 10 for 0s, 5 for 10s
		},
		"(2 record) to now: start time = second record time, use sp1": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTwapToNowInput(baseTime.Add(10*time.Second), baseAssetB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(2, 1)},
		},
		"(2 record, three asset pool) to now: start time = second record time, use sp1": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC},
			ctxTime:      tPlusOneMin,
			input:        makeSimpleTapTwapToNowInput(baseTime.Add(10*time.Second), baseAssetB),
			expTwap:      []sdk.Dec{sdk.NewDecWithPrec(2, 1), sdk.NewDecWithPrec(2, 1), sdk.NewDecWithPrec(2, 1)},
		},
		"(2 record) first record time < start time < second record time": {
			recordsToSet: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			ctxTime:      baseTime.Add(20 * time.Second),
			input:        makeSimpleTwapToNowInput(baseTime.Add(5*time.Second), baseAssetA),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: []sdk.Dec{ThreePlusOneThird.MulInt64(2)},
		},
		"(2 record, three asset pool) first record time < start time < second record time": {
			recordsToSet: []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC, tPlus10sp5TapRecordAB, tPlus10sp5TapRecordAC, tPlus10sp5TapRecordBC},
			ctxTime:      baseTime.Add(20 * time.Second),
			input:        makeSimpleTapTwapToNowInput(baseTime.Add(5*time.Second), baseAssetA),
			// 10 for 5s, 5 for 10s = 100/15 = 6 + 2/3 = 6.66666666
			expTwap: []sdk.Dec{ThreePlusOneThird.MulInt64(2), ThreePlusOneThird.MulInt64(2), ThreePlusOneThird.MulInt64(2)},
		},
		// error catching
		"start time too old": {
			recordsToSet:  []types.TwapRecord{baseRecord},
			ctxTime:       tPlusOne,
			input:         makeSimpleTwapToNowInput(baseTime.Add(-time.Hour), baseAssetA),
			expectedError: twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
		"three asset pool, start time too old": {
			recordsToSet:  []types.TwapRecord{tapRecordAB, tapRecordAC, tapRecordBC},
			ctxTime:       tPlusOne,
			input:         makeSimpleTapTwapToNowInput(baseTime.Add(-time.Hour), baseAssetA),
			expectedError: twap.TimeTooOldError{Time: baseTime.Add(-time.Hour)},
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.preSetRecords(test.recordsToSet)
			s.Ctx = s.Ctx.WithBlockTime(test.ctxTime)

			// test the values of `GetArithmeticTwapToNow` if bool in test field is true
			for i, record := range test.input {
				twap, err := s.twapkeeper.GetArithmeticTwapToNow(s.Ctx, record.poolId,
					record.quoteAssetDenom, record.baseAssetDenom,
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
