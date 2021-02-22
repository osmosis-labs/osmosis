package keeper

import (
	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Relock unlock previous lockID and create a new lock with newCoins with same duration and endtime
func (ak AdminKeeper) Relock(ctx sdk.Context, lockID uint64, newCoins sdk.Coins) error {
	lock, err := ak.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// send original coins back to owner
	if err := ak.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lock.Owner, lock.Coins); err != nil {
		return err
	}

	// lock newCoins into module account
	if err := ak.bk.SendCoinsFromAccountToModule(ctx, lock.Owner, types.ModuleName, newCoins); err != nil {
		return err
	}

	// replace to new coins
	lock.Coins = newCoins

	// reset lock record inside store
	store := ctx.KVStore(ak.storeKey)
	store.Set(lockStoreKey(lockID), ak.cdc.MustMarshalJSON(lock))
	return nil
}

// BreakLock unlock a lockID without considering time with admin priviledge
func (ak AdminKeeper) BreakLock(ctx sdk.Context, lockID uint64) error {
	lock, err := ak.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// send coins back to owner
	if err := ak.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lock.Owner, lock.Coins); err != nil {
		return err
	}

	store := ctx.KVStore(ak.storeKey)
	store.Delete(lockStoreKey(lockID)) // remove lock from store

	refKeys := lockRefKeys(*lock)
	for _, refKey := range refKeys {
		err = ak.deleteLockRefByKey(ctx, refKey, lockID)
		if err != nil {
			return err
		}
	}
	return nil
}
