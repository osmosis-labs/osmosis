package twap_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v11/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v11/x/twap"
	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

var zeroDec = sdk.ZeroDec()
var oneDec = sdk.OneDec()
var twoDec = oneDec.Add(oneDec)
var OneSec = sdk.NewDec(1e9)

func newRecord(t time.Time, sp0, accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		Asset0Denom:     defaultUniV2Coins[0].Denom,
		Asset1Denom:     defaultUniV2Coins[1].Denom,
		Time:            t,
		P0LastSpotPrice: sp0,
		P1LastSpotPrice: sdk.OneDec().Quo(sp0),
		// make new copies
		P0ArithmeticTwapAccumulator: accum0.Add(sdk.ZeroDec()),
		P1ArithmeticTwapAccumulator: accum1.Add(sdk.ZeroDec()),
	}
}

// make an expected record for math tests, we adjust other values in the test runner.
func newExpRecord(accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		Asset0Denom: defaultUniV2Coins[0].Denom,
		Asset1Denom: defaultUniV2Coins[1].Denom,
		// make new copies
		P0ArithmeticTwapAccumulator: accum0.Add(sdk.ZeroDec()),
		P1ArithmeticTwapAccumulator: accum1.Add(sdk.ZeroDec()),
	}
}

func TestRecordWithUpdatedAccumulators(t *testing.T) {
	tests := map[string]struct {
		record          types.TwapRecord
		interpolateTime time.Time
		expRecord       types.TwapRecord
	}{
		"0accum": {
			record:          newRecord(time.Unix(1, 0), sdk.NewDec(10), zeroDec, zeroDec),
			interpolateTime: time.Unix(2, 0),
			expRecord:       newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
		"small starting accumulators": {
			record:          newRecord(time.Unix(1, 0), sdk.NewDec(10), oneDec, twoDec),
			interpolateTime: time.Unix(2, 0),
			expRecord:       newExpRecord(oneDec.Add(OneSec.MulInt64(10)), twoDec.Add(OneSec.QuoInt64(10))),
		},
		"larger time interval": {
			record:          newRecord(time.Unix(11, 0), sdk.NewDec(10), oneDec, twoDec),
			interpolateTime: time.Unix(55, 0),
			expRecord:       newExpRecord(oneDec.Add(OneSec.MulInt64(44*10)), twoDec.Add(OneSec.MulInt64(44).QuoInt64(10))),
		},
		"same time": {
			record:          newRecord(time.Unix(1, 0), sdk.NewDec(10), oneDec, twoDec),
			interpolateTime: time.Unix(1, 0),
			expRecord:       newExpRecord(oneDec, twoDec),
		},
		// TODO: Overflow tests
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// correct expected record based off copy/paste values
			test.expRecord.Time = test.interpolateTime
			test.expRecord.P0LastSpotPrice = test.record.P0LastSpotPrice
			test.expRecord.P1LastSpotPrice = test.record.P1LastSpotPrice

			gotRecord := twap.RecordWithUpdatedAccumulators(test.record, test.interpolateTime)
			require.Equal(t, test.expRecord, gotRecord)
		})
	}
}

func (s *TestSuite) TestUpdateTwap() {
	poolId := s.PrepareBalancerPoolWithCoins(defaultUniV2Coins...)
	newSp := sdk.OneDec()

	tests := map[string]struct {
		record     types.TwapRecord
		updateTime time.Time
		expRecord  types.TwapRecord
	}{
		"0 accum start": {
			record:     newRecord(time.Unix(1, 0), sdk.NewDec(10), zeroDec, zeroDec),
			updateTime: time.Unix(2, 0),
			expRecord:  newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			// setup common, block time, pool Id, expected spot prices
			s.Ctx = s.Ctx.WithBlockTime(test.updateTime.UTC())
			test.record.PoolId = poolId
			test.expRecord.PoolId = poolId
			test.expRecord.P0LastSpotPrice = newSp
			test.expRecord.P1LastSpotPrice = newSp
			test.expRecord.Height = s.Ctx.BlockHeight()
			test.expRecord.Time = s.Ctx.BlockTime()

			newRecord := s.twapkeeper.UpdateRecord(s.Ctx, test.record)
			s.Require().Equal(test.expRecord, newRecord)
		})
	}
}

func newOneSidedRecord(time time.Time, accum sdk.Dec, useP0 bool) types.TwapRecord {
	record := types.TwapRecord{Time: time, Asset0Denom: denom0, Asset1Denom: denom1}
	if useP0 {
		record.P0ArithmeticTwapAccumulator = accum
	} else {
		record.P1ArithmeticTwapAccumulator = accum
	}
	record.P0LastSpotPrice = sdk.ZeroDec()
	record.P1LastSpotPrice = sdk.OneDec()
	return record
}

type computeArithmeticTwapTestCase struct {
	startRecord types.TwapRecord
	endRecord   types.TwapRecord
	quoteAsset  string
	expTwap     sdk.Dec
	expErr      bool
}

// TestComputeArithmeticTwap tests ComputeArithmeticTwap on various inputs.
// The test vectors are structured by setting up different start and records,
// based on time interval, and their accumulator values.
// Then an expected TWAP is provided in each test case, to compare against computed.
func TestComputeArithmeticTwap(t *testing.T) {
	testCaseFromDeltas := func(startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) computeArithmeticTwapTestCase {
		return computeArithmeticTwapTestCase{
			newOneSidedRecord(baseTime, startAccum, true),
			newOneSidedRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff), true),
			denom0,
			expectedTwap,
			false,
		}
	}
	testCaseFromDeltasAsset1 := func(startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) computeArithmeticTwapTestCase {
		return computeArithmeticTwapTestCase{
			newOneSidedRecord(baseTime, startAccum, false),
			newOneSidedRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff), false),
			denom1,
			expectedTwap,
			false,
		}
	}
	tenSecAccum := OneSec.MulInt64(10)
	pointOneAccum := OneSec.QuoInt64(10)
	tests := map[string]computeArithmeticTwapTestCase{
		"basic: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  denom0,
			expTwap:     sdk.OneDec(),
		},
		// this test just shows what happens in case the records are reversed.
		// It should return the correct result, even though this is incorrect internal API usage
		"invalid call: reversed records of above": {
			startRecord: newOneSidedRecord(tPlusOne, OneSec, true),
			endRecord:   newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			quoteAsset:  denom0,
			expTwap:     sdk.OneDec(),
		},
		"same record: denom0, end spot price = 0": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			quoteAsset:  denom0,
			expTwap:     sdk.ZeroDec(),
		},
		"same record: denom1, end spot price = 1": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			quoteAsset:  denom1,
			expTwap:     sdk.OneDec(),
		},
		"accumulator = 10*OneSec, t=5s. 0 base accum": testCaseFromDeltas(
			sdk.ZeroDec(), tenSecAccum, 5*time.Second, sdk.NewDec(2)),
		"accumulator = 10*OneSec, t=3s. 0 base accum": testCaseFromDeltas(
			sdk.ZeroDec(), tenSecAccum, 3*time.Second, ThreePlusOneThird),
		"accumulator = 10*OneSec, t=100s. 0 base accum": testCaseFromDeltas(
			sdk.ZeroDec(), tenSecAccum, 100*time.Second, sdk.NewDecWithPrec(1, 1)),

		// test that base accum has no impact
		"accumulator = 10*OneSec, t=5s. 10 base accum": testCaseFromDeltas(
			sdk.NewDec(10), tenSecAccum, 5*time.Second, sdk.NewDec(2)),
		"accumulator = 10*OneSec, t=3s. 10*second base accum": testCaseFromDeltas(
			tenSecAccum, tenSecAccum, 3*time.Second, ThreePlusOneThird),
		"accumulator = 10*OneSec, t=100s. .1*second base accum": testCaseFromDeltas(
			pointOneAccum, tenSecAccum, 100*time.Second, sdk.NewDecWithPrec(1, 1)),

		"accumulator = 10*OneSec, t=100s. 0 base accum (asset 1)": testCaseFromDeltasAsset1(sdk.ZeroDec(), OneSec.MulInt64(10), 100*time.Second, sdk.NewDecWithPrec(1, 1)),

		// TODO: Overflow, rounding
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualTwap, err := twap.ComputeArithmeticTwap(test.startRecord, test.endRecord, test.quoteAsset)
			require.Equal(t, test.expTwap, actualTwap)
			require.NoError(t, err)
		})
	}
}

// This test should be table driven, for testing the behavior of computeArithmeticTwap, around error returning
// when there has been an intermediate spot price error.
func TestComputeArithmeticTwapWithSpotPriceError(t *testing.T) {
	newOneSidedRecordWErrorTime := func(time time.Time, accum sdk.Dec, useP0 bool, errTime time.Time) types.TwapRecord {
		record := newOneSidedRecord(time, accum, useP0)
		record.LastErrorTime = errTime
		return record
	}
	tests := map[string]computeArithmeticTwapTestCase{
		// should error, since end time may have been used to interpolate this value
		"errAtEndTime from end record": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec, true, tPlusOne),
			quoteAsset:  denom0,
			expTwap:     sdk.OneDec(),
			expErr:      true,
		},
		// should error, since start time may have been used to interpolate this value
		"err at StartTime exactly": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec, true, baseTime),
			quoteAsset:  denom0,
			expTwap:     sdk.OneDec(),
			expErr:      true,
		},
		"err before StartTime": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec, true, tMinOne),
			quoteAsset:  denom0,
			expTwap:     sdk.OneDec(),
			expErr:      false,
		},
		// Should not happen, but if it did would error
		"err after EndTime": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec, true, baseTime.Add(20*time.Second)),
			quoteAsset:  denom0,
			expTwap:     sdk.OneDec(),
			expErr:      true,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualTwap, err := twap.ComputeArithmeticTwap(test.startRecord, test.endRecord, test.quoteAsset)
			require.Equal(t, test.expTwap, actualTwap)
			osmoassert.ConditionalError(t, test.expErr, err)
		})
	}
}

// TestPruneRecords tests that all twap records earlier than
// current block time - RecordHistoryKeepPeriod are pruned from the store.
func (s *TestSuite) TestPruneRecords() {
	recordHistoryKeepPeriod := s.twapkeeper.RecordHistoryKeepPeriod(s.Ctx)

	tMin2Record, tMin1Record, baseRecord, tPlus1Record := s.createTestRecordsFromTime(baseTime.Add(-recordHistoryKeepPeriod))

	// non-ascending insertion order.
	recordsToPreSet := []types.TwapRecord{tPlus1Record, tMin1Record, baseRecord, tMin2Record}

	expectedKeptRecords := []types.TwapRecord{baseRecord, tPlus1Record}
	s.SetupTest()
	s.preSetRecords(recordsToPreSet)

	ctx := s.Ctx
	twapKeeper := s.twapkeeper

	ctx = ctx.WithBlockTime(baseTime)

	err := twapKeeper.PruneRecords(ctx)
	s.Require().NoError(err)

	s.validateExpectedRecords(expectedKeptRecords)
}
