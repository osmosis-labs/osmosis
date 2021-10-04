package osmomath

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestAbsDifferenceWithSign(t *testing.T) {
	decA, err := sdk.NewDecFromStr("3.2")
	require.NoError(t, err)
	decB, err := sdk.NewDecFromStr("4.3432389")
	require.NoError(t, err)

	s, b := AbsDifferenceWithSign(decA, decB)
	require.True(t, b)

	expectedDec, err := sdk.NewDecFromStr("1.1432389")
	require.NoError(t, err)
	require.Equal(t, expectedDec, s)
}

func TestPowApprox(t *testing.T) {
	base, err := sdk.NewDecFromStr("0.8")
	require.NoError(t, err)
	exp, err := sdk.NewDecFromStr("0.32")
	require.NoError(t, err)

	s := PowApprox(base, exp, powPrecision)
	expectedDec, err := sdk.NewDecFromStr("0.93108385")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision),
		"expected value & actual value's difference should less than precision",
	)
}

func TestPow(t *testing.T) {
	base, err := sdk.NewDecFromStr("1.68")
	require.NoError(t, err)
	exp, err := sdk.NewDecFromStr("0.32")
	require.NoError(t, err)

	s := Pow(base, exp)
	expectedDec, err := sdk.NewDecFromStr("1.18058965")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision),
		"expected value & actual value's difference should less than precision",
	)
}
