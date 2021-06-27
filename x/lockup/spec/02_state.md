<!--
order: 2
-->

# State

## Locked coins management

Locked coins are all stored in module account for `lockup` module which is called `LockPool`.
When user lock coins within `lockup` module, it's moved from user account to `LockPool` and a record (`PeriodLock` struct) is created.

Once the period is over, user can withdraw it at anytime from `LockPool`.
User can withdraw by PeriodLock ID or withdraw all `UnlockableCoins` at a time.

### Period Lock

A `PeriodLock` is a single unit of lock by period. It's a record of locked coin at a specific time.
It stores owner, duration, unlock time and the amount of coins locked.

```go
type PeriodLock struct {
  ID         uint64
  Owner      sdk.AccAddress
  Duration   time.Duration
  UnlockTime time.Time
  Coins      sdk.Coins
}
```

All locks are stored on the KVStore as value at `{KeyPrefixPeriodLock}{ID}` key.

### Period lock reference queues

To provide time efficient queries, several reference queues are managed by denom, unlock time, and duration.
There are two big queues to store the lock references. (`a_prefix_key`)
1. Lock references that hasn't started with unlocking yet has prefix of `KeyPrefixNotUnlocking`.
2. Lock references that has started unlocking already has prefix of `KeyPrefixUnlocking`.
3. Lock references that has withdrawn, it's removed from the store.

Regardless the lock has started unlocking or not, it stores below references. (`b_prefix_key`)
1. `{KeyPrefixLockTimestamp}{LockEndTime}`
2. `{KeyPrefixLockDuration}{Duration}`
3. `{KeyPrefixAccountLockTimestamp}{Owner}{LockEndTime}`
4. `{KeyPrefixAccountLockDuration}{Owner}{Duration}`
5. `{KeyPrefixDenomLockTimestamp}{Denom}{LockEndTime}`
6. `{KeyPrefixDenomLockDuration}{Denom}{Duration}`
7. `{KeyPrefixAccountDenomLockTimestamp}{Owner}{Denom}{LockEndTime}`
8. `{KeyPrefixAccountDenomLockDuration}{Owner}{Denom}{Duration}`

For end time keys, they are converted to sortable string by using `sdk.FormatTimeBytes` function.

**Note:**
Additionally, for locks that hasn't started unlocking yet, it stores accumulation store for efficient rewards distribution mechanism.

For reference management, `addLockRefByKey` function is used a lot.
Here key is the prefix key to be used for iteration. It is combination of two prefix keys.(`{a_prefix_key}{b_prefix_key}`)

```go
// addLockRefByKey make a lockID iterable with the prefix `key`
func (k Keeper) addLockRefByKey(ctx sdk.Context, key []byte, lockID uint64) error {
	store := ctx.KVStore(k.storeKey)
	lockIDBz := sdk.Uint64ToBigEndian(lockID)
	endKey := combineKeys(key, lockIDBz)
	if store.Has(endKey) {
		return fmt.Errorf("lock with same ID exist: %d", lockID)
	}
	store.Set(endKey, lockIDBz)
	return nil
}
```
