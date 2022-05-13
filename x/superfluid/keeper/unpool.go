package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gammtypes "github.com/osmosis-labs/osmosis/v7/x/gamm/types"

	"github.com/osmosis-labs/osmosis/v7/x/superfluid/types"
)

func (k Keeper) UnpoolAllowedPools(ctx sdk.Context, sender sdk.AccAddress, poolId uint64, lockId uint64) (uint64, error) {
	// check if pool is whitelisted for unpool
	allowedPools := k.GetUnpoolAllowedPools(ctx)
	allowed := false

	for _, allowedPoolId := range allowedPools {
		if poolId == allowedPoolId {
			allowed = true
		}
	}

	if !allowed {
		return 0, types.ErrUnpoolNotAllowed
	}

	// check if lock is unlocking
	lock, err := k.lk.GetLockByID(ctx, lockId)
	if err != nil {
		return 0, err
	}

	gammShare := lock.Coins[0]
	if gammShare.Denom != gammtypes.GetPoolShareDenom(poolId) {
		return 0, types.ErrLockUnpoolNotAllowed
	}

	// check if the lock is superfluid delegated
	_, found := k.GetIntermediaryAccountFromLockId(ctx, lockId)
	if found {
		// superfluid undelegate first
		// this undelegates delegation, breaks synthetic locks and
		// create a new synthetic lock representing unstaking
		err = k.SuperfluidUndelegate(ctx, sender.String(), lock.ID)
		if err != nil {
			return 0, err
		}
		// we don't need to call `SuperfluidUnbondLock` here as we would unlock break the lock anyways
	}

	// finish unlocking directly for locked locks
	// this also unlocks locks that were in the unlocking queue
	err = k.lk.BreakLockForUnpool(ctx, *lock)
	if err != nil {
		return 0, err
	}

	exitCoins, err := k.gk.ExitPool(ctx, sender, poolId, gammShare.Amount, sdk.NewCoins())
	if err != nil {
		return 0, err
	}

	newLock, err := k.lk.CreateLock(ctx, sender, exitCoins, lock.Duration)
	if err != nil {
		return 0, err
	}

	initializedTime := time.Time{}
	// if the lock was unlocking, run separate logic to preserve lock endTime
	if lock.EndTime != initializedTime {
		err = k.lk.BeginForceUnlockWithEndTime(ctx, lockId, lock.EndTime)
		if err != nil {
			return 0, err
		}
	} else {
		err = k.lk.BeginForceUnlock(ctx, newLock.ID, newLock.Coins)
		if err != nil {
			return 0, err
		}
	}

	return newLock.ID, nil
}

func (k Keeper) GetUnpoolAllowedPools(ctx sdk.Context) []uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyUnpoolAllowedPools)
	if len(bz) == 0 {
		return []uint64{}
	}

	allowedPools := types.UnpoolWhitelistedPools{}
	k.cdc.MustUnmarshal(bz, &allowedPools)
	return allowedPools.Ids
}

func (k Keeper) SetUnpoolAllowedPools(ctx sdk.Context, poolIds []uint64) {
	store := ctx.KVStore(k.storeKey)

	allowedPools := types.UnpoolWhitelistedPools{
		Ids: poolIds,
	}

	bz := k.cdc.MustMarshal(&allowedPools)
	store.Set(types.KeyUnpoolAllowedPools, bz)
}
