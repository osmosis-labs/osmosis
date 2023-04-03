package concentrated_liquidity

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/internal/math"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	types "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
)

const MinNumPositionsToCombine = 2

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

	err = k.initOrUpdatePositionUptime(ctx, poolId, liquidity, owner, lowerTick, upperTick, liquidityDelta, joinTime, positionId)
	if err != nil {
		return err
	}

	k.setPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, liquidity, positionId)
	return nil
}

func (k Keeper) hasFullPosition(ctx sdk.Context, positionId uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPositionId(positionId)
	return store.Has(key)
}

// GetPositionLiquidity checks if the provided positionId exists. Returns position liquidity if found. Error otherwise.
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
	liquidity sdk.Dec,
	positionId uint64,
) {
	store := ctx.KVStore(k.storeKey)

	// Create a new Position object with the provided information.
	position := model.Position{
		PositionId: positionId,
		PoolId:     poolId,
		Address:    owner.String(),
		LowerTick:  lowerTick,
		UpperTick:  upperTick,
		JoinTime:   joinTime,
		Liquidity:  liquidity,
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
		return types.AddressPoolPositionIdNotFoundError{Owner: owner.String(), PoolId: poolId, PositionId: positionId}
	}
	store.Delete(key)

	// Remove the pool ID to position ID mapping.
	key = types.KeyPoolPositionPositionId(poolId, positionId)
	if !store.Has(key) {
		return types.PoolPositionIdNotFoundError{PoolId: poolId, PositionId: positionId}
	}
	store.Delete(key)

	return nil
}

// CreateFullRangePosition creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, coins, and frozen until time.
// The function returns the amounts of token 0 and token 1, and the liquidity created from the position.
func (k Keeper) CreateFullRangePosition(ctx sdk.Context, concentratedPool types.ConcentratedPoolExtension, owner sdk.AccAddress, coins sdk.Coins) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, err error) {
	// Determine the max and min ticks for the concentrated pool we are migrating to.
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(concentratedPool.GetExponentAtPriceOne())

	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, err = k.createPosition(ctx, concentratedPool.GetId(), owner, coins.AmountOf(concentratedPool.GetToken0()), coins.AmountOf(concentratedPool.GetToken1()), sdk.ZeroInt(), sdk.ZeroInt(), minTick, maxTick)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, err
	}

	return positionId, amount0, amount1, liquidity, joinTime, nil
}

func CalculateUnderlyingAssetsFromPosition(ctx sdk.Context, position model.Position, pool types.ConcentratedPoolExtension) (sdk.Dec, sdk.Dec, error) {
	// Transform the provided ticks into their corresponding sqrtPrices.
	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(position.LowerTick, position.UpperTick, pool.GetExponentAtPriceOne())
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

// fungifyChargedPosition takes in a list of positionIds and combines them into a single position.
// The previous positions are deleted from state and the new position ID is returned.
// An error is returned if the caller does not own all the positions, if the positions are all not fully charged, or if the positions are not all in the same pool / tick range.
func (k Keeper) fungifyChargedPosition(ctx sdk.Context, owner sdk.AccAddress, positionIds []uint64) (uint64, error) {
	// Check we meet the minimum number of positions to combine.
	if len(positionIds) <= MinNumPositionsToCombine {
		return 0, types.PositionQuantityTooLowError{MinNumPositions: MinNumPositionsToCombine, NumPositions: len(positionIds)}
	}

	// Check that all the positions are in the same pool, tick range, and are fully charged.
	// Sum the liquidity of all the positions.
	poolId, lowerTick, upperTick, liquidity, err := k.validatePositionsAndGetTotalLiquidity(ctx, owner, positionIds)
	if err != nil {
		return 0, err
	}

	// The new position's timestamp is the current block time minus the fully charged duration.
	joinTime := ctx.BlockTime().Add(-types.FullyChargedDuration)

	// Get the next position ID and increment the global counter.
	positionId := k.getNextPositionIdAndIncrement(ctx)

	// Initialize the fee accumulator for the new position.
	if err := k.initializeFeeAccumulatorPosition(ctx, poolId, lowerTick, upperTick, positionId); err != nil {
		return 0, err
	}

	// Check if the position already exists.
	hasFullPosition := k.hasFullPosition(ctx, positionId)
	if !hasFullPosition {
		// If the position does not exist, initialize it with the provided liquidity and tick range.
		err = k.initOrUpdatePositionUptime(ctx, poolId, liquidity, owner, lowerTick, upperTick, sdk.ZeroDec(), joinTime, positionId)
		if err != nil {
			return 0, err
		}
	} else {
		// If the position already exists, return an error.
		return 0, err
	}

	// Update the position in the pool based on the provided tick range and liquidity delta.
	_, _, err = k.updatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidity, joinTime, positionId)
	if err != nil {
		return 0, err
	}

	// Delete the previous positions.
	for _, positionId := range positionIds {
		err := k.deletePosition(ctx, positionId, owner, poolId)
		if err != nil {
			return 0, err
		}
	}

	return positionId, nil
}

// validatePositionsAndGetTotalLiquidity checks that the positions are all in the same pool and tick range, and returns the total liquidity of the positions.
func (k Keeper) validatePositionsAndGetTotalLiquidity(ctx sdk.Context, owner sdk.AccAddress, positionIds []uint64) (uint64, int64, int64, sdk.Dec, error) {
	totalLiquidity := sdk.ZeroDec()
	// Note the first position's params to use as the base for comparison.
	basePosition, err := k.GetPosition(ctx, positionIds[0])
	if err != nil {
		return 0, 0, 0, sdk.Dec{}, err
	}

	for i, positionId := range positionIds {
		position, err := k.GetPosition(ctx, positionId)
		if err != nil {
			return 0, 0, 0, sdk.Dec{}, err
		}
		// Check that the caller owns all the positions.
		if position.Address != owner.String() {
			return 0, 0, 0, sdk.Dec{}, types.PositionOwnerMismatchError{PositionOwner: position.Address, Sender: owner.String()}
		}

		// Check that all the positions are fully charged.
		fullyChargedMinTimestamp := position.JoinTime.Add(types.FullyChargedDuration)
		if fullyChargedMinTimestamp.After(ctx.BlockTime()) {
			return 0, 0, 0, sdk.Dec{}, types.PositionNotFullyChargedError{PositionId: position.PositionId, PositionJoinTime: position.JoinTime, FullyChargedMinTimestamp: fullyChargedMinTimestamp}
		}

		// Check that all the positions are in the same pool and tick range.
		if i > 0 {
			if position.PoolId != basePosition.PoolId {
				return 0, 0, 0, sdk.Dec{}, types.PositionsNotInSamePoolError{Position1PoolId: position.PoolId, Position2PoolId: basePosition.PoolId}
			}
			if position.LowerTick != basePosition.LowerTick || position.UpperTick != basePosition.UpperTick {
				return 0, 0, 0, sdk.Dec{}, types.PositionsNotInSameTickRangeError{Position1TickLower: position.LowerTick, Position1TickUpper: position.UpperTick, Position2TickLower: basePosition.LowerTick, Position2TickUpper: basePosition.UpperTick}
			}
		}

		// Add the liquidity of the position to the total liquidity.
		totalLiquidity = totalLiquidity.Add(position.Liquidity)
	}
	return basePosition.PoolId, basePosition.LowerTick, basePosition.UpperTick, totalLiquidity, nil
}
