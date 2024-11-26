package keeper

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	gammtypes "github.com/osmosis-labs/osmosis/v27/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

type MigrationType int

const (
	SuperfluidBonded MigrationType = iota
	SuperfluidUnbonding
	NonSuperfluid
	Unlocked
	Unsupported
)

// RouteLockedBalancerToConcentratedMigration routes the provided lock to the proper migration function based on the lock status.
// The testing conditions and scope for the different lock status are as follows:
// Lock Status = Superfluid delegated
// - cannot migrate partial shares
// - Instantly undelegate which will bypass unbonding time.
// - Create new CL Lock and Re-delegate it as a concentrated liquidity position.
//
// Lock Status = Superfluid undelegating
// - cannot migrate partial shares
// - Continue undelegating as superfluid unbonding CL Position.
// - Lock the tokens and create an unlocking syntheticLock (to handle cases of slashing)
//
// Lock Status = Locked or unlocking (no superfluid delegation/undelegation)
// - cannot migrate partial shares
// - Force unlock tokens from gamm shares.
// - Create new CL lock and starts unlocking or unlocking where it left off.
//
// Lock Status = Unlocked
// - can migrate partial shares
// - For ex: LP shares
// - Create new CL lock and starts unlocking or unlocking where it left off.
//
// Errors if the lock is not found, if the lock is not a balancer pool lock, or if the lock is not owned by the sender.
func (k Keeper) RouteLockedBalancerToConcentratedMigration(ctx sdk.Context, sender sdk.AccAddress, providedLockId int64, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins) (positionData cltypes.CreateFullRangePositionData, migratedPoolIDs types.MigrationPoolIDs, concentratedLockId uint64, err error) {
	synthLockBeforeMigration, migrationType, err := k.getMigrationType(ctx, providedLockId)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, types.MigrationPoolIDs{}, 0, err
	}

	// As a hack around to get frontend working, we decided to allow negative values for the provided lock ID to indicate that the user wants to migrate shares that are not locked.
	lockId := uint64(providedLockId)

	switch migrationType {
	case SuperfluidBonded:
		positionData, concentratedLockId, migratedPoolIDs, err = k.migrateSuperfluidBondedBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, synthLockBeforeMigration.SynthDenom, tokenOutMins)
	case SuperfluidUnbonding:
		positionData, concentratedLockId, migratedPoolIDs, err = k.migrateSuperfluidUnbondingBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, synthLockBeforeMigration.SynthDenom, tokenOutMins)
	case NonSuperfluid:
		positionData, concentratedLockId, migratedPoolIDs, err = k.migrateNonSuperfluidLockBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, tokenOutMins)
	case Unlocked:
		positionData, migratedPoolIDs, err = k.gk.MigrateUnlockedPositionFromBalancerToConcentrated(ctx, sender, sharesToMigrate, tokenOutMins)
		concentratedLockId = 0
	default:
		return cltypes.CreateFullRangePositionData{}, types.MigrationPoolIDs{}, 0, errors.New("unsupported migration type")
	}

	return positionData, migratedPoolIDs, concentratedLockId, err
}

// migrateSuperfluidBondedBalancerToConcentrated migrates a user's superfluid bonded balancer position to a superfluid bonded concentrated liquidity position.
// The function first undelegates the superfluid delegated position, force unlocks and exits the balancer pool, creates a full range concentrated liquidity position, locks it, then superfluid delegates it.
// Any remaining gamm shares stay locked in the original gamm pool (utilizing the same lock and lockID that the shares originated from) and remain superfluid delegated / undelegating / vanilla locked as they
// were  when the migration was initiated. The function returns the concentrated liquidity position ID, amounts of tokens in the position, the liquidity amount, join time, and IDs of the involved pools and locks.
func (k Keeper) migrateSuperfluidBondedBalancerToConcentrated(ctx sdk.Context,
	sender sdk.AccAddress,
	originalLockId uint64,
	sharesToMigrate sdk.Coin,
	synthDenomBeforeMigration string,
	tokenOutMins sdk.Coins,
) (cltypes.CreateFullRangePositionData, uint64, types.MigrationPoolIDs, error) {
	migratedPoolIDs, preMigrationLock, remainingLockTime, err := k.validateMigration(ctx, sender, originalLockId, sharesToMigrate)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Get the validator address from the synth denom and ensure it is a valid address.
	valAddr := strings.Split(synthDenomBeforeMigration, "/")[4]
	_, err = sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Superfluid undelegate the portion of shares the user is migrating from the superfluid delegated position.
	// If all shares are being migrated, this deletes the connection between the gamm lock and the intermediate account, deletes the synthetic lock, and burns the synthetic osmo.
	intermediateAccount, err := k.SuperfluidUndelegateToConcentratedPosition(ctx, sender.String(), originalLockId)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	// It also returns the lock object that contains the remaining shares that were not used in this migration.
	exitCoins, err := k.forceUnlockAndExitBalancerPool(ctx, sender, migratedPoolIDs.LeavingID, preMigrationLock, sharesToMigrate, tokenOutMins, true)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Create a full range (min to max tick) concentrated liquidity position, lock it, and superfluid delegate it.
	positionData, concentratedLockId, err := k.clk.CreateFullRangePositionLocked(ctx, migratedPoolIDs.EnteringID, sender, exitCoins, remainingLockTime)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	err = k.SuperfluidDelegate(ctx, sender.String(), concentratedLockId, intermediateAccount.ValAddr)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	return positionData, concentratedLockId, migratedPoolIDs, nil
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
) (cltypes.CreateFullRangePositionData, uint64, types.MigrationPoolIDs, error) {
	migratedPoolIDs, preMigrationLock, remainingLockTime, err := k.validateMigration(ctx, sender, lockId, sharesToMigrate)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Get the validator address from the synth denom and ensure it is a valid address.
	valAddr := strings.Split(synthDenomBeforeMigration, "/")[4]
	_, err = sdk.ValAddressFromBech32(valAddr)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	// It also returns the lock object that contains the remaining shares that were not used in this migration.
	exitCoins, err := k.forceUnlockAndExitBalancerPool(ctx, sender, migratedPoolIDs.LeavingID, preMigrationLock, sharesToMigrate, tokenOutMins, true)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Create a full range (min to max tick) concentrated liquidity position.
	positionData, concentratedLockId, err := k.clk.CreateFullRangePositionUnlocking(ctx, migratedPoolIDs.EnteringID, sender, exitCoins, remainingLockTime)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// The previous gamm intermediary account is now invalid for the new lock, since the underlying denom has changed and intermediary accounts are
	// created by validator address, denom, and gauge id.
	// We must therefore create and set a new intermediary account based on the previous validator but with the new lock's denom.
	concentratedLockupDenom := cltypes.GetConcentratedLockupDenomFromPoolId(migratedPoolIDs.EnteringID)
	clIntermediateAccount, err := k.GetOrCreateIntermediaryAccount(ctx, concentratedLockupDenom, valAddr)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Synthetic lock is created to indicate unbonding position. The synthetic lock will be in unbonding period for remainingLockTime.
	// Create a new synthetic lockup for the new intermediary account in an unlocking status for the remaining duration.
	err = k.createSyntheticLockupWithDuration(ctx, concentratedLockId, clIntermediateAccount, remainingLockTime, unlockingStatus)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	return positionData, concentratedLockId, migratedPoolIDs, nil
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
) (cltypes.CreateFullRangePositionData, uint64, types.MigrationPoolIDs, error) {
	migratedPoolIDs, preMigrationLock, remainingLockTime, err := k.validateMigration(ctx, sender, lockId, sharesToMigrate)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Force unlock, validate the provided sharesToMigrate, and exit the balancer pool.
	// This will return the coins that will be used to create the concentrated liquidity position.
	// It also returns the lock object that contains the remaining shares that were not used in this migration.
	exitCoins, err := k.forceUnlockAndExitBalancerPool(ctx, sender, migratedPoolIDs.LeavingID, preMigrationLock, sharesToMigrate, tokenOutMins, true)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	// Create a new lock that is unlocking for the remaining time of the old lock.
	// Regardless of the previous lock's status, we create a new lock that is unlocking.
	// This is because locking without superfluid is pointless in the context of concentrated liquidity.
	positionData, concentratedLockId, err := k.clk.CreateFullRangePositionUnlocking(ctx, migratedPoolIDs.EnteringID, sender, exitCoins, remainingLockTime)
	if err != nil {
		return cltypes.CreateFullRangePositionData{}, 0, types.MigrationPoolIDs{}, err
	}

	return positionData, concentratedLockId, migratedPoolIDs, nil
}

// getMigrationType determines the status of the provided lock which is used to determine the method for migration.
// It also returns the underlying synthetic locks of the provided lock, if any exist.
func (k Keeper) getMigrationType(ctx sdk.Context, providedLockId int64) (synthLockBeforeMigration lockuptypes.SyntheticLock, migrationType MigrationType, err error) {
	// As a hack around to get frontend working, we decided to allow negative values for the provided lock ID to indicate that the user wants to migrate shares that are not locked.
	if providedLockId <= 0 {
		return lockuptypes.SyntheticLock{}, Unlocked, nil
	}

	lockId := uint64(providedLockId)

	synthLockBeforeMigration, _, err = k.lk.GetSyntheticLockupByUnderlyingLockId(ctx, lockId)
	if err != nil {
		return lockuptypes.SyntheticLock{}, Unsupported, err
	}

	// TODO: Change to if !found
	if synthLockBeforeMigration == (lockuptypes.SyntheticLock{}) {
		migrationType = NonSuperfluid
	} else if strings.Contains(synthLockBeforeMigration.SynthDenom, "superbonding") {
		migrationType = SuperfluidBonded
	} else if strings.Contains(synthLockBeforeMigration.SynthDenom, "superunbonding") {
		migrationType = SuperfluidUnbonding
	} else {
		return lockuptypes.SyntheticLock{}, Unsupported, fmt.Errorf("lock %d contains an unsupported synthetic lock", lockId)
	}

	return synthLockBeforeMigration, migrationType, nil
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
func (k Keeper) validateMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (types.MigrationPoolIDs, *lockuptypes.PeriodLock, time.Duration, error) {
	// Defense in depth, ensuring the sharesToMigrate contains gamm pool share prefix.
	if !strings.HasPrefix(sharesToMigrate.Denom, gammtypes.GAMMTokenPrefix) {
		return types.MigrationPoolIDs{}, &lockuptypes.PeriodLock{}, 0, types.SharesToMigrateDenomPrefixError{Denom: sharesToMigrate.Denom, ExpectedDenomPrefix: gammtypes.GAMMTokenPrefix}
	}

	// Get the balancer poolId by parsing the gamm share denom.
	poolIdLeaving, err := gammtypes.GetPoolIdFromShareDenom(sharesToMigrate.Denom)
	if err != nil {
		return types.MigrationPoolIDs{}, &lockuptypes.PeriodLock{}, 0, err
	}

	// Ensure a governance sanctioned link exists between the balancer pool and a concentrated pool.
	poolIdEntering, err := k.gk.GetLinkedConcentratedPoolID(ctx, poolIdLeaving)
	if err != nil {
		return types.MigrationPoolIDs{}, &lockuptypes.PeriodLock{}, 0, err
	}

	// Check that lockID corresponds to sender and that the denomination of LP shares corresponds to the poolId.
	preMigrationLock, err := k.validateGammLockForSuperfluidStaking(ctx, sender, poolIdLeaving, lockId)
	if err != nil {
		return types.MigrationPoolIDs{}, &lockuptypes.PeriodLock{}, 0, err
	}

	// Before we break the lock, we must note the time remaining on the lock.
	remainingLockTime, err := k.getExistingLockRemainingDuration(ctx, preMigrationLock)
	if err != nil {
		return types.MigrationPoolIDs{}, &lockuptypes.PeriodLock{}, 0, err
	}

	return types.MigrationPoolIDs{
		EnteringID: poolIdEntering,
		LeavingID:  poolIdLeaving,
	}, preMigrationLock, remainingLockTime, nil
}

// forceUnlockAndExitBalancerPool validates the unlocking and exiting of gamm LP tokens from the Balancer pool. It performs the following steps:
//
// 1. Completes the unlocking process / deletes synthetic locks for the provided lock.
// 2. If shares to migrate are not specified, all shares in the lock are migrated.
// 3. Ensures that the number of shares to migrate is less than or equal to the number of shares in the lock.
// 4. Exits the position in the Balancer pool.
// 5. Ensures that exactly two coins are returned.
// 6. Any remaining shares that were not migrated are re-locked as a new lock for the remaining time on the lock.
func (k Keeper) forceUnlockAndExitBalancerPool(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving uint64, lock *lockuptypes.PeriodLock, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins, exitCoinsLengthIsTwo bool) (exitCoins sdk.Coins, err error) {
	// validateMigration ensures that the preMigrationLock contains coins of length 1.
	gammSharesInLock := lock.Coins[0]

	// If shares to migrate is not specified, we migrate all shares.
	if sharesToMigrate.IsZero() {
		sharesToMigrate = gammSharesInLock
	}

	// Otherwise, we must ensure that the shares to migrate is less than or equal to the shares in the lock.
	if sharesToMigrate.Amount.GT(gammSharesInLock.Amount) {
		return sdk.Coins{}, types.MigrateMoreSharesThanLockHasError{SharesToMigrate: sharesToMigrate.Amount.String(), SharesInLock: gammSharesInLock.Amount.String()}
	}

	// Finish unlocking directly for locked or unlocking locks
	if !sharesToMigrate.Equal(gammSharesInLock) {
		return sdk.Coins{}, types.MigratePartialSharesError{SharesToMigrate: sharesToMigrate.Amount.String(), SharesInLock: gammSharesInLock.Amount.String()}
	}

	// Force migrate, which breaks and deletes associated synthetic locks (if exists).
	err = k.lk.ForceUnlock(ctx, *lock)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Exit the balancer pool position.
	exitCoins, err = k.gk.ExitPool(ctx, sender, poolIdLeaving, sharesToMigrate.Amount, tokenOutMins)
	if err != nil {
		return sdk.Coins{}, err
	}

	// if exit coins length should be two, check exitCoins length
	if exitCoinsLengthIsTwo && len(exitCoins) != 2 {
		return sdk.Coins{}, types.TwoTokenBalancerPoolError{NumberOfTokens: len(exitCoins)}
	}

	return exitCoins, nil
}
