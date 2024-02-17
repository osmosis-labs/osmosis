package osmomath

import (
	"math/rand"
	"testing"
)

func BenchmarkLog2(b *testing.B) {
	tests := []BigDec{
		MustNewBigDecFromStr("1.2"),
		MustNewBigDecFromStr("1.234"),
		MustNewBigDecFromStr("1024"),
		NewBigDec(2048 * 2048 * 2048 * 2048 * 2048),
		MustNewBigDecFromStr("999999999999999999999999999999999999999999999999999999.9122181273612911"),
		MustNewBigDecFromStr("0.563289239121902491248219047129047129"),
		BigDecFromDec(MaxSpotPrice),                                        // 2^128 - 1
		MustNewBigDecFromStr("336879543251729078828740861357450529340.45"), // (2^128 - 1) * 0.99
	}

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		test := tests[rand.Int63n(int64(len(tests)))]
		b.StartTimer()
		_ = test.LogBase2()
	}
}
