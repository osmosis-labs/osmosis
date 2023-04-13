package keeper

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

// MigrateLockedPositionFromBalancerToConcentrated unlocks a balancer pool lock, exits the pool and migrates the LP position to a full range concentrated liquidity position.
// If the lock is superfluid delegated, it will undelegate the superfluid position and redelegate it as the concentrated liquidity position.
// If the lock is superfluid undelegating, it will fully undelegate the superfluid position and redelegate it as the concentrated liquidity position, but continue to unlock where it left off.
// If the lock is locked or unlocking but not superfluid delegated/undelegating, it will migrate the position and either start unlocking or continue unlocking where it left off.
// Errors if the lock is not found, if the lock is not a balancer pool lock, or if the lock is not owned by the sender.
func (k Keeper) MigrateLockedPositionFromBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, poolIdLeaving, poolIdEntering, gammLockId, concentratedLockId uint64, err error) {
	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving = gammtypes.MustGetPoolIdFromShareDenom(sharesToMigrate.Denom)

	// Ensure a governance sanctioned link exists between the balancer pool and the concentrated pool.
	poolIdEntering, err = k.gk.GetLinkedConcentratedPoolID(ctx, poolIdLeaving)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Get the concentrated pool from the provided ID and type cast it to ConcentratedPoolExtension.
	concentratedPool, err := k.clk.GetPoolFromPoolIdAndConvertToConcentrated(ctx, poolIdEntering)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Check that lockID corresponds to sender, and contains correct denomination of LP shares.
	lock, err := k.validateLockForUnpool(ctx, sender, poolIdLeaving, lockId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}
	gammSharesInLock := lock.Coins[0]
	preUnlockLock := *lock

	// Before we break the lock, we must note the time remaining on the lock.
	remainingLockTime := k.getExistingLockRemainingDuration(ctx, lock)

	// Check if the lock has a corresponding synthetic lock.
	synthLockBeforeMigration := k.lk.GetAllSyntheticLockupsByLockup(ctx, lockId)

	// If it does, check if it is superfluid delegated or undelegating.
	wasSuperfluidUndelegatingBeforeMigration := len(synthLockBeforeMigration) > 0 && strings.Contains(synthLockBeforeMigration[0].SynthDenom, "superunbonding")
	wasSuperfluidDelegatedBeforeMigration := len(synthLockBeforeMigration) > 0 && strings.Contains(synthLockBeforeMigration[0].SynthDenom, "superbonding")

	// If it is superfluid delegated or undelegating, get the validator address from the synth denom.
	valAddr := ""
	if wasSuperfluidDelegatedBeforeMigration || wasSuperfluidUndelegatingBeforeMigration {
		valAddr = strings.Split(synthLockBeforeMigration[0].SynthDenom, "/")[4]
	}

	// If the lock wassuperfluid delegated, superfluid undelegate it
	// This deletes the connection between the lock and the intermediate account, deletes the synthetic lock, and burns the synthetic osmo.
	intermediateAccount := types.SuperfluidIntermediaryAccount{}
	if wasSuperfluidDelegatedBeforeMigration {
		// superfluid undelegate and break any underlying synthetic locks
		// this is the same as SuperfluidUndelegate, but does not create a corresponding superunbonding synthetic lock
		intermediateAccount, err = k.SuperfluidUndelegateToConcentratedPosition(ctx, sender.String(), lockId)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
	}

	// Finish unlocking directly for locked locks
	// this also unlocks locks that were in the unlocking queue
	err = k.lk.ForceUnlock(ctx, *lock)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// If shares to migrate is not specified, we migrate all shares.
	if sharesToMigrate.IsZero() {
		sharesToMigrate = gammSharesInLock
	}

	// Otherwise, we must ensure that the shares to migrate is less than or equal to the shares in the lock.
	if sharesToMigrate.Amount.GT(gammSharesInLock.Amount) {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, fmt.Errorf("shares to migrate must be less than or equal to shares in lock")
	}

	// Exit the balancer pool position.
	exitCoins, err := k.gk.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, sdk.NewCoins())
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}
	// Defense in depth, ensuring we are returning exactly two coins.
	if len(exitCoins) != 2 {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, fmt.Errorf("Balancer pool must have exactly two tokens")
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	if wasSuperfluidDelegatedBeforeMigration {
		// If the lock was previously superfluid delegated, we create a new lock and keep it locked.
		positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionLocked(ctx, concentratedPool, sender, exitCoins, remainingLockTime)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
	} else {
		// If the lock was unlocking, we create a new lock that is unlocking for the remaining time of the old lock.
		positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionUnlocking(ctx, concentratedPool, sender, exitCoins, remainingLockTime)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
	}

	// If the lock was previously superfluid delegated, superfluid delegate the new concentrated lock to the same validator
	if wasSuperfluidDelegatedBeforeMigration {
		err := k.SuperfluidDelegate(ctx, sender.String(), concentratedLockId, intermediateAccount.ValAddr)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
	}

	// If the lock was superfluid undelegating at time of migration
	if wasSuperfluidUndelegatingBeforeMigration {
		// Create and set a new intermediary account based on the previous validator but with the new lock id and concentratedLockupDenom
		concentratedLockupDenom := cltypes.GetConcentratedLockupDenom(poolIdEntering, positionId)
		clIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, concentratedLockupDenom, valAddr)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}

		// Create a new synthetic lockup for the new intermediary account in an unlocking status
		err = k.createSyntheticLockup(ctx, concentratedLockId, clIntermediateAccount, unlockingStatus)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
	}

	// If there are remaining gamm shares, we must re-lock them.
	remainingGammShares := gammSharesInLock.Sub(sharesToMigrate)
	if !remainingGammShares.IsZero() {
		// Create a new lock with the remaining gamm shares for the remaining lock time.
		newLock, err := k.lk.CreateLock(ctx, sender, sdk.NewCoins(remainingGammShares), remainingLockTime)
		gammLockId = newLock.ID
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}

		// If the gamm lock was previously superfluid bonded, superfluid delegate the gamm like normal
		if wasSuperfluidDelegatedBeforeMigration {
			err := k.SuperfluidDelegate(ctx, sender.String(), gammLockId, valAddr)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
		}

		// If the gamm lock was superfluid unbonding, get the previous gamm intermediary account, create a new gamm synthetic lockup, and set it to unlocking
		if wasSuperfluidUndelegatingBeforeMigration {
			gammIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, remainingGammShares.Denom, valAddr)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
			err = k.createSyntheticLockup(ctx, gammLockId, gammIntermediateAccount, unlockingStatus)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
		}

		// If the gamm lock was neither superfluid delegated or undelegating but it was unlocking, we begin the unlock from where it left off.
		if preUnlockLock.IsUnlocking() {
			_, err := k.lk.BeginForceUnlock(ctx, newLock.ID, newLock.Coins)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
		}
	}

	return positionId, amount0, amount1, liquidity, joinTime, poolIdLeaving, poolIdEntering, gammLockId, concentratedLockId, nil
}
