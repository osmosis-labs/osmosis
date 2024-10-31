package twap_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v27/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v27/x/twap"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types/twapmock"
)

var (
	zeroDec              = osmomath.ZeroDec()
	oneDec               = osmomath.OneDec()
	twoDec               = oneDec.Add(oneDec)
	pointFiveDec         = osmomath.OneDec().Quo(twoDec)
	OneSec               = osmomath.MustNewDecFromStr("1000.000000000000000000")
	logTen               = twap.TwapLog(osmomath.NewDec(10))
	logOneOverTen        = twap.TwapLog(osmomath.OneDec().QuoInt64(10))
	tenSecAccum          = OneSec.MulInt64(10)
	geometricTenSecAccum = OneSec.Mul(logTen)
)

func (s *TestSuite) TestGetSpotPrices() {
	currTime := time.Now()
	poolID := s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins...)
	mockAMMI := twapmock.NewProgrammedAmmInterface(s.App.TwapKeeper.GetAmmInterface())
	s.App.TwapKeeper.SetAmmInterface(mockAMMI)

	ctx := s.Ctx.WithBlockTime(currTime.Add(5 * time.Second))

	testCases := map[string]struct {
		poolID                uint64
		prevErrTime           time.Time
		mockSp0               osmomath.Dec
		mockSp1               osmomath.Dec
		mockSp0Err            error
		mockSp1Err            error
		expectedSp0           osmomath.Dec
		expectedSp1           osmomath.Dec
		expectedLatestErrTime time.Time
	}{
		"zero sp": {
			poolID:                poolID,
			prevErrTime:           currTime,
			mockSp0:               osmomath.ZeroDec(),
			mockSp1:               osmomath.ZeroDec(),
			mockSp0Err:            fmt.Errorf("foo"),
			expectedSp0:           osmomath.ZeroDec(),
			expectedSp1:           osmomath.ZeroDec(),
			expectedLatestErrTime: ctx.BlockTime(),
		},
		"exceeds max spot price": {
			poolID:                poolID,
			prevErrTime:           currTime,
			mockSp0:               types.MaxSpotPrice.Add(osmomath.OneDec()),
			mockSp1:               types.MaxSpotPrice.Add(osmomath.OneDec()),
			expectedSp0:           types.MaxSpotPrice,
			expectedSp1:           types.MaxSpotPrice,
			expectedLatestErrTime: ctx.BlockTime(),
		},
		"valid spot prices": {
			poolID:                poolID,
			prevErrTime:           currTime,
			mockSp0:               osmomath.NewDecWithPrec(55, 2),
			mockSp1:               osmomath.NewDecWithPrec(6, 1),
			expectedSp0:           osmomath.NewDecWithPrec(55, 2),
			expectedSp1:           osmomath.NewDecWithPrec(6, 1),
			expectedLatestErrTime: currTime,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			mockAMMI.ProgramPoolSpotPriceOverride(tc.poolID, denom0, denom1, tc.mockSp0, tc.mockSp0Err)
			mockAMMI.ProgramPoolSpotPriceOverride(tc.poolID, denom1, denom0, tc.mockSp1, tc.mockSp1Err)

			sp0, sp1, latestErrTime := twap.GetSpotPrices(ctx, mockAMMI, tc.poolID, denom0, denom1, tc.prevErrTime)
			s.Require().Equal(tc.expectedSp0, sp0)
			s.Require().Equal(tc.expectedSp1, sp1)
			s.Require().Equal(tc.expectedLatestErrTime, latestErrTime)
		})
	}
}

func (s *TestSuite) TestNewTwapRecord() {
	// prepare pool before test

	tests := map[string]struct {
		poolId        uint64
		denom0        string
		denom1        string
		expectedErr   error
		expectedPanic bool
	}{
		"denom with lexicographical order": {
			1,
			denom0,
			denom1,
			nil,
			false,
		},
		"denom with non-lexicographical order": {
			1,
			denom1,
			denom0,
			nil,
			false,
		},
		"new record with same denom": {
			1,
			denom0,
			denom0,
			fmt.Errorf("both assets cannot be of the same denom: assetA: %s, assetB: %s", denom0, denom0),
			false,
		},
		"error in getting spot price": {
			2,
			denom1,
			denom0,
			nil,
			true,
		},
	}
	// iterate twice on the test cases, once with Balancer pools, once with Concentrated liquidity pools
	for i := 0; i < 2; i++ {
		for name, test := range tests {
			s.Run(name, func() {
				s.SetupTest()
				if i == 0 {
					s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins...)
				} else {
					s.PrepareConcentratedPoolWithCoins(denom0, denom1)
				}

				twapRecord, err := twap.NewTwapRecord(s.App.PoolManagerKeeper, s.Ctx, test.poolId, test.denom0, test.denom1)

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
					s.Require().Equal(osmomath.ZeroDec(), twapRecord.P0ArithmeticTwapAccumulator)
					s.Require().Equal(osmomath.ZeroDec(), twapRecord.P1ArithmeticTwapAccumulator)
					s.Require().Equal(osmomath.ZeroDec(), twapRecord.GeometricTwapAccumulator)
				}
			})
		}
	}
}

func (s *TestSuite) TestUpdateRecord() {
	poolId := s.PrepareBalancerPoolWithCoins(defaultTwoAssetCoins...)
	programmableAmmInterface := twapmock.NewProgrammedAmmInterface(s.App.TwapKeeper.GetAmmInterface())
	s.App.TwapKeeper.SetAmmInterface(programmableAmmInterface)

	spotPriceResOne := twapmock.SpotPriceResult{Sp: osmomath.OneDec(), Err: nil}
	spotPriceResOneErr := twapmock.SpotPriceResult{Sp: osmomath.OneDec(), Err: errors.New("dummy err")}
	spotPriceResOneErrNilDec := twapmock.SpotPriceResult{Sp: osmomath.Dec{}, Err: errors.New("dummy err")}
	baseTime := time.Unix(2, 0).UTC()
	updateTime := time.Unix(3, 0).UTC()
	baseTimeMinusOne := time.Unix(1, 0).UTC()

	zeroAccumNoErrSp10Record := newRecord(poolId, baseTime, osmomath.NewDec(10), zeroDec, zeroDec, zeroDec)
	sp10OneTimeUnitAccumRecord := newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10), geometricTenSecAccum)
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
			expRecord:        withSp0(withLastErrTime(sp10OneTimeUnitAccumRecord, updateTime), osmomath.ZeroDec()),
		},
		"0 accum start, sp1 err at update with nil dec": {
			record:           zeroAccumNoErrSp10Record,
			spotPriceResult0: spotPriceResOne,
			spotPriceResult1: spotPriceResOneErrNilDec,
			expRecord:        withSp1(withLastErrTime(sp10OneTimeUnitAccumRecord, updateTime), osmomath.ZeroDec()),
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
			if (test.expRecord.P0LastSpotPrice == osmomath.Dec{}) {
				test.expRecord.P0LastSpotPrice = test.spotPriceResult0.Sp
			}
			if (test.expRecord.P1LastSpotPrice == osmomath.Dec{}) {
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

			newRecord, err := s.twapkeeper.UpdateRecord(s.Ctx, test.record)
			s.Require().NoError(err)
			s.Equal(test.expRecord, newRecord)
		})
	}
}

func TestRecordWithUpdatedAccumulators(t *testing.T) {
	poolId := uint64(1)
	defaultRecord := newRecord(poolId, time.Unix(1, 0), osmomath.NewDec(10), oneDec, twoDec, pointFiveDec)
	tests := map[string]struct {
		record      types.TwapRecord
		newTime     time.Time
		expRecord   types.TwapRecord
		expectPanic bool
	}{
		"accum with zero value": {
			record:    newRecord(poolId, time.Unix(1, 0), osmomath.NewDec(10), zeroDec, zeroDec, zeroDec),
			newTime:   time.Unix(2, 0),
			expRecord: newExpRecord(OneSec.MulInt64(10), OneSec.QuoInt64(10), geometricTenSecAccum),
		},
		"small starting accumulators": {
			record:    defaultRecord,
			newTime:   time.Unix(2, 0),
			expRecord: newExpRecord(oneDec.Add(OneSec.MulInt64(10)), twoDec.Add(OneSec.QuoInt64(10)), pointFiveDec.Add(geometricTenSecAccum)),
		},
		"larger time interval": {
			record:    newRecord(poolId, time.Unix(11, 0), osmomath.NewDec(10), oneDec, twoDec, pointFiveDec),
			newTime:   time.Unix(55, 0),
			expRecord: newExpRecord(oneDec.Add(OneSec.MulInt64(44*10)), twoDec.Add(OneSec.MulInt64(44).QuoInt64(10)), pointFiveDec.Add(OneSec.MulInt64(44).Mul(logTen))),
		},
		"same time, accumulator should not change": {
			record:    defaultRecord,
			newTime:   time.Unix(1, 0),
			expRecord: newExpRecord(oneDec, twoDec, pointFiveDec),
		},
		"sp0 - zero spot price - accum0 unchanged, accum1 updated, geom accum unchanged, last err time set": {
			record:    withPrice0Set(defaultRecord, osmomath.ZeroDec()),
			newTime:   defaultRecord.Time.Add(time.Second),
			expRecord: withLastErrTime(newExpRecord(oneDec, twoDec.Add(osmomath.NewDecWithPrec(1, 1).Mul(OneSec)), pointFiveDec), defaultRecord.Time.Add(time.Second)),
		},
		"sp1 - zero spot price - accum0 updated, accum1 unchanged, geom accum updated correctly": {
			record:    withPrice1Set(defaultRecord, osmomath.ZeroDec()),
			newTime:   defaultRecord.Time.Add(time.Second),
			expRecord: newExpRecord(tenSecAccum.Add(oneDec), twoDec, pointFiveDec.Add(geometricTenSecAccum)),
		},
		"both sp - zero spot price - accum0 unchange, accum1 unchanged, geom accum unchanged": {
			record:    withPrice1Set(withPrice0Set(defaultRecord, osmomath.ZeroDec()), osmomath.ZeroDec()),
			newTime:   defaultRecord.Time.Add(time.Second),
			expRecord: withLastErrTime(newExpRecord(oneDec, twoDec, pointFiveDec), defaultRecord.Time.Add(time.Second)),
		},
		"spot price of one - geom accumulator 0": {
			record:    withPrice1Set(withPrice0Set(defaultRecord, osmomath.OneDec()), osmomath.OneDec()),
			newTime:   defaultRecord.Time.Add(time.Second),
			expRecord: newExpRecord(oneDec.Add(OneSec), twoDec.Add(OneSec), pointFiveDec),
		},

		"nanoseconds in time of the original record do not affect final result": {
			record:    withTime(defaultRecord, defaultRecord.Time.Add(oneHundredNanoseconds)),
			newTime:   time.Unix(2, 0),
			expRecord: newExpRecord(oneDec.Add(OneSec.MulInt64(10)), twoDec.Add(OneSec.QuoInt64(10)), pointFiveDec.Add(geometricTenSecAccum)),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// correct expected record based off copy/paste values
			test.expRecord.Time = test.newTime
			test.expRecord.PoolId = test.record.PoolId
			test.expRecord.P0LastSpotPrice = test.record.P0LastSpotPrice
			test.expRecord.P1LastSpotPrice = test.record.P1LastSpotPrice

			osmoassert.ConditionalPanic(t, test.expectPanic, func() {
				gotRecord := twap.RecordWithUpdatedAccumulators(test.record, test.newTime)
				require.Equal(t, test.expRecord, gotRecord)
			})
		})
	}
}

func TestRecordWithUpdatedAccumulators_ThreeAsset(t *testing.T) {
	poolId := uint64(2)
	tests := map[string]struct {
		record          []types.TwapRecord
		interpolateTime time.Time
		expRecord       []types.TwapRecord
	}{
		"accum with zero value": {
			record:          newThreeAssetRecord(poolId, time.Unix(1, 0), osmomath.NewDec(10), zeroDec, zeroDec, zeroDec, zeroDec, zeroDec, zeroDec),
			interpolateTime: time.Unix(2, 0),
			expRecord:       newThreeAssetExpRecord(poolId, OneSec.MulInt64(10), OneSec.QuoInt64(10), OneSec.MulInt64(20), geometricTenSecAccum, geometricTenSecAccum, OneSec.Mul(logOneOverTen)),
		},
		"small starting accumulators": {
			record:          newThreeAssetRecord(poolId, time.Unix(1, 0), osmomath.NewDec(10), twoDec, oneDec, twoDec, oneDec, twoDec, oneDec),
			interpolateTime: time.Unix(2, 0),
			expRecord:       newThreeAssetExpRecord(poolId, twoDec.Add(OneSec.MulInt64(10)), oneDec.Add(OneSec.QuoInt64(10)), twoDec.Add(OneSec.MulInt64(20)), oneDec.Add(geometricTenSecAccum), twoDec.Add(geometricTenSecAccum), oneDec.Add(OneSec.Mul(logOneOverTen))),
		},
		"larger time interval": {
			record:          newThreeAssetRecord(poolId, time.Unix(11, 0), osmomath.NewDec(10), twoDec, oneDec, twoDec, oneDec, twoDec, oneDec),
			interpolateTime: time.Unix(55, 0),
			expRecord:       newThreeAssetExpRecord(poolId, twoDec.Add(OneSec.MulInt64(44*10)), oneDec.Add(OneSec.MulInt64(44).QuoInt64(10)), twoDec.Add(OneSec.MulInt64(44*20)), oneDec.Add(OneSec.MulInt64(44).Mul(logTen)), twoDec.Add(OneSec.MulInt64(44).Mul(logTen)), oneDec.Add(OneSec.MulInt64(44).Mul(logOneOverTen))),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for i := range test.record {
				// correct expected record based off copy/paste values
				test.expRecord[i].Time = test.interpolateTime
				test.expRecord[i].P0LastSpotPrice = test.record[i].P0LastSpotPrice
				test.expRecord[i].P1LastSpotPrice = test.record[i].P1LastSpotPrice

				gotRecord := twap.RecordWithUpdatedAccumulators(test.record[i], test.interpolateTime)
				require.Equal(t, test.expRecord[i], gotRecord)
			}
		})
	}
}

func (s *TestSuite) TestGetInterpolatedRecord() {
	baseRecord := newTwoAssetPoolTwapRecordWithDefaults(baseTime, osmomath.OneDec(), osmomath.OneDec(), osmomath.OneDec(), osmomath.OneDec().Quo(twoDec))

	// all tests occur with updateTime = base time + time.Unix(1, 0)
	tests := map[string]struct {
		recordsToPreSet     types.TwapRecord
		testPoolId          uint64
		testDenom0          string
		testDenom1          string
		testTime            time.Time
		expectedAccumulator osmomath.Dec
		expectedErr         error
		expectedLastErrTime time.Time
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
			expectedAccumulator: baseRecord.P0ArithmeticTwapAccumulator.Add(osmomath.NewDec(1000)),
		},
		"call 1 second after existing record with error": {
			recordsToPreSet: withLastErrTime(baseRecord, baseTime),
			testPoolId:      baseRecord.PoolId,
			testDenom0:      baseRecord.Asset0Denom,
			testDenom1:      baseRecord.Asset1Denom,
			testTime:        baseTime.Add(time.Second),
			// 1(spot price) * 1000(one sec in milli-seconds)
			expectedAccumulator: baseRecord.P0ArithmeticTwapAccumulator.Add(osmomath.NewDec(1000)),
			expectedLastErrTime: baseTime.Add(time.Second),
		},
		"call 1 second before existing record": {
			recordsToPreSet: baseRecord,
			testPoolId:      baseRecord.PoolId,
			testDenom0:      baseRecord.Asset0Denom,
			testDenom1:      baseRecord.Asset1Denom,
			testTime:        baseTime.Add(-time.Second),
			expectedErr: fmt.Errorf("looking for a time that's too old, not in the historical index. "+
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
				if test.recordsToPreSet.Time.Equal(test.recordsToPreSet.LastErrorTime) {
					// last error time updated
					s.Require().Equal(test.testTime, interpolatedRecord.LastErrorTime)
				} else {
					// last error time unchanged
					s.Require().Equal(test.recordsToPreSet.LastErrorTime, interpolatedRecord.LastErrorTime)
				}
			}
		})
	}
}

func (s *TestSuite) TestGetInterpolatedRecord_ThreeAsset() {
	baseRecord := newThreeAssetRecord(2, baseTime, osmomath.NewDec(10), osmomath.ZeroDec(), osmomath.ZeroDec(), osmomath.ZeroDec(), osmomath.ZeroDec(), osmomath.ZeroDec(), osmomath.ZeroDec())
	// all tests occur with updateTime = base time + time.Unix(1, 0)
	tests := map[string]struct {
		recordsToPreSet       []types.TwapRecord
		testTime              time.Time
		expectedP0Accumulator []osmomath.Dec
		expectedP1Accumulator []osmomath.Dec
		expectedErr           error
	}{
		"call 1 second after existing record": {
			recordsToPreSet: baseRecord,
			testTime:        baseTime.Add(time.Second),
			// P0 and P1 TwapAccumulators both start at 0
			// A 10 spot price * 1000ms = 10000
			// A 10 spot price * 1000ms = 10000
			// B .1 spot price * 1000ms = 100
			expectedP0Accumulator: []osmomath.Dec{
				baseRecord[0].P0ArithmeticTwapAccumulator.Add(osmomath.NewDec(10000)),
				baseRecord[1].P0ArithmeticTwapAccumulator.Add(osmomath.NewDec(10000)),
				baseRecord[2].P0ArithmeticTwapAccumulator.Add(osmomath.NewDec(100)),
			},
			// B .1 spot price * 1000ms = 100
			// C 20 spot price * 1000ms = 20000
			// C 20 spot price * 1000ms = 20000
			expectedP1Accumulator: []osmomath.Dec{
				baseRecord[0].P1ArithmeticTwapAccumulator.Add(osmomath.NewDec(100)),
				baseRecord[1].P1ArithmeticTwapAccumulator.Add(osmomath.NewDec(20000)),
				baseRecord[2].P1ArithmeticTwapAccumulator.Add(osmomath.NewDec(20000)),
			},
		},
		"call 1 second after existing record with error": {
			recordsToPreSet: []types.TwapRecord{
				withLastErrTime(baseRecord[0], baseTime),
				withLastErrTime(baseRecord[1], baseTime),
				withLastErrTime(baseRecord[2], baseTime),
			},
			testTime: baseTime.Add(time.Second),
			// P0 and P1 TwapAccumulators both start at 0
			// A 10 spot price * 1000ms = 10000
			// A 10 spot price * 1000ms = 10000
			// B .1 spot price * 1000ms = 100
			expectedP0Accumulator: []osmomath.Dec{
				baseRecord[0].P0ArithmeticTwapAccumulator.Add(osmomath.NewDec(10000)),
				baseRecord[1].P0ArithmeticTwapAccumulator.Add(osmomath.NewDec(10000)),
				baseRecord[2].P0ArithmeticTwapAccumulator.Add(osmomath.NewDec(100)),
			},
			// B .1 spot price * 1000ms = 100
			// C 20 spot price * 1000ms = 20000
			// C 20 spot price * 1000ms = 20000
			expectedP1Accumulator: []osmomath.Dec{
				baseRecord[0].P1ArithmeticTwapAccumulator.Add(osmomath.NewDec(100)),
				baseRecord[1].P1ArithmeticTwapAccumulator.Add(osmomath.NewDec(20000)),
				baseRecord[2].P1ArithmeticTwapAccumulator.Add(osmomath.NewDec(20000)),
			},
		},
		"call 1 second before existing record": {
			recordsToPreSet: baseRecord,
			testTime:        baseTime.Add(-time.Second),
			expectedErr: fmt.Errorf("looking for a time that's too old, not in the historical index. "+
				" Try storing the accumulator value. (requested time %s)", baseTime.Add(-time.Second)),
		},
		"test non lexicographical order parameter": {
			recordsToPreSet: baseRecord,
			testTime:        baseTime,
		},
	}

	for name, test := range tests {
		s.Run(name, func() {
			s.SetupTest()
			for i := range test.recordsToPreSet {
				s.twapkeeper.StoreNewRecord(s.Ctx, test.recordsToPreSet[i])

				interpolatedRecord, err := s.twapkeeper.GetInterpolatedRecord(s.Ctx, baseRecord[i].PoolId, baseRecord[i].Asset0Denom, baseRecord[i].Asset1Denom, test.testTime)
				if test.expectedErr != nil {
					s.Require().Error(err)
					s.Require().Equal(test.expectedErr.Error(), err.Error())
					return
				}
				s.Require().NoError(err)

				if test.testTime.Equal(baseTime) {
					s.Require().Equal(test.recordsToPreSet[i], interpolatedRecord)
				} else {
					s.Require().Equal(test.testTime, interpolatedRecord.Time)
					s.Require().Equal(test.recordsToPreSet[i].P0LastSpotPrice, interpolatedRecord.P0LastSpotPrice)
					s.Require().Equal(test.recordsToPreSet[i].P1LastSpotPrice, interpolatedRecord.P1LastSpotPrice)
					s.Require().Equal(test.expectedP0Accumulator[i], interpolatedRecord.P0ArithmeticTwapAccumulator)
					s.Require().Equal(test.expectedP1Accumulator[i], interpolatedRecord.P1ArithmeticTwapAccumulator)
					if test.recordsToPreSet[i].Time.Equal(test.recordsToPreSet[i].LastErrorTime) {
						// last error time updated
						s.Require().Equal(test.testTime, interpolatedRecord.LastErrorTime)
					} else {
						// last error time unchanged
						s.Require().Equal(test.recordsToPreSet[i].LastErrorTime, interpolatedRecord.LastErrorTime)
					}
				}
			}
		})
	}
}

type computeThreeAssetArithmeticTwapTestCase struct {
	startRecord []types.TwapRecord
	endRecord   []types.TwapRecord
	quoteAsset  []string
	expTwap     []osmomath.Dec
	expErr      bool
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
		baseDenom   string
		quoteDenom  string
		overrideSp  osmomath.Dec
		overrideErr error
	}

	type expectedResults struct {
		spotPriceA    osmomath.Dec
		spotPriceB    osmomath.Dec
		lastErrorTime time.Time
		isMostRecent  bool
	}

	spError := errors.New("spot price error")

	validateRecords := func(expectedRecords []expectedResults, actualRecords []types.TwapRecord) {
		s.Require().Equal(len(expectedRecords), len(actualRecords))
		for i, r := range expectedRecords {
			s.Require().Equal(r.spotPriceA, actualRecords[i].P0LastSpotPrice, "record %d", i)
			s.Require().Equal(r.spotPriceB, actualRecords[i].P1LastSpotPrice, "record %d", i)
			s.Require().Equal(r.lastErrorTime, actualRecords[i].LastErrorTime, "record %d", i)
		}
	}

	tests := map[string]struct {
		preSetRecords     []types.TwapRecord
		poolId            uint64
		ammMock           twapmock.ProgrammedPoolManagerInterface
		spOverrides       []spOverride
		poolDenomOverride []string

		blockTime   time.Time
		blockHeight int64

		expectedHistoricalRecords []expectedResults
		expectError               error
	}{
		"no records pre-set; error": {
			preSetRecords: []types.TwapRecord{},
			poolId:        1,
			blockTime:     baseTime,

			expectError: poolmanagertypes.FailedToFindRouteError{PoolId: baseRecord.PoolId},
		},
		"existing records in different pool; no-op": {
			preSetRecords: []types.TwapRecord{baseRecord},
			poolId:        baseRecord.PoolId + 1,
			blockTime:     baseTime.Add(time.Second),

			expectError: poolmanagertypes.FailedToFindRouteError{PoolId: baseRecord.PoolId + 1},
		},
		"the returned number of records does not match expected": {
			preSetRecords: []types.TwapRecord{baseRecord},
			poolId:        baseRecord.PoolId,
			blockTime:     baseRecord.Time.Add(time.Second),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: osmomath.NewDec(2),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: osmomath.NewDecWithPrec(2, 1),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: "extradenom",
					overrideSp: osmomath.NewDecWithPrec(3, 1),
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
					overrideSp: osmomath.NewDec(2),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: osmomath.NewDecWithPrec(2, 1),
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
					spotPriceA:   osmomath.NewDec(2),
					spotPriceB:   osmomath.NewDecWithPrec(2, 1),
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
					overrideSp:  osmomath.NewDecWithPrec(2, 1),
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
					spotPriceA:    osmomath.ZeroDec(),
					spotPriceB:    osmomath.NewDecWithPrec(2, 1),
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
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:   baseRecord.Asset1Denom,
					quoteDenom:  baseRecord.Asset0Denom,
					overrideSp:  types.MaxSpotPrice.Add(osmomath.OneDec()),
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
					spotPriceA:    osmomath.OneDec(),
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
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: osmomath.OneDec(),
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
					spotPriceA:    osmomath.OneDec(),
					spotPriceB:    osmomath.OneDec(),
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
					overrideSp: osmomath.OneDec(),
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
					spotPriceA:    osmomath.OneDec(),
					spotPriceB:    osmomath.ZeroDec(),
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
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
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
					spotPriceA:   osmomath.OneDec(),
					spotPriceB:   osmomath.OneDec().Add(osmomath.OneDec()),
					isMostRecent: true,
				},
			},
		},
		// should always be greater than the last record's time.
		"new record can't be inserted prior to the last record's update time": {
			preSetRecords: []types.TwapRecord{baseRecord, tPlus10sp5Record},
			poolId:        baseRecord.PoolId,

			blockTime: baseRecord.Time.Add(time.Second * 5),

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
				},
			},

			expectError: types.InvalidUpdateRecordError{
				RecordBlockHeight: tPlus10sp5Record.Height,
				RecordTime:        tPlus10sp5Record.Time,
				ActualBlockHeight: (baseRecord.Height + 1),
				ActualTime:        baseRecord.Time.Add(time.Second * 5),
			},
		},
		// should always be greater than the last record's block.
		"new record can't be inserted before the last record's update block": {
			preSetRecords: []types.TwapRecord{mostRecentRecordPoolOne},
			poolId:        baseRecord.PoolId,

			// Even if lastRecord.Time < ctx.Time,
			// lastRecord.Height >= ctx.BlockHeight also throws error
			blockTime:   mostRecentRecordPoolOne.Time.Add(time.Second),
			blockHeight: mostRecentRecordPoolOne.Height - 1,

			spOverrides: []spOverride{
				{
					baseDenom:  mostRecentRecordPoolOne.Asset0Denom,
					quoteDenom: mostRecentRecordPoolOne.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  mostRecentRecordPoolOne.Asset1Denom,
					quoteDenom: mostRecentRecordPoolOne.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
				},
			},

			expectError: types.InvalidUpdateRecordError{
				RecordBlockHeight: mostRecentRecordPoolOne.Height,
				RecordTime:        mostRecentRecordPoolOne.Time,
				ActualBlockHeight: mostRecentRecordPoolOne.Height - 1,
				ActualTime:        mostRecentRecordPoolOne.Time.Add(time.Second),
			},
		},
		"new record can be update in same block with last record if accumulators are zero (afterPoolCreate hook called)": {
			preSetRecords: []types.TwapRecord{baseRecord},
			poolId:        baseRecord.PoolId,

			// Even if lastRecord.Time < ctx.Time,
			// lastRecord.Height >= ctx.BlockHeight also throws error
			blockTime: baseRecord.Time,

			spOverrides: []spOverride{
				{
					baseDenom:  baseRecord.Asset0Denom,
					quoteDenom: baseRecord.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  baseRecord.Asset1Denom,
					quoteDenom: baseRecord.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
				},
			},

			expectedHistoricalRecords: []expectedResults{
				{
					spotPriceA:   osmomath.OneDec(),
					spotPriceB:   osmomath.OneDec().Add(osmomath.OneDec()),
					isMostRecent: true,
				},
			},

			expectError: nil,
		},
		"new record can't be updated in same block with last record if accumulators not equal to zero": {
			preSetRecords: []types.TwapRecord{mostRecentRecordPoolOne},
			poolId:        mostRecentRecordPoolOne.PoolId,

			// Even if lastRecord.Time < ctx.Time,
			// lastRecord.Height >= ctx.BlockHeight also throws error
			blockTime: mostRecentRecordPoolOne.Time,

			spOverrides: []spOverride{
				{
					baseDenom:  mostRecentRecordPoolOne.Asset0Denom,
					quoteDenom: mostRecentRecordPoolOne.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  mostRecentRecordPoolOne.Asset1Denom,
					quoteDenom: mostRecentRecordPoolOne.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
				},
			},
			expectError: types.InvalidUpdateRecordError{},
		},
		"multi-asset pool; pre-set at t and t + 1; creates new records": {
			preSetRecords: []types.TwapRecord{threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC, tPlus10sp5ThreeAssetRecordAB, tPlus10sp5ThreeAssetRecordAC, tPlus10sp5ThreeAssetRecordBC},
			poolId:        threeAssetRecordAB.PoolId,
			blockTime:     threeAssetRecordAB.Time.Add(time.Second * 11),
			spOverrides: []spOverride{
				{
					baseDenom:  threeAssetRecordAB.Asset0Denom,
					quoteDenom: threeAssetRecordAB.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  threeAssetRecordAB.Asset1Denom,
					quoteDenom: threeAssetRecordAB.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
				},
				{
					baseDenom:  threeAssetRecordAC.Asset0Denom,
					quoteDenom: threeAssetRecordAC.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  threeAssetRecordAC.Asset1Denom,
					quoteDenom: threeAssetRecordAC.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()).Add(osmomath.OneDec()),
				},
				{
					baseDenom:  threeAssetRecordBC.Asset0Denom,
					quoteDenom: threeAssetRecordBC.Asset1Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
				},
				{
					baseDenom:  threeAssetRecordBC.Asset1Denom,
					quoteDenom: threeAssetRecordBC.Asset0Denom,
					overrideSp: osmomath.OneDec(),
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record AB at t.
				{
					spotPriceA: threeAssetRecordAB.P0LastSpotPrice,
					spotPriceB: threeAssetRecordAB.P1LastSpotPrice,
				},
				// The original record AB at t + 1.
				{
					spotPriceA: tPlus10sp5ThreeAssetRecordAB.P0LastSpotPrice,
					spotPriceB: tPlus10sp5ThreeAssetRecordAB.P1LastSpotPrice,
				},
				// The new record AB added.
				{
					spotPriceA:   osmomath.OneDec(),
					spotPriceB:   osmomath.OneDec().Add(osmomath.OneDec()),
					isMostRecent: true,
				},
				// The original record AC at t.
				{
					spotPriceA: threeAssetRecordAC.P0LastSpotPrice,
					spotPriceB: threeAssetRecordAC.P1LastSpotPrice,
				},
				// The original record AC at t + 1.
				{
					spotPriceA: tPlus10sp5ThreeAssetRecordAC.P0LastSpotPrice,
					spotPriceB: tPlus10sp5ThreeAssetRecordAC.P1LastSpotPrice,
				},
				// The new record AC added.
				{
					spotPriceA:   osmomath.OneDec(),
					spotPriceB:   osmomath.OneDec().Add(osmomath.OneDec()).Add(osmomath.OneDec()),
					isMostRecent: true,
				},
				// The original record BC at t.
				{
					spotPriceA: threeAssetRecordBC.P0LastSpotPrice,
					spotPriceB: threeAssetRecordBC.P1LastSpotPrice,
				},
				// The original record BC at t + 1.
				{
					spotPriceA: tPlus10sp5ThreeAssetRecordBC.P0LastSpotPrice,
					spotPriceB: tPlus10sp5ThreeAssetRecordBC.P1LastSpotPrice,
				},
				// The new record BC added.
				{
					spotPriceA:   osmomath.OneDec().Add(osmomath.OneDec()),
					spotPriceB:   osmomath.OneDec(),
					isMostRecent: true,
				},
			},
		},
		"multi-asset pool; pre-set at t and t + 1 with err, large spot price, overwrites error time": {
			preSetRecords: []types.TwapRecord{threeAssetRecordAB, threeAssetRecordAC, threeAssetRecordBC, withLastErrTime(tPlus10sp5ThreeAssetRecordAB, tPlus10sp5ThreeAssetRecordAB.Time), tPlus10sp5ThreeAssetRecordAC, tPlus10sp5ThreeAssetRecordBC},
			poolId:        threeAssetRecordAB.PoolId,
			blockTime:     threeAssetRecordAB.Time.Add(time.Second * 11),
			spOverrides: []spOverride{
				{
					baseDenom:  threeAssetRecordAB.Asset0Denom,
					quoteDenom: threeAssetRecordAB.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  threeAssetRecordAB.Asset1Denom,
					quoteDenom: threeAssetRecordAB.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
				},
				{
					baseDenom:  threeAssetRecordAC.Asset0Denom,
					quoteDenom: threeAssetRecordAC.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:   threeAssetRecordAC.Asset1Denom,
					quoteDenom:  threeAssetRecordAC.Asset0Denom,
					overrideSp:  types.MaxSpotPrice.Add(osmomath.OneDec()),
					overrideErr: nil, // twap logic should identify the large spot price and mark it as error.
				},
				{
					baseDenom:  threeAssetRecordBC.Asset0Denom,
					quoteDenom: threeAssetRecordBC.Asset1Denom,
					overrideSp: osmomath.OneDec(),
				},
				{
					baseDenom:  threeAssetRecordBC.Asset1Denom,
					quoteDenom: threeAssetRecordBC.Asset0Denom,
					overrideSp: osmomath.OneDec().Add(osmomath.OneDec()),
				},
			},

			expectedHistoricalRecords: []expectedResults{
				// The original record AB at t.
				{
					spotPriceA: threeAssetRecordAB.P0LastSpotPrice,
					spotPriceB: threeAssetRecordAB.P1LastSpotPrice,
				},
				// The original record AB at t + 1.
				{
					spotPriceA:    tPlus10sp5ThreeAssetRecordAB.P0LastSpotPrice,
					spotPriceB:    tPlus10sp5ThreeAssetRecordAB.P1LastSpotPrice,
					lastErrorTime: tPlus10sp5ThreeAssetRecordAB.Time,
				},
				// The new record AB added.
				{
					spotPriceA:    osmomath.OneDec(),
					spotPriceB:    osmomath.OneDec().Add(osmomath.OneDec()),
					lastErrorTime: tPlus10sp5ThreeAssetRecordAB.Time,
					isMostRecent:  true,
				},
				// The original record AC at t.
				{
					spotPriceA: threeAssetRecordAC.P0LastSpotPrice,
					spotPriceB: threeAssetRecordAC.P1LastSpotPrice,
				},
				// The original record AC at t + 1.
				{
					spotPriceA: tPlus10sp5ThreeAssetRecordAC.P0LastSpotPrice,
					spotPriceB: tPlus10sp5ThreeAssetRecordAC.P1LastSpotPrice,
				},
				// The new record AC added.
				{
					spotPriceA:    osmomath.OneDec(),
					spotPriceB:    types.MaxSpotPrice,                            // Although the price returned from AMM was MaxSpotPrice + 1, it is reset to just MaxSpotPrice.
					lastErrorTime: threeAssetRecordAC.Time.Add(time.Second * 11), // equals to block time
					isMostRecent:  true,
				},
				// The original record BC at t.
				{
					spotPriceA: threeAssetRecordBC.P0LastSpotPrice,
					spotPriceB: threeAssetRecordBC.P1LastSpotPrice,
				},
				// The original record BC at t + 1.
				{
					spotPriceA: tPlus10sp5ThreeAssetRecordBC.P0LastSpotPrice,
					spotPriceB: tPlus10sp5ThreeAssetRecordBC.P1LastSpotPrice,
				},
				// The new record BC added.
				{
					spotPriceA:   osmomath.OneDec(),
					spotPriceB:   osmomath.OneDec().Add(osmomath.OneDec()),
					isMostRecent: true,
				},
			},
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			twapKeeper := s.App.TwapKeeper
			ctx := s.Ctx.WithBlockTime(tc.blockTime)
			if tc.blockHeight != 0 {
				ctx = s.Ctx.WithBlockTime(tc.blockTime).WithBlockHeight(tc.blockHeight)
			}

			if len(tc.spOverrides) > 0 {
				ammMock := twapmock.NewProgrammedAmmInterface(s.App.PoolManagerKeeper)
				for _, sp := range tc.spOverrides {
					ammMock.ProgramPoolSpotPriceOverride(tc.poolId, sp.baseDenom, sp.quoteDenom, sp.overrideSp, sp.overrideErr)
					ammMock.ProgramPoolDenomsOverride(tc.poolId, []string{sp.baseDenom, sp.quoteDenom}, nil)
				}

				twapKeeper.SetAmmInterface(ammMock)
			} else if len(tc.poolDenomOverride) > 0 {
				ammMock := twapmock.NewProgrammedAmmInterface(s.App.PoolManagerKeeper)
				ammMock.ProgramPoolDenomsOverride(tc.poolId, tc.poolDenomOverride, nil)
				twapKeeper.SetAmmInterface(ammMock)
			}

			s.preSetRecords(tc.preSetRecords)

			err := twapKeeper.UpdateRecords(ctx, tc.poolId)

			if tc.expectError != nil {
				s.Require().ErrorAs(err, &tc.expectError)
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

func (s *TestSuite) TestAfterCreatePool() {
	tests := map[string]struct {
		poolId    uint64
		poolCoins sdk.Coins
		// if this field is set true, we swap in the same block with pool creation
		runSwap     bool
		expectedErr bool
	}{
		"Pool not existing": {
			poolId:      2,
			expectedErr: true,
		},
		"Default Pool, no swap on pool creation block": {
			poolId:    1,
			poolCoins: defaultTwoAssetCoins,
			runSwap:   false,
		},
		"Default Pool, swap on pool creation block": {
			poolId:    1,
			poolCoins: defaultTwoAssetCoins,
			runSwap:   true,
		},
		"Multi assets pool, no swap on pool creation block": {
			poolId:    1,
			poolCoins: defaultThreeAssetCoins,
			runSwap:   false,
		},
		"Multi assets pool, swap on pool creation block": {
			poolId:    1,
			poolCoins: defaultThreeAssetCoins,
			runSwap:   true,
		},
	}

	for name, tc := range tests {
		s.Run(name, func() {
			s.SetupTest()
			var poolId uint64

			// set up pool with input coins
			if tc.poolCoins != nil {
				poolId = s.PrepareBalancerPoolWithCoins(tc.poolCoins...)
				if tc.runSwap {
					s.RunBasicSwap(poolId)
				}
			}

			err := s.twapkeeper.AfterCreatePool(s.Ctx, tc.poolId)
			if tc.expectedErr {
				s.Require().Error(err)
				return
			}
			s.Require().Equal(tc.poolId, poolId)
			s.Require().NoError(err)

			denoms := tc.poolCoins.Denoms()
			denomPairs := types.GetAllUniqueDenomPairs(denoms)
			expectedRecords := []types.TwapRecord{}
			for _, denomPair := range denomPairs {
				expectedRecord, err := twap.NewTwapRecord(s.App.PoolManagerKeeper, s.Ctx, poolId, denomPair.Denom0, denomPair.Denom1)
				s.Require().NoError(err)
				expectedRecords = append(expectedRecords, expectedRecord)
			}

			// consistency check that the number of records is exactly equal to the number of denompairs
			allRecords, err := s.twapkeeper.GetAllMostRecentRecordsForPool(s.Ctx, poolId)
			s.Require().NoError(err)
			s.Require().Equal(len(denomPairs), len(allRecords))
			s.Require().Equal(len(expectedRecords), len(allRecords))

			// check on the correctness of all individual twap records
			for i, denomPair := range denomPairs {
				actualRecord, err := s.twapkeeper.GetMostRecentRecordStoreRepresentation(s.Ctx, poolId, denomPair.Denom0, denomPair.Denom1)
				s.Require().NoError(err)
				s.Require().Equal(expectedRecords[i], actualRecord)
				actualRecord, err = s.twapkeeper.GetRecordAtOrBeforeTime(s.Ctx, poolId, s.Ctx.BlockTime(), denomPair.Denom0, denomPair.Denom1)
				s.Require().NoError(err)
				s.Require().Equal(expectedRecords[i], actualRecord)
			}

			// test that after creating a pool
			// has triggered `trackChangedPool`,
			// and that we have the state of price impacted pools.
			changedPools := s.twapkeeper.GetChangedPools(s.Ctx)
			s.Require().Equal(1, len(changedPools))
			s.Require().Equal(tc.poolId, changedPools[0])
		})
	}
}

// This tests the behavior of computeArithmeticTwap, around error returning
// when there has been an intermediate spot price error.
func (s *TestSuite) TestComputeArithmeticTwapWithSpotPriceError() {
	newOneSidedRecordWErrorTime := func(time time.Time, accum osmomath.Dec, useP0 bool, errTime time.Time) types.TwapRecord {
		record := newOneSidedRecord(time, accum, useP0)
		record.LastErrorTime = errTime
		return record
	}

	arithStrategy := &twap.ArithmeticTwapStrategy{
		TwapKeeper: *s.App.TwapKeeper,
	}

	tests := map[string]computeTwapTestCase{
		// should error, since end time may have been used to interpolate this value
		"errAtEndTime from end record": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec, true, tPlusOne),
			quoteAsset:  denom0,
			expTwap:     osmomath.OneDec(),
			expErr:      true,
		},
		// should error, since start time may have been used to interpolate this value
		"err at StartTime exactly from end record": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec, true, baseTime),
			quoteAsset:  denom0,
			expTwap:     osmomath.OneDec(),
			expErr:      true,
		},
		// should error, since start record is erroneous
		"err at StartTime exactly from start record": {
			startRecord: newOneSidedRecordWErrorTime(baseTime, osmomath.ZeroDec(), true, baseTime),
			endRecord:   newOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  denom0,
			expTwap:     osmomath.OneDec(),
			expErr:      true,
		},
		"err before StartTime": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec, true, tMinOne),
			quoteAsset:  denom0,
			expTwap:     osmomath.OneDec(),
			expErr:      false,
		},
		// Should not happen, but if it did would error
		"err after EndTime": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecordWErrorTime(tPlusOne, OneSec.MulInt64(2), true, baseTime.Add(20*time.Second)),
			quoteAsset:  denom0,
			expTwap:     osmomath.OneDec().MulInt64(2),
			expErr:      true,
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			actualTwap, err := twap.ComputeTwap(test.startRecord, test.endRecord, test.quoteAsset, arithStrategy)
			s.Require().Equal(test.expTwap, actualTwap)
			osmoassert.ConditionalError(s.T(), test.expErr, err)
		})
	}
}

// TestTwapLog_CorrectBase tests that the base of 2 is used for the twap log function.
// log_2{16} = 4
func (s *TestSuite) TestTwapLog_CorrectBase() {
	logOf := osmomath.NewDec(16)
	expectedValue := osmomath.NewDec(4)

	result := twap.TwapLog(logOf)

	s.Require().Equal(expectedValue, result)
}

func (s *TestSuite) TestTwapLog() {
	smallestAdditiveTolerance := osmomath.ErrTolerance{
		AdditiveTolerance: osmomath.SmallestDec(),
	}

	testcases := []struct {
		name        string
		price       osmomath.Dec
		expected    osmomath.Dec
		expectPanic bool
	}{
		{
			"max spot price",
			gammtypes.MaxSpotPrice,
			// log_2{2^128 - 1} = 128
			osmomath.MustNewDecFromStr("127.999999999999999999"),
			false,
		},
		{
			"zero price - panic",
			osmomath.ZeroDec(),
			osmomath.Dec{},
			true,
		},
		{
			"smallest dec",
			osmomath.SmallestDec(),
			// https://www.wolframalpha.com/input?i=log+base+2+of+%2810%5E-18%29+with+20+digits
			osmomath.MustNewDecFromStr("59.794705707972522262").Neg(),
			false,
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			osmoassert.ConditionalPanic(s.T(), tc.expectPanic, func() {
				result := twap.TwapLog(tc.price)

				smallestAdditiveTolerance.CompareBigDec(
					osmomath.BigDecFromDec(tc.expected),
					osmomath.BigDecFromDec(result),
				)
			})
		})
	}
}

func testCaseFromDeltas(s *TestSuite, startAccum, accumDiff osmomath.Dec, timeDelta time.Duration, expectedTwap osmomath.Dec) computeTwapTestCase {
	return computeTwapTestCase{
		newOneSidedRecord(baseTime, startAccum, true),
		newOneSidedRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff), true),
		[]twap.TwapStrategy{
			&twap.ArithmeticTwapStrategy{
				TwapKeeper: *s.App.TwapKeeper,
			},
		},
		denom0,
		expectedTwap,
		false,
		false,
	}
}

func testCaseFromDeltasAsset1(s *TestSuite, startAccum, accumDiff osmomath.Dec, timeDelta time.Duration, expectedTwap osmomath.Dec) computeTwapTestCase {
	return computeTwapTestCase{
		newOneSidedRecord(baseTime, startAccum, false),
		newOneSidedRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff), false),
		[]twap.TwapStrategy{
			&twap.ArithmeticTwapStrategy{
				TwapKeeper: *s.App.TwapKeeper,
			},
		},
		denom1,
		expectedTwap,
		false,
		false,
	}
}

func geometricTestCaseFromDeltas0(s *TestSuite, startAccum, accumDiff osmomath.Dec, timeDelta time.Duration, expectedTwap osmomath.Dec) computeTwapTestCase {
	return computeTwapTestCase{
		newOneSidedGeometricRecord(baseTime, startAccum),
		newOneSidedGeometricRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff)),
		[]twap.TwapStrategy{
			&twap.GeometricTwapStrategy{
				TwapKeeper: *s.App.TwapKeeper,
			},
		},
		denom0,
		expectedTwap,
		false,
		false,
	}
}

func geometricTestCaseFromDeltas1(s *TestSuite, startAccum, accumDiff osmomath.Dec, timeDelta time.Duration, expectedTwap osmomath.Dec) computeTwapTestCase {
	return geometricTestCaseFromDeltas0(s, startAccum, accumDiff, timeDelta, osmomath.OneDec().Quo(expectedTwap))
}

func testThreeAssetCaseFromDeltas(startAccum, accumDiff osmomath.Dec, timeDelta time.Duration, expectedTwap osmomath.Dec) computeThreeAssetArithmeticTwapTestCase {
	return computeThreeAssetArithmeticTwapTestCase{
		newThreeAssetOneSidedRecord(baseTime, startAccum, true),
		newThreeAssetOneSidedRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff), true),
		[]string{denom0, denom0, denom1},
		[]osmomath.Dec{expectedTwap, expectedTwap, expectedTwap},
		false,
	}
}
