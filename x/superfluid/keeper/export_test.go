package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/osmomath"
	cltypes "github.com/osmosis-labs/osmosis/v27/x/concentrated-liquidity/types"
	lockuptypes "github.com/osmosis-labs/osmosis/v27/x/lockup/types"
	types "github.com/osmosis-labs/osmosis/v27/x/superfluid/types"
)

var (
	StakingSyntheticDenom   = stakingSyntheticDenom
	UnstakingSyntheticDenom = unstakingSyntheticDenom
)

func (k Keeper) ValidateLockForSFDelegate(ctx sdk.Context, lock *lockuptypes.PeriodLock, sender string) error {
	return k.validateLockForSFDelegate(ctx, lock, sender)
}

func (k Keeper) PrepareConcentratedLockForSlash(ctx sdk.Context, lock *lockuptypes.PeriodLock, slashAmt osmomath.Dec) (sdk.AccAddress, sdk.Coins, error) {
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

func (k Keeper) ForceUnlockAndExitBalancerPool(ctx sdk.Context, sender sdk.AccAddress, poolIdLeaving uint64, lock *lockuptypes.PeriodLock, sharesToMigrate sdk.Coin, tokenOutMins sdk.Coins, exitCoinsLengthIsTwo bool) (exitCoins sdk.Coins, err error) {
	return k.forceUnlockAndExitBalancerPool(ctx, sender, poolIdLeaving, lock, sharesToMigrate, tokenOutMins, exitCoinsLengthIsTwo)
}

func (k Keeper) GetMigrationType(ctx sdk.Context, lockId int64) (synthLockBeforeMigration lockuptypes.SyntheticLock, migrationType MigrationType, err error) {
	return k.getMigrationType(ctx, lockId)
}

func (k Keeper) ValidateMigration(ctx sdk.Context, sender sdk.AccAddress, lockId uint64, sharesToMigrate sdk.Coin) (types.MigrationPoolIDs, *lockuptypes.PeriodLock, time.Duration, error) {
	return k.validateMigration(ctx, sender, lockId, sharesToMigrate)
}

func (k Keeper) AddToConcentratedLiquiditySuperfluidPosition(ctx sdk.Context, owner sdk.AccAddress, positionId uint64, amount0Added, amount1Added osmomath.Int) (cltypes.CreateFullRangePositionData, uint64, error) {
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

func (k Keeper) ConvertLockToStake(ctx sdk.Context, sender sdk.AccAddress, valAddr string, lockId uint64,
	minAmtToStake osmomath.Int) (totalAmtConverted osmomath.Int, err error) {
	return k.convertLockToStake(ctx, sender, valAddr, lockId, minAmtToStake)
}

func (k Keeper) ConvertGammSharesToOsmoAndStake(
	ctx sdk.Context,
	sender sdk.AccAddress, valAddr string,
	poolIdLeaving uint64, exitCoins sdk.Coins, minAmtToStake osmomath.Int, originalSuperfluidValAddr string,
) (totalAmtCoverted osmomath.Int, err error) {
	return k.convertGammSharesToOsmoAndStake(ctx, sender, valAddr, poolIdLeaving, exitCoins, minAmtToStake, originalSuperfluidValAddr)
}

func (k Keeper) ConvertUnlockedToStake(ctx sdk.Context, sender sdk.AccAddress, valAddr string, sharesToStake sdk.Coin,
	minAmtToStake osmomath.Int) (totalAmtConverted osmomath.Int, err error) {
	return k.convertUnlockedToStake(ctx, sender, valAddr, sharesToStake, minAmtToStake)
}

func (k Keeper) DelegateBaseOnValsetPref(ctx sdk.Context, sender sdk.AccAddress, valAddr, originalSuperfluidValAddr string, totalAmtToStake osmomath.Int) error {
	return k.delegateBaseOnValsetPref(ctx, sender, valAddr, originalSuperfluidValAddr, totalAmtToStake)
}
