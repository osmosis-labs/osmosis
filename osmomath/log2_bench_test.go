package osmomath

import (
	"testing"
)

// Benchmark that scales x between 1 <= x < 2 -      417597  ns/op
// Benchmark that scales x between 1 <= x < 1.0001 - 3372629 ns/op
func BenchmarkLog2(b *testing.B) {
	tests := []struct {
		value BigDec
	}{
		// TODO: Choose selection here more robustly
		{
			value: MustNewDecFromStr("1.2"),
		},
		{
			value: MustNewDecFromStr("1.234"),
		},
		{
			value: MustNewDecFromStr("1024"),
		},
		{
			value: NewBigDec(2048 * 2048 * 2048 * 2048 * 2048),
		},
		{
			value: MustNewDecFromStr("999999999999999999999999999999999999999999999999999999.9122181273612911"),
		},
	}

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			test.value.ApproxLog2()
		}
	}
}
