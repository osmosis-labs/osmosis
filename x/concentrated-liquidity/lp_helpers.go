package concentrated_liquidity

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v13/x/concentrated-liquidity/types"
)

// getOrInitPosition retrieves the position for the given tick range. If it doesn't exist, it returns an initialized position with zero liquidity.
func (k Keeper) getOrInitPosition(
	ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	isIncentivized bool,
) (*model.Position, error) {
	if !k.poolExists(ctx, poolId) {
		return nil, types.PoolNotFoundError{PoolId: poolId}
	}
	if k.hasPosition(ctx, poolId, owner, lowerTick, upperTick, isIncentivized) {
		position, err := k.getPosition(ctx, poolId, owner, lowerTick, upperTick, isIncentivized)
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
	incentiveIDsCommittedTo []uint64,
) (err error) {
	var isIncentivized bool
	if len(incentiveIDsCommittedTo) > 0 {
		isIncentivized = true
	}

	position, err := k.getOrInitPosition(ctx, poolId, owner, lowerTick, upperTick, isIncentivized)
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

	// STUB: assume we are getting the max freeze time from the IDs themselves
	maxFreezeTime := time.Second * 30

	if isIncentivized {
		position.FrozenUntil = ctx.BlockTime().Add(maxFreezeTime)
	}
	fmt.Printf("position frozen until: %v \n", position.FrozenUntil)

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, position)
	return nil
}

func (k Keeper) hasPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, isIncentivized bool) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPosition(poolId, owner, lowerTick, upperTick, isIncentivized)
	return store.Has(key)
}

// getPosition checks if a position exists at the provided upper and lower ticks for the given owner. Returns position if found.
func (k Keeper) getPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, isIncentivized bool) (*model.Position, error) {
	store := ctx.KVStore(k.storeKey)
	positionStruct := &model.Position{}
	key := types.KeyPosition(poolId, owner, lowerTick, upperTick, isIncentivized)

	found, err := osmoutils.Get(store, key, positionStruct)
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
	// If the position has any entry for FrozenUntil, it is incentivized.
	positionIsIncentivized := !position.FrozenUntil.IsZero()
	fmt.Printf("positionIsIncentivized: %v \n", positionIsIncentivized)
	// We key by bool isIncentivized rather than the position's FrozenUntil field, because it is a better design
	// choice to group a user's incentivized and non-incentivized positions together rather than creating
	// a new entry for every one.
	key := types.KeyPosition(poolId, owner, lowerTick, upperTick, positionIsIncentivized)
	osmoutils.MustSet(store, key, position)
}

// checkPositionFreezeTime checks in the provided position has existed for longer than the incentivized position freeze time.
// If the position has existed for longer than the incentivized position freeze time, it returns true.
// If the position has not existed for longer than the incentivized position freeze time, it returns false.
func (k Keeper) checkPositionIsFrozen(ctx sdk.Context, position *model.Position) bool {
	// If the position has any entry for FrozenUntil, it is incentivized.
	positionIsIncentivized := !position.FrozenUntil.IsZero()
	if positionIsIncentivized {
		return ctx.BlockTime().Before(position.FrozenUntil)
	}
	return false
}
