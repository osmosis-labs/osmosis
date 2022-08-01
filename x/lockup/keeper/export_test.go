package keeper

import (
	"github.com/osmosis-labs/osmosis/v10/x/lockup/types"

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

func (k Keeper) SyntheticCoins(coins sdk.Coins, suffix string) sdk.Coins {
	return syntheticCoins(coins, suffix)
}

func (k Keeper) GetCoinsFromLocks(locks []types.PeriodLock) sdk.Coins {
	return k.getCoinsFromLocks(locks)
}
