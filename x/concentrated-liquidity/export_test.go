package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

// OrderInitialPoolDenoms sets the pool denoms of a cl pool
func OrderInitialPoolDenoms(denom0, denom1 string) (string, string, error) {
	return cltypes.OrderInitialPoolDenoms(denom0, denom1)
}

func (k Keeper) CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, amount0Desired, amount1Desired, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (amtDenom0, amtDenom1 sdk.Int, liquidityCreated sdk.Dec, err error) {
	return k.createPosition(ctx, poolId, owner, amount0Desired, amount1Desired, amount0Min, amount1Min, lowerTick, upperTick)
}

func (k Keeper) SetTickInfo(ctx sdk.Context, poolId uint64, tickIndex int64, tickInfo types.TickInfo) {
	k.setTickInfo(ctx, poolId, tickIndex, tickInfo)
}

func (k Keeper) GetPoolById(ctx sdk.Context, poolId uint64) (cltypes.ConcentratedPoolExtension, error) {
	return k.getPoolById(ctx, poolId)
}

func (k Keeper) SetPool(ctx sdk.Context, pool cltypes.ConcentratedPoolExtension) error {
	return k.setPool(ctx, pool)
}
