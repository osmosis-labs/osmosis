package stableswap

import (
	"math/rand"
	"testing"

	"github.com/osmosis-labs/osmosis/osmomath"
)

func BenchmarkCFMM(b *testing.B) {
	// Uses solveCfmm
	for i := 0; i < b.N; i++ {
		runCalcCFMM(solveCfmm)
	}
}

func BenchmarkBinarySearchMultiAsset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runCalcMultiAsset(solveCFMMBinarySearchMulti)
	}
}

func runCalcCFMM(solve func(osmomath.BigDec, osmomath.BigDec, []osmomath.BigDec, osmomath.BigDec) osmomath.BigDec) {
	xReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	yReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	yIn := osmomath.NewBigDec(rand.Int63n(100000))
	solve(xReserve, yReserve, []osmomath.BigDec{}, yIn)
}

func runCalcTwoAsset(solve func(osmomath.BigDec, osmomath.BigDec, osmomath.BigDec) osmomath.BigDec) {
	xReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	yReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	yIn := osmomath.NewBigDec(rand.Int63n(100000))
	solve(xReserve, yReserve, yIn)
}

func runCalcMultiAsset(solve func(osmomath.BigDec, osmomath.BigDec, osmomath.BigDec, osmomath.BigDec) osmomath.BigDec) {
	xReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	yReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	mReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	nReserve := osmomath.NewBigDec(rand.Int63n(100000) + 50000)
	w := mReserve.Mul(mReserve).Add(nReserve.Mul(nReserve))
	yIn := osmomath.NewBigDec(rand.Int63n(100000))
	solve(xReserve, yReserve, w, yIn)
}
