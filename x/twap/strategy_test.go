package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/osmoutils/osmoassert"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/twap"
	"github.com/osmosis-labs/osmosis/v15/x/twap/types"
)

type computeTwapTestCase struct {
	startRecord    types.TwapRecord
	endRecord      types.TwapRecord
	twapStrategies []twap.TwapStrategy
	quoteAsset     string
	expTwap        sdk.Dec
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
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  denom0,
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
			},
			expTwap: sdk.OneDec(),
		},
		// this test just shows what happens in case the records are reversed.
		// It should return the correct result, even though this is incorrect internal API usage
		"arithmetic only: invalid call: reversed records of above": {
			startRecord: newOneSidedRecord(tPlusOne, OneSec, true),
			endRecord:   newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			quoteAsset:  denom0,
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
			},
			expTwap: sdk.OneDec(),
		},
		"same record: denom0, end spot price = 0": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			quoteAsset:  denom0,
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
				geomStrategy,
			},
			expTwap: sdk.ZeroDec(),
		},
		"same record: denom1, end spot price = 1": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			quoteAsset:  denom1,
			twapStrategies: []twap.TwapStrategy{
				arithStrategy,
				geomStrategy,
			},
			expTwap: sdk.OneDec(),
		},
		"arithmetic only: accumulator = 10*OneSec, t=5s. 0 base accum": testCaseFromDeltas(
			s, sdk.ZeroDec(), tenSecAccum, 5*time.Second, sdk.NewDec(2)),
		"arithmetic only: accumulator = 10*OneSec, t=100s. 0 base accum (asset 1)": testCaseFromDeltasAsset1(s, sdk.ZeroDec(), OneSec.MulInt64(10), 100*time.Second, sdk.NewDecWithPrec(1, 1)),
		"geometric only: accumulator = log(10)*OneSec, t=5s. 0 base accum": geometricTestCaseFromDeltas0(
			s, sdk.ZeroDec(), geometricTenSecAccum, 5*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(5*1000))),
		"geometric only: accumulator = log(10)*OneSec, t=100s. 0 base accum (asset 1)": geometricTestCaseFromDeltas1(s, sdk.ZeroDec(), geometricTenSecAccum, 100*time.Second, sdk.OneDec().Quo(twap.TwapPow(geometricTenSecAccum.QuoInt64(100*1000)))),
		"three asset same record: asset1, end spot price = 1": {
			startRecord: newThreeAssetOneSidedRecord(baseTime, sdk.ZeroDec(), true)[1],
			endRecord:   newThreeAssetOneSidedRecord(baseTime, sdk.ZeroDec(), true)[1],
			quoteAsset:  denom2,
			expTwap:     sdk.OneDec(),
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
		"same record (zero time delta), division by 0 - panic": {
			startRecord: newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			quoteAsset:  denom0,
			expPanic:    true,
		},
		"accumulator = 10*OneSec, t=5s. 0 base accum": testCaseFromDeltas(
			s, sdk.ZeroDec(), tenSecAccum, 5*time.Second, sdk.NewDec(2)),
		"accumulator = 10*OneSec, t=3s. 0 base accum": testCaseFromDeltas(
			s, sdk.ZeroDec(), tenSecAccum, 3*time.Second, ThreePlusOneThird),
		"accumulator = 10*OneSec, t=100s. 0 base accum": testCaseFromDeltas(
			s, sdk.ZeroDec(), tenSecAccum, 100*time.Second, sdk.NewDecWithPrec(1, 1)),

		// test that base accum has no impact
		"accumulator = 10*OneSec, t=5s. 10 base accum": testCaseFromDeltas(
			s, sdk.NewDec(10), tenSecAccum, 5*time.Second, sdk.NewDec(2)),
		"accumulator = 10*OneSec, t=3s. 10*second base accum": testCaseFromDeltas(
			s, tenSecAccum, tenSecAccum, 3*time.Second, ThreePlusOneThird),
		"accumulator = 10*OneSec, t=100s. .1*second base accum": testCaseFromDeltas(
			s, pointOneAccum, tenSecAccum, 100*time.Second, sdk.NewDecWithPrec(1, 1)),

		"accumulator = 10*OneSec, t=100s. 0 base accum (asset 1)": testCaseFromDeltasAsset1(s, sdk.ZeroDec(), OneSec.MulInt64(10), 100*time.Second, sdk.NewDecWithPrec(1, 1)),

		"start record time with nanoseconds does not change result": {
			startRecord: newOneSidedRecord(baseTime.Add(oneHundredNanoseconds), sdk.ZeroDec(), true),
			endRecord:   newOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  denom0,
			expTwap:     sdk.OneDec(),
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
			MultiplicativeTolerance: sdk.SmallestDec(),
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
			startRecord: newOneSidedGeometricRecord(baseTime, sdk.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(tPlusOne, geometricTenSecAccum),
			quoteAsset:  denom0,
			expTwap:     sdk.NewDec(10),
		},
		"basic denom1: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newOneSidedGeometricRecord(baseTime, sdk.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(tPlusOne, geometricTenSecAccum),
			quoteAsset:  denom1,
			expTwap:     sdk.OneDec().Quo(sdk.NewDec(10)),
		},

		// basic test for both denom with non-zero start accumulator
		"denom0: start accumulator of 10 * 1s, end accumulator 10 * 1s + 20 * 2s = 20": {
			startRecord: newOneSidedGeometricRecord(baseTime, geometricTenSecAccum),
			endRecord:   newOneSidedGeometricRecord(baseTime.Add(time.Second*2), geometricTenSecAccum.Add(OneSec.MulInt64(2).Mul(twap.TwapLog(sdk.NewDec(20))))),
			quoteAsset:  denom0,
			expTwap:     sdk.NewDec(20),
		},
		"denom1 start accumulator of 10 * 1s, end accumulator 10 * 1s + 20 * 2s = 20": {
			startRecord: newOneSidedGeometricRecord(baseTime, geometricTenSecAccum),
			endRecord:   newOneSidedGeometricRecord(baseTime.Add(time.Second*2), geometricTenSecAccum.Add(OneSec.MulInt64(2).Mul(twap.TwapLog(sdk.NewDec(20))))),
			quoteAsset:  denom1,
			expTwap:     sdk.OneDec().Quo(sdk.NewDec(20)),
		},

		// toggle time delta.
		"accumulator = log(10)*OneSec, t=5s. 0 base accum": geometricTestCaseFromDeltas0(
			s, sdk.ZeroDec(), geometricTenSecAccum, 5*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(5*1000))),
		"accumulator = log(10)*OneSec, t=3s. 0 base accum": geometricTestCaseFromDeltas0(
			s, sdk.ZeroDec(), geometricTenSecAccum, 3*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(3*1000))),
		"accumulator = log(10)*OneSec, t=100s. 0 base accum": geometricTestCaseFromDeltas0(
			s, sdk.ZeroDec(), geometricTenSecAccum, 100*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(100*1000))),

		// test that base accum has no impact
		"accumulator = log(10)*OneSec, t=5s. 10 base accum": geometricTestCaseFromDeltas0(
			s, logTen, geometricTenSecAccum, 5*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(5*1000))),
		"accumulator = log(10)*OneSec, t=3s. 10*second base accum": geometricTestCaseFromDeltas0(
			s, OneSec.MulInt64(10).Mul(logTen), geometricTenSecAccum, 3*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(3*1000))),
		"accumulator = 10*OneSec, t=100s. .1*second base accum": geometricTestCaseFromDeltas0(
			s, OneSec.MulInt64(10).Mul(logOneOverTen), geometricTenSecAccum, 100*time.Second, twap.TwapPow(geometricTenSecAccum.QuoInt64(100*1000))),

		"price of 1_000_000 for an hour": {
			startRecord: newOneSidedGeometricRecord(baseTime, sdk.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseTime.Add(time.Hour), OneSec.MulInt64(60*60).Mul(twap.TwapLog(sdk.NewDec(1_000_000)))),
			quoteAsset:  denom0,
			expTwap:     sdk.NewDec(1_000_000),
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

			expTwap: sdk.ZeroDec(),
		},

		"no underflow test: spot price is smallest possible denom0 quote": {
			startRecord: newOneSidedGeometricRecord(baseTime, sdk.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseRecord.Time.Add(oneHundredYearsMin1MsDuration), underflowTestCaseAccumDiff),
			quoteAsset:  denom0,
			expTwap:     gammtypes.MinSpotPrice,
		},

		"no underflow test: spot price is smallest possible denom1 quote": {
			startRecord: newOneSidedGeometricRecord(baseTime, sdk.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseRecord.Time.Add(oneHundredYearsMin1MsDuration), underflowTestCaseAccumDiff),
			quoteAsset:  denom1,
			expTwap:     sdk.OneDec().Quo(gammtypes.MinSpotPrice),
		},

		"zero accum difference - return zero": {
			startRecord: newOneSidedGeometricRecord(baseTime, sdk.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseRecord.Time.Add(time.Millisecond), sdk.ZeroDec()),
			quoteAsset:  denom1,

			expTwap: sdk.ZeroDec(),
		},

		"start record time with nanoseconds does not change result": {
			startRecord: newOneSidedGeometricRecord(baseTime.Add(oneHundredNanoseconds), sdk.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(tPlusOne, geometricTenSecAccum),
			quoteAsset:  denom0,
			expTwap:     sdk.NewDec(10),
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

				s.Require().Equal(0, errTolerance.CompareBigDec(osmomath.BigDecFromSDKDec(tc.expTwap), osmomath.BigDecFromSDKDec(actualTwap)), "expected %s, got %s", tc.expTwap, actualTwap)
			})
		})
	}
}

func (s *TestSuite) TestComputeArithmeticStrategyTwap_ThreeAsset() {
	tenSecAccum := OneSec.MulInt64(10)
	pointOneAccum := OneSec.QuoInt64(10)
	tests := map[string]computeThreeAssetArithmeticTwapTestCase{
		"three asset basic: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newThreeAssetOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newThreeAssetOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  []string{denom0, denom0, denom1},
			expTwap:     []sdk.Dec{sdk.OneDec(), sdk.OneDec(), sdk.OneDec()},
		},
		"three asset. accumulator = 10*OneSec, t=5s. 0 base accum": testThreeAssetCaseFromDeltas(
			sdk.ZeroDec(), tenSecAccum, 5*time.Second, sdk.NewDec(2)),

		// test that base accum has no impact
		"three asset. accumulator = 10*OneSec, t=5s. 10 base accum": testThreeAssetCaseFromDeltas(
			sdk.NewDec(10), tenSecAccum, 5*time.Second, sdk.NewDec(2)),
		"three asset. accumulator = 10*OneSec, t=100s. .1*second base accum": testThreeAssetCaseFromDeltas(
			pointOneAccum, tenSecAccum, 100*time.Second, sdk.NewDecWithPrec(1, 1)),
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
		five        = sdk.NewDec(5)
		fiveFor3Sec = OneSec.MulInt64(3).Mul(twap.TwapLog(five))

		ten          = five.MulInt64(2)
		tenFor100Sec = OneSec.MulInt64(100).Mul(twap.TwapLog(ten))

		errTolerance = sdk.MustNewDecFromStr("0.00000001")
	)

	tests := map[string]computeThreeAssetArithmeticTwapTestCase{
		"three asset basic: spot price = 10 for one second, 0 init accumulator": {
			startRecord: newThreeAssetOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newThreeAssetOneSidedRecord(tPlusOne, geometricTenSecAccum, true),
			quoteAsset:  []string{denom0, denom0, denom1},
			expTwap:     []sdk.Dec{sdk.NewDec(10), sdk.NewDec(10), sdk.NewDec(10)},
		},
		"three asset. accumulator = 5*3Sec, t=3s, no start accum": testThreeAssetCaseFromDeltas(
			sdk.ZeroDec(), fiveFor3Sec, 3*time.Second, five),

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
