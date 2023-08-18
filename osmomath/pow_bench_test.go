package osmomath

import (
	"testing"
)

func BenchmarkPow(b *testing.B) {
	tests := []struct {
		base SDKDec
		exp  SDKDec
	}{
		// TODO: Choose selection here more robustly
		{
			base: MustNewSDKDecFromStr("1.2"),
			exp:  MustNewSDKDecFromStr("1.2"),
		},
		{
			base: MustNewSDKDecFromStr("0.5"),
			exp:  MustNewSDKDecFromStr("11.122"),
		},
		{
			base: MustNewSDKDecFromStr("0.1"),
			exp:  MustNewSDKDecFromStr("0.00000492"),
		},
		{
			base: MustNewSDKDecFromStr("0.0002423"),
			exp:  MustNewSDKDecFromStr("0.1234"),
		},
		{
			base: MustNewSDKDecFromStr("0.493"),
			exp:  MustNewSDKDecFromStr("0.00000121"),
		},
		{
			base: MustNewSDKDecFromStr("0.000249"),
			exp:  MustNewSDKDecFromStr("2.304"),
		},
		{
			base: MustNewSDKDecFromStr("0.2342"),
			exp:  MustNewSDKDecFromStr("32.2"),
		},
		{
			base: MustNewSDKDecFromStr("0.000999"),
			exp:  MustNewSDKDecFromStr("142.4"),
		},
		{
			base: MustNewSDKDecFromStr("1.234"),
			exp:  MustNewSDKDecFromStr("120.3"),
		},
		{
			base: MustNewSDKDecFromStr("0.00122"),
			exp:  MustNewSDKDecFromStr("123.2"),
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
		base SDKDec
	}{
		// TODO: Choose selection here more robustly
		{
			base: MustNewSDKDecFromStr("1.29847"),
		},
		{
			base: MustNewSDKDecFromStr("1.313135"),
		},
		{
			base: MustNewSDKDecFromStr("1.65976735939"),
		},
	}
	one_half := MustNewSDKDecFromStr("0.5")

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			Pow(test.base, one_half)
		}
	}
}
