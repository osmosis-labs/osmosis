package osmoassert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
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
func DecApproxEq(t *testing.T, d1 osmomath.Dec, d2 osmomath.Dec, tol osmomath.Dec, msgAndArgs ...interface{}) {
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

type Stringer interface {
	String() string
}

// Equal compares A with B and asserts that they are equal within tolerance error tolerance
func Equal[T Stringer](t *testing.T, tolerance osmomath.ErrTolerance, A, B T) {
	errMsg := fmt.Sprintf("expected %s, actual %s", A.String(), B.String())
	switch a := any(A).(type) {
	case osmomath.Int:
		b, ok := any(B).(osmomath.Int)
		failNowIfNot(t, ok)

		require.True(t, tolerance.Compare(a, b) == 0, errMsg)

	case osmomath.BigDec:
		b, ok := any(B).(osmomath.BigDec)
		failNowIfNot(t, ok)

		require.True(t, tolerance.CompareBigDec(a, b) == 0, errMsg)

	case osmomath.Dec:
		b, ok := any(B).(osmomath.Dec)
		failNowIfNot(t, ok)

		require.True(t, tolerance.CompareDec(a, b) == 0, errMsg)
	case sdk.Coin:
		b, ok := any(B).(sdk.Coin)
		failNowIfNot(t, ok)
		Equal(t, tolerance, a.Amount, b.Amount)

	case sdk.Coins:
		b, ok := any(B).(sdk.Coins)
		failNowIfNot(t, ok)

		if len(a) != len(b) {
			require.FailNow(t, errMsg)
		}

		for i, coinA := range a {
			Equal(t, tolerance, coinA, b[i])
		}

	default:
		require.FailNow(t, "unsupported types")
	}
}

func Uint64ArrayValuesAreUnique(values []uint64) bool {
	valueMap := make(map[uint64]bool)
	for _, val := range values {
		if _, exists := valueMap[val]; exists {
			return false
		}
		valueMap[val] = true
	}
	return true
}
