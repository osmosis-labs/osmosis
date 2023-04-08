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
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
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
	positionId, underlyingLockId uint64,
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

	k.SetPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, liquidity, positionId, underlyingLockId)
	return nil
}

func (k Keeper) hasFullPosition(ctx sdk.Context, positionId uint64) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPositionId(positionId)
	return store.Has(key)
}

// hasAnyPositionForPool returns true if there is at least one position
// existing for a given pool. False otherwise. Returns false and error
// on any database error.
func (k Keeper) hasAnyPositionForPool(ctx sdk.Context, poolId uint64) (bool, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.KeyPoolPosition(poolId)
	parse := func(bz []byte) (uint64, error) {
		return sdk.BigEndianToUint64(bz), nil
	}
	return osmoutils.HasAnyAtPrefix(store, key, parse)
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

// SetPosition sets the position information for a given user in a given pool.
func (k Keeper) SetPosition(ctx sdk.Context,
	poolId uint64,
	owner sdk.AccAddress,
	lowerTick, upperTick int64,
	joinTime time.Time,
	liquidity sdk.Dec,
	positionId uint64,
	underlyingLockId uint64,
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

	// Set the position ID to underlying lock ID mapping if underlyingLockId is provided.
	key = types.KeyPositionIdForLock(positionId)
	_, err := k.GetPositionIdToLock(ctx, positionId)
	if err != nil && underlyingLockId != 0 {
		// We did not find an underlying lock ID, but one was provided. Set it.
		store.Set(key, sdk.Uint64ToBigEndian(underlyingLockId))
	} else if err == nil && underlyingLockId == 0 {
		// We found an underlying lock ID, but none was provided. Delete it.
		store.Delete(key)
	}

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

// CreateFullRangePosition creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, and coins.
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

// CreateFullRangePositionUnlocking creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, and coins.
// This function is strictly used when migrating a balancer position to CL, where the balancer position is locked until a certain time.
// We lock the cl position for whatever the remaining time is from the balancer position and immediately begin unlocking it.
func (k Keeper) CreateFullRangePositionUnlocking(ctx sdk.Context, concentratedPool types.ConcentratedPoolExtension, owner sdk.AccAddress, coins sdk.Coins, remainingLockDuration time.Duration) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, concentratedLockID uint64, err error) {
	// Determine the min and max ticks for the concentrated pool we are migrating to.
	minTick, maxTick := GetMinAndMaxTicksFromExponentAtPriceOne(concentratedPool.GetExponentAtPriceOne())

	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, err = k.createPosition(ctx, concentratedPool.GetId(), owner, coins.AmountOf(concentratedPool.GetToken0()), coins.AmountOf(concentratedPool.GetToken1()), sdk.ZeroInt(), sdk.ZeroInt(), minTick, maxTick)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	// Create a coin object to represent the underlying liquidity for the cl position.
	underlyingLiquidityTokenized := sdk.NewCoins(sdk.NewCoin(fmt.Sprintf("%s/%d/%d", types.ClTokenPrefix, concentratedPool.GetId(), positionId), liquidity.TruncateInt()))

	// Mint the underlying liquidity as a token and send to the owner.
	err = k.bankKeeper.MintCoins(ctx, lockuptypes.ModuleName, underlyingLiquidityTokenized)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, lockuptypes.ModuleName, owner, underlyingLiquidityTokenized)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	// Lock the position for the specified duration.
	// Note, the endblocker for the lockup module contains an exception for this CL denom. When a lock with a denom of cl/pool/{poolId}/{positionId} is mature,
	// it does not send the coins to the owner account and instead burns them. This is strictly to use well tested pre-existing methods rather than potentially introducing bugs with more new logic and methods.
	concentratedLock, err := k.lockupKeeper.CreateLock(ctx, owner, underlyingLiquidityTokenized, remainingLockDuration)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	// Update the position to have the lock ID
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}
	positionAddress, err := sdk.AccAddressFromBech32(position.Address)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}
	_, _, err = k.updatePosition(ctx, position.PoolId, positionAddress, position.LowerTick, position.UpperTick, position.Liquidity, position.JoinTime, position.PositionId, concentratedLock.ID)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	// Begin unlocking the lock
	concentratedLockID, err = k.lockupKeeper.BeginForceUnlock(ctx, concentratedLock.ID, underlyingLiquidityTokenized)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	return positionId, amount0, amount1, liquidity, joinTime, concentratedLockID, nil
}

func CalculateUnderlyingAssetsFromPosition(ctx sdk.Context, position model.Position, pool types.ConcentratedPoolExtension) (sdk.Coin, sdk.Coin, error) {
	// Transform the provided ticks into their corresponding sqrtPrices.
	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(position.LowerTick, position.UpperTick, pool.GetExponentAtPriceOne())
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Calculate the amount of underlying assets in the position
	asset0, asset1 := pool.CalcActualAmounts(ctx, position.LowerTick, position.UpperTick, sqrtPriceLowerTick, sqrtPriceUpperTick, position.Liquidity)

	// Create coin objects from the underlying assets.
	coin0 := sdk.NewCoin(pool.GetToken0(), asset0.TruncateInt())
	coin1 := sdk.NewCoin(pool.GetToken1(), asset1.TruncateInt())

	return coin0, coin1, nil
}

// getNextPositionIdAndIncrement returns the next position Id, and increments the corresponding state entry.
func (k Keeper) getNextPositionIdAndIncrement(ctx sdk.Context) uint64 {
	nextPositionId := k.GetNextPositionId(ctx)
	k.SetNextPositionId(ctx, nextPositionId+1)
	return nextPositionId
}

// fungifyChargedPosition takes in a list of positionIds and combines them into a single position.
// The old position's unclaimed rewards are transferred to the new position.
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

	fullyChargedDuration := types.SupportedUptimes[len(types.SupportedUptimes)-1]

	// The new position's timestamp is the current block time minus the fully charged duration.
	joinTime := ctx.BlockTime().Add(-fullyChargedDuration)

	// Get the next position ID and increment the global counter.
	newPositionId := k.getNextPositionIdAndIncrement(ctx)

	// Initialize the fee accumulator for the new position.
	if err := k.initializeFeeAccumulatorPosition(ctx, poolId, lowerTick, upperTick, newPositionId); err != nil {
		return 0, err
	}

	// Check if the position already exists.
	hasFullPosition := k.hasFullPosition(ctx, newPositionId)
	if !hasFullPosition {
		// If the position does not exist, initialize it with the provided liquidity and tick range.
		err = k.initOrUpdatePositionUptime(ctx, poolId, liquidity, owner, lowerTick, upperTick, sdk.ZeroDec(), joinTime, newPositionId)
		if err != nil {
			return 0, err
		}
	} else {
		// If the position already exists, return an error.
		return 0, err
	}

	// Update the position in the pool based on the provided tick range and liquidity delta.
	// We hardcode zero for the underlying lock ID here, since we verified that all positions have no underlying lock.
	_, _, err = k.updatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidity, joinTime, newPositionId, 0)
	if err != nil {
		return 0, err
	}

	// Get the new position
	newPosition, err := k.GetPosition(ctx, newPositionId)
	if err != nil {
		return 0, err
	}

	// Get the new position's store name as well as uptime accumulators for the pool.
	newPositionName := string(types.KeyPositionId(newPositionId))
	uptimeAccumulators, err := k.getUptimeAccumulators(ctx, newPosition.PoolId)
	if err != nil {
		return 0, err
	}

	// Move unclaimed rewards from the old positions to the new position.
	// Also, delete the old positions from state.

	// Compute uptime growth outside of the range between lower tick and upper tick
	uptimeGrowthOutside, err := k.GetUptimeGrowthOutsideRange(ctx, newPosition.PoolId, newPosition.LowerTick, newPosition.UpperTick)
	if err != nil {
		return 0, err
	}

	// Move unclaimed rewards from the old positions to the new position.
	// Also, delete the old positions from state.

	// Loop through each position ID.
	for _, positionId := range positionIds {
		// Loop through each uptime accumulator for the pool.
		for uptimeIndex, uptimeAccum := range uptimeAccumulators {
			oldPositionName := string(types.KeyPositionId(positionId))
			// Check if the accumulator contains the position.
			hasPosition, err := uptimeAccum.HasPosition(oldPositionName)
			if err != nil {
				return 0, err
			}
			// If the accumulator contains the position, move the unclaimed rewards to the new position.
			if hasPosition {
				// Prepare the accumulator for the old position.
				rewards, dust, err := prepareAccumAndClaimRewards(uptimeAccum, oldPositionName, uptimeGrowthOutside[uptimeIndex])
				if err != nil {
					return 0, err
				}
				unclaimedRewardsForPosition := sdk.NewDecCoinsFromCoins(rewards...).Add(dust...)

				// Add the unclaimed rewards to the new position.
				err = uptimeAccum.AddToUnclaimedRewards(newPositionName, unclaimedRewardsForPosition)
				if err != nil {
					return 0, err
				}

				// Delete the accumulator position from state.
				uptimeAccum.DeletePosition(oldPositionName)
			}
		}
		// Remove the old cl position from state.
		err = k.deletePosition(ctx, positionId, owner, poolId)
		if err != nil {
			return 0, err
		}
	}

	return newPositionId, nil
}

// validatePositionsAndGetTotalLiquidity checks that the positions are all in the same pool and tick range, and returns the total liquidity of the positions.
func (k Keeper) validatePositionsAndGetTotalLiquidity(ctx sdk.Context, owner sdk.AccAddress, positionIds []uint64) (uint64, int64, int64, sdk.Dec, error) {
	totalLiquidity := sdk.ZeroDec()
	// Note the first position's params to use as the base for comparison.
	basePosition, err := k.GetPosition(ctx, positionIds[0])
	if err != nil {
		return 0, 0, 0, sdk.Dec{}, err
	}

	fullyChargedDuration := types.SupportedUptimes[len(types.SupportedUptimes)-1]

	for i, positionId := range positionIds {
		position, err := k.GetPosition(ctx, positionId)
		if err != nil {
			return 0, 0, 0, sdk.Dec{}, err
		}
		// Check that the caller owns all the positions.
		if position.Address != owner.String() {
			return 0, 0, 0, sdk.Dec{}, types.PositionOwnerMismatchError{PositionOwner: position.Address, Sender: owner.String()}
		}

		// Check that all the positions have no underlying lock that has not yet matured.
		underlyingLockId, _ := k.GetPositionIdToLock(ctx, positionId)
		if underlyingLockId != 0 {
			position, err = k.validateIsNotLockedAndUpdate(ctx, position, underlyingLockId)
			if err != nil {
				return 0, 0, 0, sdk.Dec{}, err
			}
		}

		// Check that all the positions are fully charged.
		fullyChargedMinTimestamp := position.JoinTime.Add(fullyChargedDuration)
		if !fullyChargedMinTimestamp.Before(ctx.BlockTime()) {
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

// GetConcentratedLockupDenom returns the concentrated lockup denom for a given pool and position.
func GetConcentratedLockupDenom(poolId, positionId uint64) string {
	return fmt.Sprintf("cl/pool/%d/%d", poolId, positionId)
}

// GetPositionIdToLock returns the positionId to lock mapping in state.
func (k Keeper) GetPositionIdToLock(ctx sdk.Context, positionId uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)

	// Get the position ID to key mapping.
	key := types.KeyPositionIdForLock(positionId)
	value := store.Get(key)
	if value == nil {
		return 0, types.PositionIdToLockNotFoundError{PositionId: positionId}
	}

	return sdk.BigEndianToUint64(value), nil
}
