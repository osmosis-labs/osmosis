package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/x/gamm/types"
)

func TestPower(t *testing.T) {
	cases := []struct{
		base string
		exp string
		expected string
	} {
		{"0.5", "3", "0.125"},
		{"0.25", "2", "0.0625"},
		{"0.3", "0.3", "0.696845320001282405"},
		{"1.25", "0.1", "1.0225651812560874"},
		{"1", "-1", "1"},
	// 	{"0.5", "-1", "2"},
	}

	for _, c := range cases {
		base, err := sdk.NewDecFromStr(c.base)
		require.NoError(t, err)
		exp, err := sdk.NewDecFromStr(c.exp)
		require.NoError(t, err)
		require.Equal(t, c.expected, types.Pow(base, exp).String()[:len(c.expected)])
	}
}

func TestPowerApprox(t *testing.T) {
	cases := []struct{
		base string
		exp string
		expected string
	} {
		{"0.5", "-1", "1.999999"},
		{"1.25", "-1.25", "0.75659"},
		{"1.75", "-0.55", "0.73507"},
		{"0.25", "-0.25", "1.414211"},
	}

	precision, _ := sdk.NewDecFromStr("0.000001")
	for _, c := range cases {
		base, err := sdk.NewDecFromStr(c.base)
		require.NoError(t, err)
		exp, err := sdk.NewDecFromStr(c.exp)
		require.NoError(t, err)
		require.Equal(t, c.expected, types.PowApprox(base, exp, precision).String()[:len(c.expected)])
	}
}
