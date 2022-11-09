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

// tickToSqrtPrice takes the tick index and returns the corresponding sqrt of the price
func tickToSqrtPrice(tickIndex sdk.Int) (sdk.Dec, error) {
	sqrtPrice, err := sdk.NewDecWithPrec(10001, 4).Power(tickIndex.Uint64()).ApproxSqrt()
	if err != nil {
		return sdk.Dec{}, err
	}

	return sqrtPrice, nil
}

// TODO: implement this
// func (k Keeper) getTickAtSqrtRatio(sqrtPrice sdk.Dec) sdk.Int {
// 	return sdk.Int{}
// }

func (k Keeper) initOrUpdateTick(ctx sdk.Context, poolId uint64, tickIndex int64, liquidityIn sdk.Dec, upper bool) (err error) {
	tickInfo, err := k.GetTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return err
	}

	// calculate liquidityGross, which does not care about whether liquidityIn is positive or negative
	liquidityBefore := tickInfo.LiquidityGross

	// note that liquidityIn can be either positive or negative.
	// If negative, this would work as a subtraction from liquidityBefore
	liquidityAfter := addLiquidity(liquidityBefore, liquidityIn)

	tickInfo.LiquidityGross = liquidityAfter

	// calculate liquidityNet, which we take into account and track depending on whether liquidityIn is positive or negative
	if upper {
		tickInfo.LiquidityNet = tickInfo.LiquidityNet.Sub(liquidityIn)
	} else {
		tickInfo.LiquidityNet = tickInfo.LiquidityNet.Add(liquidityIn)
	}

	k.setTickInfo(ctx, poolId, tickIndex, tickInfo)

	return nil
}

func (k Keeper) crossTick(ctx sdk.Context, poolId uint64, tickIndex int64) (liquidityDelta sdk.Dec, err error) {
	tickInfo, err := k.GetTickInfo(ctx, poolId, tickIndex)
	if err != nil {
		return sdk.Dec{}, err
	}

	return tickInfo.LiquidityNet, nil
}

// NextInitializedTick returns the next initialized tick index based on the
// current or provided tick index. If no initialized tick exists, <0, false>
// will be returned. The zeroForOne argument indicates if we need to find the next
// initialized tick to the left or right of the current tick index, where true
// indicates searching to the left.
func (k Keeper) NextInitializedTick(ctx sdk.Context, poolId uint64, tickIndex int64, zeroForOne bool) (next int64, initialized bool) {
	store := ctx.KVStore(k.storeKey)

	// Construct a prefix store with a prefix of <TickPrefix | poolID>, allowing
	// us to retrieve the next initialized tick without having to scan all ticks.
	prefixBz := types.KeyTickPrefix(poolId)
	prefixStore := prefix.NewStore(store, prefixBz)

	var startKey []byte
	if !zeroForOne {
		startKey = types.TickIndexToBytes(tickIndex)
	} else {
		// When looking to the left of the current tick, we need to evaluate the
		// current tick as well. The end cursor for reverse iteration is non-inclusive
		// so must add one and handle overflow.
		startKey = types.TickIndexToBytes(osmomath.Max(tickIndex, tickIndex+1))
	}

	var iter db.Iterator
	if !zeroForOne {
		iter = prefixStore.Iterator(startKey, nil)
	} else {
		iter = prefixStore.ReverseIterator(nil, startKey)
	}

	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		// Since, we constructed our prefix store with <TickPrefix | poolID>, the
		// key is the encoding of a tick index.
		tick, err := types.TickIndexFromBytes(iter.Key())
		if err != nil {
			panic(fmt.Errorf("invalid tick index (%s): %v", string(iter.Key()), err))
		}

		if !zeroForOne && tick > tickIndex {
			return tick, true
		}
		if zeroForOne && tick <= tickIndex {
			return tick, true
		}
	}

	return 0, false
}

// getTickInfo gets tickInfo given poolId and tickIndex. Returns a boolean field that returns true if value is found for given key.
func (k Keeper) GetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64) (tickInfo TickInfo, err error) {
	store := ctx.KVStore(k.storeKey)
	tickStruct := TickInfo{}
	key := types.KeyTick(poolId, tickIndex)

	found, err := osmoutils.GetIfFound(store, key, &tickStruct)
	// return 0 values if key has not been initialized
	if !found {
		return TickInfo{LiquidityGross: sdk.ZeroDec(), LiquidityNet: sdk.ZeroDec()}, err
	}
	if err != nil {
		return tickStruct, err
	}

	return tickStruct, nil
}

func (k Keeper) setTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo TickInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTick(poolId, tickIndex)
	osmoutils.MustSet(store, key, &tickInfo)
}
