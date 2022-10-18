package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) UpdateTickWithNewLiquidity(ctx sdk.Context, poolId uint64, liquidityDelta sdk.Int) {
	tickInfo := k.getTickInfoByPoolID(ctx, poolId)

	liquidityBefore := tickInfo.Liquidity
	liquidityAfter := liquidityBefore.Add(liquidityDelta)

	tickInfo.Liquidity = liquidityAfter

	if liquidityBefore == sdk.ZeroInt() {
		tickInfo.Initialized = true
	}

	k.setTickInfoByPoolID(ctx, poolId, tickInfo)
}

func (k Keeper) getTickInfoByPoolID(ctx sdk.Context, poolId uint64) types.TickInfo {
	store := ctx.KVStore(k.storeKey)
	tickInfo := types.TickInfo{}
	key := types.KeyTickByPool(poolId)
	osmoutils.MustGet(store, key, &tickInfo)
	return tickInfo
}

func (k Keeper) setTickInfoByPoolID(ctx sdk.Context, poolId uint64, tickInfo types.TickInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyTickByPool(poolId)
	osmoutils.MustSet(store, key, &tickInfo)
}
