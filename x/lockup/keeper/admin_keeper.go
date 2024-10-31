package keeper

import (
	"github.com/cosmos/gogoproto/proto"

	"github.com/osmosis-labs/osmosis/v27/x/lockup/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Relock unlock previous lockID and create a new lock with newCoins with same duration and endtime.
func (ak AdminKeeper) Relock(ctx sdk.Context, lockID uint64, newCoins sdk.Coins) error {
	lock, err := ak.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}

	// send original coins back to owner
	if err := ak.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, lock.Coins); err != nil {
		return err
	}

	// lock newCoins into module account
	if err := ak.bk.SendCoinsFromAccountToModule(ctx, owner, types.ModuleName, newCoins); err != nil {
		return err
	}

	// replace to new coins
	lock.Coins = newCoins

	// reset lock record inside store
	store := ctx.KVStore(ak.storeKey)
	bz, err := proto.Marshal(lock)
	if err != nil {
		return err
	}
	store.Set(lockStoreKey(lockID), bz)
	return nil
}

// BreakLock unlock a lockID without considering time with admin privilege.
func (ak AdminKeeper) BreakLock(ctx sdk.Context, lockID uint64) error {
	lock, err := ak.GetLockByID(ctx, lockID)
	if err != nil {
		return err
	}

	owner, err := sdk.AccAddressFromBech32(lock.Owner)
	if err != nil {
		return err
	}

	// send coins back to owner
	if err := ak.bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, owner, lock.Coins); err != nil {
		return err
	}

	ak.deleteLock(ctx, lockID)

	refKeys, err := lockRefKeys(*lock)
	if err != nil {
		return err
	}

	for _, refKey := range refKeys {
		ak.deleteLockRefByKey(ctx, refKey, lockID)
	}
	return nil
}
