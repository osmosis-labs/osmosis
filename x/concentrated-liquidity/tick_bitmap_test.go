package concentrated_liquidity_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
)

func TestTickBitmap_FlipTick(t *testing.T) {
	tb := cl.NewTickBitmap()

	require.NoError(t, tb.FlipTick(85176, 1))
	require.Error(t, tb.FlipTick(85176, 1000))
}

func TestTickBitmap_NextInitializedTickWithinOneWord(t *testing.T) {
	tb := cl.NewTickBitmap()

	// word boundaries are at 64 bits
	ticks := []int32{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, tick := range ticks {
		require.NoError(t, tb.FlipTick(tick, 1))
	}

	t.Run("lte = false; returns tick to right if at initialized tick", func(t *testing.T) {
		next, initd := tb.NextInitializedTickWithinOneWord(78, 1, false)
		require.Equal(t, int32(84), next)
		require.True(t, initd)
	})
}
