package osmomath

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	withinOne     = ErrTolerance{AdditiveTolerance: sdk.OneDec()}
	withinFactor8 = ErrTolerance{MultiplicativeTolerance: sdk.NewDec(8)}
	zero          = ZeroDec()
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
	noErrTolerance := ErrTolerance{AdditiveTolerance: sdk.ZeroDec()}
	testErrToleranceAdditive := ErrTolerance{AdditiveTolerance: sdk.NewDec(1 << 20)}
	testErrToleranceMultiplicative := ErrTolerance{AdditiveTolerance: sdk.ZeroDec(), MultiplicativeTolerance: sdk.NewDec(10)}
	testErrToleranceBoth := ErrTolerance{AdditiveTolerance: sdk.NewDec(1 << 20), MultiplicativeTolerance: sdk.NewDec(1 << 3)}
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
		"linear f, no err tolerance, converges":                          {lineF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 51, sdk.NewInt(1 + (1 << 25)), false},
		"linear f, no err tolerance, does not converge":                  {lineF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 10, sdk.Int{}, true},
		"cubic f, no err tolerance, converges":                           {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 51, sdk.NewInt(322539792367616), false},
		"cubic f, no err tolerance, does not converge":                   {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 10, sdk.Int{}, true},
		"cubic f, large additive err tolerance, converges":               {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt((1 << 15)), testErrToleranceAdditive, 51, sdk.NewInt(1 << 46), false},
		"cubic f, large additive err tolerance, does not converge":       {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt((1 << 30)), testErrToleranceAdditive, 10, sdk.Int{}, true},
		"cubic f, large multiplicative err tolerance, converges":         {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), testErrToleranceMultiplicative, 51, sdk.NewInt(322539792367616), false},
		"cubic f, large multiplicative err tolerance, does not converge": {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), testErrToleranceMultiplicative, 10, sdk.Int{}, true},
		"cubic f, both err tolerances, converges":                        {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt((1 << 15)), testErrToleranceBoth, 51, sdk.NewInt(1 << 45), false},
		"cubic f, both err tolerances, does not converge":                {cubicF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt((1 << 30)), testErrToleranceBoth, 10, sdk.Int{}, true},
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

// straight line function that returns input. Simplest to binary search on,
// binary search directly reveals one bit of the answer in each iteration with this function.
func lineF(a BigDec) BigDec {
	return a
}
func cubicF(a BigDec) BigDec {
	return a.PowerInteger(3)
}

var negCubicFConstant BigDec

func init() {
	negCubicFConstant = NewBigDec(1 << 62).PowerInteger(3).Neg()
}

func negCubicF(a BigDec) BigDec {
	return a.PowerInteger(3).Add(negCubicFConstant)
}

type searchFn func(BigDec) BigDec

type binarySearchTestCase struct {
	f             searchFn
	lowerbound    BigDec
	upperbound    BigDec
	targetOutput  BigDec
	errTolerance  ErrTolerance
	maxIterations int

	expectedSolvedInput BigDec
	expectErr           bool
	// This binary searches inputs to a monotonic increasing function F
	// We stop when the answer is within error bounds stated by errTolerance
	// First, (lowerbound + upperbound) / 2 becomes the current estimate.
	// A current output is also defined as f(current estimate). In this case f is lineF
	// We then compare the current output with the target output to see if it's within error tolerance bounds. If not, continue binary searching by iterating.
	// If it is, we return current output
	// Additive error bounds are solid addition / subtraction bounds to error, while multiplicative bounds take effect after dividing by the minimum between the two compared numbers.
}

// This test ensures that we use exactly the expected number of iterations (one bit of x at a time)
// to find the answer to binary search on a line.
func TestBinarySearchLineIterationCounts(t *testing.T) {
	tests := map[string]binarySearchTestCase{}

	generateExactTestCases := func(lowerbound, upperbound BigDec,
		errTolerance ErrTolerance, maxNumIters int) {
		tcSetName := fmt.Sprintf("simple linear case: lower %s, upper %s", lowerbound.String(), upperbound.String())
		// first pass get it working with no err tolerance or rounding direction
		target := lowerbound.Add(upperbound).QuoRaw(2)
		for expectedItersToTarget := 1; expectedItersToTarget < maxNumIters; expectedItersToTarget++ {
			// make two test cases, one at expected iter count, one at one below expected
			// to guarantee were getting error as expected.
			for subFromIter := 0; subFromIter < 2; subFromIter++ {
				testCase := binarySearchTestCase{
					f:          lineF,
					lowerbound: lowerbound, upperbound: upperbound,
					targetOutput: target, expectedSolvedInput: target,
					errTolerance:  errTolerance,
					maxIterations: expectedItersToTarget - subFromIter,
					expectErr:     subFromIter != 0,
				}
				tcName := fmt.Sprintf("%s, target %s, iters %d, expError %v",
					tcSetName, target.String(), expectedItersToTarget, testCase.expectErr)
				tests[tcName] = testCase
			}
			target = lowerbound.Add(target).QuoRaw(2)
		}
	}

	generateExactTestCases(ZeroDec(), NewBigDec(1<<20), withinOne, 20)
	// we can go further than 50, if we could specify non-integer additive err tolerance. TODO: Add this.
	generateExactTestCases(NewBigDec(1<<20), NewBigDec(1<<50), withinOne, 50)
	runBinarySearchTestCases(t, tests, exactlyEqual)
}

var fnMap = map[string]searchFn{"line": lineF, "cubic": cubicF, "neg_cubic": negCubicF}

// This function tests that any value in a given range can be reached within expected num iterations.
func TestIterationDepthRandValue(t *testing.T) {
	tests := map[string]binarySearchTestCase{}
	exactEqual := ErrTolerance{AdditiveTolerance: sdk.ZeroDec()}
	withinOne := ErrTolerance{AdditiveTolerance: sdk.OneDec()}
	within32 := ErrTolerance{AdditiveTolerance: sdk.OneDec().Mul(sdk.NewDec(32))}

	createRandInput := func(fnName string, lowerbound, upperbound int64,
		errTolerance ErrTolerance, maxNumIters int, errToleranceName string) {
		targetF := fnMap[fnName]
		targetX := int64(rand.Intn(int(upperbound-lowerbound-1))) + lowerbound + 1
		target := targetF(NewBigDec(targetX))
		testCase := binarySearchTestCase{
			f:          lineF,
			lowerbound: NewBigDec(lowerbound), upperbound: NewBigDec(upperbound),
			targetOutput: target, expectedSolvedInput: target,
			errTolerance:  errTolerance,
			maxIterations: maxNumIters,
			expectErr:     false,
		}
		tcname := fmt.Sprintf("%s: lower %d, upper %d, in %d iter of %s, rand target %d",
			fnName, lowerbound, upperbound, maxNumIters, errToleranceName, target)
		tests[tcname] = testCase
	}

	for i := 0; i < 1000; i++ {
		// Takes a 21st iteration to guaranteeably get 0
		createRandInput("line", 0, 1<<20, exactEqual, 21, "exactly equal")
		// Takes 20 iterations to guaranteeably get 1 within 0.
		createRandInput("line", 0, 1<<20, withinOne, 20, "within one")
		// Takes 15 iterations to guaranteeably get to 32. Needed to reach any number in [0, 31]
		createRandInput("line", 0, 1<<20, within32, 15, "within 32")
	}
	runBinarySearchTestCases(t, tests, errToleranceEqual)
}

type equalityMode int

const (
	exactlyEqual      equalityMode = iota
	errToleranceEqual equalityMode = iota
	equalWithinOne    equalityMode = iota
)

func withRoundingDir(e ErrTolerance, r RoundingDirection) ErrTolerance {
	return ErrTolerance{
		AdditiveTolerance:       e.AdditiveTolerance,
		MultiplicativeTolerance: e.MultiplicativeTolerance,
		RoundingDir:             r,
	}
}

func runBinarySearchTestCases(t *testing.T, tests map[string]binarySearchTestCase,
	equality equalityMode) {
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actualSolvedInput, err := BinarySearchBigDec(
				tc.f, tc.lowerbound, tc.upperbound, tc.targetOutput, tc.errTolerance, tc.maxIterations)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if equality == exactlyEqual {
					require.True(DecEq(t, tc.expectedSolvedInput, actualSolvedInput))
				} else if equality == errToleranceEqual {
					require.True(t, tc.errTolerance.CompareBigDec(tc.expectedSolvedInput, actualSolvedInput) == 0)
				} else {
					_, valid, msg, dec1, dec2 := DecApproxEq(t, tc.expectedSolvedInput, actualSolvedInput, OneDec())
					require.True(t, valid, msg+" \n d1 = %s, d2 = %s", dec1, dec2,
						tc.expectedSolvedInput, actualSolvedInput)
				}
			}
		})
	}
}

func TestBinarySearchBigDec(t *testing.T) {
	testErrToleranceAdditive := ErrTolerance{AdditiveTolerance: sdk.NewDec(1 << 30)}
	errToleranceBoth := ErrTolerance{AdditiveTolerance: sdk.NewDec(1 << 30), MultiplicativeTolerance: sdk.NewDec(1 << 3)}

	twoTo50 := NewBigDec(1 << 50)
	twoTo25PlusOne := NewBigDec(1 + (1 << 25))
	twoTo25PlusOneCubed := twoTo25PlusOne.PowerInteger(3)

	tests := map[string]binarySearchTestCase{
		"cubic f, no err tolerance, converges":     {cubicF, zero, twoTo50, twoTo25PlusOneCubed, withinOne, 51, twoTo25PlusOne, false},
		"cubic f, no err tolerance, not converges": {cubicF, zero, twoTo50, twoTo25PlusOneCubed, withinOne, 10, twoTo25PlusOne, true},
		// Target = 2^33 - 2^29, so correct input is 2^11 - e, for 0 < e < 2^10.
		// additive error tolerance is 2^30. So we converge at first value
		// whose cube is within 2^30 of answer. As were in binary search with power of two bounds
		// we go through powers of two first.
		// hence we hit at 2^11 first, since (2^11)^3 - target is 2^29, which is within additive err tolerance.
		"cubic f, within 2^30, target 2^33 - 2^29": {
			cubicF,
			zero, twoTo50,
			NewBigDec((1 << 33) - (1 << 29)),
			testErrToleranceAdditive, 51, NewBigDec(1 << 11), false},
		// basically same as above, but due to needing to roundup, we converge at a value > 2^11.
		// We try (1<<11 + 1<<10)^3 which is way too large.
		// notice by trial, that (1 << 11 + 1<<7)^3 - target > 2^30, but that
		// (1 << 11 + 1<<6)^3 - target < 2^30, so that is the answer.
		"cubic f, within 2^30, roundup, target 2^33 + 2^29": {
			cubicF,
			zero, twoTo50,
			NewBigDec((1 << 33) + (1 << 29)),
			withRoundingDir(testErrToleranceAdditive, RoundUp),
			51, NewBigDec(1<<11 + 1<<6), false},
		"cubic f, large multiplicative err tolerance, converges": {
			cubicF,
			zero, twoTo50,
			NewBigDec(1 << 30), withinFactor8, 51, NewBigDec(1 << 11), false},
		"cubic f, both err tolerances, converges": {
			cubicF,
			zero, twoTo50,
			NewBigDec((1 << 33) - (1 << 29)),
			errToleranceBoth, 51, NewBigDec(1 << 11), false},
		"neg cubic f, no err tolerance, converges": {negCubicF, zero, twoTo50,
			twoTo25PlusOneCubed.Add(negCubicFConstant), withinOne, 51, twoTo25PlusOne, false},
		// "neg cubic f, large multiplicative err tolerance, converges": {
		// 	negCubicF,
		// 	zero, twoTo50,
		// 	NewBigDec(1 << 30).Add(negCubicFConstant),
		// 	withinFactor8, 51, NewBigDec(1 << 11), false},
	}

	runBinarySearchTestCases(t, tests, equalWithinOne)
}

func TestBinarySearchRoundingBehavior(t *testing.T) {
	withinTwoTo30 := ErrTolerance{AdditiveTolerance: sdk.NewDec(1 << 30)}

	twoTo50 := NewBigDec(1 << 50)
	// twoTo25PlusOne := NewBigDec(1 + (1 << 25))
	// twoTo25PlusOneCubed := twoTo25PlusOne.Power(3)

	tests := map[string]binarySearchTestCase{
		"lineF, roundup within 2^30, target 2^32 + 2^30 + 1, expected=2^32 + 2^31": {f: lineF,
			lowerbound: zero, upperbound: twoTo50,
			targetOutput:        NewBigDec((1 << 32) + (1 << 30) + 1),
			errTolerance:        withRoundingDir(withinTwoTo30, RoundUp),
			maxIterations:       51,
			expectedSolvedInput: NewBigDec(1<<32 + 1<<31)},
		"lineF, roundup within 2^30, target 2^32 + 2^30 - 1, expected=2^32 + 2^30": {f: lineF,
			lowerbound: zero, upperbound: twoTo50,
			targetOutput:        NewBigDec((1 << 32) + (1 << 30) - 1),
			errTolerance:        withRoundingDir(withinTwoTo30, RoundUp),
			maxIterations:       51,
			expectedSolvedInput: NewBigDec(1<<32 + 1<<30)},
		"lineF, rounddown within 2^30, target 2^32 + 2^30 + 1, expected=2^32 + 2^31": {f: lineF,
			lowerbound: zero, upperbound: twoTo50,
			targetOutput:        NewBigDec((1 << 32) + (1 << 30) + 1),
			errTolerance:        withRoundingDir(withinTwoTo30, RoundDown),
			maxIterations:       51,
			expectedSolvedInput: NewBigDec(1<<32 + 1<<30)},
		"lineF, rounddown within 2^30, target 2^32 + 2^30 - 1, expected=2^32 + 2^30": {f: lineF,
			lowerbound: zero, upperbound: twoTo50,
			targetOutput:        NewBigDec((1 << 32) + (1 << 30) - 1),
			errTolerance:        withRoundingDir(withinTwoTo30, RoundDown),
			maxIterations:       51,
			expectedSolvedInput: NewBigDec(1 << 32)},
	}

	runBinarySearchTestCases(t,
		tests,
		equalWithinOne)
}

func TestErrTolerance_Compare(t *testing.T) {
	ZeroErrTolerance := ErrTolerance{AdditiveTolerance: sdk.ZeroDec(), MultiplicativeTolerance: sdk.Dec{}}
	NonZeroErrAdditive := ErrTolerance{AdditiveTolerance: sdk.NewDec(10), MultiplicativeTolerance: sdk.Dec{}}
	NonZeroErrMultiplicative := ErrTolerance{AdditiveTolerance: sdk.Dec{}, MultiplicativeTolerance: sdk.NewDec(10)}
	NonZeroErrBoth := ErrTolerance{AdditiveTolerance: sdk.NewDec(1), MultiplicativeTolerance: sdk.NewDec(10)}
	tests := []struct {
		name         string
		tol          ErrTolerance
		intInput     int64
		intReference int64

		expectedCompareResult int
	}{
		{"0 tolerance: <", ZeroErrTolerance, 1000, 1001, -1},
		{"0 tolerance: =", ZeroErrTolerance, 1001, 1001, 0},
		{"0 tolerance: >", ZeroErrTolerance, 1002, 1001, 1},
		{"Nonzero additive tolerance: <", NonZeroErrAdditive, 420, 1001, -1},
		{"Nonzero additive tolerance: =", NonZeroErrAdditive, 1011, 1001, 0},
		{"Nonzero additive tolerance: >", NonZeroErrAdditive, 1230, 1001, 1},
		{"Nonzero multiplicative tolerance within bounds: <", NonZeroErrMultiplicative, 1000, 1001, 0},
		{"Nonzero multiplicative tolerance within bounds: =", NonZeroErrMultiplicative, 1001, 1001, 0},
		{"Nonzero multiplicative tolerance within bounds: >", NonZeroErrMultiplicative, 1002, 1001, 0},
		{"Nonzero multiplicative tolerance out of bounds: <", NonZeroErrMultiplicative, 100000, 1001, 1},
		{"Nonzero multiplicative tolerance out of bounds: =", NonZeroErrMultiplicative, 100100, 1001, 1},
		{"Nonzero multiplicative tolerance out of bounds: >", NonZeroErrMultiplicative, 100200, 1001, 1},
		{"multiplicative tolerance with negative: >", NonZeroErrMultiplicative,
			-20020, -1001, -1},
		{"Nonzero both tolerance: <", NonZeroErrBoth, 990, 1001, -1},
		{"Nonzero both tolerance: =", NonZeroErrBoth, 1002, 1001, 0},
		{"Nonzero both tolerance: >", NonZeroErrBoth, 1011, 1001, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInt := tt.tol.Compare(sdk.NewInt(tt.intInput), sdk.NewInt(tt.intReference))
			if gotInt != tt.expectedCompareResult {
				t.Errorf("ErrTolerance.Compare() = %v, want %v", gotInt, tt.expectedCompareResult)
			}
			gotIntRev := tt.tol.Compare(sdk.NewInt(tt.intReference), sdk.NewInt(tt.intInput))
			if gotIntRev != -tt.expectedCompareResult {
				t.Errorf("ErrTolerance.Compare() = %v, want %v", gotIntRev, -tt.expectedCompareResult)
			}
			gotBigDec := tt.tol.CompareBigDec(NewBigDec(tt.intInput), NewBigDec(tt.intReference))
			if gotBigDec != tt.expectedCompareResult {
				t.Errorf("ErrTolerance.CompareBigDec() = %v, want %v", gotBigDec, tt.expectedCompareResult)
			}
			gotBigDecRev := tt.tol.CompareBigDec(NewBigDec(tt.intReference), NewBigDec(tt.intInput))
			if gotBigDecRev != -tt.expectedCompareResult {
				t.Errorf("ErrTolerance.CompareBigDec() = %v, want %v", gotBigDecRev, -tt.expectedCompareResult)
			}
		})
	}
}
