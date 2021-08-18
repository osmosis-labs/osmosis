package osmomath

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkPow(b *testing.B) {
	tests := []struct {
		base sdk.Dec
		exp  sdk.Dec
	}{
		// TODO: Choose selection here more robustly
		{
			base: sdk.MustNewDecFromStr("1.2"),
			exp:  sdk.MustNewDecFromStr("1.2"),
		},
		{
			base: sdk.MustNewDecFromStr("0.5"),
			exp:  sdk.MustNewDecFromStr("11.122"),
		},
		{
			base: sdk.MustNewDecFromStr("0.1"),
			exp:  sdk.MustNewDecFromStr("0.00000492"),
		},
		{
			base: sdk.MustNewDecFromStr("0.0002423"),
			exp:  sdk.MustNewDecFromStr("0.1234"),
		},
		{
			base: sdk.MustNewDecFromStr("0.493"),
			exp:  sdk.MustNewDecFromStr("0.00000121"),
		},
		{
			base: sdk.MustNewDecFromStr("0.000249"),
			exp:  sdk.MustNewDecFromStr("2.304"),
		},
		{
			base: sdk.MustNewDecFromStr("0.2342"),
			exp:  sdk.MustNewDecFromStr("32.2"),
		},
		{
			base: sdk.MustNewDecFromStr("0.000999"),
			exp:  sdk.MustNewDecFromStr("142.4"),
		},
		{
			base: sdk.MustNewDecFromStr("1.234"),
			exp:  sdk.MustNewDecFromStr("120.3"),
		},
		{
			base: sdk.MustNewDecFromStr("0.00122"),
			exp:  sdk.MustNewDecFromStr("123.2"),
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
		base sdk.Dec
	}{
		// TODO: Choose selection here more robustly
		{
			base: sdk.MustNewDecFromStr("1.29847"),
		},
		{
			base: sdk.MustNewDecFromStr("1.313135"),
		},
		{
			base: sdk.MustNewDecFromStr("1.65976735939"),
		},
	}
	one_half := sdk.MustNewDecFromStr("0.5")

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			Pow(test.base, one_half)
		}
	}
}
