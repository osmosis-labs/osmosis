package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) UpdateTickWithNewLiquidity(ctx sdk.Context, poolId uint64, tickIndex sdk.Int, liquidityDelta sdk.Int) {
	tickInfo := k.getTickInfo(ctx, poolId, tickIndex)

	liquidityBefore := tickInfo.Liquidity
	liquidityAfter := liquidityBefore.Add(liquidityDelta)

	tickInfo.Liquidity = liquidityAfter

	if liquidityBefore == sdk.ZeroInt() {
		tickInfo.Initialized = true
	}

	k.setTickInfo(ctx, poolId, tickIndex, tickInfo)
}

func (k Keeper) getTickInfo(ctx sdk.Context, poolId uint64, tickIndex sdk.Int) TickInfo {
	store := ctx.KVStore(k.storeKey)
	tickInfo := TickInfo{}
	key := types.KeyTick(poolId, tickIndex)
	osmoutils.MustGet(store, key, &tickInfo)
	return tickInfo
}

func (k Keeper) setTickInfo(ctx sdk.Context, poolId uint64, tickIndex sdk.Int, tickInfo TickInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTick(poolId, tickIndex)
	osmoutils.MustSet(store, key, &tickInfo)
}
