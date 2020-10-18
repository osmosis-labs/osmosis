package pool

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func TestPowApprox(t *testing.T) {
	base, err := sdk.NewDecFromStr("1.5")
	require.NoError(t, err)
	exp, err := sdk.NewDecFromStr("0.4")
	require.NoError(t, err)
	precision, err := sdk.NewDecFromStr("0.00000001")
	require.NoError(t, err)

	s := powApprox(base, exp, precision)
	expectedDec, err := sdk.NewDecFromStr("1.17607902")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(precision),
		"expected value & actual value's difference should less than precision",
	)
}

func TestPow(t *testing.T) {
	base, err := sdk.NewDecFromStr("1.4")
	require.NoError(t, err)
	exp, err := sdk.NewDecFromStr("3.2")
	require.NoError(t, err)

	s := pow(base, exp)
	expectedDec, err := sdk.NewDecFromStr("2.93501087")
	require.NoError(t, err)

	require.True(
		t,
		expectedDec.Sub(s).Abs().LTE(powPrecision),
		"expected value & actual value's difference should less than precision",
	)
}
