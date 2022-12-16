package twap_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v13/app/apptesting/osmoassert"
	"github.com/osmosis-labs/osmosis/v13/osmomath"
	"github.com/osmosis-labs/osmosis/v13/x/twap"
	"github.com/osmosis-labs/osmosis/v13/x/twap/types"
)

// TestComputeArithmeticStrategyTwap tests computeArithmeticTwap on various inputs.
// Contrary to computeTwap that handles the cases with zero delta correctly,
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

func (s *TestSuite) TestComputeGeometricStrategyTwap() {
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

		// TODO: this is the highest price we currently support with the given precision bounds.
		// Need to choose better base and potentially improve math functions to mitigate.
		"price of 1_000_000 for an hour": {
			startRecord: newOneSidedGeometricRecord(baseTime, sdk.ZeroDec()),
			endRecord:   newOneSidedGeometricRecord(baseTime.Add(time.Hour), OneSec.MulInt64(60*60).Mul(twap.TwapLog(sdk.NewDec(1_000_000)))),
			quoteAsset:  denom0,
			expTwap:     sdk.NewDec(1_000_000),
		},
		// TODO: overflow tests
		// - max spot price
		// - large time delta
		// - both

		// TODO: hand calculated tests
	}

	for name, tc := range tests {
		s.Run(name, func() {
			osmoassert.ConditionalPanic(s.T(), tc.expPanic, func() {
				geometricStrategy := &twap.GeometricTwapStrategy{TwapKeeper: *s.App.TwapKeeper}
				actualTwap := geometricStrategy.ComputeTwap(tc.startRecord, tc.endRecord, tc.quoteAsset)
				osmoassert.DecApproxEq(s.T(), tc.expTwap, actualTwap, osmomath.GetPowPrecision())
			})
		})
	}
}

// TODO: split up this test case to cover both arithmetic and geometric twap
func (s *TestSuite) TestComputeArithmeticTwap_ThreeAsset() {
	testThreeAssetCaseFromDeltas := func(startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) computeThreeAssetArithmeticTwapTestCase {
		return computeThreeAssetArithmeticTwapTestCase{
			newThreeAssetOneSidedRecord(baseTime, startAccum, true),
			newThreeAssetOneSidedRecord(baseTime.Add(timeDelta), startAccum.Add(accumDiff), true),
			[]string{denom0, denom0, denom1},
			[]sdk.Dec{expectedTwap, expectedTwap, expectedTwap},
		}
	}

	tenSecAccum := OneSec.MulInt64(10)
	pointOneAccum := OneSec.QuoInt64(10)
	tests := map[string]computeThreeAssetArithmeticTwapTestCase{
		"three asset basic: spot price = 1 for one second, 0 init accumulator": {
			startRecord: newThreeAssetOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newThreeAssetOneSidedRecord(tPlusOne, OneSec, true),
			quoteAsset:  []string{denom0, denom0, denom1},
			expTwap:     []sdk.Dec{sdk.OneDec(), sdk.OneDec(), sdk.OneDec()},
		},
		"three asset same record: asset1, end spot price = 1": {
			startRecord: newThreeAssetOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			endRecord:   newThreeAssetOneSidedRecord(baseTime, sdk.ZeroDec(), true),
			quoteAsset:  []string{denom1, denom2, denom2},
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

type computeThreeAssetArithmeticTwapTestCase struct {
	startRecord []types.TwapRecord
	endRecord   []types.TwapRecord
	quoteAsset  []string
	expTwap     []sdk.Dec
}

func testCaseFromDeltas(s *TestSuite, startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) computeTwapTestCase {
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
	}
}

func testCaseFromDeltasAsset1(s *TestSuite, startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) computeTwapTestCase {
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
	}
}

func geometricTestCaseFromDeltas0(s *TestSuite, startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) computeTwapTestCase {
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
	}
}

func geometricTestCaseFromDeltas1(s *TestSuite, startAccum, accumDiff sdk.Dec, timeDelta time.Duration, expectedTwap sdk.Dec) computeTwapTestCase {
	return geometricTestCaseFromDeltas0(s, startAccum, accumDiff, timeDelta, sdk.OneDec().Quo(expectedTwap))
}
