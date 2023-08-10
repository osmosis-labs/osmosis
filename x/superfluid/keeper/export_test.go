package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cltypes "github.com/osmosis-labs/osmosis/v17/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v17/x/lockup/types"
	types "github.com/osmosis-labs/osmosis/v17/x/superfluid/types"
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

func (k Keeper) MigrateSuperfluidBondedBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin, synthDenomBeforeMigration string, tokenOutMins sdk.Coins) (cltypes.CreateFullRangePositionData, uint64, types.MigrationPoolIDs, error) {
	return k.migrateSuperfluidBondedBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, synthDenomBeforeMigration, tokenOutMins)
}

func (k Keeper) MigrateSuperfluidUnbondingBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin, synthDenomBeforeMigration string, tokenOutMins sdk.Coins) (cltypes.CreateFullRangePositionData, uint64, types.MigrationPoolIDs, error) {
	return k.migrateSuperfluidUnbondingBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, synthDenomBeforeMigration, tokenOutMins)
}

func (k Keeper) MigrateNonSuperfluidLockBalancerToConcentrated(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins) (cltypes.CreateFullRangePositionData, uint64, types.MigrationPoolIDs, error) {
	return k.migrateNonSuperfluidLockBalancerToConcentrated(ctx, sender, lockId, sharesToMigrate, tokenOutMins)
}

func (k Keeper) ValidateSharesToMigrateUnlockAndExitBalancerPool(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving uint64, lock *lockuptypes.PeriodLock, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins) (exitCoins sdk.Coins, err error) {
	return k.validateSharesToMigrateUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, lock, sharesToMigrate, tokenOutMins)
}

func (k Keeper) RouteMigration(ctx sdk.Context, sender sdk.AccAddress, lockId int64, sharesToMigrate sdk.Coin) (synthLockBeforeMigration lockuptypes.SyntheticLock, migrationType MigrationType, err error) {
	return k.routeMigration(ctx, sender, lockId, sharesToMigrate)
}

func (k Keeper) ValidateMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (types.MigrationPoolIDs, *lockuptypes.PeriodLock, time.Duration, error) {
	return k.validateMigration(ctx, sender, lockId, sharesToMigrate)
}

func (k Keeper) AddToConcentratedLiquiditySuperfluidPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, amount0Added, amount1Added sdk.Int) (cltypes.CreateFullRangePositionData, uint64, error) {
	return k.addToConcentratedLiquiditySuperfluidPosition(ctx, owner, positionId, amount0Added, amount1Added)
}

func (k Keeper) ValidateGammLockForSuperfluidStaking(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, lockId uint64) (*lockuptypes.PeriodLock, error) {
	return k.validateGammLockForSuperfluidStaking(ctx, sender, poolId, lockId)
}

func (k Keeper) GetExistingLockRemainingDuration(ctx sdk.Context, lock *lockuptypes.PeriodLock) (time.Duration, error) {
	return k.getExistingLockRemainingDuration(ctx, lock)
}

func (k Keeper) DistributeSuperfluidGauges(ctx sdk.Context) {
	k.distributeSuperfluidGauges(ctx)
}
