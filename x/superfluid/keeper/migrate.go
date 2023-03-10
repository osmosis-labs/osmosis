package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
)

// UnlockAndMigrate unlocks a balancer pool lock, exits the pool and migrates the LP position to a full range concentrated liquidity position.
// If the lock is superfluid delegated, it will undelegate the superfluid position.
// Errors if the lock is not found, if the lock is not a balancer pool lock, or if the lock is not owned by the sender.
func (k Keeper) UnlockAndMigrate(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (amount0, amount1 sdk.Int, liquidity sdk.Dec, poolIdLeaving, poolIdEntering, newLockId uint64, freezeDuration time.Duration, err error) {
	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving = gammtypes.MustGetPoolIdFromShareDenom(sharesToMigrate.Denom)

	// Ensure a governance sanctioned link exists between the balancer pool and the concentrated pool.
	poolIdEntering, err = k.gk.GetLinkedConcentratedPoolID(ctx, poolIdLeaving)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
	}

	// Get the concentrated pool from the provided ID and type cast it to ConcentratedPoolExtension.
	concentratedPool, err := k.clk.GetPoolFromPoolIdAndConvertToConcentrated(ctx, poolIdEntering)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
	}

	// Check that lockID corresponds to sender, and contains correct denomination of LP shares.
	lock, err := k.validateLockForUnpool(ctx, sender, poolIdLeaving, lockId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
	}
	gammSharesInLock := lock.Coins[0]
	preUnlockLock := *lock

	// Before we break the lock, we must note the time remaining on the lock.
	// We will be freezing the concentrated liquidity position for this duration.
	freezeDuration = k.getExistingLockRemainingDuration(ctx, lock)

	// If superfluid delegated, superfluid undelegate
	// This also burns the underlying synthetic osmo
	err = k.unbondSuperfluidIfExists(ctx, sender, lockId)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
	}

	// Finish unlocking directly for locked locks
	// this also unlocks locks that were in the unlocking queue
	err = k.lk.ForceUnlock(ctx, *lock)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
	}

	// If shares to migrate is not specified, we migrate all shares.
	if sharesToMigrate.IsZero() {
		sharesToMigrate = gammSharesInLock
	}

	// Otherwise, we must ensure that the shares to migrate is less than or equal to the shares in the lock.
	if sharesToMigrate.Amount.GT(gammSharesInLock.Amount) {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, fmt.Errorf("shares to migrate must be less than or equal to shares in lock")
	}

	// Exit the balancer pool position.
	exitCoins, err := k.gk.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, sdk.NewCoins())
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
	}
	// Defense in depth, ensuring we are returning exactly two coins.
	if len(exitCoins) != 2 {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, fmt.Errorf("Balancer pool must have exactly two tokens")
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	amount0, amount1, liquidity, err = k.clk.CreateFullRangePosition(ctx, concentratedPool, sender, exitCoins, freezeDuration)
	if err != nil {
		return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
	}

	// If there are remaining gamm shares, we must re-lock them.
	remainingGammShares := gammSharesInLock.Sub(sharesToMigrate)
	if !remainingGammShares.IsZero() {
		newLock, err := k.lk.CreateLock(ctx, sender, sdk.NewCoins(remainingGammShares), freezeDuration)
		newLockId = newLock.ID
		if err != nil {
			return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
		}
		// If the lock was unlocking, we begin the unlock from where it left off.
		if preUnlockLock.IsUnlocking() {
			_, err := k.lk.BeginForceUnlock(ctx, newLock.ID, newLock.Coins)
			if err != nil {
				return sdk.Int{}, sdk.Int{}, sdk.Dec{}, 0, 0, 0, 0, err
			}
		}
	}

	return amount0, amount1, liquidity, poolIdLeaving, poolIdEntering, newLockId, freezeDuration, nil
}
