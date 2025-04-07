package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
)

// TestReverseRelationTickIndexToBytes tests if TickIndexToBytes and TickIndexFromBytes
// successfully converts back to the original value.
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
			tickIndex: types.MinInitializedTick,
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

func TestTickIndexFromBytes_ErrorCases(t *testing.T) {
	tests := map[string]struct {
		incorrectByteLength     bool
		incorrectNegativePrefix bool
		testZero                bool
	}{
		"use incorrect byte length": {
			incorrectByteLength: true,
		},
		"use incorrect negative prefix": {
			incorrectNegativePrefix: true,
		},
		"use incorrect positive prefix for positive number": {
			incorrectNegativePrefix: false,
		},
		"use incorrect positive prefix for zero": {
			incorrectNegativePrefix: false,
			testZero:                true,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			var tickIndexByte []byte
			if tc.incorrectByteLength {
				tickIndexByte = make([]byte, 10)
			} else if tc.incorrectNegativePrefix {
				// first create correct negative tick index byte using TickIndexToBytes
				correctTickIndexByte := types.TickIndexToBytes(-2)

				// now manually change it to have positive prefix
				correctTickIndexByte[0] = types.TickPositivePrefix[0]
				tickIndexByte = correctTickIndexByte
			} else {
				var correctTickIndexByte []byte
				if tc.testZero {
					correctTickIndexByte = types.TickIndexToBytes(0)
				} else {
					correctTickIndexByte = types.TickIndexToBytes(2)
				}

				// now manually change it to have negative prefix
				correctTickIndexByte[0] = types.TickNegativePrefix[0]
				tickIndexByte = correctTickIndexByte
			}
			tickIndex, err := types.TickIndexFromBytes(tickIndexByte)
			require.Error(t, err)
			require.Equal(t, int64(0), tickIndex)
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

func BenchmarkKeyPool(b *testing.B) {
	maxPoolId := 65536
	for i := 0; i < b.N; i++ {
		types.KeyPool(uint64(i % maxPoolId))
	}
}
