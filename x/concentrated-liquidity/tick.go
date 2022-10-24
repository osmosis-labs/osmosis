package concentrated_liquidity

import (
	fmt "fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"

	"github.com/osmosis-labs/osmosis/v12/osmomath"
	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) tickToSqrtPrice(tickIndex sdk.Int) (sdk.Dec, error) {
	price, err := sdk.NewDecWithPrec(10001, 4).Power(tickIndex.Uint64()).ApproxSqrt()
	if err != nil {
		return sdk.Dec{}, err
	}

	return price, nil
}

// TODO: implement this
// func (k Keeper) getTickAtSqrtRatio(sqrtPrice sdk.Dec) sdk.Int {
// 	return sdk.Int{}
// }

func (k Keeper) UpdateTickWithNewLiquidity(ctx sdk.Context, poolId uint64, tickIndex int64, liquidityDelta sdk.Int) {
	tickInfo := k.getTickInfo(ctx, poolId, tickIndex)

	liquidityBefore := tickInfo.Liquidity
	liquidityAfter := liquidityBefore.Add(liquidityDelta)
	tickInfo.Liquidity = liquidityAfter

	k.setTickInfo(ctx, poolId, tickIndex, tickInfo)
}

// NextInitializedTick returns the next initialized tick index based on the
// current or provided tick index. If no initialized tick exists, <0, false>
// will be returned. The lte argument indicates if we need to find the next
// initialized tick to the left or right of the current tick index, where true
// indicates searching to the left.
func (k Keeper) NextInitializedTick(ctx sdk.Context, poolId uint64, tickIndex int64, lte bool) (next int64, initialized bool) {
	store := ctx.KVStore(k.storeKey)

	// Construct a prefix store with a prefix of <TickPrefix | poolID>, allowing
	// us to retrieve the next initialized tick without having to scan all ticks.
	prefixBz := types.KeyTickPrefix(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)

	var startKey []byte
	if lte {
		// When looking to the left of the current tick, we need to evaluate the
		// current tick as well. The end cursor for reverse iteration is non-inclusive
		// so must add one and handle overflow.
		startKey = types.TickIndexToBytes(osmomath.Max(tickIndex, tickIndex+1))
	} else {
		startKey = types.TickIndexToBytes(tickIndex)
	}

	var iter db.Iterator
	if lte {
		iter = prefixStore.ReverseIterator(nil, startKey)
	} else {
		iter = prefixStore.Iterator(startKey, nil)
	}

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the encoding of a tick index.
		tick, err := types.TickIndexFromBytes(iter.Key())
		if err != nil {
			panic(fmt.Errorf("invalid tick index (%s): %v", string(iter.Key()), err))
		}

		if !lte && tick > tickIndex {
			return tick, true
		}
		if lte && tick <= tickIndex {
			return tick, true
		}
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
