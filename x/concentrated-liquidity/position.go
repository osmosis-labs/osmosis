package concentrated_liquidity

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

var emptyOptions = &accum.Options{}

// getOrInitPosition retrieves the position's liquidity for the given tick range.
// If it doesn't exist, it returns zero.
func (k Keeper) getOrInitPosition(
	ctx sdk.Context,
	positionId uint64,
) (sdk.Dec, error) {
	if k.hasFullPosition(ctx, positionId) {
		positionLiquidity, err := k.GetPositionLiquidity(ctx, positionId)
		if err != nil {
			return sdk.Dec{}, err
		}
		return positionLiquidity, nil
	}
	return sdk.ZeroDec(), nil
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
	positionId uint64,
) (err error) {
	liquidity, err := k.getOrInitPosition(ctx, positionId)
	if err != nil {
		return err
	}

	// note that liquidityIn can be either positive or negative.
	// If negative, this would work as a subtraction from liquidityBefore
	liquidity = liquidity.Add(liquidityDelta)
	if liquidity.IsNegative() {
		return types.NegativeLiquidityError{Liquidity: liquidity}
	}

	err = k.initOrUpdatePositionUptime(ctx, poolId, liquidity, owner, lowerTick, upperTick, liquidityDelta, joinTime, freezeDuration, positionId)
	if err != nil {
		return err
	}

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, freezeDuration, liquidity, positionId)
	return nil
}

func (k Keeper) hasFullPosition(ctx sdk.Context, positionId uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPositionId(positionId)
	return store.Has(key)
}

// GetPositionLiquidity checks if the provided positionId exists. Returns position if found.
func (k Keeper) GetPositionLiquidity(ctx sdk.Context, positionId uint64) (sdk.Dec, error) {
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Dec{}, err
	}

	return position.Liquidity, nil
}

// GetPosition checks if the given position id exists. Returns position if found.
func (k Keeper) GetPosition(ctx sdk.Context, positionId uint64) (model.Position, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPositionId(positionId)

	positionStruct := &model.Position{}
	found, err := osmoutils.Get(store, key, positionStruct)
	if err != nil {
		return model.Position{}, err
	}

	if !found {
		return model.Position{}, types.PositionIdNotFoundError{PositionId: positionId}
	}

	return *positionStruct, nil
}

// GetUserPositions gets all the existing user positions, with the option to filter by a specific pool.
func (k Keeper) GetUserPositions(ctx sdk.Context, addr sdk.AccAddress, poolId uint64) ([]model.Position, error) {
	var prefix []byte
	if poolId == 0 {
		prefix = types.KeyUserPositions(addr)
	} else {
		prefix = types.KeyAddressAndPoolId(addr, poolId)
	}

	positions := []model.Position{}

	// Gather all position IDs for the given user and pool ID.
	positionIds, err := osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), prefix, ParsePositionIdFromBz)
	if err != nil {
		return nil, err
	}

	// Retrieve each position from the store using its ID and add it to the result slice.
	for _, positionId := range positionIds {
		position, err := k.GetPosition(ctx, positionId)
		if err != nil {
			return nil, err
		}
		positions = append(positions, position)
	}

	return positions, nil
}

// setPosition sets the position information for a given user in a given pool.
func (k Keeper) setPosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	joinTime time.Time,
	freezeDuration time.Duration,
	liquidity sdk.Dec,
	positionId uint64,
) {
	store := ctx.KVStore(k.storeKey)

	// Create a new Position object with the provided information.
	position := model.Position{
		PositionId:     positionId,
		PoolId:         poolId,
		Address:        owner.String(),
		LowerTick:      lowerTick,
		UpperTick:      upperTick,
		JoinTime:       joinTime,
		FreezeDuration: freezeDuration,
		Liquidity:      liquidity,
	}

	// Set the position ID to position mapping.
	key := types.KeyPositionId(positionId)
	osmoutils.MustSet(store, key, &position)

	// Set the address-pool-position ID to position mapping.
	key = types.KeyAddressPoolIdPositionId(owner, poolId, positionId)
	store.Set(key, sdk.Uint64ToBigEndian(positionId))

	// Set the pool ID to position ID mapping.
	key = types.KeyPoolPositionPositionId(poolId, positionId)
	store.Set(key, sdk.Uint64ToBigEndian(positionId))
}

func (k Keeper) deletePosition(ctx sdk.Context,
	positionId uint64,
	owner sdk.AccAddress,
	poolId uint64,
) error {
	store := ctx.KVStore(k.storeKey)

	// Remove the position ID to position mapping.
	key := types.KeyPositionId(positionId)
	if !store.Has(key) {
		return types.PositionIdNotFoundError{PositionId: positionId}
	}
	store.Delete(key)

	// Remove the address-pool-position ID to position mapping.
	key = types.KeyAddressPoolIdPositionId(owner, poolId, positionId)
	if !store.Has(key) {
		return fmt.Errorf("position id %d not found for address %s and pool id %d", positionId, owner.String(), poolId)
	}
	store.Delete(key)

	// Remove the pool ID to position ID mapping.
	key = types.KeyPoolPositionPositionId(poolId, positionId)
	if !store.Has(key) {
		return fmt.Errorf("position id %d not found for pool id %d", positionId, poolId)
	}
	store.Delete(key)

	return nil
}

// CreateFullRangePosition creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, coins, and frozen until time.
// The function returns the amounts of token 0 and token 1, and the liquidity created from the position.
func (k Keeper) CreateFullRangePosition(ctx sdk.Context, concentratedPool types.ConcentratedPoolExtension, owner sdk.AccAddress, coins sdk.Coins, freezeDuration time.Duration) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, err error) {
	// Determine the max and min ticks for the concentrated pool we are migrating to.
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(concentratedPool.GetPrecisionFactorAtPriceOne())

	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, err = k.createPosition(ctx, concentratedPool.GetId(), owner, coins.AmountOf(concentratedPool.GetToken0()), coins.AmountOf(concentratedPool.GetToken1()), sdk.ZeroInt(), sdk.ZeroInt(), minTick, maxTick, freezeDuration)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, err
	}

	return positionId, amount0, amount1, liquidity, joinTime, nil
}

func CalculateUnderlyingAssetsFromPosition(ctx sdk.Context, position model.Position, pool types.ConcentratedPoolExtension) (sdk.Dec, sdk.Dec, error) {
	// Transform the provided ticks into their corresponding sqrtPrices.
	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(position.LowerTick, position.UpperTick, pool.GetPrecisionFactorAtPriceOne())
	if err != nil {
		return sdk.Dec{}, sdk.Dec{}, err
	}

	// Calculate the amount of underlying assets in the position
	asset0, asset1 := pool.CalcActualAmounts(ctx, position.LowerTick, position.UpperTick, sqrtPriceLowerTick, sqrtPriceUpperTick, position.Liquidity)
	return asset0, asset1, nil
}

// getNextPositionIdAndIncrement returns the next position Id, and increments the corresponding state entry.
func (k Keeper) getNextPositionIdAndIncrement(ctx sdk.Context) uint64 {
	nextPositionId := k.GetNextPositionId(ctx)
	k.SetNextPositionId(ctx, nextPositionId+1)
	return nextPositionId
}
