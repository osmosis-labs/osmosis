package concentrated_liquidity

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var emptyOptions = &accum.Options{}

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
	if k.hasFullPosition(ctx, poolId, owner, lowerTick, upperTick, frozenUntil) {
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
// If a position is present, it combines the existing liquidity in that position with the provided liquidity delta. It also
// bumps up all uptime accumulators to current time, including the ones the new position isn't eligible for.
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

	// We update accumulators _prior_ to any position-related updates to ensure
	// past rewards aren't distributed to new liquidity. We also update pool's
	// LastLiquidityUpdate here.
	err = k.updateUptimeAccumulatorsToNow(ctx, poolId)
	if err != nil {
		return err
	}

	// Create records for relevant uptime accumulators here.
	uptimeAccumulators, err := k.getUptimeAccumulators(ctx, poolId)
	if err != nil {
		return err
	}

	for uptimeIndex, uptime := range types.SupportedUptimes {
		// We assume every position update requires the position to be frozen for the
		// min uptime again. Thus, the difference between the position's `FrozenUntil`
		// and the blocktime when the update happens should be greater than or equal
		// to the required uptime.
		if position.FrozenUntil.Sub(ctx.BlockTime()) >= uptime {
			curUptimeAccum := uptimeAccumulators[uptimeIndex]

			// If a record does not exist for this uptime accumulator, create a new position.
			// Otherwise, add to existing record.
			positionName := string(types.KeyFullPosition(poolId, owner, lowerTick, upperTick, frozenUntil))
			recordExists, err := curUptimeAccum.HasPosition(positionName)
			if err != nil {
				return err
			}

			if !recordExists {
				err = curUptimeAccum.NewPosition(positionName, position.Liquidity, emptyOptions)
			} else if !liquidityDelta.IsNegative() {
				err = curUptimeAccum.AddToPosition(positionName, liquidityDelta)
			} else {
				err = curUptimeAccum.RemoveFromPosition(positionName, liquidityDelta.Neg())
			}
			if err != nil {
				return err
			}
		}
	}

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, position, frozenUntil)
	return nil
}

func (k Keeper) hasFullPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, frozenUntil time.Time) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, frozenUntil)
	return store.Has(key)
}

// GetPosition checks if a position exists at the provided upper and lower ticks and frozenUntil time for the given owner. Returns position if found.
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

// GetUserPositions gets all the existing user positions across many pools.
func (k Keeper) GetUserPositions(ctx sdk.Context, addr sdk.AccAddress) ([]types.FullPositionByOwnerResult, error) {
	return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyUserPositions(addr), ParseFullPositionFromBytes)
}

// ParsePositionFromBz parses bytes into a position struct. Returns a parsed position and nil on success.
// Returns error if bytes length is zero or if fails to parse the given bytes into the position struct.
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

// CreateFullRangePosition creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, coins, and frozen until time.
// The function returns the amounts of token 0 and token 1, and the liquidity created from the position.
func (k Keeper) CreateFullRangePosition(ctx sdk.Context, concentratedPool types.ConcentratedPoolExtension, owner sdk.AccAddress, coins sdk.Coins, frozenUntil time.Time) (amount0, amount1 sdk.Int, liquidity sdk.Dec, err error) {
	// Determine the max and min ticks for the concentrated pool we are migrating to.
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(concentratedPool.GetPrecisionFactorAtPriceOne())

	// Create a full range (min to max tick) concentrated liquidity position.
	amount0, amount1, liquidity, err = k.createPosition(ctx, concentratedPool.GetId(), owner, coins.AmountOf(concentratedPool.GetToken0()), coins.AmountOf(concentratedPool.GetToken1()), sdk.ZeroInt(), sdk.ZeroInt(), minTick, maxTick, frozenUntil)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	return amount0, amount1, liquidity, nil
}
