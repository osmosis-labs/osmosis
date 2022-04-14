package stableswap

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkCFMM(b *testing.B) {
	// Uses solveCfmm
	for i := 0; i < b.N; i++ {
		runCalc(solveCfmm)
	}
}

func BenchmarkBinarySearch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runCalc(solveCFMMBinarySearch(cfmmConstant))
	}
}

func runCalc(solve func(sdk.Dec, sdk.Dec, sdk.Dec) sdk.Dec) {
	xReserve := sdk.NewDec(rand.Int63n(100000) + 50000)
	yReserve := sdk.NewDec(rand.Int63n(100000) + 50000)
	yIn := sdk.NewDec(rand.Int63n(100000))
	solve(xReserve, yReserve, yIn)
}
