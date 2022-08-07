package osmoassert

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// ConditionalPanic checks if expectPanic is true, asserts that sut (system under test)
// panics. If expectPanic is false, asserts that sut does not panic.
// returns true if sut panics and false it it does not
func ConditionalPanic(t *testing.T, expectPanic bool, sut func()) {
	if expectPanic {
		require.Panics(t, sut)
		return
	}
	require.NotPanics(t, sut)
}

// DecApproxEq is a helper function to compare two decimals.
// It validates the two decimal are within a certain tolerance.
// If not, it fails with a message.
func DecApproxEq(t *testing.T, d1 sdk.Dec, d2 sdk.Dec, tol sdk.Dec) {
	diff := d1.Sub(d2).Abs()
	require.True(t, diff.LTE(tol), "expected |d1 - d2| <:\t%s\ngot |d1 - d2| = \t\t%s\nd1: %s, d2: %s", tol, diff, d1, d2)
}
