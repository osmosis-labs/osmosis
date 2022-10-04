package osmoutils

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
)

func TestBinarySearch(t *testing.T) {
	// straight line function that returns input. Simplest to binary search on,
	// binary search directly reveals one bit of the answer in each iteration with this function.
	lineF := func(a sdk.Int) (sdk.Int, error) {
		return a, nil
	}
	cubicF := func(a sdk.Int) (sdk.Int, error) {
		calculation := sdk.Dec(a)
		result := calculation.Power(3)
		output := sdk.Int(result)
		return output, nil
	}
	noErrTolerance := ErrTolerance{AdditiveTolerance: sdk.ZeroInt()}
	testErrToleranceAdditive := ErrTolerance{AdditiveTolerance: sdk.NewInt(1 << 20)}
	testErrToleranceMultiplicative := ErrTolerance{AdditiveTolerance: sdk.ZeroInt(), MultiplicativeTolerance: sdk.NewDec(10)}
	testErrToleranceBoth := ErrTolerance{AdditiveTolerance: sdk.NewInt(1 << 20), MultiplicativeTolerance: sdk.NewDec(1 << 3)}
	tests := map[string]struct {
		f             func(sdk.Int) (sdk.Int, error)
		lowerbound    sdk.Int
		upperbound    sdk.Int
		targetOutput  sdk.Int
		errTolerance  ErrTolerance
		maxIterations int

		expectedSolvedInput sdk.Int
		expectErr           bool
		// This binary searches inputs to a monotonic increasing function F
		// We stop when the answer is within error bounds stated by errTolerance
		// First, (lowerbound + upperbound) / 2 becomes the current estimate.
		// A current output is also defined as f(current estimate). In this case f is lineF
		// We then compare the current output with the target output to see if it's within error tolerance bounds. If not, continue binary searching by iterating.
		// If it is, we return current output
		// Additive error bounds are solid addition / subtraction bounds to error, while multiplicative bounds take effect after dividing by the minimum between the two compared numbers.
	}{
		"linear f, no err tolerance, converges": {lineF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 51, sdk.NewInt(1 + (1 << 25)), false},
		"linear f, no err tolerance, does not converge": {lineF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 10, sdk.Int{}, true},
		"cubic f, no err tolerance, converges": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 51, sdk.NewInt(322539792367616), false},
		"cubic f, no err tolerance, does not converge": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 10, sdk.Int{}, true},
		"cubic f, large additive err tolerance, converges": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt((1 << 15)), testErrToleranceAdditive, 51, sdk.NewInt(1 << 46), false},
		"cubic f, large additive err tolerance, does not converge": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt((1 << 30)), testErrToleranceAdditive, 10, sdk.Int{}, true},
		"cubic f, large multiplicative err tolerance, converges": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), testErrToleranceMultiplicative, 51, sdk.NewInt(322539792367616), false},
		"cubic f, large multiplicative err tolerance, does not converge": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), testErrToleranceMultiplicative, 10, sdk.Int{}, true},
		"cubic f, both err tolerances, converges": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt((1 << 15)), testErrToleranceBoth, 51, sdk.NewInt(1 << 45), false},
		"cubic f, both err tolerances, does not converge": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt((1 << 30)), testErrToleranceBoth, 10, sdk.Int{}, true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualSolvedInput, err := BinarySearch(tc.f, tc.lowerbound, tc.upperbound, tc.targetOutput, tc.errTolerance, tc.maxIterations)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(sdk.IntEq(t, tc.expectedSolvedInput, actualSolvedInput))
			}
		})
	}
}

func TestBinarySearchBigDec(t *testing.T) {
	// straight line function that returns input. Simplest to binary search on,
	// binary search directly reveals one bit of the answer in each iteration with this function.
	lineF := func(a osmomath.BigDec) (osmomath.BigDec, error) {
		return a, nil
	}
	cubicF := func(a osmomath.BigDec) (osmomath.BigDec, error) {
		// these precision shifts are done implicitly in the int binary search tests
		// we keep them here to maintain parity between test cases across implementations
		calculation := a.Quo(osmomath.NewBigDec(10).Power(18))
		result := calculation.Power(3)
		output := result.Mul(osmomath.NewBigDec(10).Power(18))
		return output, nil
	}
	lowErrTolerance := ErrTolerance{AdditiveTolerance: sdk.OneInt()}
	testErrToleranceAdditive := ErrTolerance{AdditiveTolerance: sdk.NewInt(1 << 20)}
	testErrToleranceMultiplicative := ErrTolerance{AdditiveTolerance: sdk.OneInt(), MultiplicativeTolerance: sdk.NewDec(10)}
	testErrToleranceBoth := ErrTolerance{AdditiveTolerance: sdk.NewInt(1 << 20), MultiplicativeTolerance: sdk.NewDec(1 << 3)}
	tests := map[string]struct {
		f             func(osmomath.BigDec) (osmomath.BigDec, error)
		lowerbound    osmomath.BigDec
		upperbound    osmomath.BigDec
		targetOutput  osmomath.BigDec
		errTolerance  ErrTolerance
		maxIterations int

		expectedSolvedInput osmomath.BigDec
		expectErr           bool
		// This binary searches inputs to a monotonic increasing function F
		// We stop when the answer is within error bounds stated by errTolerance
		// First, (lowerbound + upperbound) / 2 becomes the current estimate.
		// A current output is also defined as f(current estimate). In this case f is lineF
		// We then compare the current output with the target output to see if it's within error tolerance bounds. If not, continue binary searching by iterating.
		// If it is, we return current output
		// Additive error bounds are solid addition / subtraction bounds to error, while multiplicative bounds take effect after dividing by the minimum between the two compared numbers.
	}{
		"linear f, no err tolerance, converges": {lineF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec(1 + (1 << 25)), lowErrTolerance, 51, osmomath.NewBigDec(1 + (1 << 25)), false},
		"linear f, no err tolerance, does not converge": {lineF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec(1 + (1 << 25)), lowErrTolerance, 10, osmomath.BigDec{}, true},
		"cubic f, no err tolerance, converges": {cubicF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec(1 + (1 << 25)), lowErrTolerance, 51, osmomath.NewBigDec(322539792367616), false},
		"cubic f, no err tolerance, does not converge": {cubicF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec(1 + (1 << 25)), lowErrTolerance, 10, osmomath.BigDec{}, true},
		"cubic f, large additive err tolerance, converges": {cubicF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec((1 << 15)), testErrToleranceAdditive, 51, osmomath.NewBigDec(1 << 46), false},
		"cubic f, large additive err tolerance, does not converge": {cubicF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec((1 << 30)), testErrToleranceAdditive, 10, osmomath.BigDec{}, true},
		"cubic f, large multiplicative err tolerance, converges": {cubicF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec(1 + (1 << 25)), testErrToleranceMultiplicative, 51, osmomath.NewBigDec(322539792367616), false},
		"cubic f, large multiplicative err tolerance, does not converge": {cubicF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec(1 + (1 << 25)), testErrToleranceMultiplicative, 10, osmomath.BigDec{}, true},
		"cubic f, both err tolerances, converges": {cubicF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec((1 << 15)), testErrToleranceBoth, 51, osmomath.NewBigDec(1 << 45), false},
		"cubic f, both err tolerances, does not converge": {cubicF, osmomath.ZeroDec(), osmomath.NewBigDec(1 << 50), osmomath.NewBigDec((1 << 30)), testErrToleranceBoth, 10, osmomath.BigDec{}, true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualSolvedInput, err := BinarySearchBigDec(tc.f, tc.lowerbound, tc.upperbound, tc.targetOutput, tc.errTolerance, tc.maxIterations)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(osmomath.DecApproxEq(t, tc.expectedSolvedInput, actualSolvedInput, osmomath.OneDec()))
			}
		})
	}
}

func TestErrTolerance_Compare(t *testing.T) {
	ZeroErrTolerance := ErrTolerance{AdditiveTolerance: sdk.ZeroInt(), MultiplicativeTolerance: sdk.Dec{}}
	NonZeroErrAdditive := ErrTolerance{AdditiveTolerance: sdk.NewInt(10), MultiplicativeTolerance: sdk.Dec{}}
	NonZeroErrMultiplicative := ErrTolerance{AdditiveTolerance: sdk.ZeroInt(), MultiplicativeTolerance: sdk.NewDec(10)}
	NonZeroErrBoth := ErrTolerance{AdditiveTolerance: sdk.NewInt(1), MultiplicativeTolerance: sdk.NewDec(10)}
	tests := []struct {
		name      string
		tol       ErrTolerance
		intInput     sdk.Int
		intReference sdk.Int

		bigDecInput     osmomath.BigDec
		bigDecReference osmomath.BigDec

		expectedCompareResult int
	}{
		{"0 tolerance: <", ZeroErrTolerance, sdk.NewInt(1000), sdk.NewInt(1001), osmomath.NewBigDec(1000), osmomath.NewBigDec(1001), -1},
		{"0 tolerance: =", ZeroErrTolerance, sdk.NewInt(1001), sdk.NewInt(1001), osmomath.NewBigDec(1001), osmomath.NewBigDec(1001), 0},
		{"0 tolerance: >", ZeroErrTolerance, sdk.NewInt(1002), sdk.NewInt(1001), osmomath.NewBigDec(1002), osmomath.NewBigDec(1001), 1},
		{"Nonzero additive tolerance: <", NonZeroErrAdditive, sdk.NewInt(420), sdk.NewInt(1001), osmomath.NewBigDec(420), osmomath.NewBigDec(1001), -1},
		{"Nonzero additive tolerance: =", NonZeroErrAdditive, sdk.NewInt(1011), sdk.NewInt(1001), osmomath.NewBigDec(1011), osmomath.NewBigDec(1001), 0},
		{"Nonzero additive tolerance: >", NonZeroErrAdditive, sdk.NewInt(1230), sdk.NewInt(1001), osmomath.NewBigDec(1230), osmomath.NewBigDec(1001), 1},
		{"Nonzero multiplicative tolerance: <", NonZeroErrMultiplicative, sdk.NewInt(1000), sdk.NewInt(1001), osmomath.NewBigDec(1000), osmomath.NewBigDec(1001), -1},
		{"Nonzero multiplicative tolerance: =", NonZeroErrMultiplicative, sdk.NewInt(1001), sdk.NewInt(1001), osmomath.NewBigDec(1001), osmomath.NewBigDec(1001), 0},
		{"Nonzero multiplicative tolerance: >", NonZeroErrMultiplicative, sdk.NewInt(1002), sdk.NewInt(1001), osmomath.NewBigDec(1002), osmomath.NewBigDec(1001), 1},
		{"Nonzero both tolerance: <", NonZeroErrBoth, sdk.NewInt(990), sdk.NewInt(1001), osmomath.NewBigDec(990), osmomath.NewBigDec(1001), -1},
		{"Nonzero both tolerance: =", NonZeroErrBoth, sdk.NewInt(1002), sdk.NewInt(1001), osmomath.NewBigDec(1002), osmomath.NewBigDec(1001), 0},
		{"Nonzero both tolerance: >", NonZeroErrBoth, sdk.NewInt(1011), sdk.NewInt(1001), osmomath.NewBigDec(1011), osmomath.NewBigDec(1001), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tol.Compare(tt.intInput, tt.intReference); got != tt.expectedCompareResult {
				t.Errorf("ErrTolerance.Compare() = %v, want %v", got, tt.expectedCompareResult)
			}
			if got := tt.tol.CompareBigDec(tt.bigDecInput, tt.bigDecReference); got != tt.expectedCompareResult {
				t.Errorf("ErrTolerance.CompareBigDec() = %v, want %v", got, tt.expectedCompareResult)
			}
		})
	}
}
