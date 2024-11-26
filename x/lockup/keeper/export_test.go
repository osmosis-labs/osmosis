package keeper

import (
	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) AddLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	return k.addLockRefByKey(ctx, key, lockID)
}

func (k Keeper) DeleteLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) {
	k.deleteLockRefByKey(ctx, key, lockID)
}

func (k Keeper) GetLockRefs(ctx sdk.Context, key []byte) []uint64 {
	return k.getLockRefs(ctx, key)
}

func (k Keeper) GetCoinsFromLocks(locks []types.PeriodLock) sdk.Coins {
	return k.getCoinsFromLocks(locks)
}

func (k Keeper) Lock(ctx sdk.Context, lock types.PeriodLock, tokensToLock sdk.Coins) error {
	return k.lock(ctx, lock, tokensToLock)
}

func (k Keeper) UnlockMaturedLockInternalLogic(ctx sdk.Context, lock types.PeriodLock) error {
	return k.unlockMaturedLockInternalLogic(ctx, lock)
}
