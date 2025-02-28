package keeper

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v29/x/gamm/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v29/x/lockup/types"
	"github.com/osmosis-labs/osmosis/v29/x/superfluid/types"
)

type MigrationType int

const (
	SuperfluidBonded MigrationType = iota
	SuperfluidUnbonding
	NonSuperfluid
	Unlocked
	Unsupported
)

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
