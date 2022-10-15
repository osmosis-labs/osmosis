package osmomath

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMulCoinAmtsByU64(t *testing.T) {
	type testcase struct {
		i    sdk.Int
		u    uint64
		want sdk.Int
	}
	tests := map[string]testcase{
		"mul with zero": {
			sdk.NewInt(5), 0, sdk.ZeroInt(),
		},
		"5 * 1": {
			sdk.NewInt(5), 1, sdk.NewInt(5),
		},
		"5 * 10": {
			sdk.NewInt(5), 10, sdk.NewInt(50),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := MulIntByU64(tt.i, tt.u)
			require.Equal(t, tt.want, got)
		})
	}
}
