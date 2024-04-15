package osmomath

import "testing"

// BenchmarkChopPrecisionAndRoundUpMutNoRounding benchmarks the chopPrecisionAndRoundUpMut function under default precision
// as perfect integers
func BenchmarkChopPrecisionAndRoundUpMutNoRounding(b *testing.B) {
	testInts := []int64{2, 1234, 1234567890, 1234567890123456789}
	testBigInts := []BigDec{}
	for _, i := range testInts {
		d := NewBigDec(i)
		d2 := d.Mul(d)
		d3 := d2.Mul(d)
		testBigInts = append(testBigInts, d, d2, d3)
	}

	b.StartTimer()

	// Run the benchmark function b.N times
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(testBigInts); j++ {
			// Call the function with the input value
			chopPrecisionAndRoundUpMut(testBigInts[j].BigInt(), defaultBigDecPrecisionReuse)
		}
	}
}

func BenchmarkChopPrecisionAndRoundUpMutSlightRounding(b *testing.B) {
	testInts := []int64{2, 1234, 1234567890, 1234567890123456789}
	testBigInts := []BigDec{}
	roundingFactor := NewBigDec(1).QuoInt64(123)
	for _, i := range testInts {
		d := NewBigDec(i).MulMut(roundingFactor)
		d2 := d.Mul(d)
		d3 := d2.Mul(d)
		testBigInts = append(testBigInts, d, d2, d3)
	}

	b.StartTimer()

	// Run the benchmark function b.N times
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(testBigInts); j++ {
			// Call the function with the input value
			chopPrecisionAndRoundUpMut(testBigInts[j].BigInt(), defaultBigDecPrecisionReuse)
		}
	}
}
