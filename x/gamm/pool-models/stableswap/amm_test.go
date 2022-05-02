package stableswap

import (
	fmt "fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Replace with https://github.com/cosmos/cosmos-sdk/blob/master/types/decimal.go#L892-L895
// once our SDK branch is up to date with it
func decApproxEq(t *testing.T, exp sdk.Dec, actual sdk.Dec, errTolerance sdk.Dec) {
	// We want |exp - actual| < errTolerance
	diff := exp.Sub(actual).Abs()
	require.True(t, diff.LTE(errTolerance), "expected %s, got %s, maximum errTolerance %s", exp, actual, errTolerance)
}

func TestCFMMInvariant(t *testing.T) {
	kErrTolerance := sdk.OneDec()

	tests := []struct {
		xReserve sdk.Dec
		yReserve sdk.Dec
		yIn      sdk.Dec
	}{
		{
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(1),
		},
		{
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(1000),
		},
		// {
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(10000),
		// },
	}

	for _, test := range tests {
		k0 := cfmmConstant(test.xReserve, test.yReserve)
		xOut := solveCfmm(test.xReserve, test.yReserve, test.yIn)
		fmt.Println(xOut)
		k1 := cfmmConstant(test.xReserve.Sub(xOut), test.yReserve.Add(test.yIn))
		decApproxEq(t, k0, k1, kErrTolerance)
	}
}
