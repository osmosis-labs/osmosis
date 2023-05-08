package keeper

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
)

// RouteLockedBalancerToConcentratedMigration routes the provided lock to the proper migration function based on the lock status.
// If the lock is superfluid delegated, it will undelegate the superfluid position and redelegate it as a concentrated liquidity position.
// If the lock is superfluid undelegating, it will instantly undelegate the superfluid position and redelegate it as a concentrated liquidity position, but continue to unlock where it left off.
// If the lock is locked or unlocking but not superfluid delegated/undelegating, it will migrate the position and either start unlocking or continue unlocking where it left off.
// Errors if the lock is not found, if the lock is not a balancer pool lock, or if the lock is not owned by the sender.
func (k Keeper) RouteLockedBalancerToConcentratedMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, poolIdLeaving, poolIdEntering, gammLockId, concentratedLockId uint64, err error) {
	// Validate and retrieve pertinent data required for migration
	poolIdLeaving, poolIdEntering, concentratedPool, preMigrationLock, remainingLockTime, synthLockBeforeMigration, isSuperfluidBonded, isSuperfluidUnbonding, err := k.prepareMigration(ctx, sender, lockId, sharesToMigrate)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	if isSuperfluidBonded {
		// Migration logic for superfluid bonded locks
		positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, err = k.migrateSuperfluidBondedBalancerToConcentrated(ctx, sender, poolIdLeaving, poolIdEntering, preMigrationLock, lockId, sharesToMigrate, synthLockBeforeMigration[0].SynthDenom, concentratedPool, remainingLockTime, tokenOutMins)
	} else if isSuperfluidUnbonding {
		// Migration logic for superfluid unbonding locks
		positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, err = k.migrateSuperfluidUnbondingBalancerToConcentrated(ctx, sender, poolIdLeaving, poolIdEntering, preMigrationLock, sharesToMigrate, synthLockBeforeMigration[0].SynthDenom, concentratedPool, remainingLockTime, tokenOutMins)
	} else if !isSuperfluidBonded && !isSuperfluidUnbonding && len(synthLockBeforeMigration) == 0 {
		// Migration logic for non-superfluid locks
		positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, err = k.migrateNonSuperfluidLockBalancerToConcentrated(ctx, sender, poolIdLeaving, poolIdEntering, preMigrationLock, sharesToMigrate, concentratedPool, remainingLockTime, tokenOutMins)
	} else {
		// Unsupported migration
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, fmt.Errorf("unexpected synth lock state for lock %d", lockId)
	}
	return positionId, amount0, amount1, liquidity, joinTime, poolIdLeaving, poolIdEntering, gammLockId, concentratedLockId, err
}

// migrateSuperfluidBondedBalancerToConcentrated migrates a user's superfluid bonded balancer position to a superfluid bonded concentrated liquidity position.
// The function first undelegates the superfluid delegated position, force unlocks and exits the balancer pool, creates a full range concentrated liquidity position, and locks it while superfluid delegating it.
// If there are any remaining gamm shares, they are re-locked and superfluid delegated as normal. The function returns the concentrated liquidity position ID, amounts of
// tokens in the position, the liquidity amount, join time, and IDs of the involved pools and locks.
func (k Keeper) migrateSuperfluidBondedBalancerToConcentrated(ctx sdk.Context,
	sender sdk.AccAddress,
	poolIdLeaving, poolIdEntering uint64,
	preMigrationLock *lockuptypes.PeriodLock,
	lockId uint64,
	sharesToMigrate sdk.Coin,
	synthDenomBeforeMigration string,
	concentratedPool cltypes.ConcentratedPoolExtension,
	remainingLockTime time.Duration,
	tokenOutMins sdk.Coins,
) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId uint64, err error) {
	// Get the validator address from the synth denom and ensure it is a valid address.
	valAddr := strings.Split(synthDenomBeforeMigration, "/")[4]
	_, err = sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	gammSharesInLock := preMigrationLock.Coins[0]

	// Superfluid undelegate the superfluid delegated position.
	// This deletes the connection between the lock and the intermediate account, deletes the synthetic lock, and burns the synthetic osmo.
	intermediateAccount, err := k.SuperfluidUndelegateToConcentratedPosition(ctx, sender.String(), lockId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	exitCoins, err := k.validateSharesToMigrateUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, preMigrationLock, sharesToMigrate, tokenOutMins)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// Create a full range (min to max tick) concentrated liquidity position, lock it, and superfluid delegate it.
	positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionLocked(ctx, poolIdEntering, sender, exitCoins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}
	err = k.SuperfluidDelegate(ctx, sender.String(), concentratedLockId, intermediateAccount.ValAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// If there are remaining gamm shares, we must re-lock them.
	remainingGammShares := gammSharesInLock.Sub(sharesToMigrate)
	if !remainingGammShares.IsZero() {
		// Create a new lock with the remaining gamm shares for the remaining lock time.
		newLock, err := k.lk.CreateLock(ctx, sender, sdk.NewCoins(remainingGammShares), remainingLockTime)
		gammLockId = newLock.ID
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
		}

		// If the gamm lock was previously superfluid bonded, superfluid delegate the gamm like normal
		err = k.SuperfluidDelegate(ctx, sender.String(), gammLockId, valAddr)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
		}
	}

	return positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, nil
}

// migrateSuperfluidUnbondingBalancerToConcentrated migrates a user's superfluid unbonding balancer position to a superfluid unbonding concentrated liquidity position.
// The function force unlocks and exits the balancer pool, creates a full range concentrated liquidity position, and locks it. If there are any remaining gamm shares, they are re-locked and begin unlocking where they left off.
// A new intermediate account and in turn synthetic lock based on the new cl share denom are created, since the old intermediate account and synthetic lock were based on the old gamm share denom.
// The remaining duration of the new lock equals to the duration of the pre-existing lock.
// The function returns the concentrated liquidity position ID, amounts of tokens in the position, the liquidity amount, join time, and IDs of the involved pools and locks.
func (k Keeper) migrateSuperfluidUnbondingBalancerToConcentrated(ctx sdk.Context,
	sender sdk.AccAddress,
	poolIdLeaving, poolIdEntering uint64,
	preMigrationLock *lockuptypes.PeriodLock,
	sharesToMigrate sdk.Coin,
	synthDenomBeforeMigration string,
	concentratedPool cltypes.ConcentratedPoolExtension,
	remainingLockTime time.Duration,
	tokenOutMins sdk.Coins,
) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId uint64, err error) {
	// Get the validator address from the synth denom and ensure it is a valid address.
	valAddr := strings.Split(synthDenomBeforeMigration, "/")[4]
	_, err = sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	gammSharesInLock := preMigrationLock.Coins[0]

	// Save unlocking state of lock before force unlocking
	wasUnlocking := preMigrationLock.IsUnlocking()

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	exitCoins, err := k.validateSharesToMigrateUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, preMigrationLock, sharesToMigrate, tokenOutMins)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	// If the lock was unlocking, we create a new lock that is unlocking for the remaining time of the old lock.
	positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionUnlocking(ctx, poolIdEntering, sender, exitCoins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// The previous gamm intermediary account is now invalid for the new lock, since the underlying denom has changed and intermediary accounts are
	// created by validator address, denom, and gauge id.
	// We must therefore create and set a new intermediary account based on the previous validator but with the new lock's denom.
	concentratedLockupDenom := cltypes.GetConcentratedLockupDenomFromPoolId(poolIdEntering)
	clIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, concentratedLockupDenom, valAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// Create a new synthetic lockup for the new intermediary account in an unlocking status
	err = k.createSyntheticLockup(ctx, concentratedLockId, clIntermediateAccount, unlockingStatus)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// If there are remaining gamm shares, we must re-lock them.
	remainingGammShares := gammSharesInLock.Sub(sharesToMigrate)
	if !remainingGammShares.IsZero() {
		// Create a new lock with the remaining gamm shares for the remaining lock time.
		newLock, err := k.lk.CreateLock(ctx, sender, sdk.NewCoins(remainingGammShares), remainingLockTime)
		gammLockId = newLock.ID
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
		}

		// Get the previous gamm intermediary account, create a new gamm synthetic lockup, and set it to unlocking
		gammIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, remainingGammShares.Denom, valAddr)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
		}
		err = k.createSyntheticLockup(ctx, gammLockId, gammIntermediateAccount, unlockingStatus)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
		}

		// If the gamm lock was unlocking, we begin the unlock from where it left off.
		if wasUnlocking {
			_, err := k.lk.BeginForceUnlock(ctx, newLock.ID, newLock.Coins)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
			}
		}
	}

	return positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, nil
}

// migrateNonSuperfluidLockBalancerToConcentrated migrates a user's non-superfluid locked or unlocking balancer position to an unlocking concentrated liquidity position.
// The function force unlocks and exits the balancer pool, creates a full range concentrated liquidity position, locks it, and begins unlocking from where the locked or unlocking lock left off.
// If there are any remaining gamm shares, they are re-locked. The function returns the concentrated liquidity position ID, amounts of tokens in the position, the liquidity amount, join time, and IDs of the involved pools and locks.
func (k Keeper) migrateNonSuperfluidLockBalancerToConcentrated(ctx sdk.Context,
	sender sdk.AccAddress,
	poolIdLeaving, poolIdEntering uint64,
	preMigrationLock *lockuptypes.PeriodLock,
	sharesToMigrate sdk.Coin,
	concentratedPool cltypes.ConcentratedPoolExtension,
	remainingLockTime time.Duration,
	tokenOutMins sdk.Coins,
) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId uint64, err error) {
	// Save unlocking state of lock before force unlocking
	wasUnlocking := preMigrationLock.IsUnlocking()

	gammSharesInLock := preMigrationLock.Coins[0]

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	exitCoins, err := k.validateSharesToMigrateUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, preMigrationLock, sharesToMigrate, tokenOutMins)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// Create a new lock that is unlocking for the remaining time of the old lock.
	// Regardless of the previous lock's status, we create a new lock that is unlocking.
	positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionUnlocking(ctx, poolIdEntering, sender, exitCoins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
	}

	// If there are remaining gamm shares, we must re-lock them.
	remainingGammShares := gammSharesInLock.Sub(sharesToMigrate)
	if !remainingGammShares.IsZero() {
		// Create a new lock with the remaining gamm shares for the remaining lock time.
		newLock, err := k.lk.CreateLock(ctx, sender, sdk.NewCoins(remainingGammShares), remainingLockTime)
		gammLockId = newLock.ID
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
		}

		// If the gamm lock was unlocking, we begin the unlock from where it left off.
		if wasUnlocking {
			_, err := k.lk.BeginForceUnlock(ctx, newLock.ID, newLock.Coins)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, err
			}
		}
	}

	return positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, nil
}

// prepareMigration prepares for the migration of gamm LP tokens from the Balancer pool to the Concentrated pool. It performs the following steps:
//
// 1. Gets the pool ID of the Balancer pool from the gamm share denomination.
// 2. Ensures a governance-sanctioned link exists between the Balancer pool and the Concentrated pool.
// 3. Validates that the lock corresponds to the sender, contains the correct denomination of LP shares, and retrieves the gamm shares from the lock.
// 4. Determines the remaining time on the lock.
// 5. Checks if the lock has a corresponding synthetic lock, indicating it is superfluid delegated or undelegating.
//
// The function returns the following values:
//
// poolIdLeaving: The ID of the balancer pool being migrated from.
// poolIdEntering: The ID of the concentrated pool being migrated to.
// gammSharesInLock: The GAMM shares contained in the lock.
// concentratedPool: The concentrated pool that will be entered.
// preMigrationLock: The original lock before migration.
// remainingLockTime: The remaining time on the lock before it expires.
// synthLockBeforeMigration: The synthetic lock associated with the lock before migration, if any.
// isSuperfluidBonded: A boolean indicating if the lock is superfluid delegated.
// isSuperfluidUnbonding: A boolean indicating if the lock is superfluid undelegating.
// err: An error, if any occurred.
func (k Keeper) prepareMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (poolIdLeaving, poolIdEntering uint64, concentratedPool cltypes.ConcentratedPoolExtension, preMigrationLock *lockuptypes.PeriodLock, remainingLockTime time.Duration, synthLockBeforeMigration []lockuptypes.SyntheticLock, isSuperfluidBonded, isSuperfluidUnbonding bool, err error) {
	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving = gammtypes.MustGetPoolIdFromShareDenom(sharesToMigrate.Denom)

	// Ensure a governance sanctioned link exists between the balancer pool and the concentrated pool.
	poolIdEntering, err = k.gk.GetLinkedConcentratedPoolID(ctx, poolIdLeaving)
	if err != nil {
		return 0, 0, nil, &lockuptypes.PeriodLock{}, 0, nil, false, false, err
	}

	// Get the concentrated pool from the provided ID and type cast it to ConcentratedPoolExtension.
	concentratedPool, err = k.clk.GetPoolFromPoolIdAndConvertToConcentrated(ctx, poolIdEntering)
	if err != nil {
		return 0, 0, nil, &lockuptypes.PeriodLock{}, 0, nil, false, false, err
	}

	// Check that lockID corresponds to sender, and contains correct denomination of LP shares.
	preMigrationLock, err = k.validateLockForUnpool(ctx, sender, poolIdLeaving, lockId)
	if err != nil {
		return 0, 0, nil, &lockuptypes.PeriodLock{}, 0, nil, false, false, err
	}

	// Before we break the lock, we must note the time remaining on the lock.
	remainingLockTime = k.getExistingLockRemainingDuration(ctx, preMigrationLock)

	// Check if the lock has a corresponding synthetic lock.
	// Synthetic lock existence implies that the lock is superfluid delegated or undelegating.
	synthLockBeforeMigration = k.lk.GetAllSyntheticLockupsByLockup(ctx, lockId)

	isSuperfluidBonded = len(synthLockBeforeMigration) > 0 && strings.Contains(synthLockBeforeMigration[0].SynthDenom, "superbonding")
	isSuperfluidUnbonding = len(synthLockBeforeMigration) > 0 && strings.Contains(synthLockBeforeMigration[0].SynthDenom, "superunbonding")
	if isSuperfluidBonded && isSuperfluidUnbonding {
		// This should never happen, but if it does, we don't support it.
		return 0, 0, nil, &lockuptypes.PeriodLock{}, 0, nil, false, false, fmt.Errorf("synthetic lock %d must be either superfluid delegated or superfluid undelegating, not both", lockId)
	}

	return poolIdLeaving, poolIdEntering, concentratedPool, preMigrationLock, remainingLockTime, synthLockBeforeMigration, isSuperfluidBonded, isSuperfluidUnbonding, nil
}

// validateSharesToMigrateUnlockAndExitBalancerPool validates the unlocking and exiting of gamm LP tokens from the Balancer pool. It performs the following steps:
//
// 1. Completes the unlocking process / deletes synthetic locks for the provided lock.
// 2. If shares to migrate are not specified, all shares in the lock are migrated.
// 3. Ensures that the number of shares to migrate is less than or equal to the number of shares in the lock.
// 4. Exits the position in the Balancer pool.
// 5. Ensures that exactly two coins are returned.
func (k Keeper) validateSharesToMigrateUnlockAndExitBalancerPool(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving uint64, lock *lockuptypes.PeriodLock, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins) (exitCoins sdk.Coins, err error) {
	// Finish unlocking directly for locked or unlocking locks
	// This also breaks and deletes associated synthetic locks.
	err = k.lk.ForceUnlock(ctx, *lock)
	if err != nil {
		return sdk.Coins{}, err
	}

	gammSharesInLock := lock.Coins[0]

	// If shares to migrate is not specified, we migrate all shares.
	if sharesToMigrate.IsZero() {
		sharesToMigrate = gammSharesInLock
	}

	// Otherwise, we must ensure that the shares to migrate is less than or equal to the shares in the lock.
	if sharesToMigrate.Amount.GT(gammSharesInLock.Amount) {
		return sdk.Coins{}, fmt.Errorf("shares to migrate must be less than or equal to shares in lock")
	}

	// Exit the balancer pool position.
	exitCoins, err = k.gk.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, tokenOutMins)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Defense in depth, ensuring we are returning exactly two coins.
	if len(exitCoins) != 2 {
		return sdk.Coins{}, fmt.Errorf("Balancer pool must have exactly two tokens")
	}
	return exitCoins, nil
}
