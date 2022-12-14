package concentrated_liquidity

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"

	"github.com/osmosis-labs/osmosis/v13/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

// getOrInitPosition retrieves the position for the given tick range. If it doesn't exist, it returns an initialized position with zero liquidity.
func (k Keeper) getOrInitPosition(
	ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
) (*model.Position, error) {
	if !k.poolExists(ctx, poolId) {
		return nil, types.PoolNotFoundError{PoolId: poolId}
	}
	if k.hasPosition(ctx, poolId, owner, lowerTick, upperTick) {
		position, err := k.getPosition(ctx, poolId, owner, lowerTick, upperTick)
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
) (err error) {
	position, err := k.getOrInitPosition(ctx, poolId, owner, lowerTick, upperTick)
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

	// TODO: consider deleting position if liquidity becomes zero

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, position)
	return nil
}

func (k Keeper) hasPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPosition(poolId, owner, lowerTick, upperTick)
	return store.Has(key)
}

// getPosition checks if a position exists at the provided upper and lower ticks for the given owner. Returns position if found.
func (k Keeper) getPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64) (*model.Position, error) {
	store := ctx.KVStore(k.storeKey)
	positionStruct := &model.Position{}
	key := types.KeyPosition(poolId, owner, lowerTick, upperTick)

	found, err := osmoutils.GetIfFound(store, key, positionStruct)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, types.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick}
	}

	return positionStruct, nil
}

func (k Keeper) setPosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	position *model.Position,
) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPosition(poolId, owner, lowerTick, upperTick)
	osmoutils.MustSet(store, key, position)
}

func (k Keeper) setRangePosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	rangePosition *model.RangePosition,
) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyRangePosition(poolId)
	osmoutils.MustSet(store, key, rangePosition)
}

func (k Keeper) RangePositionIterator(ctx sdk.Context,
	poolId uint64,
) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.KeyRangePosition(poolId))
}

func (k Keeper) finalizeRangeOrderPositions(ctx sdk.Context, poolId uint64) error {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	currentTick := pool.GetCurrentTick()

	iterator := k.RangePositionIterator(ctx, poolId)
	defer iterator.Close()

	// iterate over all range positions and withdraw expired positions
	for ; iterator.Valid(); iterator.Next() {
		rangePosition := model.RangePosition{}
		err := proto.Unmarshal(iterator.Value(), &rangePosition)
		if err != nil {
			panic(err)
		}

		// If the range position is not expired, skip it
		if (rangePosition.ZeroForOne && rangePosition.UpperTick < currentTick.Uint64()) ||
			(!rangePosition.ZeroForOne && currentTick.Uint64() < rangePosition.LowerTick) {
			_, _, err = k.withdrawPosition(ctx, poolId, sdk.AccAddress(rangePosition.Address), int64(rangePosition.LowerTick), int64(rangePosition.UpperTick), rangePosition.Liquidity)
			if err != nil {
				return err
			}
		}
	}

	// todo: delete range positions
	return nil
}
