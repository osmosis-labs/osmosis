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
	joinTime time.Time,
	freezeDuration time.Duration,
) (*model.Position, error) {
	if !k.poolExists(ctx, poolId) {
		return nil, types.PoolNotFoundError{PoolId: poolId}
	}
	if k.hasFullPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration) {
		position, err := k.GetPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)
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
	joinTime time.Time,
	freezeDuration time.Duration,
) (err error) {
	position, err := k.getOrInitPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)
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
	position.FreezeDuration = freezeDuration
	position.JoinTime = joinTime

	// TODO: consider deleting position if liquidity becomes zero

	err = k.initOrUpdatePositionUptime(ctx, poolId, position, owner, lowerTick, upperTick, liquidityDelta, joinTime, freezeDuration)
	if err != nil {
		return err
	}

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, position, joinTime, freezeDuration)
	return nil
}

func (k Keeper) hasFullPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)
	return store.Has(key)
}

// GetPosition checks if a position exists at the provided upper and lower ticks and freezeDuration time for the given owner. Returns position if found.
func (k Keeper) GetPosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, joinTime time.Time, freezeDuration time.Duration) (*model.Position, error) {
	store := ctx.KVStore(k.storeKey)
	positionStruct := &model.Position{}
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)

	found, err := osmoutils.Get(store, key, positionStruct)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, types.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick, JoinTime: joinTime, FreezeDuration: freezeDuration}
	}

	return positionStruct, nil
}

// GetUserPositions gets all the existing user positions, with the option to filter by a specific pool.
func (k Keeper) GetUserPositions(ctx sdk.Context, addr sdk.AccAddress, poolId uint64) ([]types.FullPositionByOwnerResult, error) {
	if poolId == 0 {
		return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyUserPositions(addr), ParseFullPositionFromBytes)
	} else {
		return osmoutils.GatherValuesFromStorePrefixWithKeyParser(ctx.KVStore(k.storeKey), types.KeyAddressAndPoolId(addr, poolId), ParseFullPositionFromBytes)
	}
}

// ParsePositionFromBz parses bytes into a position struct. Returns a parsed position and nil on success.
// Returns error if bytes length is zero or if fails to parse the given bytes into the position struct.
func (k Keeper) setPosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	position *model.Position,
	joinTime time.Time,
	freezeDuration time.Duration,
) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)
	osmoutils.MustSet(store, key, position)
}

func (k Keeper) deletePosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	joinTime time.Time,
	freezeDuration time.Duration,
) error {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyFullPosition(poolId, owner, lowerTick, upperTick, joinTime, freezeDuration)

	if !store.Has(key) {
		return types.PositionNotFoundError{PoolId: poolId, LowerTick: lowerTick, UpperTick: upperTick, JoinTime: joinTime, FreezeDuration: freezeDuration}
	}

	store.Delete(key)
	return nil
}

// CreateFullRangePosition creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, coins, and frozen until time.
// The function returns the amounts of token 0 and token 1, and the liquidity created from the position.
func (k Keeper) CreateFullRangePosition(ctx sdk.Context, concentratedPool types.ConcentratedPoolExtension, owner sdk.AccAddress, coins sdk.Coins, freezeDuration time.Duration) (amount0, amount1 sdk.Int, liquidity sdk.Dec, err error) {
	// Determine the max and min ticks for the concentrated pool we are migrating to.
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(concentratedPool.GetPrecisionFactorAtPriceOne())

	// Create a full range (min to max tick) concentrated liquidity position.
	amount0, amount1, liquidity, err = k.createPosition(ctx, concentratedPool.GetId(), owner, coins.AmountOf(concentratedPool.GetToken0()), coins.AmountOf(concentratedPool.GetToken1()), sdk.ZeroInt(), sdk.ZeroInt(), minTick, maxTick, freezeDuration)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, err
	}

	return amount0, amount1, liquidity, nil
}
