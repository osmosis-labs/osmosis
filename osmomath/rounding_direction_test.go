package osmomath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDivIntByU64ToBigDec(t *testing.T) {
	type testcase struct {
		i      Int
		u      uint64
		round  RoundingDirection
		want   BigDec
		expErr bool
	}
	tests := map[string]testcase{
		"div by zero": {NewInt(5), 0, RoundUp, BigDec{}, true},
		"5/3 round up": {NewInt(5), 3, RoundUp,
			MustNewBigDecFromStr("1.666666666666666666666666666666666667"), false},
		"5/3 round down": {NewInt(5), 3, RoundDown,
			MustNewBigDecFromStr("1.666666666666666666666666666666666666"), false},
		"5/3 round banker": {NewInt(5), 3, RoundBankers,
			MustNewBigDecFromStr("1.666666666666666666666666666666666667"), false},
		"7/3 round up": {NewInt(7), 3, RoundUp,
			MustNewBigDecFromStr("2.333333333333333333333333333333333334"), false},
		"7/3 round down": {NewInt(7), 3, RoundDown,
			MustNewBigDecFromStr("2.333333333333333333333333333333333333"), false},
		"7/3 round banker": {NewInt(7), 3, RoundBankers,
			MustNewBigDecFromStr("2.333333333333333333333333333333333333"), false},
	}
	addTCForAllRoundingModes := func(prefix string, i Int, u uint64, want BigDec) {
		for round := 1; round < 4; round++ {
			tests[fmt.Sprintf("%s rounding=%d", prefix, round)] =
				testcase{i, u, RoundingDirection(round), want, false}
		}
	}
	addTCForAllRoundingModes("odd divided by 2", NewInt(5), 2, NewBigDecWithPrec(25, 1))

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := DivIntByU64ToBigDec(tt.i, tt.u, tt.round)
			require.Equal(t, tt.want, got)
			if tt.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
