package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

// addToConcentratedLiquiditySuperfluidPosition adds the specified amounts of tokens to an existing superfluid staked
// concentrated liquidity position. Under the hood, it withdraws the current position, adds funds to the withdrawn position,
// and then creates a new position with the new liquidity.
//
// Returns:
// newPositionId: ID of the newly created concentrated liquidity position.
// actualAmount0: Actual amount of token 0 added.
// actualAmount1: Actual amount of token 1 added.
// newLiquidity: The new liquidity value.
// newLockId: ID of the lock associated with the new position.
// error: Error, if any.
//
// An error is returned if:
// - The position does not exist.
// - The amount added is negative.
// - The position is not superfluid staked.
// - The position is the last position in the pool.
func (k Keeper) addToConcentratedLiquiditySuperfluidPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, amount0ToAdd, amount1ToAdd sdk.Int) (uint64, sdk.Int, sdk.Int, sdk.Dec, uint64, error) {
	position, err := k.clk.GetPosition(ctx, positionId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}

	if amount0ToAdd.IsNegative() || amount1ToAdd.IsNegative() {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, cltypes.NegativeAmountAddedError{PositionId: position.PositionId, Asset0Amount: amount0ToAdd, Asset1Amount: amount1ToAdd}
	}

	// If the position is not superfluid staked, return error.
	positionHasActiveUnderlyingLock, lockId, err := k.clk.PositionHasActiveUnderlyingLock(ctx, positionId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}
	if !positionHasActiveUnderlyingLock || lockId == 0 {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, types.PositionNotSuperfluidStakedError{PositionId: position.PositionId}
	}

	lock, err := k.lk.GetLockByID(ctx, lockId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}

	// Defense in depth. Require the underlying lock:
	// - duration is equal to unbonding time
	// - end time is zero (not unbonding)
	unbondingDuration := k.sk.UnbondingTime(ctx)
	if lock.Duration != unbondingDuration || !lock.EndTime.IsZero() {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, types.LockImproperStateError{LockId: lockId, UnbondingDuration: unbondingDuration.String()}
	}

	// Superfluid undelegate the superfluid delegated position.
	// This deletes the connection between the lock and the intermediate account, deletes the synthetic lock, and burns the synthetic osmo.
	intermediateAccount, err := k.SuperfluidUndelegateToConcentratedPosition(ctx, owner.String(), lockId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}

	// Finish unlocking directly for the lock.
	// This also breaks and deletes associated synthetic locks.
	err = k.lk.ForceUnlock(ctx, *lock)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}

	// Withdraw full position.
	amount0Withdrawn, amount1Withdrawn, err := k.clk.WithdrawPosition(ctx, owner, positionId, position.Liquidity)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}

	// If this is the last position in the pool, error.
	anyPositionsRemainingInPool, err := k.clk.HasAnyPositionForPool(ctx, position.PoolId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}
	if !anyPositionsRemainingInPool {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, cltypes.AddToLastPositionInPoolError{PoolId: position.PoolId, PositionId: position.PositionId}
	}

	// Create a coins object that includes the old position coins and the new position coins.
	concentratedPool, err := k.clk.GetPoolFromPoolIdAndConvertToConcentrated(ctx, position.PoolId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}
	newPositionCoins := sdk.NewCoins(sdk.NewCoin(concentratedPool.GetToken0(), amount0Withdrawn.Add(amount0ToAdd)), sdk.NewCoin(concentratedPool.GetToken1(), amount1Withdrawn.Add(amount1ToAdd)))

	// Create a full range (min to max tick) concentrated liquidity position, lock it, and superfluid delegate it.
	newPositionId, actualAmount0, actualAmount1, newLiquidity, _, newLockId, err := k.clk.CreateFullRangePositionLocked(ctx, position.PoolId, owner, newPositionCoins, unbondingDuration)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}
	err = k.SuperfluidDelegate(ctx, owner.String(), newLockId, intermediateAccount.ValAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, err
	}

	return newPositionId, actualAmount0, actualAmount1, newLiquidity, newLockId, nil
}
