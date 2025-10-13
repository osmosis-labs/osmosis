package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	lockuptypes "github.com/osmosis-labs/osmosis/v31/x/lockup/types"
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
