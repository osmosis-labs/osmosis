package concentrated_liquidity_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	cl "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func TestKeeper_TickOrdering(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("concentrated_liquidity")
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	k := cl.NewKeeper(storeKey)

	liquidityTicks := []int64{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, t := range liquidityTicks {
		k.SetTickInfo(ctx, 1, t, cl.TickInfo{})
	}

	store := ctx.KVStore(storeKey)
	prefixBz := types.KeyTickPrefix(1)
	prefixStore := prefix.NewStore(store, prefixBz)

	// Pick a value and ensure ordering is correct for lte=false, i.e. increasing
	// ticks.
	startKey := types.TickIndexToBytes(-4)
	iter := prefixStore.Iterator(startKey, nil)
	defer iter.Close()

	var vals []int64
	for ; iter.Valid(); iter.Next() {
		tick, err := types.TickIndexFromBytes(iter.Key())
		require.NoError(t, err)

		vals = append(vals, tick)
	}

	require.Equal(t, []int64{-4, 70, 78, 84, 139, 240, 535}, vals)

	// Pick a value and ensure ordering is correct for lte=true, i.e. decreasing
	// ticks.
	startKey = types.TickIndexToBytes(84)
	revIter := prefixStore.ReverseIterator(nil, startKey)
	defer revIter.Close()

	vals = nil
	for ; revIter.Valid(); revIter.Next() {
		tick, err := types.TickIndexFromBytes(revIter.Key())
		require.NoError(t, err)

		vals = append(vals, tick)
	}

	require.Equal(t, []int64{78, 70, -4, -55, -200}, vals)
}

func TestKeeper_NextInitializedTick(t *testing.T) {
	storeKey := sdk.NewKVStoreKey("concentrated_liquidity")
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	k := cl.NewKeeper(storeKey)

	liquidityTicks := []int64{-200, -55, -4, 70, 78, 84, 139, 240, 535}
	for _, t := range liquidityTicks {
		k.SetTickInfo(ctx, 1, t, cl.TickInfo{})
	}

	t.Run("lte=true", func(t *testing.T) {
		t.Run("returns tick to right if at initialized tick", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 78, true)
			require.Equal(t, int64(84), n)
			require.True(t, initd)
		})
		t.Run("returns tick to right if at initialized tick", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, -55, true)
			require.Equal(t, int64(-4), n)
			require.True(t, initd)
		})
		t.Run("returns the tick directly to the right", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 77, true)
			require.Equal(t, int64(78), n)
			require.True(t, initd)
		})
		t.Run("returns the tick directly to the right", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, -56, true)
			require.Equal(t, int64(-55), n)
			require.True(t, initd)
		})
		t.Run("returns the next words initialized tick if on the right boundary", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, -257, true)
			require.Equal(t, int64(-200), n)
			require.True(t, initd)
		})
		t.Run("returns the next initialized tick from the next word", func(t *testing.T) {
			k.SetTickInfo(ctx, 1, 340, cl.TickInfo{})

			n, initd := k.NextInitializedTick(ctx, 1, 328, true)
			require.Equal(t, int64(340), n)
			require.True(t, initd)
		})
	})

	t.Run("lte=false", func(t *testing.T) {
		t.Run("returns tick directly to the left of input tick if not initialized", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 79, false)
			require.Equal(t, int64(78), n)
			require.True(t, initd)
		})
		t.Run("returns same tick if initialized", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 78, false)
			require.Equal(t, int64(78), n)
			require.True(t, initd)
		})
		t.Run("returns next initialized tick far away", func(t *testing.T) {
			n, initd := k.NextInitializedTick(ctx, 1, 100, false)
			require.Equal(t, int64(84), n)
			require.True(t, initd)
		})
	})
}

func (suite *KeeperTestSuite) TestTickToSqrtPrice() {
	testCases := []struct {
		name              string
		tickIndex         sdk.Int
		sqrtPriceExpected string
		expectErr         bool
	}{
		{
			"happy path",
			sdk.NewInt(85176),
			"70.710004849206351867",
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			sqrtPrice, err := suite.App.ConcentratedLiquidityKeeper.TickToSqrtPrice(tc.tickIndex)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().Equal(tc.sqrtPriceExpected, sqrtPrice.String())
			}
		})
	}
}
