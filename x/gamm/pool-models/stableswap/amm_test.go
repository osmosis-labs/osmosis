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
		uReserve sdk.Dec
		wSumSquares sdk.Dec
		yIn      sdk.Dec
	}{
		{
			sdk.NewDec(100),
			sdk.NewDec(100),
			// represents a 4-asset pool with 100 in each reserve
			sdk.NewDec(200),
			sdk.NewDec(20000),
			sdk.NewDec(1),
		},
		{
			sdk.NewDec(100),
			sdk.NewDec(100),
			sdk.NewDec(200),
			sdk.NewDec(20000),
			sdk.NewDec(1000),
		},
		// {
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(100000),
		// 	sdk.NewDec(10000),
		// },
	}

	for _, test := range tests {
		// two-asset stableswap tests
		k0 := cfmmConstant(test.xReserve, test.yReserve)
		xOut := solveCfmm(test.xReserve, test.yReserve, test.yIn)
		fmt.Println(xOut)
		k1 := cfmmConstant(test.xReserve.Sub(xOut), test.yReserve.Add(test.yIn))
		decApproxEq(t, k0, k1, kErrTolerance)

		// multi-asset stableswap tests
		k2 := cfmmConstantMulti(test.xReserve, test.yReserve, test.uReserve, test.wSumSquares)
		xOut2 := solveCfmmMulti(test.xReserve, test.yReserve, test.wSumSquares, test.yIn)
		fmt.Println(xOut2)
		k3 := cfmmConstantMulti(test.xReserve.Sub(xOut2), test.yReserve.Add(test.yIn), test.uReserve, test.wSumSquares)
		decApproxEq(t, k2, k3, kErrTolerance)
	}
}
