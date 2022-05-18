```{=html}
<!--
order: 2
-->
```

State
=====

Locked coins management
-----------------------

Locked coins are all stored in module account for `lockup` module which
is called `LockPool`. When user lock coins within `lockup` module, it's
moved from user account to `LockPool` and a record (`PeriodLock` struct)
is created.

Once the period is over, user can withdraw it at anytime from
`LockPool`. User can withdraw by PeriodLock ID or withdraw all
`UnlockableCoins` at a time.

### Period Lock

A `PeriodLock` is a single unit of lock by period. It's a record of
locked coin at a specific time. It stores owner, duration, unlock time
and the amount of coins locked.

``` {.go}
type PeriodLock struct {
  ID         uint64
  Owner      sdk.AccAddress
  Duration   time.Duration
  UnlockTime time.Time
  Coins      sdk.Coins
}
```

All locks are stored on the KVStore as value at
`{KeyPrefixPeriodLock}{ID}` key.

### Period lock reference queues

To provide time efficient queries, several reference queues are managed
by denom, unlock time, and duration. There are two big queues to store
the lock references. (`a_prefix_key`)

1. Lock references that hasn't started with unlocking yet has prefix of
    `KeyPrefixNotUnlocking`.
2. Lock references that has started unlocking already has prefix of
    `KeyPrefixUnlocking`.
3. Lock references that has withdrawn, it's removed from the store.

Regardless the lock has started unlocking or not, it stores below
references. (`b_prefix_key`)

1. `{KeyPrefixLockDuration}{Duration}`
2. `{KeyPrefixAccountLockDuration}{Owner}{Duration}`
3. `{KeyPrefixDenomLockDuration}{Denom}{Duration}`
4. `{KeyPrefixAccountDenomLockDuration}{Owner}{Denom}{Duration}`

If the lock is unlocking, it also stores the below referneces.

1. `{KeyPrefixLockTimestamp}{LockEndTime}`
2. `{KeyPrefixAccountLockTimestamp}{Owner}{LockEndTime}`
3. `{KeyPrefixDenomLockTimestamp}{Denom}{LockEndTime}`
4. `{KeyPrefixAccountDenomLockTimestamp}{Owner}{Denom}{LockEndTime}`

For end time keys, they are converted to sortable string by using
`sdk.FormatTimeBytes` function.

**Note:** Additionally, for locks that hasn't started unlocking yet, it
stores accumulation store for efficient rewards distribution mechanism.

For reference management, `addLockRefByKey` function is used a lot. Here
key is the prefix key to be used for iteration. It is combination of two
prefix keys.(`{a_prefix_key}{b_prefix_key}`)

``` {.go}
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

### Synthetic Lockup

Synthetic Lockups are a concept that serve the following roles:

- Add "restrictions" to an underlying PeriodLock, so that its bond
    status must be managed by a module rather than a BeginUnlockMessage
- Allow issuing of a locked, "synthetic" denom type
- Allow distribution of rewards to locked synthetic denominations.

The first goal can eventually be pushed into a new data structure, as it
doesn't really relate to the synthetic component.

This is then used for superfluid staking. (Old docs below):

The goal of synthetic lockup is to support the querying of locks by
denom especially for delegated staking. By combining native denom and
synthetic suffix, lockup supports querying with synthetic denom with
existing denom querying functions.

Synthetic lockup is creating virtual lockup where new denom is
combination of original denom and synthetic suffix. At the time of
synthetic lockup creation and deletion, accumulation store is also being
updated and on querier side, they can query as freely as native lockup.

Note: The staking, distribution, slashing, superfluid module would be
refactored to use lockup module and synthetic lockup. The following
changes for synthetic lockup on native lockup change could be defined as
per use case. For now we assume this change is made on hook receiver
side which manages synthetic lockup, e.g. use cases are when user start
/ pause superfluid staking on a lockup, redelegation event, unbonding
event etc.

External modules are managing synthetic locks to use it on their own
logic implementation. (e.g. delegated staking and superfluid staking)

A `SyntheticLock` is a single unit of synthetic lockup. Each synthetic
lockup has reference `PeriodLock` ID, synthetic suffix (`Suffix`) and
synthetic lock's removal time (`EndTime`).

``` {.go}
type SyntheticLock struct {
 LockId  uint64
 Suffix  string
 EndTime time.Time
}
```

All synthetic locks are stored on the KVStore as value at
`{KeyPrefixPeriodLock}{LockID}{Suffix}` key.

### Synthetic lock reference queues

To provide time efficient queries, several reference queues are managed
by denom, unlock time, and duration.

1. `{KeyPrefixDenomLockTimestamp}{SyntheticDenom}{LockEndTime}`
2. `{KeyPrefixDenomLockDuration}{SyntheticDenom}{Duration}`
3. `{KeyPrefixAccountDenomLockTimestamp}{Owner}{SyntheticDenom}{LockEndTime}`
4. `{KeyPrefixAccountDenomLockDuration}{Owner}{SyntheticDenom}{Duration}`

SyntheticDenom is expressed as `{Denom}{Suffix}`. (Note: we can change
this to `{Prefix}{Denom}` as per discussion with Dev)

For end time keys, they are converted to sortable string by using
`sdk.FormatTimeBytes` function.

**Note:** To implement the auto removal of synthetic lockups that is
already finished, we manage a separate time basis queue at
`{KeyPrefixSyntheticLockTimestamp}{EndTime}{LockId}{Suffix}`
