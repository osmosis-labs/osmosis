package keeper

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v15/x/superfluid/types"
)

type MigrationType int

const (
	SuperfluidBonded MigrationType = iota
	SuperfluidUnbonding
	NonSuperfluid
	Unsupported
)

// RouteLockedBalancerToConcentratedMigration routes the provided lock to the proper migration function based on the lock status.
// If the lock is superfluid delegated, it will instantly undelegate the superfluid position and redelegate it as a concentrated liquidity position.
// If the lock is superfluid undelegating, it will instantly undelegate the superfluid position and redelegate it as a concentrated liquidity position, but continue to unlock where it left off.
// If the lock is locked or unlocking but not superfluid delegated/undelegating, it will migrate the position and either start unlocking or continue unlocking where it left off.
// Errors if the lock is not found, if the lock is not a balancer pool lock, or if the lock is not owned by the sender.
func (k Keeper) RouteLockedBalancerToConcentratedMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, poolIdLeaving, poolIdEntering, gammLockId, concentratedLockId uint64, err error) {
	synthLocksBeforeMigration, migrationType, err := k.routeMigration(ctx, sender, lockId, sharesToMigrate)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	switch migrationType {
	case SuperfluidBonded:
		positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering, err = k.migrateSuperfluidBondedBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, synthLocksBeforeMigration[0].SynthDenom, tokenOutMins)
	case SuperfluidUnbonding:
		positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering, err = k.migrateSuperfluidUnbondingBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, synthLocksBeforeMigration[0].SynthDenom, tokenOutMins)
	case NonSuperfluid:
		positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering, err = k.migrateNonSuperfluidLockBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, tokenOutMins)
	default:
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, fmt.Errorf("unsupported migration type")
	}

	return positionId, amount0, amount1, liquidity, joinTime, poolIdLeaving, poolIdEntering, gammLockId, concentratedLockId, err
}

// migrateSuperfluidBondedBalancerToConcentrated migrates a user's superfluid bonded balancer position to a superfluid bonded concentrated liquidity position.
// The function first undelegates the superfluid delegated position, force unlocks and exits the balancer pool, creates a full range concentrated liquidity position, locks it, then superfluid delegates it.
// If there are any remaining gamm shares, they are re-locked back in the gamm pool and superfluid delegated as normal. The function returns the concentrated liquidity position ID, amounts of
// tokens in the position, the liquidity amount, join time, and IDs of the involved pools and locks.
func (k Keeper) migrateSuperfluidBondedBalancerToConcentrated(ctx sdk.Context,
	sender sdk.AccAddress,
	lockId uint64,
	sharesToMigrate sdk.Coin,
	synthDenomBeforeMigration string,
	tokenOutMins sdk.Coins,
) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering uint64, err error) {
	poolIdLeaving, poolIdEntering, preMigrationLock, remainingLockTime, err := k.validateMigration(ctx, sender, lockId, sharesToMigrate)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Get the validator address from the synth denom and ensure it is a valid address.
	valAddr := strings.Split(synthDenomBeforeMigration, "/")[4]
	_, err = sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Superfluid undelegate the superfluid delegated position.
	// This deletes the connection between the gamm lock and the intermediate account, deletes the synthetic lock, and burns the synthetic osmo.
	intermediateAccount, err := k.SuperfluidUndelegateToConcentratedPosition(ctx, sender.String(), lockId)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	// It also returns the lock object that contains the remaining shares that were not used in this migration.
	exitCoins, remainingSharesLock, err := k.validateSharesToMigrateUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, preMigrationLock, sharesToMigrate, tokenOutMins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Create a full range (min to max tick) concentrated liquidity position, lock it, and superfluid delegate it.
	positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionLocked(ctx, poolIdEntering, sender, exitCoins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}
	err = k.SuperfluidDelegate(ctx, sender.String(), concentratedLockId, intermediateAccount.ValAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// If there are any remaining gamm shares after the migration, we must re-superfluid delegate them as they were previously in the gamm pool.
	if remainingSharesLock.ID != 0 {
		gammLockId = remainingSharesLock.ID
		// Superfluid delegate the gamm lock.
		err = k.SuperfluidDelegate(ctx, sender.String(), remainingSharesLock.ID, valAddr)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
	}

	return positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering, nil
}

// migrateSuperfluidUnbondingBalancerToConcentrated migrates a user's superfluid unbonding balancer position to a superfluid unbonding concentrated liquidity position.
// The function force unlocks and exits the balancer pool, creates a full range concentrated liquidity position, and locks it. If there are any remaining gamm shares, they are re-locked and begin unlocking where they left off.
// A new intermediate account and in turn synthetic lock based on the new cl share denom are created, since the old intermediate account and synthetic lock were based on the old gamm share denom.
// The remaining duration of the new lock equals to the duration of the pre-existing lock.
// The function returns the concentrated liquidity position ID, amounts of tokens in the position, the liquidity amount, join time, and IDs of the involved pools and locks.
func (k Keeper) migrateSuperfluidUnbondingBalancerToConcentrated(ctx sdk.Context,
	sender sdk.AccAddress,
	lockId uint64,
	sharesToMigrate sdk.Coin,
	synthDenomBeforeMigration string,
	tokenOutMins sdk.Coins,
) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering uint64, err error) {
	poolIdLeaving, poolIdEntering, preMigrationLock, remainingLockTime, err := k.validateMigration(ctx, sender, lockId, sharesToMigrate)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Get the validator address from the synth denom and ensure it is a valid address.
	valAddr := strings.Split(synthDenomBeforeMigration, "/")[4]
	_, err = sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Save unlocking state of lock before force unlocking
	wasUnlocking := preMigrationLock.IsUnlocking()

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	// It also returns the lock object that contains the remaining shares that were not used in this migration.
	exitCoins, remainingSharesLock, err := k.validateSharesToMigrateUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, preMigrationLock, sharesToMigrate, tokenOutMins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	// If the lock was unlocking, we create a new lock that is unlocking for the remaining time of the old lock.
	positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionUnlocking(ctx, poolIdEntering, sender, exitCoins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// The previous gamm intermediary account is now invalid for the new lock, since the underlying denom has changed and intermediary accounts are
	// created by validator address, denom, and gauge id.
	// We must therefore create and set a new intermediary account based on the previous validator but with the new lock's denom.
	concentratedLockupDenom := cltypes.GetConcentratedLockupDenomFromPoolId(poolIdEntering)
	clIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, concentratedLockupDenom, valAddr)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Create a new synthetic lockup for the new intermediary account in an unlocking status
	err = k.createSyntheticLockup(ctx, concentratedLockId, clIntermediateAccount, unlockingStatus)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// If there are any remaining gamm shares after the migration, we must re-create the synthetic lock and begin unlocking it from where it left off.
	if remainingSharesLock.ID != 0 {
		gammLockId = remainingSharesLock.ID
		// Get the previous gamm intermediary account, create a new gamm synthetic lockup, and set it to unlocking.
		gammIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, remainingSharesLock.Coins[0].Denom, valAddr)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}
		err = k.createSyntheticLockup(ctx, gammLockId, gammIntermediateAccount, unlockingStatus)
		if err != nil {
			return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
		}

		// If lock was previously unlocking, begin the unlock from where it left off.
		if wasUnlocking {
			_, err = k.lk.BeginForceUnlock(ctx, remainingSharesLock.ID, remainingSharesLock.Coins)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
		}
	}

	return positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering, nil
}

// migrateNonSuperfluidLockBalancerToConcentrated migrates a user's non-superfluid locked or unlocking balancer position to an unlocking concentrated liquidity position.
// The function force unlocks and exits the balancer pool, creates a full range concentrated liquidity position, locks it, and begins unlocking from where the locked or unlocking lock left off.
// If there are any remaining gamm shares, they are re-locked back in the gamm pool. The function returns the concentrated liquidity position ID, amounts of tokens in the position,
// the liquidity amount, join time, and IDs of the involved pools and locks.
func (k Keeper) migrateNonSuperfluidLockBalancerToConcentrated(ctx sdk.Context,
	sender sdk.AccAddress,
	lockId uint64,
	sharesToMigrate sdk.Coin,
	tokenOutMins sdk.Coins,
) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering uint64, err error) {
	poolIdLeaving, poolIdEntering, preMigrationLock, remainingLockTime, err := k.validateMigration(ctx, sender, lockId, sharesToMigrate)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}
	// Save unlocking state of lock before force unlocking
	wasUnlocking := preMigrationLock.IsUnlocking()

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	// It also returns the lock object that contains the remaining shares that were not used in this migration.
	exitCoins, remainingSharesLock, err := k.validateSharesToMigrateUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, preMigrationLock, sharesToMigrate, tokenOutMins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// Create a new lock that is unlocking for the remaining time of the old lock.
	// Regardless of the previous lock's status, we create a new lock that is unlocking.
	// This is because locking without superfluid is pointless in the context of concentrated liquidity.
	positionId, amount0, amount1, liquidity, joinTime, concentratedLockId, err = k.clk.CreateFullRangePositionUnlocking(ctx, poolIdEntering, sender, exitCoins, remainingLockTime)
	if err != nil {
		return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
	}

	// If there are remaining gamm shares, we must re-lock them.
	if remainingSharesLock.ID != 0 {
		gammLockId = remainingSharesLock.ID
		// If the gamm lock was unlocking, we begin the unlock from where it left off.
		if wasUnlocking {
			_, err := k.lk.BeginForceUnlock(ctx, remainingSharesLock.ID, remainingSharesLock.Coins)
			if err != nil {
				return 0, sdk.Int{}, sdk.Int{}, sdk.Dec{}, time.Time{}, 0, 0, 0, 0, err
			}
		}
	}

	return positionId, amount0, amount1, liquidity, joinTime, gammLockId, concentratedLockId, poolIdLeaving, poolIdEntering, nil
}

// routeMigration determines the status of the provided lock which is used to determine the method for migration.
// It also returns the underlying synthetic locks of the provided lock, if any exist.
func (k Keeper) routeMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (synthLocksBeforeMigration []lockuptypes.SyntheticLock, migrationType MigrationType, err error) {
	synthLocksBeforeMigration = k.lk.GetAllSyntheticLockupsByLockup(ctx, lockId)
	migrationType = NonSuperfluid

	for _, synthLockBeforeMigration := range synthLocksBeforeMigration {
		if strings.Contains(synthLockBeforeMigration.SynthDenom, "superbonding") {
			migrationType = SuperfluidBonded
		}
		if strings.Contains(synthLockBeforeMigration.SynthDenom, "superunbonding") {
			migrationType = SuperfluidUnbonding
		}
		if strings.Contains(synthLockBeforeMigration.SynthDenom, "superbonding") && strings.Contains(synthLockBeforeMigration.SynthDenom, "superunbonding") {
			return nil, Unsupported, fmt.Errorf("lock %d contains both superfluid bonded and unbonded tokens", lockId)
		}
	}

	return synthLocksBeforeMigration, migrationType, nil
}

// validateMigration performs validation for the migration of gamm LP tokens from a Balancer pool to the canonical Concentrated pool. It performs the following steps:
//
// 1. Gets the pool ID of the Balancer pool from the gamm share denomination.
// 2. Ensures a governance-sanctioned link exists between the Balancer pool and the Concentrated pool.
// 3. Validates that the provided lock corresponds to the sender and contains the correct denomination of LP shares.
// 4. Determines the remaining time on the lock.
//
// The function returns the following values:
//
// poolIdLeaving: The ID of the balancer pool being migrated from.
// poolIdEntering: The ID of the concentrated pool being migrated to.
// preMigrationLock: The original lock before migration.
// remainingLockTime: The remaining time on the lock before it expires.
// err: An error, if any occurred.
func (k Keeper) validateMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (poolIdLeaving, poolIdEntering uint64, preMigrationLock *lockuptypes.PeriodLock, remainingLockTime time.Duration, err error) {
	// Defense in depth, ensuring the sharesToMigrate contains gamm pool share prefix.
	if !strings.HasPrefix(sharesToMigrate.Denom, gammtypes.GAMMTokenPrefix) {
		return 0, 0, &lockuptypes.PeriodLock{}, 0, types.SharesToMigrateDenomPrefixError{Denom: sharesToMigrate.Denom, ExpectedDenomPrefix: gammtypes.GAMMTokenPrefix}
	}

	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving, err = gammtypes.GetPoolIdFromShareDenom(sharesToMigrate.Denom)
	if err != nil {
		return 0, 0, &lockuptypes.PeriodLock{}, 0, err
	}

	// Ensure a governance sanctioned link exists between the balancer pool and a concentrated pool.
	poolIdEntering, err = k.gk.GetLinkedConcentratedPoolID(ctx, poolIdLeaving)
	if err != nil {
		return 0, 0, &lockuptypes.PeriodLock{}, 0, err
	}

	// Further defense in depth, ensuring that the pool ID we are entering can be type cased to a concentrated pool extension.
	_, err = k.clk.GetConcentratedPoolById(ctx, poolIdEntering)
	if err != nil {
		return 0, 0, &lockuptypes.PeriodLock{}, 0, err
	}

	// Check that lockID corresponds to sender, and contains correct denomination of LP shares.
	preMigrationLock, err = k.validateLockForUnpool(ctx, sender, poolIdLeaving, lockId)
	if err != nil {
		return 0, 0, &lockuptypes.PeriodLock{}, 0, err
	}

	// Before we break the lock, we must note the time remaining on the lock.
	remainingLockTime = k.getExistingLockRemainingDuration(ctx, preMigrationLock)

	return poolIdLeaving, poolIdEntering, preMigrationLock, remainingLockTime, nil
}

// validateSharesToMigrateUnlockAndExitBalancerPool validates the unlocking and exiting of gamm LP tokens from the Balancer pool. It performs the following steps:
//
// 1. Completes the unlocking process / deletes synthetic locks for the provided lock.
// 2. If shares to migrate are not specified, all shares in the lock are migrated.
// 3. Ensures that the number of shares to migrate is less than or equal to the number of shares in the lock.
// 4. Exits the position in the Balancer pool.
// 5. Ensures that exactly two coins are returned.
// 6. Any remaining shares that were not migrated are re-locked as a new lock for the remaining time on the lock.
func (k Keeper) validateSharesToMigrateUnlockAndExitBalancerPool(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving uint64, lock *lockuptypes.PeriodLock, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins, remainingLockTime time.Duration) (exitCoins sdk.Coins, remainingSharesLock lockuptypes.PeriodLock, err error) {
	// validateMigration ensures that the preMigrationLock contains coins of length 1.
	gammSharesInLock := lock.Coins[0]

	// If shares to migrate is not specified, we migrate all shares.
	if sharesToMigrate.IsZero() {
		sharesToMigrate = gammSharesInLock
	}

	// Otherwise, we must ensure that the shares to migrate is less than or equal to the shares in the lock.
	if sharesToMigrate.Amount.GT(gammSharesInLock.Amount) {
		return sdk.Coins{}, lockuptypes.PeriodLock{}, types.MigrateMoreSharesThanLockHasError{SharesToMigrate: sharesToMigrate.Amount.String(), SharesInLock: gammSharesInLock.Amount.String()}
	}

	// Determine if there will be any remaining gamm shares after migration.
	remainingGammShares := gammSharesInLock.Sub(sharesToMigrate)

	// Finish unlocking directly for locked or unlocking locks
	// This also breaks and deletes associated synthetic locks.
	err = k.lk.ForceUnlock(ctx, *lock)
	if err != nil {
		return sdk.Coins{}, lockuptypes.PeriodLock{}, err
	}

	// Exit the balancer pool position.
	exitCoins, err = k.gk.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, tokenOutMins)
	if err != nil {
		return sdk.Coins{}, lockuptypes.PeriodLock{}, err
	}

	// Defense in depth, ensuring we are returning exactly two coins.
	if len(exitCoins) != 2 {
		return sdk.Coins{}, lockuptypes.PeriodLock{}, types.TwoTokenBalancerPoolError{NumberOfTokens: len(exitCoins)}
	}

	// If there is a remainder of gamm shares, create a new lock with the remaining gamm shares for the remaining lock time.
	if !remainingGammShares.IsZero() {
		remainingSharesLock, err = k.lk.CreateLock(ctx, sender, sdk.NewCoins(remainingGammShares), remainingLockTime)
		if err != nil {
			return sdk.Coins{}, lockuptypes.PeriodLock{}, err
		}
	}

	return exitCoins, remainingSharesLock, nil
}
