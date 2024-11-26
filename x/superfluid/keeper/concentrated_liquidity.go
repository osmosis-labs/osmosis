package keeper

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

// addToConcentratedLiquiditySuperfluidPosition adds the specified amounts of tokens to an existing superfluid staked
// concentrated liquidity position. Under the hood, it withdraws the current position, adds funds to the withdrawn position,
// and then creates a new position with the new liquidity.
//
// Returns:
// newPositionId: ID of the newly created concentrated liquidity position.
// actualAmount0: Actual amount of token 0 existing in the updated position.
// actualAmount1: Actual amount of token 1 existing in the updated position.
// newLiquidity: The new liquidity value.
// newLockId: ID of the lock associated with the new position.
// error: Error, if any.
//
// An error is returned if:
// - The position does not exist.
// - The amount added is negative.
// - The provided sender does not own the lock.
// - The provided sender does not own the position.
// - The position is not superfluid staked.
// - The position is the last position in the pool.
// - The lock duration does not match the unbonding duration.
func (k Keeper) addToConcentratedLiquiditySuperfluidPosition(ctx sdk.Context, sender sdk.AccAddress, positionId uint64, amount0ToAdd, amount1ToAdd osmomath.Int) (cltypes.CreateFullRangePositionData, uint64, error) {
	position, err := k.clk.GetPosition(ctx, positionId)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}

	if amount0ToAdd.IsNegative() || amount1ToAdd.IsNegative() {
		return cltypes.CreateFullRangePositionData{}, 0, cltypes.NegativeAmountAddedError{PositionId: position.PositionId, Asset0Amount: amount0ToAdd, Asset1Amount: amount1ToAdd}
	}

	// If the position is not superfluid staked, return error.
	positionHasActiveUnderlyingLock, lockId, err := k.clk.PositionHasActiveUnderlyingLock(ctx, positionId)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}
	if !positionHasActiveUnderlyingLock || lockId == 0 {
		return cltypes.CreateFullRangePositionData{}, 0, types.PositionNotSuperfluidStakedError{PositionId: position.PositionId}
	}

	// Defense in depth making sure that the position is full-range.
	if position.LowerTick != cltypes.MinInitializedTick || position.UpperTick != cltypes.MaxTick {
		return cltypes.CreateFullRangePositionData{}, 0, types.ConcentratedTickRangeNotFullError{ActualLowerTick: position.LowerTick, ActualUpperTick: position.UpperTick}
	}

	lock, err := k.lk.GetLockByID(ctx, lockId)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}

	// Defense in depth. Require the underlying lock:
	// - owner matches the owner of the position (this should always be true)
	// - owner matches the function caller
	// - duration is equal to unbonding time
	// - end time is zero (not unbonding)
	if lock.Owner != position.Address {
		return cltypes.CreateFullRangePositionData{}, 0, types.LockOwnerMismatchError{LockId: lockId, LockOwner: lock.Owner, ProvidedOwner: position.Address}
	}
	if lock.Owner != sender.String() {
		return cltypes.CreateFullRangePositionData{}, 0, types.LockOwnerMismatchError{LockId: lockId, LockOwner: lock.Owner, ProvidedOwner: sender.String()}
	}
	unbondingDuration, err := k.sk.UnbondingTime(ctx)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}
	if lock.Duration != unbondingDuration || !lock.EndTime.IsZero() {
		return cltypes.CreateFullRangePositionData{}, 0, types.LockImproperStateError{LockId: lockId, UnbondingDuration: unbondingDuration.String()}
	}

	// Superfluid undelegate the superfluid delegated position.
	// This deletes the connection between the lock and the intermediate account, deletes the synthetic lock, and burns the synthetic osmo.
	intermediateAccount, err := k.SuperfluidUndelegateToConcentratedPosition(ctx, sender.String(), lockId)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}

	// Finish unlocking directly for the lock.
	// This also breaks and deletes associated synthetic locks.
	err = k.lk.ForceUnlock(ctx, *lock)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}

	// Withdraw full liquidity from the position.
	amount0Withdrawn, amount1Withdrawn, err := k.clk.WithdrawPosition(ctx, sender, positionId, position.Liquidity)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}

	// If this is the last position in the pool, error.
	anyPositionsRemainingInPool, err := k.clk.HasAnyPositionForPool(ctx, position.PoolId)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}
	if !anyPositionsRemainingInPool {
		return cltypes.CreateFullRangePositionData{}, 0, cltypes.AddToLastPositionInPoolError{PoolId: position.PoolId, PositionId: position.PositionId}
	}

	// Create a coins object that includes the old position coins and the new position coins.
	concentratedPool, err := k.clk.GetConcentratedPoolById(ctx, position.PoolId)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}
	newPositionCoins := sdk.NewCoins(sdk.NewCoin(concentratedPool.GetToken0(), amount0Withdrawn.Add(amount0ToAdd)), sdk.NewCoin(concentratedPool.GetToken1(), amount1Withdrawn.Add(amount1ToAdd)))

	// Create a full range (min to max tick) concentrated liquidity position, lock it, and superfluid delegate it.
	positionData, newLockId, err := k.clk.CreateFullRangePositionLocked(ctx, position.PoolId, sender, newPositionCoins, unbondingDuration)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}
	err = k.SuperfluidDelegate(ctx, sender.String(), newLockId, intermediateAccount.ValAddr)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, err
	}

	// Emit events.
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToConcentratedLiquiditySuperfluidPosition,
			sdk.NewAttribute(sdk.AttributeKeySender, sender.String()),
			sdk.NewAttribute(types.AttributeKeyPoolId, strconv.FormatUint(position.PoolId, 10)),
			sdk.NewAttribute(types.AttributePositionId, strconv.FormatUint(positionId, 10)),
			sdk.NewAttribute(types.AttributeNewPositionId, strconv.FormatUint(positionData.ID, 10)),
			sdk.NewAttribute(types.AttributeAmount0, positionData.Amount0.String()),
			sdk.NewAttribute(types.AttributeAmount1, positionData.Amount1.String()),
			sdk.NewAttribute(types.AttributeConcentratedLockId, strconv.FormatUint(newLockId, 10)),
			sdk.NewAttribute(types.AttributeLiquidity, positionData.Liquidity.String()),
		),
	})

	return positionData, newLockId, nil
}
