package twap_test

import (
	"time"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v27/x/twap"
	"github.com/osmosis-labs/osmosis/v27/x/twap/types"
)

type computeTwapTestCase struct {
	startRecord    types.TwapRecord
	endRecord      types.TwapRecord
	twapStrategies []twap.TwapStrategy
	quoteAsset     string
	expTwap        osmomath.Dec
	expErr         bool
	expPanic       bool
}

// TestComputeArithmeticTwap tests computeTwap on various inputs.
// The test vectors are structured by setting up different start and records,
// based on time interval, and their accumulator values.
// Then an expected TWAP is provided in each test case, to compare against computed.
func (s *TestSuite) TestComputeTwap() {
	arithStrategy := &twap.ArithmeticTwapStrategy{
		TwapKeeper: *s.App.TwapKeeper,
	}

	geomStrategy := &twap.GeometricTwapStrategy{
		TwapKeeper: *s.App.TwapKeeper,
	}

	tests := map[string]computeTwapTestCase{
		"arithmetic only, basic: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  denom0,
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
			},
			expTwap: osmomath.OneDec(),
		},
		// this test just shows what happens in case the records are reversed.
		// It should return the correct result, even though this is incorrect internal API usage
		"arithmetic only: invalid call: reversed records of above": {
			startRecord: newOneSidedRecord(tPlusOne, OneSec, true),
			endRecord:   newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			quoteAsset:  denom0,
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
			},
			expTwap: osmomath.OneDec(),
		},
		"same record: denom0, end spot price = 0": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			quoteAsset:  denom0,
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
				geomStrategy,
			},
			expTwap: osmomath.ZeroDec(),
		},
		"same record: denom1, end spot price = 1": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			quoteAsset:  denom1,
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
				geomStrategy,
			},
			expTwap: osmomath.OneDec(),
		},
		"arithmetic only: accumulator = 10*OneSec, t=5s. 0 base accum": testCaseFromDeltas(
			s, osmomath.ZeroDec(), tenSecAccum, 5*time.Second, osmomath.NewDec(2)),
		"arithmetic only: accumulator = 10*OneSec, t=100s. 0 base accum (asset 1)": testCaseFromDeltasAsset1(s, osmomath.ZeroDec(), OneSec.MulInt64(10), 100*time.Second, osmomath.NewDecWithPrec(1, 1)),
		"geometric only: accumulator = log(10)*OneSec, t=5s. 0 base accum": geometricTestCaseFromDeltas0(
			s, osmomath.ZeroDec(), geometricTenSecAccum, 5*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(5*1000))),
		"geometric only: accumulator = log(10)*OneSec, t=100s. 0 base accum (asset 1)": geometricTestCaseFromDeltas1(s, osmomath.ZeroDec(), geometricTenSecAccum, 100*time.Second, osmomath.OneDec().Quo(twap.TwapPow(geometricTenSecAccum.QuoInt64(100*1000)))),
		"three asset same record: asset1, end spot price = 1": {
			startRecord: newThreeAssetOneSidedRecord(baseTime, osmomath.ZeroDec(), true)[1],
			endRecord:   newThreeAssetOneSidedRecord(baseTime, osmomath.ZeroDec(), true)[1],
			quoteAsset:  denom2,
			expTwap:     osmomath.OneDec(),
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
				geomStrategy,
			},
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			for _, twapStrategy := range test.twapStrategies {
				actualTwap, err := twap.ComputeTwap(test.startRecord, test.endRecord, test.quoteAsset, twapStrategy)
				s.Require().NoError(err)
				osmoassert.DecApproxEq(s.T(), test.expTwap, actualTwap, osmomath.GetPowPrecision())
			}
		})
	}
}

// TestComputeArithmeticStrategyTwap tests arithmetic strategy's computeTwap
// Contrary to computeTwap function (logic.go) that handles the cases with zero delta correctly,
// this function should panic in case of zero delta.
func (s *TestSuite) TestComputeArithmeticStrategyTwap() {
	pointOneAccum := OneSec.QuoInt64(10)
	tests := map[string]computeTwapTestCase{
		"basic: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  denom0,
			expTwap:     osmomath.OneDec(),
		},
		// this test just shows what happens in case the records are reversed.
		// It should return the correct result, even though this is incorrect internal API usage
		"invalid call: reversed records of above": {
			startRecord: newOneSidedRecord(tPlusOne, OneSec, true),
			endRecord:   newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			quoteAsset:  denom0,
			expTwap:     osmomath.OneDec(),
		},
		"same record (zero time delta), division by 0 - panic": {
			startRecord: newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			quoteAsset:  denom0,
			expPanic:    true,
		},
		"accumulator = 10*OneSec, t=5s. 0 base accum": testCaseFromDeltas(
			s, osmomath.ZeroDec(), tenSecAccum, 5*time.Second, osmomath.NewDec(2)),
		"accumulator = 10*OneSec, t=3s. 0 base accum": testCaseFromDeltas(
			s, osmomath.ZeroDec(), tenSecAccum, 3*time.Second, ThreePlusOneThird),
		"accumulator = 10*OneSec, t=100s. 0 base accum": testCaseFromDeltas(
			s, osmomath.ZeroDec(), tenSecAccum, 100*time.Second, osmomath.NewDecWithPrec(1, 1)),

		// test that base accum has no impact
		"accumulator = 10*OneSec, t=5s. 10 base accum": testCaseFromDeltas(
			s, osmomath.NewDec(10), tenSecAccum, 5*time.Second, osmomath.NewDec(2)),
		"accumulator = 10*OneSec, t=3s. 10*second base accum": testCaseFromDeltas(
			s, tenSecAccum, tenSecAccum, 3*time.Second, ThreePlusOneThird),
		"accumulator = 10*OneSec, t=100s. .1*second base accum": testCaseFromDeltas(
			s, pointOneAccum, tenSecAccum, 100*time.Second, osmomath.NewDecWithPrec(1, 1)),

		"accumulator = 10*OneSec, t=100s. 0 base accum (asset 1)": testCaseFromDeltasAsset1(s, osmomath.ZeroDec(), OneSec.MulInt64(10), 100*time.Second, osmomath.NewDecWithPrec(1, 1)),

		"start record time with nanoseconds does not change result": {
			startRecord: newOneSidedRecord(baseTime.Add(oneHundredNanoseconds), osmomath.ZeroDec(), true),
			endRecord:   newOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  denom0,
			expTwap:     osmomath.OneDec(),
		},
	}
	for name, test := range tests {
		s.Run(name, func() {
			osmoassert.ConditionalPanic(s.T(), test.expPanic, func() {
				arithmeticStrategy := &twap.ArithmeticTwapStrategy{TwapKeeper: *s.App.TwapKeeper}
				actualTwap := arithmeticStrategy.ComputeTwap(test.startRecord, test.endRecord, test.quoteAsset)
				s.Require().Equal(test.expTwap, actualTwap)
			})
		})
	}
}

// TestComputeGeometricStrategyTwap tests geometric strategy's computeTwap
// Contrary to computeTwap function (logic.go) that handles the cases with zero delta correctly,
// this function should panic in case of zero delta.
func (s *TestSuite) TestComputeGeometricStrategyTwap() {
	var (
		errTolerance = osmomath.ErrTolerance{
			MultiplicativeTolerance: osmomath.SmallestDec(),
			RoundingDir:             osmomath.RoundDown,
		}

		// Compute accumulator difference for the underflow test case by
		// taking log base 2 of the min spot price
		minSpotPriceLogBase2 = twap.TwapLog(gammtypes.MinSpotPrice)

		// Compute accumulator difference for the overflow test case by
		// taking log base 2 of the max spot price
		maxSpotPriceLogBase2 = twap.TwapLog(gammtypes.MaxSpotPrice)

		oneHundredYearsInHours        int64 = 100 * 365 * 24
		oneHundredYears                     = OneSec.MulInt64(60 * 60 * oneHundredYearsInHours)
		oneHundredYearsMin1MsDuration       = time.Duration(oneHundredYearsInHours)*time.Hour - time.Millisecond

		// Subtract 1ms from 100 years to assume that we interpolate.
		oneHundredYearsMin1Ms = oneHundredYears.Sub(oneDec)

		// Calculate the geometric accumulator difference for overflow test case.
		overflowTestCaseAccumDiff = oneHundredYearsMin1Ms.Mul(maxSpotPriceLogBase2)

		// Calculate the geometric accumulator difference for underflow test case.
		underflowTestCaseAccumDiff = oneHundredYearsMin1Ms.Mul(minSpotPriceLogBase2)
	)

	tests := map[string]computeTwapTestCase{
		// basic test for both denom with zero start accumulator
		"basic denom0: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newOneSidedGeometricRecord(baseTime, osmomath.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(tPlusOne, geometricTenSecAccum),
			quoteAsset:  denom0,
			expTwap:     osmomath.NewDec(10),
		},
		"basic denom1: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newOneSidedGeometricRecord(baseTime, osmomath.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(tPlusOne, geometricTenSecAccum),
			quoteAsset:  denom1,
			expTwap:     osmomath.OneDec().Quo(osmomath.NewDec(10)),
		},

		// basic test for both denom with non-zero start accumulator
		"denom0: start accumulator of 10 * 1s, end accumulator 10 * 1s + 20 * 2s = 20": {
			startRecord: newOneSidedGeometricRecord(baseTime, geometricTenSecAccum),
			endRecord:   newOneSidedGeometricRecord(baseTime.Add(time.Second*2), geometricTenSecAccum.Add(OneSec.MulInt64(2).Mul(twap.TwapLog(osmomath.NewDec(20))))),
			quoteAsset:  denom0,
			expTwap:     osmomath.NewDec(20),
		},
		"denom1 start accumulator of 10 * 1s, end accumulator 10 * 1s + 20 * 2s = 20": {
			startRecord: newOneSidedGeometricRecord(baseTime, geometricTenSecAccum),
			endRecord:   newOneSidedGeometricRecord(baseTime.Add(time.Second*2), geometricTenSecAccum.Add(OneSec.MulInt64(2).Mul(twap.TwapLog(osmomath.NewDec(20))))),
			quoteAsset:  denom1,
			expTwap:     osmomath.OneDec().Quo(osmomath.NewDec(20)),
		},

		// toggle time delta.
		"accumulator = log(10)*OneSec, t=5s. 0 base accum": geometricTestCaseFromDeltas0(
			s, osmomath.ZeroDec(), geometricTenSecAccum, 5*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(5*1000))),
		"accumulator = log(10)*OneSec, t=3s. 0 base accum": geometricTestCaseFromDeltas0(
			s, osmomath.ZeroDec(), geometricTenSecAccum, 3*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(3*1000))),
		"accumulator = log(10)*OneSec, t=100s. 0 base accum": geometricTestCaseFromDeltas0(
			s, osmomath.ZeroDec(), geometricTenSecAccum, 100*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(100*1000))),

		// test that base accum has no impact
		"accumulator = log(10)*OneSec, t=5s. 10 base accum": geometricTestCaseFromDeltas0(
			s, logTen, geometricTenSecAccum, 5*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(5*1000))),
		"accumulator = log(10)*OneSec, t=3s. 10*second base accum": geometricTestCaseFromDeltas0(
			s, OneSec.MulInt64(10).Mul(logTen), geometricTenSecAccum, 3*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(3*1000))),
		"accumulator = 10*OneSec, t=100s. .1*second base accum": geometricTestCaseFromDeltas0(
			s, OneSec.MulInt64(10).Mul(logOneOverTen), geometricTenSecAccum, 100*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(100*1000))),

		"price of 1_000_000 for an hour": {
			startRecord: newOneSidedGeometricRecord(baseTime, osmomath.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseTime.Add(time.Hour), OneSec.MulInt64(60*60).Mul(twap.TwapLog(osmomath.NewDec(1_000_000)))),
			quoteAsset:  denom0,
			expTwap:     osmomath.NewDec(1_000_000),
		},

		"no overflow test: at max spot price denom0 quote - get max spot price": {
			startRecord: withSp0(baseRecord, gammtypes.MaxSpotPrice),
			endRecord:   newOneSidedGeometricRecord(baseRecord.Time.Add(oneHundredYearsMin1MsDuration), overflowTestCaseAccumDiff),
			quoteAsset:  denom0,
			expTwap:     gammtypes.MaxSpotPrice,
		},

		"expected precision loss test: - at spot price denom1 quote - return zero": {
			startRecord: withSp0(baseRecord, gammtypes.MaxSpotPrice),
			endRecord:   newOneSidedGeometricRecord(baseRecord.Time.Add(oneHundredYearsMin1MsDuration), overflowTestCaseAccumDiff),
			quoteAsset:  denom1,

			expTwap: osmomath.ZeroDec(),
		},

		"no underflow test: spot price is smallest possible denom0 quote": {
			startRecord: newOneSidedGeometricRecord(baseTime, osmomath.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseRecord.Time.Add(oneHundredYearsMin1MsDuration), underflowTestCaseAccumDiff),
			quoteAsset:  denom0,
			expTwap:     gammtypes.MinSpotPrice,
		},

		"no underflow test: spot price is smallest possible denom1 quote": {
			startRecord: newOneSidedGeometricRecord(baseTime, osmomath.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseRecord.Time.Add(oneHundredYearsMin1MsDuration), underflowTestCaseAccumDiff),
			quoteAsset:  denom1,
			expTwap:     osmomath.OneDec().Quo(gammtypes.MinSpotPrice),
		},

		"zero accum difference - return zero": {
			startRecord: newOneSidedGeometricRecord(baseTime, osmomath.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseRecord.Time.Add(time.Millisecond), osmomath.ZeroDec()),
			quoteAsset:  denom1,

			expTwap: osmomath.ZeroDec(),
		},

		"start record time with nanoseconds does not change result": {
			startRecord: newOneSidedGeometricRecord(baseTime.Add(oneHundredNanoseconds), osmomath.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(tPlusOne, geometricTenSecAccum),
			quoteAsset:  denom0,
			expTwap:     osmomath.NewDec(10),
		},
	}

	for name, tc := range tests {
		tc := tc
		s.Run(name, func() {
			osmoassert.ConditionalPanic(s.T(), tc.expPanic, func() {
				geometricStrategy := &twap.GeometricTwapStrategy{TwapKeeper: *s.App.TwapKeeper}
				actualTwap := geometricStrategy.ComputeTwap(tc.startRecord, tc.endRecord, tc.quoteAsset)

				// Sig fig round the expected value.
				tc.expTwap = osmomath.SigFigRound(tc.expTwap, gammtypes.SpotPriceSigFigs)

				osmoassert.Equal(s.T(), errTolerance, osmomath.BigDecFromDec(tc.expTwap), osmomath.BigDecFromDec(actualTwap))
			})
		})
	}
}

func (s *TestSuite) TestComputeArithmeticStrategyTwap_ThreeAsset() {
	tenSecAccum := OneSec.MulInt64(10)
	pointOneAccum := OneSec.QuoInt64(10)
	tests := map[string]computeThreeAssetArithmeticTwapTestCase{
		"three asset basic: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newThreeAssetOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newThreeAssetOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  []string{denom0, denom0, denom1},
			expTwap:     []osmomath.Dec{osmomath.OneDec(), osmomath.OneDec(), osmomath.OneDec()},
		},
		"three asset. accumulator = 10*OneSec, t=5s. 0 base accum": testThreeAssetCaseFromDeltas(
			osmomath.ZeroDec(), tenSecAccum, 5*time.Second, osmomath.NewDec(2)),

		// test that base accum has no impact
		"three asset. accumulator = 10*OneSec, t=5s. 10 base accum": testThreeAssetCaseFromDeltas(
			osmomath.NewDec(10), tenSecAccum, 5*time.Second, osmomath.NewDec(2)),
		"three asset. accumulator = 10*OneSec, t=100s. .1*second base accum": testThreeAssetCaseFromDeltas(
			pointOneAccum, tenSecAccum, 100*time.Second, osmomath.NewDecWithPrec(1, 1)),
	}
	for name, test := range tests {
		s.Run(name, func() {
			for i, startRec := range test.startRecord {
				arithmeticStrategy := &twap.ArithmeticTwapStrategy{TwapKeeper: *s.App.TwapKeeper}
				actualTwap := arithmeticStrategy.ComputeTwap(startRec, test.endRecord[i], test.quoteAsset[i])
				s.Require().Equal(test.expTwap[i], actualTwap)
			}
		})
	}
}

func (s *TestSuite) TestComputeGeometricStrategyTwap_ThreeAsset() {
	var (
		five        = osmomath.NewDec(5)
		fiveFor3Sec = OneSec.MulInt64(3).Mul(twap.TwapLog(five))

		ten          = five.MulInt64(2)
		tenFor100Sec = OneSec.MulInt64(100).Mul(twap.TwapLog(ten))

		errTolerance = osmomath.MustNewDecFromStr("0.00000001")
	)

	tests := map[string]computeThreeAssetArithmeticTwapTestCase{
		"three asset basic: spot price = 10 for one second, 0 init accumulator": {
			startRecord: newThreeAssetOneSidedRecord(baseTime, osmomath.ZeroDec(), true),
			endRecord:   newThreeAssetOneSidedRecord(tPlusOne, geometricTenSecAccum, true),
			quoteAsset:  []string{denom0, denom0, denom1},
			expTwap:     []osmomath.Dec{osmomath.NewDec(10), osmomath.NewDec(10), osmomath.NewDec(10)},
		},
		"three asset. accumulator = 5*3Sec, t=3s, no start accum": testThreeAssetCaseFromDeltas(
			osmomath.ZeroDec(), fiveFor3Sec, 3*time.Second, five),

		// test that base accum has no impact
		"three asset. accumulator = 5*3Sec, t=3s. 10 base accum": testThreeAssetCaseFromDeltas(
			geometricTenSecAccum, fiveFor3Sec, 3*time.Second, five),
		"three asset. accumulator = 100*100s, t=100s. .1*second base accum": testThreeAssetCaseFromDeltas(
			twap.TwapLog(OneSec.Quo(ten)), tenFor100Sec, 100*time.Second, ten),
	}
	for name, test := range tests {
		s.Run(name, func() {
			for i, startRec := range test.startRecord {
				geometricStrategy := &twap.GeometricTwapStrategy{TwapKeeper: *s.App.TwapKeeper}
				actualTwap := geometricStrategy.ComputeTwap(startRec, test.endRecord[i], test.quoteAsset[i])
				osmoassert.DecApproxEq(s.T(), test.expTwap[i], actualTwap, errTolerance)
			}
		})
	}
}
