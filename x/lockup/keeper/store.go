package keeper

import (
	"time"

	"github.com/c-osmosis/osmosis/x/lockup/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetPeriodLockKey returns the prefix key used for getting a set of period locks
// where unlockTime is after a specific time
func (k Keeper) GetPeriodLockKey(prefix []byte, timestamp time.Time) []byte {
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(prefix)

	bz := make([]byte, prefixL+8+timeBzL)

	// copy the prefix
	copy(bz[:prefixL], prefix)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)
	return bz
}

// GetLastLockID returns ID used last time
func (k Keeper) GetLastLockID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.KeyLastLockID)
	if bz == nil {
		return 0
	}

	return sdk.BigEndianToUint64(bz)
}

// SetLastLockID save ID used by last lock
func (k Keeper) SetLastLockID(ctx sdk.Context, ID uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyLastLockID, sdk.Uint64ToBigEndian(ID))
}

// LockStoreKey returns action store key from ID
func LockStoreKey(ID uint64) []byte {
	return append(types.KeyPrefixPeriodLock, sdk.Uint64ToBigEndian(ID)...)
}

// Lock is a utility to lock coins into module account
func (k Keeper) Lock(ctx sdk.Context, lock types.PeriodLock) error {
	if err := k.bk.SendCoinsFromAccountToModule(ctx, lock.Owner, types.ModuleName, lock.Coins); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, lock.Owner)
	lockID := k.GetLastLockID(ctx) + 1
	prefixStore.Set(LockStoreKey(lockID), k.cdc.MustMarshalJSON(&lock))
	k.SetLastLockID(ctx, lockID)

	// TODO set iterators for timestamps
	// TODO set array of lock IDs per iterator to consider duplication
	// KeyPrefixLockTimestamp
	// KeyPrefixAccountLockTimestamp
	// KeyPrefixDenomLockTimestamp
	// KeyPrefixAccountDenomLockTimestamp
	return nil
}
