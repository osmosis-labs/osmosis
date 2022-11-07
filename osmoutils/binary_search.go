package osmoutils

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
)

// ErrTolerance is used to define a compare function, which checks if two
// ints are within a certain error tolerance of one another,
// and (optionally) that they are rounding in the correct direction.
// ErrTolerance.Compare(a, b) returns true iff:
// * RoundingMode = RoundUp, then b >= a
// * RoundingMode = RoundDown, then b <= a
// * |a - b| <= AdditiveTolerance
// * |a - b| / min(a, b) <= MultiplicativeTolerance
//
// Each check is respectively ignored if the entry is nil.
// So AdditiveTolerance = sdk.Int{} or sdk.ZeroInt()
// MultiplicativeTolerance = sdk.Dec{}
// RoundingDir = RoundUnconstrained.
// Note that if AdditiveTolerance == 0, then this is equivalent to a standard compare.
type ErrTolerance struct {
	AdditiveTolerance       sdk.Int
	MultiplicativeTolerance sdk.Dec
	RoundingDir             osmomath.RoundingDirection
}

// Compare returns if actual is within errTolerance of expected.
// returns 0 if it is
// returns 1 if not, and expected > actual.
// returns -1 if not, and expected < actual
func (e ErrTolerance) Compare(expected sdk.Int, actual sdk.Int) int {
	diff := expected.Sub(actual).Abs()

	comparisonSign := 0
	if expected.GT(actual) {
		comparisonSign = 1
	} else {
		comparisonSign = -1
	}

	// Check additive tolerance equations
	if !e.AdditiveTolerance.IsNil() {
		// if no error accepted, do a direct compare.
		if e.AdditiveTolerance.IsZero() {
			if expected.Equal(actual) {
				return 0
			}
		}

		if diff.GT(e.AdditiveTolerance) {
			return comparisonSign
		}
	}
	// Check multiplicative tolerance equations
	if !e.MultiplicativeTolerance.IsNil() && !e.MultiplicativeTolerance.IsZero() {
		errTerm := diff.ToDec().Quo(sdk.MinInt(expected, actual).ToDec())
		if errTerm.GT(e.MultiplicativeTolerance) {
			return comparisonSign
		}
	}

	return 0
}

// CompareBigDec validates if actual is within errTolerance of expected.
// returns 0 if it is
// returns 1 if not, and expected > actual.
// returns -1 if not, and expected < actual
func (e ErrTolerance) CompareBigDec(expected osmomath.BigDec, actual osmomath.BigDec) int {
	diff := expected.Sub(actual).Abs()

	comparisonSign := 0
	if expected.GT(actual) {
		comparisonSign = 1
	} else {
		comparisonSign = -1
	}

	// Ensure that even if expected is within tolerance of actual, we don't count it as equal if its in the wrong direction.
	// so if were supposed to round down, it must be that `expected >= actual`.
	// likewise if were supposed to round up, it must be that `expected <= actual`.
	// If neither of the above, then rounding direction does not enforce a constraint.
	if e.RoundingDir == osmomath.RoundDown {
		if expected.LT(actual) {
			return -1
		}
	} else if e.RoundingDir == osmomath.RoundUp {
		if expected.GT(actual) {
			return 1
		}
	}

	// Check additive tolerance equations
	if !e.AdditiveTolerance.IsNil() {
		// if no error accepted, do a direct compare.
		if e.AdditiveTolerance.IsZero() {
			if expected.Equal(actual) {
				return 0
			}
		}

		if diff.GT(osmomath.BigDecFromSDKDec(e.AdditiveTolerance.ToDec())) {
			return comparisonSign
		}
	}
	// Check multiplicative tolerance equations
	if !e.MultiplicativeTolerance.IsNil() && !e.MultiplicativeTolerance.IsZero() {
		errTerm := diff.Quo(osmomath.MinDec(expected, actual))
		if errTerm.GT(osmomath.BigDecFromSDKDec(e.MultiplicativeTolerance)) {
			return comparisonSign
		}
	}

	return 0
}

// Binary search inputs between [lowerbound, upperbound] to a monotonic increasing function f.
// We stop once f(found_input) meets the ErrTolerance constraints.
// If we perform more than maxIterations (or equivalently lowerbound = upperbound), we return an error.
func BinarySearch(f func(input sdk.Int) (sdk.Int, error),
	lowerbound sdk.Int,
	upperbound sdk.Int,
	targetOutput sdk.Int,
	errTolerance ErrTolerance,
	maxIterations int,
) (sdk.Int, error) {
	// Setup base case of loop
	curEstimate := lowerbound.Add(upperbound).QuoRaw(2)
	curOutput, err := f(curEstimate)
	if err != nil {
		return sdk.Int{}, err
	}
	curIteration := 0
	for ; curIteration < maxIterations; curIteration += 1 {
		compRes := errTolerance.Compare(curOutput, targetOutput)
		if compRes > 0 {
			upperbound = curEstimate
		} else if compRes < 0 {
			lowerbound = curEstimate
		} else {
			return curEstimate, nil
		}
		curEstimate = lowerbound.Add(upperbound).QuoRaw(2)
		curOutput, err = f(curEstimate)
		if err != nil {
			return sdk.Int{}, err
		}
	}

	return sdk.Int{}, errors.New("hit maximum iterations, did not converge fast enough")
}

// SdkDec
type SdkDec[D any] interface {
	Add(SdkDec[D]) SdkDec[D]
	Quo(SdkDec[D]) SdkDec[D]
	QuoRaw(int64) SdkDec[D]
}

// BinarySearchBigDec takes as input:
// * an input range [lowerbound, upperbound]
// * an increasing function f
// * a target output x
// * max number of iterations (for gas control / handling does-not-converge cases)
//
// It binary searches on the input range, until it finds an input y s.t. f(y) meets the err tolerance constraints for how close it is to x.
// If we perform more than maxIterations (or equivalently lowerbound = upperbound), we return an error.
func BinarySearchBigDec(f func(input osmomath.BigDec) (osmomath.BigDec, error),
	lowerbound osmomath.BigDec,
	upperbound osmomath.BigDec,
	targetOutput osmomath.BigDec,
	errTolerance ErrTolerance,
	maxIterations int,
) (osmomath.BigDec, error) {
	// Setup base case of loop
	curEstimate := lowerbound.Add(upperbound).Quo(osmomath.NewBigDec(2))
	curOutput, err := f(curEstimate)
	if err != nil {
		return osmomath.BigDec{}, err
	}
	curIteration := 0
	for ; curIteration < maxIterations; curIteration += 1 {
		// fmt.Println(targetOutput, curOutput)
		compRes := errTolerance.CompareBigDec(targetOutput, curOutput)
		if compRes < 0 {
			upperbound = curEstimate
		} else if compRes > 0 {
			lowerbound = curEstimate
		} else {
			return curEstimate, nil
		}
		curEstimate = lowerbound.Add(upperbound).Quo(osmomath.NewBigDec(2))
		curOutput, err = f(curEstimate)
		if err != nil {
			return osmomath.BigDec{}, err
		}
	}

	return osmomath.BigDec{}, errors.New("hit maximum iterations, did not converge fast enough")
}
