package twap_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v11/x/twap"
	"github.com/osmosis-labs/osmosis/v11/x/twap/types"
)

var zeroDec = sdk.ZeroDec()
var oneDec = sdk.OneDec()
var twoDec = oneDec.Add(oneDec)

// once second of duration when converted to decimal
var OneSec = sdk.NewDec(1e9)

func newRecord(poolId uint64, t time.Time, sp0, accum0, accum1 sdk.Dec) types.TwapRecord {
	return types.TwapRecord{
		PoolId:          poolId,
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

func (s *TestSuite) TestNewTwapRecord() {
	// prepare pool before test
	poolId := s.PrepareBalancerPoolWithCoins(defaultUniV2Coins...)

	tests := map[string]struct {
		poolId        uint64
		denom0        string
		denom1        string
		expectedErr   bool
		expectedPanic bool
	}{
		"denom with lexicographical order": {
			poolId,
			denom1,
			denom0,
			false,
			false,
		},
		"denom with non-lexicographical order": {
			poolId,
			denom0,
			denom1,
			false,
			false,
		},
		"new record with same denom": {
			poolId,
			denom0,
			denom0,
			true,
			false,
		},
		"non existent pool id": {
			poolId + 1,
			denom1,
			denom0,
			false,
			true,
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			if test.expectedPanic {
				s.Require().Panics(func() {
					twap.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, test.poolId, test.denom0, test.denom1)
				})
				return
			}

			twapRecord, err := twap.NewTwapRecord(s.App.GAMMKeeper, s.Ctx, test.poolId, test.denom0, test.denom1)

			if test.expectedErr {
				s.Require().Error(err)
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
	tests := map[string]struct {
		inputRecord       types.TwapRecord
		updateBlockHeight bool
		updateSpotPrice   bool
		updateBlockTime   bool
		expRecord         types.TwapRecord
	}{
		"happy path, zero accumulator": {
			inputRecord:       newRecord(1, s.Ctx.BlockTime(), sdk.NewDec(10), zeroDec, zeroDec),
			updateBlockHeight: false,
			updateSpotPrice:   false,
			expRecord:         newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
		"update block height": {
			inputRecord:       newRecord(1, s.Ctx.BlockTime(), sdk.NewDec(10), zeroDec, zeroDec),
			updateBlockHeight: true,
			expRecord:         newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
		"update block time": {
			inputRecord:     newRecord(1, s.Ctx.BlockTime(), sdk.NewDec(10), zeroDec, zeroDec),
			updateSpotPrice: true,
			updateBlockTime: true,
			expRecord:       newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
		"spot price changed": {
			inputRecord:     newRecord(1, s.Ctx.BlockTime(), sdk.NewDec(10), zeroDec, zeroDec),
			updateSpotPrice: true,
			expRecord:       newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10)),
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			poolId := s.PrepareBalancerPoolWithCoins(defaultUniV2Coins...)

			if test.updateBlockHeight {
				s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1)
			}

			if test.updateBlockTime {
				s.Ctx = s.Ctx.WithBlockTime(s.Ctx.BlockTime().Add(time.Minute))
			}

			if test.updateSpotPrice {
				fmt.Println(name)
				s.RunBasicSwap(poolId)
			}

			newRecord := s.twapkeeper.UpdateRecord(s.Ctx, test.inputRecord)

			spotPrice, err := s.App.GAMMKeeper.CalculateSpotPrice(s.Ctx, poolId, newRecord.Asset0Denom, newRecord.Asset1Denom)
			s.Require().NoError(err)

			s.Require().Equal(spotPrice, newRecord.P0LastSpotPrice)
			s.Require().Equal(s.Ctx.BlockHeight(), newRecord.Height)
			s.Require().Equal(s.Ctx.BlockTime(), newRecord.Time)
			s.Require().Equal(poolId, newRecord.PoolId)

			// we test specific math within deeper unit tests, test only for accumulator increase
			if test.updateBlockTime {
				s.Require().True(newRecord.P0ArithmeticTwapAccumulator.GT(zeroDec))
				s.Require().True(newRecord.P1ArithmeticTwapAccumulator.GT(zeroDec))
			} else {
				s.Require().False(newRecord.P0ArithmeticTwapAccumulator.GT(zeroDec))
				s.Require().False(newRecord.P1ArithmeticTwapAccumulator.GT(zeroDec))
			}
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

func TestRecordWithUpdatedAccumulators(t *testing.T) {
	poolId := uint64(1)
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
			record:    newRecord(poolId, time.Unix(1, 0), sdk.NewDec(10), oneDec, twoDec),
			newTime:   time.Unix(2, 0),
			expRecord: newExpRecord(oneDec.Add(OneSec.MulInt64(10)), twoDec.Add(OneSec.QuoInt64(10))),
		},
		"larger time interval": {
			record:    newRecord(poolId, time.Unix(11, 0), sdk.NewDec(10), oneDec, twoDec),
			newTime:   time.Unix(55, 0),
			expRecord: newExpRecord(oneDec.Add(OneSec.MulInt64(44*10)), twoDec.Add(OneSec.MulInt64(44).QuoInt64(10))),
		},
		"same time, accumulator should not change": {
			record:    newRecord(poolId, time.Unix(1, 0), sdk.NewDec(10), oneDec, twoDec),
			newTime:   time.Unix(1, 0),
			expRecord: newExpRecord(oneDec, twoDec),
		},
		// TODO: Overflow tests
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// correct expected record based off copy/paste values
			test.expRecord.Time = test.newTime
			test.expRecord.P0LastSpotPrice = test.record.P0LastSpotPrice
			test.expRecord.P1LastSpotPrice = test.record.P1LastSpotPrice

			gotRecord := twap.RecordWithUpdatedAccumulators(test.record, test.newTime)
			require.Equal(t, test.expRecord, gotRecord)
		})
	}
}

func (s *TestSuite) TestGetInterpolatedRecord() {
	baseRecord := newTwapRecordWithDefaults(baseTime, sdk.OneDec(), sdk.OneDec(), sdk.OneDec())
	// tMin2Record, tMin1Record, baseRecord, tPlus1Record := s.createTestRecordsFromTime(baseTime)

	tests := map[string]struct {
		recordsToPreSet     types.TwapRecord
		testPoolId          uint64
		testDenom0          string
		testDenom1          string
		testTime            time.Time
		expectedAccumulator sdk.Dec
		expectedErr         bool
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
			// 1000000000(spot price) * 1(time)
			expectedAccumulator: baseRecord.P0ArithmeticTwapAccumulator.Add(sdk.NewDec(1000000000)),
		},
		"call 1 second before existing record": {
			recordsToPreSet: baseRecord,
			testPoolId:      baseRecord.PoolId,
			testDenom0:      baseRecord.Asset0Denom,
			testDenom1:      baseRecord.Asset1Denom,
			testTime:        baseTime.Add(-time.Second),
			expectedErr:     true,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			s.twapkeeper.StoreNewRecord(s.Ctx, test.recordsToPreSet)

			interpolatedRecord, err := s.twapkeeper.GetInterpolatedRecord(s.Ctx, test.testPoolId, test.testDenom0, test.testDenom1, test.testTime)
			if test.expectedErr {
				s.Require().Error(err)
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

// TestComputeArithmeticTwap tests ComputeArithmeticTwap on various inputs.
// The test vectors are structured by setting up different start and records,
// based on time interval, and their accumulator values.
// Then an expected TWAP is provided in each test case, to compare against computed.
func TestComputeArithmeticTwap(t *testing.T) {
	newOneSidedRecord := func(time time.Time, accum sdk.Dec, useP0 bool) types.TwapRecord {
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

	type testCase struct {
		startRecord types.TwapRecord
		endRecord   types.TwapRecord
		quoteAsset  string
		expTwap     sdk.Dec
	}

	testCaseFromDeltas := func(startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) testCase {
		return testCase{
			newOneSidedRecord(baseTime, startAccum, true),
			newOneSidedRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff), true),
			denom0,
			expectedTwap,
		}
	}
	testCaseFromDeltasAsset1 := func(startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) testCase {
		return testCase{
			newOneSidedRecord(baseTime, startAccum, false),
			newOneSidedRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff), false),
			denom1,
			expectedTwap,
		}
	}
	tenSecAccum := OneSec.MulInt64(10)
	pointOneAccum := OneSec.QuoInt64(10)
	tests := map[string]testCase{
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
			actualTwap := twap.ComputeArithmeticTwap(test.startRecord, test.endRecord, test.quoteAsset)
			require.Equal(t, test.expTwap, actualTwap)
		})
	}
}
