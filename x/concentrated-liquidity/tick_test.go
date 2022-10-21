package concentrated_liquidity

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestNextInitializedTick(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("concentrated_liquidity")
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	k := Keeper{storeKey: storeKey}

	liquidityTicks := []int64{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, t := range liquidityTicks {
		k.setTickInfo(ctx, 1, t, TickInfo{})
	}

	t.Run("lte=false", func(t *testing.T) {
		t.Run("returns tick to right if at initialized tick", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 78, false)
			require.Equal(t, int64(84), n)
			require.True(t, initd)
		})
		t.Run("returns tick to right if at initialized tick", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, -55, false)
			require.Equal(t, int64(-4), n)
			require.True(t, initd)
		})
		t.Run("returns the tick directly to the right", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 77, false)
			require.Equal(t, int64(78), n)
			require.True(t, initd)
		})
		t.Run("returns the tick directly to the right", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, -56, false)
			require.Equal(t, int64(-55), n)
			require.True(t, initd)
		})
		t.Run("returns the next words initialized tick if on the right boundary", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, -257, false)
			require.Equal(t, int64(-200), n)
			require.True(t, initd)
		})
		t.Run("returns the next initialized tick from the next word", func(t *testing.T) {
			k.setTickInfo(ctx, 1, 340, TickInfo{})

			n, initd := k.NextInitializedTick(ctx, 1, 328, false)
			require.Equal(t, int64(340), n)
			require.True(t, initd)
		})
	})

	t.Run("lte=true", func(t *testing.T) {
		t.Run("returns tick directly to the left of input tick if not initialized", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 79, true)
			require.Equal(t, int64(78), n)
			require.True(t, initd)
		})
		t.Run("returns same tick if initialized", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 78, true)
			require.Equal(t, int64(78), n)
			require.True(t, initd)
		})
	})
}
