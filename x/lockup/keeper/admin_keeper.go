package keeper

import (
	"github.com/c-osmosis/osmosis/x/lockup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RageQuit unlock previous lockID and create a new lock with newCoins with same duration and endtime
func (ak AdminKeeper) RageQuit(ctx sdk.Context, lockID uint64, newCoins sdk.Coins) error {
	lock, err := ak.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	// send original coins back to owner
	if err := ak.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, lock.Owner, lock.Coins); err != nil {
		return err
	}

	// lock newCoins into module account
	if ak.bk.SendCoinsFromAccountToModule(ctx, lock.Owner, types.ModuleName, newCoins); err != nil {
		return err
	}

	// replace to new coins
	lock.Coins = newCoins

	// reset lock record inside store
	store := ctx.KVStore(ak.storeKey)
	store.Set(LockStoreKey(lockID), ak.cdc.MustMarshalJSON(lock))
	return nil
}
