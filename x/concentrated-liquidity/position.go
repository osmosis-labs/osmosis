package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v12/osmoutils"
	types "github.com/osmosis-labs/osmosis/v12/x/concentrated-liquidity/types"
)

func (k Keeper) initOrUpdatePosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	liquidityDelta sdk.Dec,
) (err error) {
	position, err := k.GetPosition(ctx, poolId, owner, lowerTick, upperTick)
	if err != nil {
		return err
	}

	liquidityBefore := position.Liquidity

	// note that liquidityIn can be either positive or negative.
	// If negative, this would work as a subtraction from liquidityBefore
	liquidityAfter := liquidityBefore.Add(liquidityDelta)

	position.Liquidity = liquidityAfter

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, position)
	return nil
}

func (k Keeper) GetPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) (position Position, err error) {
	store := ctx.KVStore(k.storeKey)
	positionStruct := Position{}
	key := types.KeyPosition(poolId, owner, lowerTick, upperTick)

	found, err := osmoutils.GetIfFound(store, key, &positionStruct)
	// return 0 values if key has not been initialized
	if !found {
		return Position{Liquidity: sdk.ZeroDec()}, nil
	}
	if err != nil {
		return positionStruct, err
	}

	return positionStruct, nil
}

func (k Keeper) setPosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	position Position,
) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPosition(poolId, owner, lowerTick, upperTick)
	osmoutils.MustSet(store, key, &position)
}
