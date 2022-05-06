package osmoutils

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestBinarySearch(t *testing.T) {
	// straight line function that returns input. Simplest to binary search on,
	// binary search directly reveals one bit of the answer in each iteration with this function.
	lineF := func(a sdk.Int) (sdk.Int, error) {
		return a, nil
	}
	noErrTolerance := ErrTolerance{AdditiveTolerance: sdk.ZeroInt()}
	tests := []struct {
		f             func(sdk.Int) (sdk.Int, error)
		lowerbound    sdk.Int
		upperbound    sdk.Int
		targetOutput  sdk.Int
		errTolerance  ErrTolerance
		maxIterations int

		expectedSolvedInput sdk.Int
		expectErr           bool
	}{
		{lineF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 51, sdk.NewInt(1 + (1 << 25)), false},
		{lineF, sdk.ZeroInt(), sdk.NewInt(1 << 50), sdk.NewInt(1 + (1 << 25)), noErrTolerance, 10, sdk.Int{}, true},
	}

	for _, tc := range tests {
		actualSolvedInput, err := BinarySearch(tc.f, tc.lowerbound, tc.upperbound, tc.targetOutput, tc.errTolerance, tc.maxIterations)
		if tc.expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.True(sdk.IntEq(t, tc.expectedSolvedInput, actualSolvedInput))
		}
	}
}
