package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v15/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v15/x/lockup/types"
)

var (
	StakingSyntheticDenom   = stakingSyntheticDenom
	UnstakingSyntheticDenom = unstakingSyntheticDenom
)

func (k Keeper) ValidateLockForSFDelegate(ctx sdk.Context, lock *lockuptypes.PeriodLock, sender string) error {
	return k.validateLockForSFDelegate(ctx, lock, sender)
}

func (k Keeper) PrepareConcentratedLockForSlash(ctx sdk.Context, lock *lockuptypes.PeriodLock, slashAmt sdk.Dec) (sdk.AccAddress, sdk.Coins, error) {
	return k.prepareConcentratedLockForSlash(ctx, lock, slashAmt)
}

func (k Keeper) MigrateSuperfluidBondedBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving, poolIdEntering uint64, preMigrationLock *lockuptypes.PeriodLock, lockId uint64, sharesToMigrate sdk.Coin, synthDenomBeforeMigration string, concentratedPool cltypes.ConcentratedPoolExtension, remainingLockTime time.Duration, tokenOutMins sdk.Coins) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId uint64, err error) {
	return k.migrateSuperfluidBondedBalancerToConcentrated(ctx, sender, poolIdLeaving, poolIdEntering, preMigrationLock, lockId, sharesToMigrate, synthDenomBeforeMigration, concentratedPool, remainingLockTime, tokenOutMins)
}

func (k Keeper) MigrateSuperfluidUnbondingBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving, poolIdEntering uint64, preMigrationLock *lockuptypes.PeriodLock, sharesToMigrate sdk.Coin, synthDenomBeforeMigration string, concentratedPool cltypes.ConcentratedPoolExtension, remainingLockTime time.Duration, tokenOutMins sdk.Coins) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId uint64, err error) {
	return k.migrateSuperfluidUnbondingBalancerToConcentrated(ctx, sender, poolIdLeaving, poolIdEntering, preMigrationLock, sharesToMigrate, synthDenomBeforeMigration, concentratedPool, remainingLockTime, tokenOutMins)
}

func (k Keeper) MigrateNonSuperfluidLockBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving, poolIdEntering uint64, preMigrationLock *lockuptypes.PeriodLock, sharesToMigrate sdk.Coin, concentratedPool cltypes.ConcentratedPoolExtension, remainingLockTime time.Duration, tokenOutMins sdk.Coins) (positionId uint64, amount0, amount1 sdk.Int, liquidity sdk.Dec, joinTime time.Time, gammLockId, concentratedLockId uint64, err error) {
	return k.migrateNonSuperfluidLockBalancerToConcentrated(ctx, sender, poolIdLeaving, poolIdEntering, preMigrationLock, sharesToMigrate, concentratedPool, remainingLockTime, tokenOutMins)
}

func (k Keeper) ValidateSharesToMigrateUnlockAndExitBalancerPool(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving uint64, lock *lockuptypes.PeriodLock, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins) (exitCoins sdk.Coins, err error) {
	return k.validateSharesToMigrateUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, lock, sharesToMigrate, tokenOutMins)
}

func (k Keeper) PrepareMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (poolIdLeaving, poolIdEntering uint64, concentratedPool cltypes.ConcentratedPoolExtension, preMigrationLock *lockuptypes.PeriodLock, remainingLockTime time.Duration, synthLockBeforeMigration []lockuptypes.SyntheticLock, isSuperfluidBonded, isSuperfluidUnbonding bool, err error) {
	return k.prepareMigration(ctx, sender, lockId, sharesToMigrate)
}

func (k Keeper) AddToConcentratedLiquiditySuperfluidPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, amount0Added, amount1Added sdk.Int) (uint64, sdk.Int, sdk.Int, sdk.Dec, uint64, error) {
	return k.addToConcentratedLiquiditySuperfluidPosition(ctx, owner, positionId, amount0Added, amount1Added)
}
