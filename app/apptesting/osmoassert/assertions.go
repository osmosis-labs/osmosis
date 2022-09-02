package osmoassert

import (
	"fmt"
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

// ConditionalError checks if expectError is true, asserts that err is an error
// If expectError is false, asserts that err is nil
func ConditionalError(t *testing.T, expectError bool, err error) {
	if expectError {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
}

// DecApproxEq is a helper function to compare two decimals.
// It validates the two decimal are within a certain tolerance.
// If not, it fails with a message.
func DecApproxEq(t *testing.T, d1 sdk.Dec, d2 sdk.Dec, tol sdk.Dec, msgAndArgs ...interface{}) {
	diff := d1.Sub(d2).Abs()
	msg := messageFromMsgAndArgs(msgAndArgs...)
	require.True(t, diff.LTE(tol), "expected |d1 - d2| <:\t%s\ngot |d1 - d2| = \t\t%s\nd1: %s, d2: %s\n%s", tol, diff, d1, d2, msg)
}

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		msg := msgAndArgs[0]
		if msgAsStr, ok := msg.(string); ok {
			return msgAsStr
		}
		return fmt.Sprintf("%+v", msg)
	}
	if len(msgAndArgs) > 1 {
		msgFormat, ok := msgAndArgs[0].(string)
		if !ok {
			return "error formatting additional arguments for DecApproxEq, please disregard."
		}
		return fmt.Sprintf(msgFormat, msgAndArgs[1:]...)
	}
	return ""
}
