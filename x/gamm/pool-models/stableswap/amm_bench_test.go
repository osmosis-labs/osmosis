package stableswap

import (
	"math/rand"
	"testing"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
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

func runCalc(solve func(osmomath.BigDec, osmomath.BigDec, osmomath.BigDec) osmomath.BigDec) {
	xReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	yReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	yIn := osmomath.NewBigDec(rand.Int63n(100000))
	solve(xReserve, yReserve, yIn)
}
