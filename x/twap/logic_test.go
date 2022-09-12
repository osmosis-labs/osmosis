package twap_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/app/apptesting/osmoassert"
	gammtypes "github.com/osmosis-labs/osmosis/v12/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v12/x/twap"
	"github.com/osmosis-labs/osmosis/v12/x/twap/types"
	"github.com/osmosis-labs/osmosis/v12/x/twap/types/twapmock"
)

var zeroDec = sdk.ZeroDec()
var oneDec = sdk.OneDec()
var twoDec = oneDec.Add(oneDec)
var OneSec = sdk.MustNewDecFromStr("1000.000000000000000000")

func newRecord(poolId uint64, t time.Time, sp0, accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		PoolId:          poolId,
		Asset0Denom:     defaultTwoAssetCoins[0].Denom,
		Asset1Denom:     defaultTwoAssetCoins[1].Denom,
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
		Asset0Denom: defaultTwoAssetCoins[0].Denom,
		Asset1Denom: defaultTwoAssetCoins[1].Denom,
		// make new copies
		P0ArithmeticTwapAccumulator: accum0.Add(sdk.ZeroDec()),
		P1ArithmeticTwapAccumulator: accum1.Add(sdk.ZeroDec()),
	}
}

func (s *TestSuite) TestNewTwapRecord() {
	// prepare pool before test
	poolId := s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins...)

	tests := map[string]struct {
		poolId        uint64
		denom0        string
		denom1        string
		expectedErr   error
		expectedPanic bool
	}{
		"denom with lexicographical order": {
			poolId,
			denom0,
			denom1,
			nil,
			false,
		},
		"denom with non-lexicographical order": {
			poolId,
			denom1,
			denom0,
			nil,
			false,
		},
		"new record with same denom": {
			poolId,
			denom0,
			denom0,
			fmt.Errorf("both assets cannot be of the same denom: assetA: %s, assetB: %s", denom0, denom0),
			false,
		},
		"error in getting spot price": {
			poolId + 1,
			denom1,
			denom0,
			nil,
			true,
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			twapRecord, err := twap.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, test.poolId, test.denom0, test.denom1)

			if test.expectedPanic {
				s.Require().Equal(twapRecord.LastErrorTime, s.Ctx.BlockTime())
			} else if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().Equal(test.expectedErr.Error(), err.Error())
			} else {
				s.Require().NoError(err)

				s.Require().Equal(test.poolId, twapRecord.PoolId)
				s.Require().Equal(s.Ctx.BlockHeight(), twapRecord.Height)
				s.Require().Equal(s.Ctx.BlockTime(), twapRecord.Time)
				s.Require().Equal(sdk.ZeroDec(), twapRecord.P0ArithmeticTwapAccumulator)
				s.Require().Equal(sdk.ZeroDec(), twapRecord.P1ArithmeticTwapAccumulator)
			}

		})
	}
}

func (s *TestSuite) TestUpdateRecord() {
	poolId := s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins...)
	programmableAmmInterface := twapmock.NewProgrammedAmmInterface(s.App.TwapKeeper.GetAmmInterface())
	s.App.TwapKeeper.SetAmmInterface(programmableAmmInterface)

	spotPriceResOne := twapmock.SpotPriceResult{Sp: sdk.OneDec(), Err: nil}
	spotPriceResOneErr := twapmock.SpotPriceResult{Sp: sdk.OneDec(), Err: errors.New("dummy err")}
	spotPriceResOneErrNilDec := twapmock.SpotPriceResult{Sp: sdk.Dec{}, Err: errors.New("dummy err")}
	baseTime := time.Unix(2, 0).UTC()
	updateTime := time.Unix(3, 0).UTC()
	baseTimeMinusOne := time.Unix(1, 0).UTC()

	zeroAccumNoErrSp10Record := newRecord(poolId, baseTime, sdk.NewDec(10), zeroDec, zeroDec)
	sp10OneTimeUnitAccumRecord := newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10))
	// all tests occur with updateTime = base time + time.Unix(1, 0)
	tests := map[string]struct {
		record           types.TwapRecord
		spotPriceResult0 twapmock.SpotPriceResult
		spotPriceResult1 twapmock.SpotPriceResult
		expRecord        types.TwapRecord
	}{
		"0 accum start, sp change": {
			record:           zeroAccumNoErrSp10Record,
			spotPriceResult0: spotPriceResOne,
			spotPriceResult1: spotPriceResOne,
			expRecord:        sp10OneTimeUnitAccumRecord,
		},
		"0 accum start, sp0 err at update": {
			record:           zeroAccumNoErrSp10Record,
			spotPriceResult0: spotPriceResOneErr,
			spotPriceResult1: spotPriceResOne,
			expRecord:        withLastErrTime(sp10OneTimeUnitAccumRecord, updateTime),
		},
		"0 accum start, sp0 err at update with nil dec": {
			record:           zeroAccumNoErrSp10Record,
			spotPriceResult0: spotPriceResOneErrNilDec,
			spotPriceResult1: spotPriceResOne,
			expRecord:        withSp0(withLastErrTime(sp10OneTimeUnitAccumRecord, updateTime), sdk.ZeroDec()),
		},
		"0 accum start, sp1 err at update with nil dec": {
			record:           zeroAccumNoErrSp10Record,
			spotPriceResult0: spotPriceResOne,
			spotPriceResult1: spotPriceResOneErrNilDec,
			expRecord:        withSp1(withLastErrTime(sp10OneTimeUnitAccumRecord, updateTime), sdk.ZeroDec()),
		},
		"startRecord err time preserved": {
			record:           withLastErrTime(zeroAccumNoErrSp10Record, baseTimeMinusOne),
			spotPriceResult0: spotPriceResOne,
			spotPriceResult1: spotPriceResOne,
			expRecord:        withLastErrTime(sp10OneTimeUnitAccumRecord, baseTimeMinusOne),
		},
		"err time bumped with start": {
			record:           withLastErrTime(zeroAccumNoErrSp10Record, baseTimeMinusOne),
			spotPriceResult0: spotPriceResOne,
			spotPriceResult1: spotPriceResOneErr,
			expRecord:        withLastErrTime(sp10OneTimeUnitAccumRecord, updateTime),
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			// setup common, block time, pool Id, expected spot prices
			s.Ctx = s.Ctx.WithBlockTime(updateTime.UTC())
			test.record.PoolId = poolId
			test.expRecord.PoolId = poolId
			if (test.expRecord.P0LastSpotPrice == sdk.Dec{}) {
				test.expRecord.P0LastSpotPrice = test.spotPriceResult0.Sp
			}
			if (test.expRecord.P1LastSpotPrice == sdk.Dec{}) {
				test.expRecord.P1LastSpotPrice = test.spotPriceResult1.Sp
			}
			test.expRecord.Height = s.Ctx.BlockHeight()
			test.expRecord.Time = s.Ctx.BlockTime()

			programmableAmmInterface.ProgramPoolSpotPriceOverride(poolId,
				defaultTwoAssetCoins[0].Denom, defaultTwoAssetCoins[1].Denom,
				test.spotPriceResult0.Sp, test.spotPriceResult0.Err)
			programmableAmmInterface.ProgramPoolSpotPriceOverride(poolId,
				defaultTwoAssetCoins[1].Denom, defaultTwoAssetCoins[0].Denom,
				test.spotPriceResult1.Sp, test.spotPriceResult1.Err)

			newRecord := s.twapkeeper.UpdateRecord(s.Ctx, test.record)
			s.Equal(test.expRecord, newRecord)
		})
	}
}

func TestRecordWithUpdatedAccumulators(t *testing.T) {
	poolId := uint64(1)
	defaultRecord := newRecord(poolId, time.Unix(1, 0), sdk.NewDec(10), oneDec, twoDec)
	tests := map[string]struct {
		record    types.TwapRecord
		newTime   time.Time
		expRecord types.TwapRecord
	}{
		"accum with zero value": {
			record:    newRecord(poolId, time.Unix(1, 0), sdk.NewDec(10), zeroDec, zeroDec),
			newTime:   time.Unix(2, 0),
			expRecord: newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
		"small starting accumulators": {
			record:    defaultRecord,
			newTime:   time.Unix(2, 0),
			expRecord: newExpRecord(oneDec.Add(OneSec.MulInt64(10)), twoDec.Add(OneSec.QuoInt64(10))),
		},
		"larger time interval": {
			record:    newRecord(poolId, time.Unix(11, 0), sdk.NewDec(10), oneDec, twoDec),
			newTime:   time.Unix(55, 0),
			expRecord: newExpRecord(oneDec.Add(OneSec.MulInt64(44*10)), twoDec.Add(OneSec.MulInt64(44).QuoInt64(10))),
		},
		"same time, accumulator should not change": {
			record:    defaultRecord,
			newTime:   time.Unix(1, 0),
			expRecord: newExpRecord(oneDec, twoDec),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// correct expected record based off copy/paste values
			test.expRecord.Time = test.newTime
			test.expRecord.PoolId = test.record.PoolId
			test.expRecord.P0LastSpotPrice = test.record.P0LastSpotPrice
			test.expRecord.P1LastSpotPrice = test.record.P1LastSpotPrice

			gotRecord := twap.RecordWithUpdatedAccumulators(test.record, test.newTime)
			require.Equal(t, test.expRecord, gotRecord)
		})
	}
}

func (s *TestSuite) TestGetInterpolatedRecord() {
	baseRecord := newTwoAssetPoolTwapRecordWithDefaults(baseTime, sdk.OneDec(), sdk.OneDec(), sdk.OneDec())

	// all tests occur with updateTime = base time + time.Unix(1, 0)
	tests := map[string]struct {
		recordsToPreSet     types.TwapRecord
		testPoolId          uint64
		testDenom0          string
		testDenom1          string
		testTime            time.Time
		expectedAccumulator sdk.Dec
		expectedErr         error
	}{
		"same time with existing record": {
			recordsToPreSet: baseRecord,
			testPoolId:      baseRecord.PoolId,
			testDenom0:      baseRecord.Asset0Denom,
			testDenom1:      baseRecord.Asset1Denom,
			testTime:        baseTime,
		},
		"call 1 second after existing record": {
			recordsToPreSet: baseRecord,
			testPoolId:      baseRecord.PoolId,
			testDenom0:      baseRecord.Asset0Denom,
			testDenom1:      baseRecord.Asset1Denom,
			testTime:        baseTime.Add(time.Second),
			// 1(spot price) * 1000(one sec in milli-seconds)
			expectedAccumulator: baseRecord.P0ArithmeticTwapAccumulator.Add(sdk.NewDec(1000)),
		},
		"call 1 second before existing record": {
			recordsToPreSet: baseRecord,
			testPoolId:      baseRecord.PoolId,
			testDenom0:      baseRecord.Asset0Denom,
			testDenom1:      baseRecord.Asset1Denom,
			testTime:        baseTime.Add(-time.Second),
			expectedErr: fmt.Errorf("looking for a time thats too old, not in the historical index. "+
				" Try storing the accumulator value. (requested time %s)", baseTime.Add(-time.Second)),
		},
		"on lexicographical order denom parameters": {
			recordsToPreSet: baseRecord,
			testPoolId:      baseRecord.PoolId,
			testDenom0:      baseRecord.Asset0Denom,
			testDenom1:      baseRecord.Asset1Denom,
			testTime:        baseTime,
		},
		"test non lexicographical order parameter": {
			recordsToPreSet: baseRecord,
			testPoolId:      baseRecord.PoolId,
			testDenom0:      baseRecord.Asset1Denom,
			testDenom1:      baseRecord.Asset0Denom,
			testTime:        baseTime,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.twapkeeper.StoreNewRecord(s.Ctx, test.recordsToPreSet)

			interpolatedRecord, err := s.twapkeeper.GetInterpolatedRecord(s.Ctx, test.testPoolId, test.testDenom0, test.testDenom1, test.testTime)
			if test.expectedErr != nil {
				s.Require().Error(err)
				s.Require().Equal(test.expectedErr.Error(), err.Error())
				return
			}
			s.Require().NoError(err)

			if test.testTime.Equal(baseTime) {
				s.Require().Equal(test.recordsToPreSet, interpolatedRecord)
			} else {
				s.Require().Equal(test.testTime, interpolatedRecord.Time)
				s.Require().Equal(test.recordsToPreSet.P0LastSpotPrice, interpolatedRecord.P0LastSpotPrice)
				s.Require().Equal(test.recordsToPreSet.P1LastSpotPrice, interpolatedRecord.P1LastSpotPrice)
				s.Require().Equal(test.expectedAccumulator, interpolatedRecord.P0ArithmeticTwapAccumulator)
				s.Require().Equal(test.expectedAccumulator, interpolatedRecord.P1ArithmeticTwapAccumulator)
			}
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
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualTwap, err := twap.ComputeArithmeticTwap(test.startRecord, test.endRecord, test.quoteAsset)
			require.Equal(t, test.expTwap, actualTwap)
			require.NoError(t, err)
		})
	}
}

// This tests the behavior of computeArithmeticTwap, around error returning
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
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec.MulInt64(2), true, baseTime.Add(20*time.Second)),
			quoteAsset:  denom0,
			expTwap:     sdk.OneDec().MulInt64(2),
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

// TestPruneRecords tests that twap records earlier than
// current block time - RecordHistoryKeepPeriod are pruned from the store
// while keeping the newest record before the above time threshold.
// Such record is kept for each pool.
func (s *TestSuite) TestPruneRecords() {
	recordHistoryKeepPeriod := s.twapkeeper.RecordHistoryKeepPeriod(s.Ctx)

	pool1OlderMin2MsRecord, // deleted
		pool2OlderMin1MsRecord,  // deleted
		pool3OlderBaseRecord,    // kept as newest under keep period
		pool4OlderPlus1Record := // kept as newest under keep period
		s.createTestRecordsFromTime(baseTime.Add(2 * -recordHistoryKeepPeriod))

	pool1Min2MsRecord, // kept as newest under keep period
		pool2Min1MsRecord,  // kept as newest under keep period
		pool3BaseRecord,    // kept as it is at the keep period boundary
		pool4Plus1Record := // kept as it is above the keep period boundary
		s.createTestRecordsFromTime(baseTime.Add(-recordHistoryKeepPeriod))

	// non-ascending insertion order.
	recordsToPreSet := []types.TwapRecord{
		pool2OlderMin1MsRecord,
		pool4Plus1Record,
		pool4OlderPlus1Record,
		pool3OlderBaseRecord,
		pool2Min1MsRecord,
		pool3BaseRecord,
		pool1Min2MsRecord,
		pool1OlderMin2MsRecord,
	}

	// tMin2Record is before the threshold and is pruned away.
	// tmin1Record is the newest record before current block time - record history keep period.
	// All other records happen after the threshold and are kept.
	expectedKeptRecords := []types.TwapRecord{
		pool3OlderBaseRecord,
		pool4OlderPlus1Record,
		pool1Min2MsRecord,
		pool2Min1MsRecord,
		pool3BaseRecord,
		pool4Plus1Record,
	}
	s.SetupTest()
	s.preSetRecords(recordsToPreSet)

	ctx := s.Ctx
	twapKeeper := s.twapkeeper

	ctx = ctx.WithBlockTime(baseTime)

	err := twapKeeper.PruneRecords(ctx)
	s.Require().NoError(err)

	s.validateExpectedRecords(expectedKeptRecords)
}

// TestUpdateRecords tests that the records are updated correctly.
// It tests the following:
// - two-asset pools
// - multi-asset pools
// - with spot price errors
// - without spot price errors
// - that new records are created
// - older historical records are not updated
// - spot price error times are either propagated from
// older records or set to current block time in case error occurred.
func (s *TestSuite) TestUpdateRecords() {
	type spOverride struct {
		poolId      uint64
		baseDenom   string
		quoteDenom  string
		overrideSp  sdk.Dec
		overrideErr error
	}

	type expectedResults struct {
		spotPriceA    sdk.Dec
		spotPriceB    sdk.Dec
		lastErrorTime time.Time
		isMostRecent  bool
	}

	var spError = errors.New("spot price error")

	validateRecords := func(expectedRecords []expectedResults, actualRecords []types.TwapRecord) {
		s.Require().Equal(len(expectedRecords), len(actualRecords))
		for i, r := range expectedRecords {
			s.Require().Equal(r.spotPriceA, actualRecords[i].P0LastSpotPrice, "record %d", i)
			s.Require().Equal(r.spotPriceB, actualRecords[i].P1LastSpotPrice, "record %d", i)
			s.Require().Equal(r.lastErrorTime, actualRecords[i].LastErrorTime, "record %d", i)
		}
	}

	tests := map[string]struct {
		preSetRecords []types.TwapRecord
		poolId        uint64
		ammMock       twapmock.ProgrammedAmmInterface
		spOverrides   []spOverride
		blockTime     time.Time

		expectedHistoricalRecords []expectedResults
		expectError               error
	}{
		"no records pre-set; error": {
			preSetRecords: []types.TwapRecord{},
			poolId:        1,
			blockTime:     baseTime,

			expectError: gammtypes.PoolDoesNotExistError{PoolId: 1},
		},
		"existing records in different pool; no-op": {
			preSetRecords: []types.TwapRecord{baseRecord},
			poolId:        baseRecord.PoolId + 1,
			blockTime:     baseTime.Add(time.Second),

			expectError: gammtypes.PoolDoesNotExistError{PoolId: baseRecord.PoolId + 1},
		},
		"the returned number of records does not match expected": {
			preSetRecords: []types.TwapRecord{baseRecord},
			poolId:        baseRecord.PoolId,
			blockTime:     baseRecord.Time.Add(time.Second),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: sdk.NewDec(2),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: sdk.NewDecWithPrec(2, 1),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: "extradenom",
					overrideSp: sdk.NewDecWithPrec(3, 1),
				},
			},

			expectError: types.InvalidRecordCountError{Expected: 3, Actual: 1},
		},
		"two-asset; pre-set record at t; updated valid spot price": {
			preSetRecords: []types.TwapRecord{baseRecord},
			poolId:        baseRecord.PoolId,
			blockTime:     baseRecord.Time.Add(time.Second),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: sdk.NewDec(2),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: sdk.NewDecWithPrec(2, 1),
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record.
				{
					spotPriceA: baseRecord.P0LastSpotPrice,
					spotPriceB: baseRecord.P1LastSpotPrice,
				},
				// The new record added.
				{
					spotPriceA:   sdk.NewDec(2),
					spotPriceB:   sdk.NewDecWithPrec(2, 1),
					isMostRecent: true,
				},
			},
		},
		"two-asset; pre-set record at t; updated with spot price error in both denom pairs": {
			preSetRecords: []types.TwapRecord{baseRecord},
			poolId:        baseRecord.PoolId,
			blockTime:     baseRecord.Time.Add(time.Second),

			spOverrides: []spOverride{
				{
					baseDenom:   baseRecord.Asset0Denom,
					quoteDenom:  baseRecord.Asset1Denom,
					overrideErr: spError,
				},
				{
					baseDenom:   baseRecord.Asset1Denom,
					quoteDenom:  baseRecord.Asset0Denom,
					overrideSp:  sdk.NewDecWithPrec(2, 1),
					overrideErr: spError,
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record.
				{
					spotPriceA: baseRecord.P0LastSpotPrice,
					spotPriceB: baseRecord.P1LastSpotPrice,
				},
				// The new record added.
				{
					spotPriceA:    sdk.ZeroDec(),
					spotPriceB:    sdk.NewDecWithPrec(2, 1),
					lastErrorTime: baseRecord.Time.Add(time.Second), // equals to block time
					isMostRecent:  true,
				},
			},
		},
		"two-asset; pre-set record at t; large spot price in one of the pairs": {
			preSetRecords: []types.TwapRecord{baseRecord},
			poolId:        baseRecord.PoolId,
			blockTime:     baseRecord.Time.Add(time.Second),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: sdk.OneDec(),
				},
				{
					baseDenom:   baseRecord.Asset1Denom,
					quoteDenom:  baseRecord.Asset0Denom,
					overrideSp:  types.MaxSpotPrice.Add(sdk.OneDec()),
					overrideErr: nil, // twap logic should identify the large spot price and mark it as error.
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record.
				{
					spotPriceA: baseRecord.P0LastSpotPrice,
					spotPriceB: baseRecord.P1LastSpotPrice,
				},
				// The new record added.
				{
					spotPriceA:    sdk.OneDec(),
					spotPriceB:    types.MaxSpotPrice,               // Although the price returned from AMM was MaxSpotPrice + 1, it is reset to just MaxSpotPrice.
					lastErrorTime: baseRecord.Time.Add(time.Second), // equals to block time
					isMostRecent:  true,
				},
			},
		},
		"two-asset; pre-set record at t with sp error; new record with no sp error; new record has old sp error": {
			preSetRecords: []types.TwapRecord{withLastErrTime(baseRecord, baseRecord.Time)},
			poolId:        baseRecord.PoolId,
			blockTime:     baseRecord.Time.Add(time.Second),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: sdk.OneDec(),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: sdk.OneDec(),
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record.
				{
					spotPriceA:    baseRecord.P0LastSpotPrice,
					spotPriceB:    baseRecord.P1LastSpotPrice,
					lastErrorTime: baseRecord.Time,
				},
				// The new record added.
				{
					spotPriceA:    sdk.OneDec(),
					spotPriceB:    sdk.OneDec(),
					lastErrorTime: baseRecord.Time,
					isMostRecent:  true,
				},
			},
		},
		"two-asset; pre-set record at t with sp error; new record with sp error and has its sp err time updated": {
			preSetRecords: []types.TwapRecord{withLastErrTime(baseRecord, baseRecord.Time)},
			poolId:        baseRecord.PoolId,
			blockTime:     baseRecord.Time.Add(time.Second),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: sdk.OneDec(),
				},
				{
					baseDenom:   baseRecord.Asset1Denom,
					quoteDenom:  baseRecord.Asset0Denom,
					overrideErr: spError,
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record.
				{
					spotPriceA:    baseRecord.P0LastSpotPrice,
					spotPriceB:    baseRecord.P1LastSpotPrice,
					lastErrorTime: baseRecord.Time,
				},
				// The new record added.
				{
					spotPriceA:    sdk.OneDec(),
					spotPriceB:    sdk.ZeroDec(),
					lastErrorTime: baseRecord.Time.Add(time.Second), // equals to block time
					isMostRecent:  true,
				},
			},
		},
		"two-asset; pre-set at t and t + 1, new record with updated spot price created": {
			preSetRecords: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			poolId:        baseRecord.PoolId,

			blockTime: baseRecord.Time.Add(time.Second * 11),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: sdk.OneDec(),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: sdk.OneDec().Add(sdk.OneDec()),
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record at t.
				{
					spotPriceA: baseRecord.P0LastSpotPrice,
					spotPriceB: baseRecord.P1LastSpotPrice,
				},
				// The original record at t + 1.
				{
					spotPriceA: tPlus10sp5Record.P0LastSpotPrice,
					spotPriceB: tPlus10sp5Record.P1LastSpotPrice,
				},
				// The new record added.
				{
					spotPriceA:   sdk.OneDec(),
					spotPriceB:   sdk.OneDec().Add(sdk.OneDec()),
					isMostRecent: true,
				},
			},
		},
		// This case should never happen in-practice since ctx.BlockTime
		// should always be greater than the last record's time.
		"two-asset; pre-set at t and t + 1, new record inserted between existing": {
			preSetRecords: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			poolId:        baseRecord.PoolId,

			blockTime: baseRecord.Time.Add(time.Second * 5),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: sdk.OneDec(),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: sdk.OneDec().Add(sdk.OneDec()),
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record at t.
				{
					spotPriceA: baseRecord.P0LastSpotPrice,
					spotPriceB: baseRecord.P1LastSpotPrice,
				},
				// The new record added.
				// TODO: it should not be possible to add a record between existing records.
				// https://github.com/osmosis-labs/osmosis/issues/2686
				{
					spotPriceA:   sdk.OneDec(),
					spotPriceB:   sdk.OneDec().Add(sdk.OneDec()),
					isMostRecent: true,
				},
				// The original record at t + 1.
				{
					spotPriceA: tPlus10sp5Record.P0LastSpotPrice,
					spotPriceB: tPlus10sp5Record.P1LastSpotPrice,
				},
			},
		},
		// TODO: complete multi-asset pool tests:
		// "multi-asset pool; pre-set at t and t + 1; creates new records": {},
		// "multi-asset pool; pre-set at t and t + 1; pre-existing records some with error and some with too large spot price, overwrites erorr time":                        {},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			twapKeeper := s.App.TwapKeeper
			ctx := s.Ctx.WithBlockTime(tc.blockTime)

			if len(tc.spOverrides) > 0 {
				ammMock := twapmock.NewProgrammedAmmInterface(s.App.GAMMKeeper)

				for _, sp := range tc.spOverrides {
					ammMock.ProgramPoolSpotPriceOverride(tc.poolId, sp.baseDenom, sp.quoteDenom, sp.overrideSp, sp.overrideErr)
					ammMock.ProgramPoolDenomsOverride(tc.poolId, []string{sp.baseDenom, sp.quoteDenom}, nil)
				}

				twapKeeper.SetAmmInterface(ammMock)
			}

			s.preSetRecords(tc.preSetRecords)

			err := twapKeeper.UpdateRecords(ctx, tc.poolId)

			if tc.expectError != nil {
				s.Require().ErrorIs(err, tc.expectError)
				return
			}

			s.Require().NoError(err)

			poolMostRecentRecords, err := twapKeeper.GetAllMostRecentRecordsForPool(ctx, tc.poolId)
			s.Require().NoError(err)

			expectedMostRecentRecords := make([]expectedResults, 0)
			for _, historical := range tc.expectedHistoricalRecords {
				if historical.isMostRecent {
					expectedMostRecentRecords = append(expectedMostRecentRecords, historical)
				}
			}

			validateRecords(expectedMostRecentRecords, poolMostRecentRecords)

			poolHistoricalRecords := s.getAllHistoricalRecordsForPool(tc.poolId)
			s.Require().NoError(err)
			validateRecords(tc.expectedHistoricalRecords, poolHistoricalRecords)
		})
	}
}
