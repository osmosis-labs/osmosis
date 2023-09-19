package osmomath

import (
	"testing"
)

func BenchmarkPow(b *testing.B) {
	tests := []struct {
		base Dec
		exp  Dec
	}{
		// TODO: Choose selection here more robustly
		{
			base: MustNewDecFromStr("1.2"),
			exp:  MustNewDecFromStr("1.2"),
		},
		{
			base: MustNewDecFromStr("0.5"),
			exp:  MustNewDecFromStr("11.122"),
		},
		{
			base: MustNewDecFromStr("0.1"),
			exp:  MustNewDecFromStr("0.00000492"),
		},
		{
			base: MustNewDecFromStr("0.0002423"),
			exp:  MustNewDecFromStr("0.1234"),
		},
		{
			base: MustNewDecFromStr("0.493"),
			exp:  MustNewDecFromStr("0.00000121"),
		},
		{
			base: MustNewDecFromStr("0.000249"),
			exp:  MustNewDecFromStr("2.304"),
		},
		{
			base: MustNewDecFromStr("0.2342"),
			exp:  MustNewDecFromStr("32.2"),
		},
		{
			base: MustNewDecFromStr("0.000999"),
			exp:  MustNewDecFromStr("142.4"),
		},
		{
			base: MustNewDecFromStr("1.234"),
			exp:  MustNewDecFromStr("120.3"),
		},
		{
			base: MustNewDecFromStr("0.00122"),
			exp:  MustNewDecFromStr("123.2"),
		},
	}

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			Pow(test.base, test.exp)
		}
	}
}

func BenchmarkSqrtPow(b *testing.B) {
	tests := []struct {
		base Dec
	}{
		// TODO: Choose selection here more robustly
		{
			base: MustNewDecFromStr("1.29847"),
		},
		{
			base: MustNewDecFromStr("1.313135"),
		},
		{
			base: MustNewDecFromStr("1.65976735939"),
		},
	}
	one_half := MustNewDecFromStr("0.5")

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			Pow(test.base, one_half)
		}
	}
}
