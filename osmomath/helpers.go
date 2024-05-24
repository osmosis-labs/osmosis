package osmomath

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var diffTypesErrorMessage = "cannot compare variables of different types"

type Stringer interface {
	String() string
}

func failNowIfNot(t *testing.T, ok bool) {
	if !ok {
		require.FailNow(t, diffTypesErrorMessage)
	}
}

func Equal[T Stringer](t *testing.T, tolerance ErrTolerance, A, B T) {
	errMsg := fmt.Sprintf("expected %s, actual %s", A.String(), B.String())
	switch a := any(A).(type) {
	case Int:
		b, ok := any(B).(Int)
		failNowIfNot(t, ok)

		require.True(t, tolerance.Compare(a, b) == 0, errMsg)

	case BigDec:
		b, ok := any(B).(BigDec)
		failNowIfNot(t, ok)

		require.True(t, tolerance.CompareBigDec(a, b) == 0, errMsg)

	case Dec:
		b, ok := any(B).(Dec)
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
