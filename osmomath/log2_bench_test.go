package osmomath

import (
	"math/rand"
	"testing"
)

func BenchmarkLog2(b *testing.B) {
	tests := []BigDec{
		MustNewDecFromStr("1.2"),
		MustNewDecFromStr("1.234"),
		MustNewDecFromStr("1024"),
		NewBigDec(2048 * 2048 * 2048 * 2048 * 2048),
		MustNewDecFromStr("999999999999999999999999999999999999999999999999999999.9122181273612911"),
		MustNewDecFromStr("0.563289239121902491248219047129047129"),
		BigDecFromSDKDec(MaxSpotPrice),                                  // 2^128 - 1
		MustNewDecFromStr("336879543251729078828740861357450529340.45"), // (2^128 - 1) * 0.99
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		test := tests[rand.Int63n(int64(len(tests)))]
		b.StartTimer()
		_ = test.LogBase2()
	}
}
