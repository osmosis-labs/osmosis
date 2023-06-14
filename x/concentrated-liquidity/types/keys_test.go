package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
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

// This is just to sanity check that keys used as accumulator names don't contain `||`
func TestAccumulatorNameKeys(t *testing.T) {
	tests := map[string]struct {
		poolId uint64
	}{
		"basic": {1},
		"zero":  {0},
		"pipe":  {124},
		"pipe2": {124*256 + 124},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			name1 := types.KeySpreadRewardPoolAccumulator(tc.poolId)
			require.NotContains(t, string(name1), accum.KeySeparator)
			for i := 0; i < 125; i++ {
				name2 := types.KeyUptimeAccumulator(tc.poolId, uint64(i))
				require.NotContains(t, string(name2), accum.KeySeparator)
			}
		})
	}
}

// sanity test to show that addresses are hex encoded.
func TestAddrKeyEncoding(t *testing.T) {
	addr := "bytes_underlying_address"
	accAddr := sdk.AccAddress(addr)
	bz := types.KeyUserPositions(accAddr)
	require.Equal(t, "\x02|62797465735f756e6465726c79696e675f61646472657373|", string(bz))
}
