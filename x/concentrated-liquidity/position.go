package concentrated_liquidity

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmoutils"
	"github.com/osmosis-labs/osmosis/osmoutils/accum"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/model"
	"github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
)

const MinNumPositions = 2

var emptyOptions = &accum.Options{}

// getOrInitPosition retrieves the position's liquidity for the given tick range.
// If it doesn't exist, it returns zero.
func (k Keeper) getOrInitPosition(
	ctx sdk.Context,
	positionId uint64,
) (sdk.Dec, error) {
	if k.hasPosition(ctx, positionId) {
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
// Errors if given liquidityDelta is negative and exceeds the amount of liquidity in the position.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
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

	err = k.initOrUpdatePositionUptimeAccumulators(ctx, poolId, liquidity, owner, lowerTick, upperTick, liquidityDelta, positionId)
	if err != nil {
		return err
	}

	err = k.SetPosition(ctx, poolId, owner, lowerTick, upperTick, joinTime, liquidity, positionId, noUnderlyingLockId)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) hasPosition(ctx sdk.Context, positionId uint64) bool {
	store := ctx.KVStore(k.storeKey)
	positionIdKey := types.KeyPositionId(positionId)
	return store.Has(positionIdKey)
}

// HasAnyPositionForPool returns true if there is at least one position
// existing for a given pool. False otherwise. Returns false and error
// on any database error.
func (k Keeper) HasAnyPositionForPool(ctx sdk.Context, poolId uint64) (bool, error) {
	store := ctx.KVStore(k.storeKey)
	poolPositionKey := types.KeyPoolPosition(poolId)
	parse := func(bz []byte) (uint64, error) {
		return sdk.BigEndianToUint64(bz), nil
	}
	return osmoutils.HasAnyAtPrefix(store, poolPositionKey, parse)
}

// isPositionOwner returns true if the given positionId is owned by the given sender inside the given pool.
func (k Keeper) isPositionOwner(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, positionId uint64) (bool, error) {
	parse := func(bz []byte) (uint64, error) {
		return sdk.BigEndianToUint64(bz), nil
	}
	isOwner, err := osmoutils.HasAnyAtPrefix(ctx.KVStore(k.storeKey), types.KeyAddressPoolIdPositionId(sender, poolId, positionId), parse)
	if err != nil {
		return false, err
	}

	return isOwner, nil
}

// GetAllPositionsForPoolId gets all the position for a specific poolId.
func (k Keeper) GetAllPositionIdsForPoolId(ctx sdk.Context, poolId uint64) ([]uint64, error) {
	parse := func(bz []byte) (uint64, error) {
		return sdk.BigEndianToUint64(bz), nil
	}

	return osmoutils.GatherValuesFromStorePrefix(ctx.KVStore(k.storeKey), types.KeyPoolPosition(poolId), parse)
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
	positionIdKey := types.KeyPositionId(positionId)

	positionStruct := &model.Position{}
	found, err := osmoutils.Get(store, positionIdKey, positionStruct)
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
) error {
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

	// TODO: The following state mappings are not properly implemented in genState.
	// (i.e. if you state export, these mappings are not retained.)
	// https://github.com/osmosis-labs/osmosis/issues/4875

	// Set the position ID to position mapping.
	positionIdKey := types.KeyPositionId(positionId)
	osmoutils.MustSet(store, positionIdKey, &position)

	// Set the address-pool-position ID to position mapping.
	addressPoolIdPositionIdKey := types.KeyAddressPoolIdPositionId(owner, poolId, positionId)
	store.Set(addressPoolIdPositionIdKey, sdk.Uint64ToBigEndian(positionId))

	// Set the pool ID to position ID mapping.
	poolIdKey := types.KeyPoolPositionPositionId(poolId, positionId)
	store.Set(poolIdKey, sdk.Uint64ToBigEndian(positionId))

	// Set the position ID to underlying lock ID mapping if underlyingLockId is provided.
	positionHasUnderlyingLock, _, err := k.positionHasActiveUnderlyingLockAndUpdate(ctx, positionId)
	if err != nil {
		return err
	}
	if !positionHasUnderlyingLock && underlyingLockId != 0 {
		// We did not find an underlying lock ID, but one was provided. Set it.
		k.setPositionIdToLock(ctx, positionId, underlyingLockId)
	}

	// If position is full range, update the pool ID to total full range liquidity mapping.
	if lowerTick == types.MinTick && upperTick == types.MaxTick {
		err := k.updateFullRangeLiquidityInPool(ctx, poolId, liquidity)
		if err != nil {
			return err
		}
	}

	return nil
}

// deletePosition deletes the position information for a given position id, user, and pool.
// Besides deleting the position, it also deletes the following mappings:
// - owner-pool-id-position-id to position id
// - pool-id-position-id to position id
// - position-id to underlying lock id if such mapping exists
// Returns error if:
// - the position with the given id does not exist.
// - the owner-pool-id-position-id to position id mapping does not exist.
// - the pool-id-position-id to position id mapping does not exist.
func (k Keeper) deletePosition(ctx sdk.Context,
	positionId uint64,
	owner sdk.AccAddress,
	poolId uint64,
) error {
	store := ctx.KVStore(k.storeKey)

	// Remove the position ID to position mapping.
	positionIdKey := types.KeyPositionId(positionId)
	if !store.Has(positionIdKey) {
		return types.PositionIdNotFoundError{PositionId: positionId}
	}
	store.Delete(positionIdKey)

	// Remove the address-pool-position ID to position mapping.
	addressPoolIdPositionIdKey := types.KeyAddressPoolIdPositionId(owner, poolId, positionId)
	if !store.Has(addressPoolIdPositionIdKey) {
		return types.AddressPoolPositionIdNotFoundError{Owner: owner.String(), PoolId: poolId, PositionId: positionId}
	}
	store.Delete(addressPoolIdPositionIdKey)

	// Remove the pool ID to position ID mapping.
	poolIdKey := types.KeyPoolPositionPositionId(poolId, positionId)
	if !store.Has(poolIdKey) {
		return types.PoolPositionIdNotFoundError{PoolId: poolId, PositionId: positionId}
	}
	store.Delete(poolIdKey)

	// Remove the position ID to underlying lock ID mapping (if it exists)
	positionIdLockKey := types.KeyPositionIdForLock(positionId)
	if store.Has(positionIdLockKey) {
		underlyingLockId, err := k.GetLockIdFromPositionId(ctx, positionId)
		if err != nil {
			return err
		}
		store.Delete(positionIdLockKey)
		lockIdPositionKey := types.KeyLockIdForPositionId(underlyingLockId)
		store.Delete(lockIdPositionKey)
	}

	return nil
}

// CreateFullRangePosition creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, and coins.
// The function returns the amounts of token 0 and token 1, and the liquidity created from the position.
func (k Keeper) CreateFullRangePosition(ctx sdk.Context, clPoolId uint64, owner sdk.AccAddress, coins sdk.Coins) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, err error) {
	// Check that exactly two coins are provided.
	if len(coins) != 2 {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, types.NumCoinsError{NumCoins: len(coins)}
	}

	concentratedPool, err := k.GetPoolFromPoolIdAndConvertToConcentrated(ctx, clPoolId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, err
	}

	// Defense in depth, ensure coins provided match the pool's token denominations.
	if coins.AmountOf(concentratedPool.GetToken0()).LTE(sdk.ZeroInt()) {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, types.Amount0IsNegativeError{Amount0: coins.AmountOf(concentratedPool.GetToken0())}
	}
	if coins.AmountOf(concentratedPool.GetToken1()).LTE(sdk.ZeroInt()) {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, types.Amount1IsNegativeError{Amount1: coins.AmountOf(concentratedPool.GetToken1())}
	}
	if len(coins) != 2 {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, types.NumCoinsError{NumCoins: len(coins)}
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, err = k.createPosition(ctx, concentratedPool.GetId(), owner, coins, sdk.ZeroInt(), sdk.ZeroInt(), types.MinTick, types.MaxTick)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, err
	}

	return positionId, amount0, amount1, liquidity, joinTime, nil
}

// CreateFullRangePositionLocked creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, and coins.
// CL shares are minted which represent the underlying liquidity and are locked for the given duration.
// State entries are also created to map the position ID to the underlying lock ID.
func (k Keeper) CreateFullRangePositionLocked(ctx sdk.Context, clPoolId uint64, owner sdk.AccAddress, coins sdk.Coins, remainingLockDuration time.Duration) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, concentratedLockID uint64, err error) {
	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, err = k.CreateFullRangePosition(ctx, clPoolId, owner, coins)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	// Mint cl shares for the position and lock them for the remaining lock duration.
	// Also sets the position ID to underlying lock ID mapping.
	concentratedLockId, _, err := k.mintSharesLockAndUpdate(ctx, clPoolId, positionId, owner, remainingLockDuration)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	return positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, nil
}

// CreateFullRangePositionUnlocking creates a full range (min to max tick) concentrated liquidity position for the given pool ID, owner, and coins.
// This function is strictly used when migrating a balancer position to CL, where the balancer position is currently unlocking.
// We lock the cl position for whatever the remaining time is from the balancer position and immediately begin unlocking from where it left off.
func (k Keeper) CreateFullRangePositionUnlocking(ctx sdk.Context, clPoolId uint64, owner sdk.AccAddress, coins sdk.Coins, remainingLockDuration time.Duration) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, concentratedLockID uint64, err error) {
	// Create a full range (min to max tick) concentrated liquidity position.
	positionId, amount0, amount1, liquidity, joinTime, err = k.CreateFullRangePosition(ctx, clPoolId, owner, coins)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	// Mint cl shares for the position and lock them for the remaining lock duration.
	// Also sets the position ID to underlying lock ID mapping.
	concentratedLockId, underlyingLiquidityTokenized, err := k.mintSharesLockAndUpdate(ctx, clPoolId, positionId, owner, remainingLockDuration)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	// Begin unlocking the newly created concentrated lock.
	concentratedLockID, err = k.lockupKeeper.BeginForceUnlock(ctx, concentratedLockId, underlyingLiquidityTokenized)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, err
	}

	return positionId, amount0, amount1, liquidity, joinTime, concentratedLockID, nil
}

// mintSharesLockAndUpdate mints the shares for the full range concentrated liquidity position and locks them for the given duration. It also updates the position ID to underlying lock ID mapping.
// In the context of concentrated liquidity, shares need to be minted in order for a lock in its current form to be utilized (we cannot lock non-coin objects).
// In turn, the locks are a prerequisite for superfluid to be enabled.
// Additionally, the cl share gets sent to the lockup module account, which, in order to be sent via bank, must be minted.
func (k Keeper) mintSharesLockAndUpdate(ctx sdk.Context, concentratedPoolId, positionId uint64, owner sdk.AccAddress, remainingLockDuration time.Duration) (concentratedLockID uint64, underlyingLiquidityTokenized sdk.Coins, err error) {
	// Ensure the provided position is full range.
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return 0, sdk.Coins{}, err
	}
	if position.LowerTick != types.MinTick || position.UpperTick != types.MaxTick {
		return 0, sdk.Coins{}, types.PositionNotFullRangeError{PositionId: positionId, LowerTick: position.LowerTick, UpperTick: position.UpperTick}
	}

	// Create a coin object to represent the underlying liquidity for the cl position.
	underlyingLiquidityTokenized = sdk.NewCoins(sdk.NewCoin(types.GetConcentratedLockupDenomFromPoolId(concentratedPoolId), position.Liquidity.TruncateInt()))

	// Mint the underlying liquidity as a token and send to the owner.
	err = k.bankKeeper.MintCoins(ctx, lockuptypes.ModuleName, underlyingLiquidityTokenized)
	if err != nil {
		return 0, sdk.Coins{}, err
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, lockuptypes.ModuleName, owner, underlyingLiquidityTokenized)
	if err != nil {
		return 0, sdk.Coins{}, err
	}

	// Lock the position for the specified duration.
	// Note, the end blocker for the lockup module contains an exception for this CL denom. When a lock with a denom of cl/pool/{poolId} is mature,
	// it does not send the coins to the owner account and instead burns them. This is strictly to use well tested pre-existing methods rather than potentially introducing bugs with new logic and methods.
	concentratedLock, err := k.lockupKeeper.CreateLock(ctx, owner, underlyingLiquidityTokenized, remainingLockDuration)
	if err != nil {
		return 0, sdk.Coins{}, err
	}

	// Set the position ID to underlying lock ID mapping.
	k.setPositionIdToLock(ctx, positionId, concentratedLock.ID)

	return concentratedLock.ID, underlyingLiquidityTokenized, nil
}

func CalculateUnderlyingAssetsFromPosition(ctx sdk.Context, position model.Position, pool types.ConcentratedPoolExtension) (sdk.Coin, sdk.Coin, error) {
	token0 := pool.GetToken0()
	token1 := pool.GetToken1()

	if position.Liquidity.IsZero() {
		return sdk.NewCoin(token0, sdk.ZeroInt()), sdk.NewCoin(token1, sdk.ZeroInt()), nil
	}

	// Calculate the amount of underlying assets in the position
	asset0, asset1, err := pool.CalcActualAmounts(ctx, position.LowerTick, position.UpperTick, position.Liquidity)
	if err != nil {
		return sdk.Coin{}, sdk.Coin{}, err
	}

	// Create coin objects from the underlying assets.
	coin0 := sdk.NewCoin(token0, asset0.TruncateInt())
	coin1 := sdk.NewCoin(token1, asset1.TruncateInt())

	return coin0, coin1, nil
}

// getNextPositionIdAndIncrement returns the next position Id, and increments the corresponding state entry.
func (k Keeper) getNextPositionIdAndIncrement(ctx sdk.Context) uint64 {
	nextPositionId := k.GetNextPositionId(ctx)
	k.SetNextPositionId(ctx, nextPositionId+1)
	return nextPositionId
}

// fungifyChargedPosition takes in a list of positionIds and combines them into a single position.
// It validates that all positions belong to the same owner, are in the same ticks, in the same pool and fully charged. Fails if not.
// Otherwise, it creates a completely new position P. P's liquidity equals to the sum of all
// liquidities of positions given by positionIds. The uptime of the join time of the new position equals
// to current block time - max authorized uptime duration (to signify that it is fully charged).
// The previous positions are deleted from state. Prior to deleting, the rewards are claimed.
// The old position's unclaimed rewards are transferred to the new position.
// The new position ID is returned.
// An error is returned if:
// - the caller does not own all the positions
// - positions are not in the same pool
// - positions are all not fully charged
// - positions are not in the same tick range
// - all positions are unlocked
func (k Keeper) fungifyChargedPosition(ctx sdk.Context, owner sdk.AccAddress, positionIds []uint64) (uint64, error) {
	// Check we meet the minimum number of positions to combine.
	if len(positionIds) < MinNumPositions {
		return 0, types.PositionQuantityTooLowError{MinNumPositions: MinNumPositions, NumPositions: len(positionIds)}
	}

	// Check that all the positions are in the same pool, tick range, and are fully charged.
	// Sum the liquidity of all the positions.
	poolId, lowerTick, upperTick, liquidity, err := k.validatePositionsAndGetTotalLiquidity(ctx, owner, positionIds)
	if err != nil {
		return 0, err
	}

	// Get the fully charged duration for the pool.
	fullyChargedDuration := k.getLargestAuthorizedUptimeDuration(ctx)

	// The new position's timestamp is the current block time minus the fully charged duration.
	joinTime := ctx.BlockTime().Add(-fullyChargedDuration)

	// Get the next position ID and increment the global counter.
	newPositionId := k.getNextPositionIdAndIncrement(ctx)

	// Update pool uptime acccumulators to now
	if err := k.updatePoolUptimeAccumulatorsToNow(ctx, poolId); err != nil {
		return 0, err
	}

	// Update the position in the pool based on the provided tick range and liquidity delta.
	// This also initializes the fee accumulator and the uptime accumulators for the new position.
	_, _, err = k.UpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidity, joinTime, newPositionId)
	if err != nil {
		return 0, err
	}

	// Get the new position's name in the pool's uptime accumulators.
	newPositionUptimeAccName := string(types.KeyPositionId(newPositionId))
	uptimeAccumulators, err := k.GetUptimeAccumulators(ctx, poolId)
	if err != nil {
		return 0, err
	}

	// Get the new position's name in the pool's fee accumulator.
	newPositionFeeAccName := types.KeyFeePositionAccumulator(newPositionId)
	feeAccumulator, err := k.GetFeeAccumulator(ctx, poolId)
	if err != nil {
		return 0, err
	}

	// Compute uptime growth outside of the range between lower tick and upper tick
	uptimeGrowthOutside, err := k.GetUptimeGrowthOutsideRange(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return 0, err
	}

	// Compute the fee growth outside of the range between lower tick and upper tick
	feeGrowthOutside, err := k.getFeeGrowthOutside(ctx, poolId, lowerTick, upperTick)
	if err != nil {
		return 0, err
	}

	// Move unclaimed rewards from the old positions to the new position.
	// Also, delete the old positions from state.

	// Loop through each position ID.
	for _, oldPositionId := range positionIds {
		// Loop through each uptime accumulator for the pool.
		for uptimeIndex, uptimeAccum := range uptimeAccumulators {
			oldPositionName := string(types.KeyPositionId(oldPositionId))
			// Check if the accumulator contains the position.
			hasPosition, err := uptimeAccum.HasPosition(oldPositionName)
			if err != nil {
				return 0, err
			}
			if !hasPosition {
				return 0, types.PositionIdNotFoundError{PositionId: oldPositionId}
			}
			// If the accumulator contains the position, move the unclaimed rewards to the new position.
			if err := moveRewardsToNewPositionAndDeleteOldAcc(ctx, uptimeAccum, oldPositionName, newPositionUptimeAccName, uptimeGrowthOutside[uptimeIndex]); err != nil {
				return 0, err
			}
		}

		// Move fees into the new fee accumulator and delete the old accumulator.
		oldPositionFeeName := types.KeyFeePositionAccumulator(oldPositionId)
		if err := moveRewardsToNewPositionAndDeleteOldAcc(ctx, feeAccumulator, oldPositionFeeName, newPositionFeeAccName, feeGrowthOutside); err != nil {
			return 0, err
		}

		// Remove the old cl position from state.
		err = k.deletePosition(ctx, oldPositionId, owner, poolId)
		if err != nil {
			return 0, err
		}
	}

	// query claimable incentives for events.
	claimableIncentives, _, err := k.GetClaimableIncentives(ctx, newPositionId)
	if err != nil {
		return 0, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtFungifyChargedPosition,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, owner.String()),
			sdk.NewAttribute(types.AttributeInputPositionIds, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(positionIds)), ","), "[]")),
			sdk.NewAttribute(types.AttributeOutputPositionId, strconv.FormatUint(newPositionId, 10)),
			sdk.NewAttribute(types.AttributeClaimableIncentives, claimableIncentives.String()),
		),
	})

	return newPositionId, nil
}

// validatePositionsAndGetTotalLiquidity validates a list of positions owned by the caller and returns their total liquidity.
// It also returns the pool ID, lower tick, and upper tick that all the provided positions are confirmed to share.
// Returns error if:
// - the caller does not own all the positions
// - positions are not in the same pool
// - positions are all not fully charged
// - positions are not in the same tick range
// - all positions are unlocked
func (k Keeper) validatePositionsAndGetTotalLiquidity(ctx sdk.Context, owner sdk.AccAddress, positionIds []uint64) (uint64, int64, int64, sdk.Dec, error) {
	totalLiquidity := sdk.ZeroDec()

	if len(positionIds) < 1 {
		return 0, 0, 0, sdk.Dec{}, types.PositionQuantityTooLowError{MinNumPositions: MinNumPositions, NumPositions: 1}
	}

	// Note the first position's params to use as the base for comparison.
	basePosition, err := k.GetPosition(ctx, positionIds[0])
	if err != nil {
		return 0, 0, 0, sdk.Dec{}, err
	}

	fullyChargedDuration := k.getLargestAuthorizedUptimeDuration(ctx)

	for _, positionId := range positionIds {
		position, err := k.GetPosition(ctx, positionId)
		if err != nil {
			return 0, 0, 0, sdk.Dec{}, err
		}
		// Check that the caller owns all the positions.
		if position.Address != owner.String() {
			return 0, 0, 0, sdk.Dec{}, types.PositionOwnerMismatchError{PositionOwner: position.Address, Sender: owner.String()}
		}

		// Check that all the positions have no underlying lock that has not yet matured.
		// Note, this is calling a non mutative method, so if the position's lock is no longer mature,
		// it will return an error but the connection will still be persisted.
		positionHasActiveUnderlyingLock, lockId, err := k.PositionHasActiveUnderlyingLock(ctx, positionId)
		if err != nil {
			return 0, 0, 0, sdk.Dec{}, err
		}
		if positionHasActiveUnderlyingLock {
			// Lock is not mature, return error.
			return 0, 0, 0, sdk.Dec{}, types.LockNotMatureError{PositionId: position.PositionId, LockId: lockId}
		}

		// Check that all the positions are fully charged.
		fullyChargedMinTimestamp := position.JoinTime.Add(fullyChargedDuration)
		if fullyChargedMinTimestamp.After(ctx.BlockTime()) {
			return 0, 0, 0, sdk.Dec{}, types.PositionNotFullyChargedError{PositionId: position.PositionId, PositionJoinTime: position.JoinTime, FullyChargedMinTimestamp: fullyChargedMinTimestamp}
		}

		// Check that all the positions are in the same pool and tick range.
		if position.PoolId != basePosition.PoolId {
			return 0, 0, 0, sdk.Dec{}, types.PositionsNotInSamePoolError{Position1PoolId: position.PoolId, Position2PoolId: basePosition.PoolId}
		}
		if position.LowerTick != basePosition.LowerTick || position.UpperTick != basePosition.UpperTick {
			return 0, 0, 0, sdk.Dec{}, types.PositionsNotInSameTickRangeError{Position1TickLower: position.LowerTick, Position1TickUpper: position.UpperTick, Position2TickLower: basePosition.LowerTick, Position2TickUpper: basePosition.UpperTick}
		}

		// Add the liquidity of the position to the total liquidity.
		totalLiquidity = totalLiquidity.Add(position.Liquidity)
	}
	return basePosition.PoolId, basePosition.LowerTick, basePosition.UpperTick, totalLiquidity, nil
}

// GetLockIdFromPositionId returns the lock id associated with the given position id.
func (k Keeper) GetLockIdFromPositionId(ctx sdk.Context, positionId uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)

	// Get the position ID to lock ID mapping.
	positionIdLockKey := types.KeyPositionIdForLock(positionId)
	value := store.Get(positionIdLockKey)
	if value == nil {
		return 0, types.PositionIdToLockNotFoundError{PositionId: positionId}
	}

	return sdk.BigEndianToUint64(value), nil
}

// GetPositionIdToLockId returns the position id associated with the given lock id.
func (k Keeper) GetPositionIdToLockId(ctx sdk.Context, underlyingLockId uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)

	// Get the lock ID to position ID mapping.
	positionIdLockKey := types.KeyLockIdForPositionId(underlyingLockId)
	value := store.Get(positionIdLockKey)
	if value == nil {
		return 0, types.LockIdToPositionIdNotFoundError{LockId: underlyingLockId}
	}

	return sdk.BigEndianToUint64(value), nil
}

// setPositionIdToLock sets both the positionId to lock mapping and the lock to positionId mapping in state.
func (k Keeper) setPositionIdToLock(ctx sdk.Context, positionId, underlyingLockId uint64) {
	store := ctx.KVStore(k.storeKey)

	// Get the position ID to key mappings and set them in state.
	positionIdLockKey, lockIdKey := types.PositionIdForLockIdKeys(positionId, underlyingLockId)
	store.Set(positionIdLockKey, sdk.Uint64ToBigEndian(underlyingLockId))
	store.Set(lockIdKey, sdk.Uint64ToBigEndian(positionId))
}

// RemovePositionIdToLock removes both the positionId to lock mapping and the lock to positionId mapping in state.
func (k Keeper) RemovePositionIdToLock(ctx sdk.Context, positionId, underlyingLockId uint64) {
	store := ctx.KVStore(k.storeKey)

	// Get the position ID to lock mappings.
	positionIdLockKey, lockIdKey := types.PositionIdForLockIdKeys(positionId, underlyingLockId)

	// Delete the mappings from state.
	store.Delete(positionIdLockKey)
	store.Delete(lockIdKey)
}

// PositionHasActiveUnderlyingLock is a non mutative method that checks if a given positionId has a corresponding lock in state.
// If it has a lock in state, checks if that lock is still active.
// If lock is still active, returns true.
// If lock is no longer active, returns false.
func (k Keeper) PositionHasActiveUnderlyingLock(ctx sdk.Context, positionId uint64) (hasActiveUnderlyingLock bool, lockId uint64, err error) {
	// Get the lock ID for the position.
	lockId, err = k.GetLockIdFromPositionId(ctx, positionId)
	if errors.Is(err, types.PositionIdToLockNotFoundError{PositionId: positionId}) {
		return false, 0, nil
	} else if err != nil {
		return false, 0, err
	}

	// Check if the underlying lock is mature.
	lockIsMature, err := k.isLockMature(ctx, lockId)
	if err != nil {
		return false, 0, err
	}
	if lockIsMature {
		return false, lockId, nil
	}
	return true, lockId, nil
}

// positionHasActiveUnderlyingLockAndUpdate is a mutative method that checks if a given positionId has a corresponding lock in state.
// If it has a lock in state, checks if that lock is still active.
// If lock is still active, returns true.
// If lock is no longer active, removes the lock ID from the position ID to lock ID mapping and returns false.
func (k Keeper) positionHasActiveUnderlyingLockAndUpdate(ctx sdk.Context, positionId uint64) (hasActiveUnderlyingLock bool, lockId uint64, err error) {
	hasActiveUnderlyingLock, lockId, err = k.PositionHasActiveUnderlyingLock(ctx, positionId)
	if err != nil {
		return false, 0, err
	}
	if hasActiveUnderlyingLock && lockId == 0 {
		// Defense in depth check. If we have an active underlying lock but no lock ID, return an error.
		return false, 0, types.PositionIdToLockNotFoundError{PositionId: positionId}
	}
	if !hasActiveUnderlyingLock && lockId != 0 {
		// If the position does not have an active underlying lock but still has a lock ID associated with it,
		// remove the link between the position and the underlying lock since the lock is mature.
		k.RemovePositionIdToLock(ctx, positionId, lockId)
		return false, 0, nil
	}
	return hasActiveUnderlyingLock, lockId, nil
}

// MustGetFullRangeLiquidityInPool returns the total liquidity that is currently in the full range of the pool.
func (k Keeper) MustGetFullRangeLiquidityInPool(ctx sdk.Context, poolId uint64) sdk.Dec {
	store := ctx.KVStore(k.storeKey)
	poolIdLiquidityKey := types.KeyPoolIdForLiquidity(poolId)
	currentTotalFullRangeLiquidity := osmoutils.MustGetDec(store, poolIdLiquidityKey)
	return currentTotalFullRangeLiquidity
}

// updateFullRangeLiquidityInPool updates the total liquidity store that is currently in the full range of the pool.
func (k Keeper) updateFullRangeLiquidityInPool(ctx sdk.Context, poolId uint64, liquidity sdk.Dec) error {
	store := ctx.KVStore(k.storeKey)
	// Get previous total liquidity.
	poolIdLiquidityKey := types.KeyPoolIdForLiquidity(poolId)
	currentTotalFullRangeLiquidityDecProto := sdk.DecProto{}
	found, err := osmoutils.Get(store, poolIdLiquidityKey, &currentTotalFullRangeLiquidityDecProto)
	if err != nil {
		return err
	}
	currentTotalFullRangeLiquidity := currentTotalFullRangeLiquidityDecProto.Dec
	// If position not found error, then we are creating the first full range liquidity position for a pool.
	if !found {
		currentTotalFullRangeLiquidity = sdk.ZeroDec()
	}

	// Add the liquidity of the new position to the total liquidity.
	newTotalFullRangeLiquidity := currentTotalFullRangeLiquidity.Add(liquidity)

	osmoutils.MustSetDec(store, poolIdLiquidityKey, newTotalFullRangeLiquidity)
	return nil
}
