package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func BenchmarkPow(b *testing.B) {
	tests := []struct {
		base sdk.Dec
		exp  sdk.Dec
	}{
		{
			base: sdk.NewDecFromStr(12, 1),
			exp:  sdk.NewDecWithPrec(12, 1),
		},
		{
			base: sdk.NewDecWithPrec(5, 1),
			exp:  sdk.NewDecWithPrec(11122, 3),
		},
		{
			base: sdk.NewDecWithPrec(1, 1),
			exp:  sdk.NewDecWithPrec(492, 8),
		},
		{
			base: sdk.NewDecWithPrec(2423, 7),
			exp:  sdk.NewDecWithPrec(1213, 1),
		},
		{
			base: sdk.NewDecWithPrec(493, 3),
			exp:  sdk.NewDecWithPrec(121, 8),
		},
		{
			base: sdk.NewDecWithPrec(249, 6),
			exp:  sdk.NewDecWithPrec(2304, 1),
		},
		{
			base: sdk.NewDecWithPrec(2342, 4),
			exp:  sdk.NewDecWithPrec(322, 1),
		},
		{
			base: sdk.NewDecWithPrec(999, 6),
			exp:  sdk.NewDecWithPrec(1424, 1),
		},
		{
			base: sdk.NewDecWithPrec(1234, 3),
			exp:  sdk.NewDecWithPrec(1203, 1),
		},
		{
			base: sdk.NewDecWithPrec(122, 5),
			exp:  sdk.NewDecWithPrec(1232, 1),
		},
	}

	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			pow(test.base, test.exp)
		}
	}
}
