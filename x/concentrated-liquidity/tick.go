package concentrated_liquidity

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) UpdateTickWithNewLiquidity(ctx sdk.Context, poolId uint64, tickIndex int64, liquidityDelta sdk.Int) {
	tickInfo := k.getTickInfo(ctx, poolId, tickIndex)

	liquidityBefore := tickInfo.Liquidity
	liquidityAfter := liquidityBefore.Add(liquidityDelta)
	tickInfo.Liquidity = liquidityAfter

	k.setTickInfo(ctx, poolId, tickIndex, tickInfo)
}

// NextInitializedTick returns the next initialized tick index based on the
// current or provided tick index. If no initialized tick exists, <0, false>
// will be returned.
func (k Keeper) NextInitializedTick(ctx sdk.Context, poolId uint64, tickIndex int64, lte bool) (next int64, initialized bool) {
	store := ctx.KVStore(k.storeKey)

	// Construct a prefix store with a prefix of <TickPrefix | poolID>, allowing
	// us to retrieve the next initialized tick without having to scan all ticks.
	prefixBz := types.KeyTickPrefix(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)

	var iter db.Iterator
	if lte {
		iter = prefixStore.ReverseIterator(sdk.Uint64ToBigEndian(uint64(tickIndex)), nil)
	} else {
		iter = prefixStore.Iterator(sdk.Uint64ToBigEndian(uint64(tickIndex)), nil)
	}

	defer iter.Close()

	i := 0
	for ; iter.Valid() && i < 2; iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the BigEndianToUint64 encoding of a tick index.
		tick := int64(sdk.BigEndianToUint64(iter.Key()))

		if tick > tickIndex {
			return tick, true
		}

		i++
	}

	return 0, false
}

func (k Keeper) getTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) TickInfo {
	store := ctx.KVStore(k.storeKey)
	tickInfo := TickInfo{}
	key := types.KeyTick(poolId, tickIndex)
	osmoutils.MustGet(store, key, &tickInfo)
	return tickInfo
}

func (k Keeper) setTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo TickInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTick(poolId, tickIndex)
	osmoutils.MustSet(store, key, &tickInfo)
}
