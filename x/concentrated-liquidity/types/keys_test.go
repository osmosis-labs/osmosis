package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/v16/x/concentrated-liquidity/types"
)

// TestReverseRelationTickIndexToBytes tests if TickIndexToBytes and TickIndexFromBytes
// succesfully converts back to the original value.
func TestReverseRelationTickIndexToBytes(t *testing.T) {
	tests := map[string]struct {
		tickIndex int64
	}{
		"positive tick index": {
			tickIndex: 3,
		},
		"negative tick index": {
			tickIndex: -3,
		},
		"zero tick index": {
			tickIndex: 0,
		},
		"maximum tick index": {
			tickIndex: types.MaxTick,
		},
		"minimum tick index": {
			tickIndex: types.MinTick,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tickIndexBytes := types.TickIndexToBytes(tc.tickIndex)

			// now we convert it back to tick index from bytes
			tickIndexConverted, err := types.TickIndexFromBytes(tickIndexBytes)
			require.NoError(t, err)
			require.Equal(t, tc.tickIndex, tickIndexConverted)
		})
	}
}
