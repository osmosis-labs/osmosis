package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

func TestValidateTicks(t *testing.T) {
	tests := map[string]struct {
		i           interface{}
		expectError bool
	}{
		"happy path": {
			i: []uint64{1, 100},
		},
		"error: zero tick spacing": {
			i:           []uint64{1, 0},
			expectError: true,
		},
		"error: wrong type": {
			i:           []int64{1, 0},
			expectError: true,
		},
		"error: not a multiple of max tick": {
			i:           []int64{types.MaxTick - 1},
			expectError: true,
		},
		"error: not a multiple of min tick": {
			i:           []int64{types.MinTick + 1},
			expectError: true,
		},
		"error: greater than max tick": {
			i:           []int64{types.MaxTick * 2},
			expectError: true,
		},
		"error: smaller than min tick": {
			i:           []int64{types.MinTick * 2},
			expectError: true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {

			err := types.ValidateTicks(tc.i)

			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
