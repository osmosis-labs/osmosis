package concentrated_liquidity

import (
	"errors"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	"github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/math"
	types "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v19/x/lockup/types"
)

const noUnderlyingLockId = uint64(0)

// CreatePositionData represents the return data from CreatePosition.
type CreatePositionData struct {
	ID        uint64
	Amount0   sdk.Int
	Amount1   sdk.Int
	Liquidity sdk.Dec
	LowerTick int64
	UpperTick int64
}

// createPosition creates a concentrated liquidity position in range between lowerTick and upperTick
// in a given poolId with the desired amount of each token. Since LPs are only allowed to provide
// liquidity proportional to the existing reserves, the actual amount of tokens used might differ from requested.
// As a result, LPs may also provide the minimum amount of each token to be used so that the system fails
// to create position if the desired amounts cannot be satisfied.
// For every initial position within a pool, it calls an AfterInitialPoolPositionCreated listener
// Currently, it creates TWAP records. Assuming that pool had all liquidity drained and then re-initialized,
// the TWAP records are updated with the valid spot price. This is needed because when there is no liquidity in pool,
// the spot price is undefined.
// On success, returns an actual amount of each token used and liquidity created.
// Returns error if:
// - the provided ticks are out of range / invalid
// - if one of the provided min amounts are negative
// - the pool provided does not exist
// - the liquidity delta is zero
// - the amount0 or amount1 returned from the position update is less than the given minimums
// - the pool or user does not have enough tokens to satisfy the requested amount
func (k Keeper) CreatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, tokensProvided sdk.Coins, amount0Min, amount1Min sdk.Int, lowerTick, upperTick int64) (positionDate CreatePositionData, err error) {
	// Use the current blockTime as the position's join time.
	joinTime := ctx.BlockTime()

	// Retrieve the pool associated with the given pool ID.
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return CreatePositionData{}, err
	}

	for _, token := range tokensProvided {
		if token.Denom != pool.GetToken0() && token.Denom != pool.GetToken1() {
			return CreatePositionData{}, errors.New("token provided is not one of the pool tokens")
		}
	}

	// Check if the provided tick range is valid according to the pool's tick spacing and module parameters.
	if err := validateTickRangeIsValid(pool.GetTickSpacing(), lowerTick, upperTick); err != nil {
		return CreatePositionData{}, err
	}
	amount0Desired := tokensProvided.AmountOf(pool.GetToken0())
	amount1Desired := tokensProvided.AmountOf(pool.GetToken1())
	if amount0Desired.IsZero() && amount1Desired.IsZero() {
		return CreatePositionData{}, errors.New("cannot create a position with zero amounts of both pool tokens")
	}

	// sanity check that both given minimum accounts are not negative amounts.
	if amount0Min.IsNegative() {
		return CreatePositionData{}, types.NotPositiveRequireAmountError{Amount: amount0Min.String()}
	}
	if amount1Min.IsNegative() {
		return CreatePositionData{}, types.NotPositiveRequireAmountError{Amount: amount1Min.String()}
	}

	// Transform the provided ticks into their corresponding sqrtPrices.
	sqrtPriceLowerTick, sqrtPriceUpperTick, err := math.TicksToSqrtPrice(lowerTick, upperTick)
	if err != nil {
		return CreatePositionData{}, err
	}

	// If multiple ticks can represent the same spot price, ensure we are using the largest of those ticks.
	lowerTick, upperTick, err = roundTickToCanonicalPriceTick(lowerTick, upperTick, sqrtPriceLowerTick, sqrtPriceUpperTick, pool.GetTickSpacing())
	if err != nil {
		return CreatePositionData{}, err
	}

	positionId := k.getNextPositionIdAndIncrement(ctx)

	// If this is the first position created in this pool, ensure that the position includes both asset0 and asset1
	// in order to assign an initial spot price.
	hasPositions, err := k.HasAnyPositionForPool(ctx, poolId)
	if err != nil {
		return CreatePositionData{}, err
	}

	if !hasPositions {
		err := k.initializeInitialPositionForPool(ctx, pool, amount0Desired, amount1Desired)
		if err != nil {
			return CreatePositionData{}, err
		}
	}

	// Calculate the amount of liquidity that will be added to the pool when this position is created.
	liquidityDelta := math.GetLiquidityFromAmounts(pool.GetCurrentSqrtPrice(), sqrtPriceLowerTick, sqrtPriceUpperTick, amount0Desired, amount1Desired)
	if liquidityDelta.IsZero() {
		return CreatePositionData{}, errors.New("liquidityDelta calculated equals zero")
	}

	// Initialize / update the position in the pool based on the provided tick range and liquidity delta.
	updateData, err := k.UpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta, joinTime, positionId)
	if err != nil {
		return CreatePositionData{}, err
	}

	// Check if the actual amounts of tokens 0 and 1 are greater than or equal to the given minimum amounts.
	if updateData.Amount0.LT(amount0Min) {
		return CreatePositionData{}, types.InsufficientLiquidityCreatedError{Actual: updateData.Amount0, Minimum: amount0Min, IsTokenZero: true}
	}
	if updateData.Amount1.LT(amount1Min) {
		return CreatePositionData{}, types.InsufficientLiquidityCreatedError{Actual: updateData.Amount1, Minimum: amount1Min}
	}

	// Transfer the actual amounts of tokens 0 and 1 from the position owner to the pool.
	err = k.sendCoinsBetweenPoolAndUser(ctx, pool.GetToken0(), pool.GetToken1(), updateData.Amount0, updateData.Amount1, owner, pool.GetAddress())
	if err != nil {
		return CreatePositionData{}, err
	}

	event := &liquidityChangeEvent{
		eventType:      types.TypeEvtCreatePosition,
		positionId:     positionId,
		sender:         owner,
		poolId:         poolId,
		lowerTick:      lowerTick,
		upperTick:      upperTick,
		joinTime:       joinTime,
		liquidityDelta: liquidityDelta,
		actualAmount0:  updateData.Amount0,
		actualAmount1:  updateData.Amount1,
	}
	event.emit(ctx)

	if !hasPositions {
		// N.B. calling this listener propagates to x/twap for twap record creation.
		// This is done after initial pool position only because only the first position
		// initializes the pool's spot price. After the initial position is created, only
		// swaps update the spot price.
		k.listeners.AfterInitialPoolPositionCreated(ctx, owner, poolId)
	}

	tokensAdded := sdk.Coins{}
	if updateData.Amount0.IsPositive() {
		tokensAdded = tokensAdded.Add(sdk.NewCoin(pool.GetToken0(), updateData.Amount0))
	}
	if updateData.Amount1.IsPositive() {
		tokensAdded = tokensAdded.Add(sdk.NewCoin(pool.GetToken1(), updateData.Amount1))
	}
	k.RecordTotalLiquidityIncrease(ctx, tokensAdded)

	return CreatePositionData{
		ID:        positionId,
		Amount0:   updateData.Amount0,
		Amount1:   updateData.Amount1,
		Liquidity: liquidityDelta,
		LowerTick: lowerTick,
		UpperTick: upperTick,
	}, nil
}

// WithdrawPosition attempts to withdraw liquidityAmount from a position with the given pool id in the given tick range.
// On success, returns a positive amount of each token withdrawn.
// If we are attempting to withdraw all liquidity available in the position, we also collect spread factors and incentives for the position.
// When the last position within a pool is removed, this function calls an AfterLastPoolPosistionRemoved listener
// Currently, it creates twap records. Assumming that pool had all liqudity drained and then re-initialized,
// the whole twap state is completely reset. This is because when there is no liquidity in pool, spot price
// is undefined. When the last position is removed by calling this method, the current sqrt price and current
// tick of the pool are set to zero. Lastly, if the tick being withdrawn from is now empty due to the withdrawal,
// it is deleted from state.
// Returns error if
// - the provided owner does not own the position being withdrawn
// - there is no position in the given tick ranges
// - if the position's underlying lock is not mature
// - if tick ranges are invalid
// - if attempts to withdraw an amount higher than originally provided in createPosition for a given range.
func (k Keeper) WithdrawPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, requestedLiquidityAmountToWithdraw sdk.Dec) (amtDenom0, amtDenom1 sdk.Int, err error) {
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// Check if the provided owner owns the position being withdrawn.
	if owner.String() != position.Address {
		return sdk.Int{}, sdk.Int{}, types.NotPositionOwnerError{PositionId: positionId, Address: owner.String()}
	}

	// Defense in depth, requestedLiquidityAmountToWithdraw should always be a value that is GE than 0.
	if requestedLiquidityAmountToWithdraw.IsNegative() {
		return sdk.Int{}, sdk.Int{}, types.InsufficientLiquidityError{Actual: requestedLiquidityAmountToWithdraw, Available: position.Liquidity}
	}

	// If underlying lock exists in state, validate unlocked conditions are met before withdrawing liquidity.
	// If the underlying lock for the position has been matured, remove the link between the position and the underlying lock.
	positionHasActiveUnderlyingLock, lockId, err := k.positionHasActiveUnderlyingLockAndUpdate(ctx, positionId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// If an underlying lock for the position exists, and the lock is not mature, return error.
	if positionHasActiveUnderlyingLock {
		return sdk.Int{}, sdk.Int{}, types.LockNotMatureError{PositionId: position.PositionId, LockId: lockId}
	}

	// Retrieve the pool associated with the given pool ID.
	pool, err := k.getPoolById(ctx, position.PoolId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// Check if the requested liquidity amount to withdraw is less than or equal to the available liquidity for the position.
	// If it is greater than the available liquidity, return an error.
	if requestedLiquidityAmountToWithdraw.GT(position.Liquidity) {
		return sdk.Int{}, sdk.Int{}, types.InsufficientLiquidityError{Actual: requestedLiquidityAmountToWithdraw, Available: position.Liquidity}
	}

	_, _, err = k.collectIncentives(ctx, owner, positionId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// Calculate the change in liquidity for the pool based on the requested amount to withdraw.
	// This amount is negative because that liquidity is being withdrawn from the pool.
	liquidityDelta := requestedLiquidityAmountToWithdraw.Neg()

	// Update the position in the pool based on the provided tick range and liquidity delta.
	updateData, err := k.UpdatePosition(ctx, position.PoolId, owner, position.LowerTick, position.UpperTick, liquidityDelta, position.JoinTime, positionId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// Transfer the actual amounts of tokens 0 and 1 from the pool to the position owner.
	err = k.sendCoinsBetweenPoolAndUser(ctx, pool.GetToken0(), pool.GetToken1(), updateData.Amount0.Abs(), updateData.Amount1.Abs(), pool.GetAddress(), owner)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, err
	}

	// If the requested liquidity amount to withdraw is equal to the available liquidity, delete the position from state.
	// Ensure we collect any outstanding spread factors and incentives prior to deleting the position from state. This claiming
	// process also clears position records from spread factor and incentive accumulators.
	if requestedLiquidityAmountToWithdraw.Equal(position.Liquidity) {
		if _, err := k.collectSpreadRewards(ctx, owner, positionId); err != nil {
			return sdk.Int{}, sdk.Int{}, err
		}

		if _, _, err := k.collectIncentives(ctx, owner, positionId); err != nil {
			return sdk.Int{}, sdk.Int{}, err
		}

		if err := k.deletePosition(ctx, positionId, owner, position.PoolId); err != nil {
			return sdk.Int{}, sdk.Int{}, err
		}

		anyPositionsRemainingInPool, err := k.HasAnyPositionForPool(ctx, position.PoolId)
		if err != nil {
			return sdk.Int{}, sdk.Int{}, err
		}

		if !anyPositionsRemainingInPool {
			// Reset the current tick and current square root price to initial values of zero since there is no
			// liquidity left.
			if err := k.uninitializePool(ctx, pool.GetId()); err != nil {
				return sdk.Int{}, sdk.Int{}, err
			}

			// N.B. since removing the liquidity of the last position in-full
			// implies invalidating spot price and current tick, we must
			// call this listener so that it updates twap module with the
			// invalid spot price for this pool.
			k.listeners.AfterLastPoolPositionRemoved(ctx, owner, pool.GetId())
		}
	}

	// If lowertick/uppertick has no liquidity in it, delete it from state.
	if updateData.LowerTickIsEmpty {
		k.RemoveTickInfo(ctx, position.PoolId, position.LowerTick)
	}
	if updateData.UpperTickIsEmpty {
		k.RemoveTickInfo(ctx, position.PoolId, position.UpperTick)
	}

	tokensRemoved := sdk.Coins{}
	if updateData.Amount0.IsPositive() {
		tokensRemoved = tokensRemoved.Add(sdk.NewCoin(pool.GetToken0(), updateData.Amount0))
	}
	if updateData.Amount1.IsPositive() {
		tokensRemoved = tokensRemoved.Add(sdk.NewCoin(pool.GetToken1(), updateData.Amount1))
	}
	k.RecordTotalLiquidityDecrease(ctx, tokensRemoved)

	event := &liquidityChangeEvent{
		eventType:      types.TypeEvtWithdrawPosition,
		positionId:     positionId,
		sender:         owner,
		poolId:         position.PoolId,
		lowerTick:      position.LowerTick,
		upperTick:      position.UpperTick,
		joinTime:       position.JoinTime,
		liquidityDelta: liquidityDelta,
		actualAmount0:  updateData.Amount0,
		actualAmount1:  updateData.Amount1,
	}
	event.emit(ctx)

	return updateData.Amount0.Neg(), updateData.Amount1.Neg(), nil
}

// addToPosition attempts to add amount0Added and amount1Added to a position with the given position id.
// For the sake of backwards-compatibility with future implementations of charging, this function deletes the old position and creates
// a new one with the resulting amount after addition. Note that due to truncation after `withdrawPosition`, there is some rounding error
// that is upper bounded by 1 unit of the more valuable token.
// Uses the amount0MinGiven + withdrawn amount0, amount1MinGiven + withdrawn amount1 as the minimum token out for creating the new position.
// Note that these field indicates the min amount corresponding to the total liquidity of the position,
// not only for the liquidity amount that is being added.
// Uses amounts withdrawn from the original position if provided min amount is zero.
// Returns error if
// - Withdrawing full position fails
// - Creating new position with added liquidity fails
// - Position with `positionId` is the last position in the pool
// - Position is superfluid staked
func (k Keeper) addToPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, amount0Added, amount1Added, amount0MinGiven, amount1MinGiven sdk.Int) (uint64, sdk.Int, sdk.Int, error) {
	position, err := k.GetPosition(ctx, positionId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, err
	}

	// Check if the provided owner owns the position being added to.
	if owner.String() != position.Address {
		return 0, sdk.Int{}, sdk.Int{}, types.NotPositionOwnerError{PositionId: positionId, Address: owner.String()}
	}

	// if one of the liquidity is negative, or both liquidity being added is zero, error
	if amount0Added.IsNegative() || amount1Added.IsNegative() {
		return 0, sdk.Int{}, sdk.Int{}, types.NegativeAmountAddedError{PositionId: position.PositionId, Asset0Amount: amount0Added, Asset1Amount: amount1Added}
	}

	if amount0Added.IsZero() && amount1Added.IsZero() {
		return 0, sdk.Int{}, sdk.Int{}, types.ErrZeroLiquidity
	}

	// If the position is superfluid staked, return error.
	// This path is handled separately in the superfluid module.
	positionHasUnderlyingLock, _, err := k.positionHasActiveUnderlyingLockAndUpdate(ctx, positionId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, err
	}
	if positionHasUnderlyingLock {
		return 0, sdk.Int{}, sdk.Int{}, types.PositionSuperfluidStakedError{PositionId: position.PositionId}
	}

	// Withdraw full position.
	amount0Withdrawn, amount1Withdrawn, err := k.WithdrawPosition(ctx, owner, positionId, position.Liquidity)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, err
	}

	anyPositionsRemainingInPool, err := k.HasAnyPositionForPool(ctx, position.PoolId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, err
	}

	if !anyPositionsRemainingInPool {
		return 0, sdk.Int{}, sdk.Int{}, types.AddToLastPositionInPoolError{PoolId: position.PoolId, PositionId: position.PositionId}
	}

	// Create new position with updated liquidity.
	amount0Desired := amount0Withdrawn.Add(amount0Added)
	amount1Desired := amount1Withdrawn.Add(amount1Added)
	pool, err := k.GetConcentratedPoolById(ctx, position.PoolId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, err
	}
	tokensProvided := sdk.NewCoins(sdk.NewCoin(pool.GetToken0(), amount0Desired), sdk.NewCoin(pool.GetToken1(), amount1Desired))
	minimumAmount0 := amount0Withdrawn
	minimumAmount1 := amount1Withdrawn

	if !amount0MinGiven.IsZero() {
		minimumAmount0 = amount0Withdrawn.Add(amount0MinGiven)
	}
	if !amount1MinGiven.IsZero() {
		minimumAmount1 = amount1Withdrawn.Add(amount1MinGiven)
	}
	newPositionData, err := k.CreatePosition(ctx, position.PoolId, owner, tokensProvided, minimumAmount0, minimumAmount1, position.LowerTick, position.UpperTick)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, err
	}

	// Emit an event indicating that a position was added to.
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToPosition,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, owner.String()),
			sdk.NewAttribute(types.AttributeKeyPositionId, strconv.FormatUint(positionId, 10)),
			sdk.NewAttribute(types.AttributeKeyNewPositionId, strconv.FormatUint(newPositionData.ID, 10)),
			sdk.NewAttribute(types.AttributeAmount0, newPositionData.Amount0.String()),
			sdk.NewAttribute(types.AttributeAmount1, newPositionData.Amount1.String()),
		),
	})

	return newPositionData.ID, newPositionData.Amount0, newPositionData.Amount1, nil
}

// UpdatePosition updates the position in the given pool id and in the given tick range and liquidityAmount.
// Negative liquidityDelta implies withdrawing liquidity.
// Positive liquidityDelta implies adding liquidity.
// Updates ticks and pool liquidity. Returns how much of each token is either added or removed.
// Negative returned amounts imply that tokens are removed from the pool.
// Positive returned amounts imply that tokens are added to the pool.
// If the lower and/or upper ticks are being updated to have zero liquidity, a boolean is returned to flag the tick as empty to be deleted at the end of the withdrawPosition method.
// WARNING: this method may mutate the pool, make sure to refetch the pool after calling this method.
func (k Keeper) UpdatePosition(ctx sdk.Context, poolId uint64, owner sdk.AccAddress, lowerTick, upperTick int64, liquidityDelta sdk.Dec, joinTime time.Time, positionId uint64) (types.UpdatePositionData, error) {
	if err := k.validatePositionUpdateById(ctx, positionId, owner, lowerTick, upperTick, liquidityDelta, joinTime, poolId); err != nil {
		return types.UpdatePositionData{}, err
	}

	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return types.UpdatePositionData{}, err
	}

	currentTick := pool.GetCurrentTick()

	// update lower tickInfo state
	lowerTickIsEmpty, err := k.initOrUpdateTick(ctx, poolId, currentTick, lowerTick, liquidityDelta, false)
	if err != nil {
		return types.UpdatePositionData{}, err
	}

	// update upper tickInfo state
	upperTickIsEmpty, err := k.initOrUpdateTick(ctx, poolId, currentTick, upperTick, liquidityDelta, true)
	if err != nil {
		return types.UpdatePositionData{}, err
	}

	// update position state
	err = k.initOrUpdatePosition(ctx, poolId, owner, lowerTick, upperTick, liquidityDelta, joinTime, positionId)
	if err != nil {
		return types.UpdatePositionData{}, err
	}

	// Refetch pool to get the updated pool.
	// Note that updateUptimeAccumulatorsToNow may modify the pool state and rewrite it to the store.
	pool, err = k.getPoolById(ctx, poolId)
	if err != nil {
		return types.UpdatePositionData{}, err
	}

	// calculate the actual amounts of tokens 0 and 1 that were added or removed from the pool.
	actualAmount0, actualAmount1, err := pool.CalcActualAmounts(ctx, lowerTick, upperTick, liquidityDelta)
	if err != nil {
		return types.UpdatePositionData{}, err
	}

	// the pool's liquidity value is only updated if this position is active
	pool.UpdateLiquidityIfActivePosition(ctx, lowerTick, upperTick, liquidityDelta)

	if err := k.setPool(ctx, pool); err != nil {
		return types.UpdatePositionData{}, err
	}

	if err := k.initOrUpdatePositionSpreadRewardAccumulator(ctx, poolId, lowerTick, upperTick, positionId, liquidityDelta); err != nil {
		return types.UpdatePositionData{}, err
	}

	// The returned amounts are rounded down to avoid returning more to clients than they actually deposited.
	return types.UpdatePositionData{
		Amount0:          actualAmount0.TruncateInt(),
		Amount1:          actualAmount1.TruncateInt(),
		LowerTickIsEmpty: lowerTickIsEmpty,
		UpperTickIsEmpty: upperTickIsEmpty,
	}, nil
}

// sendCoinsBetweenPoolAndUser takes the amounts calculated from a join/exit position and executes the send between pool and user
func (k Keeper) sendCoinsBetweenPoolAndUser(ctx sdk.Context, denom0, denom1 string, amount0, amount1 sdk.Int, sender, receiver sdk.AccAddress) error {
	if amount0.IsNegative() {
		return types.Amount0IsNegativeError{Amount0: amount0}
	}
	if amount1.IsNegative() {
		return types.Amount1IsNegativeError{Amount1: amount1}
	}

	finalCoinsToSend := sdk.NewCoins(sdk.NewCoin(denom1, amount1), sdk.NewCoin(denom0, amount0))
	err := k.bankKeeper.SendCoins(ctx, sender, receiver, finalCoinsToSend)
	if err != nil {
		return err
	}
	return nil
}

// initializeInitialPositionForPool ensures that the first position created on this pool includes both asset0 and asset1
// This is required so we can set the pool's sqrtPrice and calculate it's initial tick from this.
// Additionally, it initializes the current sqrt price and current tick from the initial reserve values.
func (k Keeper) initializeInitialPositionForPool(ctx sdk.Context, pool types.ConcentratedPoolExtension, amount0Desired, amount1Desired sdk.Int) error {
	// Check that the position includes some amount of both asset0 and asset1
	if !amount0Desired.GT(sdk.ZeroInt()) || !amount1Desired.GT(sdk.ZeroInt()) {
		return types.InitialLiquidityZeroError{Amount0: amount0Desired, Amount1: amount1Desired}
	}

	// Calculate the spot price and sqrt price from the amount provided
	initialSpotPrice := amount1Desired.ToDec().Quo(amount0Desired.ToDec())
	// TODO: any concerns with this being an sdk.Dec?
	initialCurSqrtPrice, err := osmomath.MonotonicSqrt(initialSpotPrice)
	if err != nil {
		return err
	}

	// Calculate the initial tick from the initial spot price
	// We round down here so that the tick is rounded to
	// the nearest possible value given the tick spacing.
	initialTick, err := math.SqrtPriceToTickRoundDownSpacing(initialCurSqrtPrice, pool.GetTickSpacing())
	if err != nil {
		return err
	}

	// Set the pool's current sqrt price and current tick to the above calculated values
	// Note that initial initial cur sqrt price might not fall directly on the initial tick.
	// For example, if we have tick spacing of 1, default exponent at price one of -6, and
	// the current spot price of 100_000_025.123 X/Y.
	// However, there are ticks only at 100_000_000 X/Y and 100_000_100 X/Y.
	// In such a case, we do not want to round the sqrt price to 100_000_000 X/Y, but rather
	// let it float within the possible tick range.
	pool.SetCurrentSqrtPrice(osmomath.BigDecFromSDKDec(initialCurSqrtPrice))
	pool.SetCurrentTick(initialTick)
	err = k.setPool(ctx, pool)
	if err != nil {
		return err
	}
	return nil
}

// uninitializePool reinitializes a pool if it has no liquidity.
// It does so by setting the current square root price and tick to zero.
// This is necessary for the twap to correctly detect a spot price error
// when there is no liquidity in the pool.
func (k Keeper) uninitializePool(ctx sdk.Context, poolId uint64) error {
	pool, err := k.getPoolById(ctx, poolId)
	if err != nil {
		return err
	}

	hasAnyPosition, err := k.HasAnyPositionForPool(ctx, poolId)
	if err != nil {
		return err
	}

	if hasAnyPosition {
		return types.UninitializedPoolWithLiquidityError{PoolId: poolId}
	}

	pool.SetCurrentSqrtPrice(osmomath.ZeroDec())
	pool.SetCurrentTick(0)

	if err := k.setPool(ctx, pool); err != nil {
		return err
	}

	return nil
}

// isLockMature checks if the underlying lock has expired.
// If the lock doesn't exist, it returns true.
// If the lock exists, it checks if the lock has expired.
// If the lock has expired, it returns true.
// If the lock is still active, it returns false.
func (k Keeper) isLockMature(ctx sdk.Context, underlyingLockId uint64) (bool, error) {
	// Query the underlying lock
	underlyingLock, err := k.lockupKeeper.GetLockByID(ctx, underlyingLockId)
	if errors.Is(err, lockuptypes.ErrLockupNotFound) {
		// Lock doesn't exist, so we can withdraw from this position
		return true, nil
	} else if err != nil {
		// Unexpected error, return false to prevent any further action and return the error
		return false, err
	}

	if underlyingLock.EndTime.IsZero() {
		// Lock is still active, so we can't withdraw from this position
		return false, nil
	}

	// Return if the lock has expired
	return underlyingLock.EndTime.Before(ctx.BlockTime()), nil
}

// validatePositionUpdateById validates the parameters for updating an existing position.
// Returns nil on success. Returns nil if position with the given id does not exist.
// Returns an error if any of the parameters are invalid or mismatched.
// If the position ID is zero, returns types.ErrZeroPositionId.
// If the position owner does not match the update initiator, returns types.PositionOwnerMismatchError.
// If the lower tick provided does not match the position's lower tick, returns types.LowerTickMismatchError.
// If the upper tick provided does not match the position's upper tick, returns types.UpperTickMismatchError.
// If the liquidity to withdraw is greater than the current liquidity of the position, returns types.LiquidityWithdrawalError.
// If the join time provided does not match the position's join time, returns types.JoinTimeMismatchError.
// If the position does not belong to the pool with the provided pool ID, returns types.PositionsNotInSamePoolError.
func (k Keeper) validatePositionUpdateById(ctx sdk.Context, positionId uint64, updateInitiator sdk.AccAddress, lowerTickGiven int64, upperTickGiven int64, liquidityDeltaGiven sdk.Dec, joinTimeGiven time.Time, poolIdGiven uint64) error {
	if positionId == 0 {
		return types.ErrZeroPositionId
	}

	if hasPosition := k.hasPosition(ctx, positionId); hasPosition {
		position, err := k.GetPosition(ctx, positionId)
		if err != nil {
			return err
		}

		if position.Address != updateInitiator.String() {
			return types.PositionOwnerMismatchError{PositionOwner: position.Address, Sender: updateInitiator.String()}
		}

		if position.LowerTick != lowerTickGiven {
			return types.LowerTickMismatchError{PositionId: positionId, Expected: position.LowerTick, Got: lowerTickGiven}
		}

		if position.UpperTick != upperTickGiven {
			return types.UpperTickMismatchError{PositionId: positionId, Expected: position.UpperTick, Got: upperTickGiven}
		}

		if liquidityDeltaGiven.IsNegative() && position.Liquidity.LT(liquidityDeltaGiven.Abs()) {
			return types.LiquidityWithdrawalError{PositionID: positionId, RequestedAmount: liquidityDeltaGiven.Abs(), CurrentLiquidity: position.Liquidity}
		}

		if position.JoinTime.UTC() != joinTimeGiven.UTC() {
			return types.JoinTimeMismatchError{PositionId: positionId, Expected: position.JoinTime, Got: joinTimeGiven}
		}

		if position.PoolId != poolIdGiven {
			return types.PositionsNotInSamePoolError{Position1PoolId: position.PoolId, Position2PoolId: poolIdGiven}
		}
	}

	return nil
}
