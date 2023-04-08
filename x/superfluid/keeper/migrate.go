package keeper

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cl "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

// UnlockAndMigrate unlocks a balancer pool lock, exits the pool and migrates the LP position to a full range concentrated liquidity position.
// If the lock is superfluid delegated, it will undelegate the superfluid position.
// Errors if the lock is not found, if the lock is not a balancer pool lock, or if the lock is not owned by the sender.
func (k Keeper) UnlockAndMigrate(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, poolIdLeaving, poolIdEntering, gammLockId, concentratedLockId uint64, err error) {
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

	// We also need to note the synthetic lock before we break the lock, because the synthetic lock denom will
	// be removed, which is the only way we can tell which validator the lock was previously delegated to.
	synthLockBeforeMigration := k.lk.GetAllSyntheticLockupsByLockup(ctx, lockId)

	// If superfluid delegated, superfluid undelegate
	// This deletes the connection between the lock and the intermediate account, deletes the synthetic lock, and burns the synthetic osmo.

	// TODO: If we want to allow migrating a SFS balancer position to a SFS concentrated position, we need to remove this logic.
	intermediateAccount := types.SuperfluidIntermediaryAccount{}
	_, isCurrentlySuperfluidDelegated := k.GetIntermediaryAccountFromLockId(ctx, lockId)
	if isCurrentlySuperfluidDelegated {
		// superfluid undelegate and break any underlying synthetic locks
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
	// This will also create a lock for the time remaining from the gamm lock and begin unlocking from where it left off.
	//
	// TODO: If we eventually want to support migrating a SFS position from balancer to CL and continue staking,
	// we need to make a createPosition that does not start unlocking automatically. Right now, this is just being used
	// as a pseudo unfreeze logic.
	positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionUnlocking(ctx, concentratedPool, sender, exitCoins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Link the newly created concentratedLockId to the intermediate account if is currently superfluid delegated

	// TODO: I feel like this is wrong, we need to either send the funds from the old intermediary account to the new one
	// or somehow ensure the slashing is done on the old intermediary account
	if isCurrentlySuperfluidDelegated {
		// Create and set a new intermediary account based on the previous validator but with the new lock id and concentratedLockupDenom
		concentratedLockupDenom := cl.GetConcentratedLockupDenom(poolIdEntering, positionId)
		clIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, concentratedLockupDenom, intermediateAccount.ValAddr)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}

		// Create a new synthetic lockup for the new intermediary account in an unlocking status
		err = k.createSyntheticLockup(ctx, concentratedLockId, clIntermediateAccount, unlockingStatus)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
	}

	// If the lock was superfluid unbonding at time of migration
	if len(synthLockBeforeMigration) > 0 && strings.Contains(synthLockBeforeMigration[0].SynthDenom, "superunbonding") {
		// Create and set a new intermediary account based on the previous validator but with the new lock id and concentratedLockupDenom
		concentratedLockupDenom := cl.GetConcentratedLockupDenom(poolIdEntering, positionId)
		valAddr := strings.Split(synthLockBeforeMigration[0].SynthDenom, "/")[4]
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
		newLock, err := k.lk.CreateLock(ctx, sender, sdk.NewCoins(remainingGammShares), remainingLockTime)
		gammLockId = newLock.ID
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
		// If the gamm lock was superfluid unbonding, get the previous gamm intermediary account, create a new gamm synthetic lockup, and set it to unlocking
		if len(synthLockBeforeMigration) > 0 && strings.Contains(synthLockBeforeMigration[0].SynthDenom, "superunbonding") {
			valAddr := strings.Split(synthLockBeforeMigration[0].SynthDenom, "/")[4]
			gammIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, remainingGammShares.Denom, valAddr)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
			err = k.createSyntheticLockup(ctx, gammLockId, gammIntermediateAccount, unlockingStatus)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
		}
		// If the gamm lock was superfluid bonded, superfluid delegate the gamm like like normal
		if len(synthLockBeforeMigration) > 0 && strings.Contains(synthLockBeforeMigration[0].SynthDenom, "superbonding") {
			valAddr := strings.Split(synthLockBeforeMigration[0].SynthDenom, "/")[4]
			err := k.SuperfluidDelegate(ctx, sender.String(), gammLockId, valAddr)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
		}
		// If the gamm lock was unlocking, we begin the unlock from where it left off.
		if preUnlockLock.IsUnlocking() {
			_, err := k.lk.BeginForceUnlock(ctx, newLock.ID, newLock.Coins)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
		}
	}

	return positionId, amount0, amount1, liquidity, joinTime, poolIdLeaving, poolIdEntering, gammLockId, concentratedLockId, nil
}
