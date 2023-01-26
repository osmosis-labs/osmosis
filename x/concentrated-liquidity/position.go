package concentrated_liquidity

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v14/x/concentrated-liquidity/types"
)

// getOrInitPosition retrieves the position for the given tick range. If it doesn't exist, it returns an initialized position with zero liquidity.
func (k Keeper) getOrInitPosition(
	ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	frozenUntil time.Time,
) (*model.Position, error) {
	if !k.poolExists(ctx, poolId) {
		return nil, types.PoolNotFoundError{PoolId: poolId}
	}
	if k.hasPosition(ctx, poolId, owner, lowerTick, upperTick, frozenUntil) {
		position, err := k.GetPosition(ctx, poolId, owner, lowerTick, upperTick, frozenUntil)
		if err != nil {
			return nil, err
		}
		return position, nil
	}
	return &model.Position{Liquidity: sdk.ZeroDec()}, nil
}

// initOrUpdatePosition checks to see if the specified owner has an existing position at the given tick range.
// If a position is not present, it initializes the position with the provided liquidity delta.
// If a position is present, it combines the existing liquidity in that position with the provided liquidity delta.
func (k Keeper) initOrUpdatePosition(
	ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	liquidityDelta sdk.Dec,
	frozenUntil time.Time,
) (err error) {
	position, err := k.getOrInitPosition(ctx, poolId, owner, lowerTick, upperTick, frozenUntil)
	if err != nil {
		return err
	}

	liquidityBefore := position.Liquidity

	// note that liquidityIn can be either positive or negative.
	// If negative, this would work as a subtraction from liquidityBefore
	liquidityAfter := liquidityBefore.Add(liquidityDelta)
	if liquidityAfter.IsNegative() {
		return types.NegativeLiquidityError{Liquidity: liquidityAfter}
	}

	position.Liquidity = liquidityAfter

	position.FrozenUntil = frozenUntil

	// TODO: consider deleting position if liquidity becomes zero

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, position, frozenUntil)
	return nil
}

func (k Keeper) hasPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, frozenUntil time.Time) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, frozenUntil)
	return store.Has(key)
}

// GetPosition checks if a position exists at the provided upper and lower ticks for the given owner. Returns position if found.
func (k Keeper) GetPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, frozenUntil time.Time) (*model.Position, error) {
	store := ctx.KVStore(k.storeKey)
	positionStruct := &model.Position{}
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, frozenUntil)

	found, err := osmoutils.Get(store, key, positionStruct)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, types.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick, FrozenUntil: frozenUntil}
	}

	return positionStruct, nil
}

func (k Keeper) setPosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	position *model.Position,
	frozenUntil time.Time,
) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, frozenUntil)
	osmoutils.MustSet(store, key, position)
}

func (k Keeper) deletePosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	frozenUntil time.Time,
) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, frozenUntil)

	if !store.Has(key) {
		return types.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick, FrozenUntil: frozenUntil}
	}

	store.Delete(key)
	return nil
}
