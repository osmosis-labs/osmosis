# Lockup

## Abstract

Lockup module provides an interface for users to lock tokens (also known as bonding) into the module to get incentives.

After tokens have been added to a specific pool and turned into LP shares through the GAMM module, users can then lock these LP shares with a specific duration in order to begin earing rewards.

To unlock these LP shares, users must trigger the unlock timer and wait for the unlock period that was set initially to be completed. After the unlock period is over, users can turn LP shares back into their respective share of tokens.

This module provides interfaces for other modules to iterate the locks efficiently and grpc query to check the status of locked coins.

## Contents

1. **[Concept](#concepts)**
2. **[State](#state)**
3. **[Messages](#messages)**
4. **[Events](#events)**
5. **[Keeper](#keeper)**
6. **[Hooks](#hooks)**
7. **[Queries](#queries)**
8. **[Transactions](#transactions)**
9. **[Params](#params)**
10. **[Endblocker](#endblocker)**

## Concepts

The purpose of `lockup` module is to provide the functionality to lock
tokens for specific period of time for LP token stakers to get
incentives.

To unlock these LP shares, users must trigger the unlock timer and wait for the unlock period that was set initially to be completed. After the unlock period is over, users can turn LP shares back into their respective share of tokens.

This module provides interfaces for other modules to iterate the locks efficiently and grpc query to check the status of locked coins.

There are currently three incentivize lockup periods; `1 day` (24h), `1 week` (168h), and `2 weeks` (336h). When locking tokens in the 2 week period, the liquidity provider is effectively earning rewards for a combination of the 1 day, 1 week, and 2 week bonding periods.

The 2 week period refers to how long it takes to unbond the LP shares. The liquidity provider can keep their LP shares bonded to the 2 week lockup period indefinitely. Unbonding is only required when the liquidity provider desires access to the underlying assets.

If the liquidity provider begins the unbonding process for their 2 week bonded LP shares, they will earn rewards for all three bonding periods during the first day of unbonding.

After the first day passes, they will only receive rewards for the 1 day and 1 week lockup periods. After seven days pass, they will only receive the 1 day rewards until the 2 weeks is complete and their LP shares are unlocked. The below chart is a visual example of what was just explained.

<br/>
<p style="text-align:center;">
<img src="/img/bonding.png" height="300"/>
</p>

</br>
</br>

## State

### Locked coins management

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

## Messages

### Lock Tokens

`MsgLockTokens` can be submitted by any token holder via a
`MsgLockTokens` transaction.

``` {.go}
type MsgLockTokens struct {
 Owner    sdk.AccAddress
 Duration time.Duration
 Coins    sdk.Coins
}
```

**State modifications:**

- Validate `Owner` has enough tokens
- Generate new `PeriodLock` record
- Save the record inside the keeper's time basis unlock queue
- Transfer the tokens from the `Owner` to lockup `ModuleAccount`.

### Begin Unlock of all locks

Once time is over, users can withdraw unlocked coins from lockup
`ModuleAccount`.

``` {.go}
type MsgBeginUnlockingAll struct {
 Owner string
}
```

**State modifications:**

- Fetch all unlockable `PeriodLock`s that has not started unlocking
    yet
- Set `PeriodLock`'s unlock time
- Remove lock references from `NotUnlocking` queue
- Add lock references to `Unlocking` queue

### Begin unlock for a lock

Once time is over, users can withdraw unlocked coins from lockup
`ModuleAccount`.

``` {.go}
type MsgBeginUnlocking struct {
 Owner string
 ID    uint64
}
```

**State modifications:**

- Check `PeriodLock` with `ID` specified by `MsgBeginUnlocking` is not
    started unlocking yet
- Set `PeriodLock`'s unlock time
- Remove lock references from `NotUnlocking` queue
- Add lock references to `Unlocking` queue

Note: If another module needs past `PeriodLock` item, it can log the
details themselves using the hooks.

## Events

The lockup module emits the following events:

### Handlers

#### MsgLockTokens

|  Type          | Attribute Key     | Attribute Value  |
|  --------------| ------------------| -----------------|
|  lock\_tokens  | period\_lock\_id  | {periodLockID}   |
|  lock\_tokens  | owner             | {owner}          |
|  lock\_tokens  | amount            | {amount}         |
|  lock\_tokens  | duration          | {duration}       |
|  lock\_tokens  | unlock\_time      | {unlockTime}     |
|  message       | action            | lock\_tokens     |
|  message       | sender            | {owner}          |
|  transfer      | recipient         | {moduleAccount}  |
|  transfer      | sender            | {owner}          |
|  transfer      | amount            | {amount}         |

#### MsgBeginUnlocking

|  Type           | Attribute Key     | Attribute Value   |
|  ---------------| ------------------| ------------------|
|  begin\_unlock  | period\_lock\_id  | {periodLockID}    |
|  begin\_unlock  | owner             | {owner}           |
|  begin\_unlock  | amount            | {amount}          |
|  begin\_unlock  | duration          | {duration}        |
|  begin\_unlock  | unlock\_time      | {unlockTime}      |
|  message        | action            | begin\_unlocking  |
|  message        | sender            | {owner}           |

#### MsgBeginUnlockingAll

|  Type                | Attribute Key     | Attribute Value        |
|  --------------------| ------------------| -----------------------|
|  begin\_unlock\_all  | owner             | {owner}                |
|  begin\_unlock\_all  | unlocked\_coins   | {unlockedCoins}        |
|  begin\_unlock       | period\_lock\_id  | {periodLockID}         |
|  begin\_unlock       | owner             | {owner}                |
|  begin\_unlock       | amount            | {amount}               |
|  begin\_unlock       | duration          | {duration}             |
|  begin\_unlock       | unlock\_time      | {unlockTime}           |
|  message             | action            | begin\_unlocking\_all  |
|  message             | sender            | {owner}                |

### Endblocker

#### Automatic withdraw when unlock time mature

|  Type            | Attribute Key     | Attribute Value  |
|  ----------------| ------------------| -----------------|
|  message         | action            | unlock\_tokens   |
|  message         | sender            | {owner}          |
|  transfer\[\]    | recipient         | {owner}          |
|  transfer\[\]    | sender            | {moduleAccount}  |
|  transfer\[\]    | amount            | {unlockAmount}   |
|  unlock\[\]      | period\_lock\_id  | {owner}          |
|  unlock\[\]      | owner             | {lockID}         |
|  unlock\[\]      | duration          | {lockDuration}   |
|  unlock\[\]      | unlock\_time      | {unlockTime}     |
|  unlock\_tokens  | owner             | {owner}          |
|  unlock\_tokens  | unlocked\_coins   | {totalAmount}    |

## Keepers

### Lockup Keeper

Lockup keeper provides utility functions to store lock queues and query
locks.

```go
// Keeper is the interface for lockup module keeper
type Keeper interface {
    // GetModuleBalance Returns full balance of the module
    GetModuleBalance(sdk.Context) sdk.Coins
    // GetModuleLockedCoins Returns locked balance of the module
    GetModuleLockedCoins(sdk.Context) sdk.Coins
    // GetAccountUnlockableCoins Returns whole unlockable coins which are not withdrawn yet
    GetAccountUnlockableCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
    // GetAccountUnlockingCoins Returns whole unlocking coins
    GetAccountUnlockingCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
    // GetAccountLockedCoins Returns a locked coins that can't be withdrawn
    GetAccountLockedCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
    // GetAccountLockedPastTime Returns the total locks of an account whose unlock time is beyond timestamp
    GetAccountLockedPastTime(sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock
    // GetAccountUnlockedBeforeTime Returns the total unlocks of an account whose unlock time is before timestamp
    GetAccountUnlockedBeforeTime(sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock
    // GetAccountLockedPastTimeDenom is equal to GetAccountLockedPastTime but denom specific
    GetAccountLockedPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock

    // GetAccountLockedLongerDuration Returns account locked with duration longer than specified
    GetAccountLockedLongerDuration(sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock
    // GetAccountLockedLongerDurationDenom Returns account locked with duration longer than specified with specific denom
    GetAccountLockedLongerDurationDenom(sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock
    // GetLocksPastTimeDenom Returns the locks whose unlock time is beyond timestamp
    GetLocksPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock
    // GetLocksLongerThanDurationDenom Returns the locks whose unlock duration is longer than duration
    GetLocksLongerThanDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock
    // GetLockByID Returns lock from lockID
    GetLockByID(sdk.Context, lockID uint64) (*types.PeriodLock, error)
    // GetPeriodLocks Returns the period locks on pool
    GetPeriodLocks(sdk.Context) ([]types.PeriodLock, error)
    // UnlockAllUnlockableCoins Unlock all unlockable coins
    UnlockAllUnlockableCoins(sdk.Context, account sdk.AccAddress) (sdk.Coins, error)
    // LockTokens lock tokens from an account for specified duration
    LockTokens(sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodLock, error)
    // AddTokensToLock locks more tokens into a lockup
    AddTokensToLock(ctx sdk.Context, owner sdk.AccAddress, lockID uint64, coins sdk.Coins) (*types.PeriodLock, error)
    // Lock is a utility to lock coins into module account
    Lock(sdk.Context, lock types.PeriodLock) error
    // Unlock is a utility to unlock coins from module account
    Unlock(sdk.Context, lock types.PeriodLock) error
    GetSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) (*types.SyntheticLock, error)
    GetAllSyntheticLockupsByLockup(ctx sdk.Context, lockID uint64) []types.SyntheticLock
    GetAllSyntheticLockups(ctx sdk.Context) []types.SyntheticLock
    // CreateSyntheticLockup create synthetic lockup with lock id and denom suffix
    CreateSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string, unlockDuration time.Duration) error
    // DeleteSyntheticLockup delete synthetic lockup with lock id and suffix
    DeleteSyntheticLockup(ctx sdk.Context, lockID uint64, suffix string) error
    DeleteAllMaturedSyntheticLocks(ctx sdk.Context)
```

### Lock Admin Keeper

Lockup admin keeper provides god privilege functions to remove tokens
from locks and create new locks.

```go
// AdminKeeper defines a god priviledge keeper functions to remove tokens from locks and create new locks
// For the governance system of token pools, we want a "ragequit" feature
// So governance changes will take 1 week to go into effect
// During that time, people can choose to "ragequit" which means they would leave the original pool
// and form a new pool with the old parameters but if they still had 2 months of lockup left,
// their liquidity still needs to be 2 month lockup-ed, just in the new pool
// And we need to replace their pool1 LP tokens with pool2 LP tokens with the same lock duration and end time

type AdminKeeper interface {
    Keeper

    // this unlock previous lockID and create a new lock with newCoins with same duration and endtime
    Relock(sdk.Context, lockID uint64, newCoins sdk.Coins) error
    // this unlock without time check with an admin priviledge
    BreakLock(sdk.Context, lockID uint64) error
}
```

## Hooks

In this section we describe the "hooks" that `lockup` module provide for
other modules.

### Tokens Locked

On lock/unlock events, lockup module execute hooks for other modules to
make following actions.

``` go
  OnTokenLocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time)
  OnTokenUnlocked(ctx sdk.Context, address sdk.AccAddress, lockID uint64, amount sdk.Coins, lockDuration time.Duration, unlockTime time.Time)
```

## Parameters

The lockup module contains the following parameters:

| Key                    | Type            | Example |
| ---------------------- | --------------- | ------- |

Note: Currently no parameters are set for `lockup` module, we will need
to move lockable durations from incentives module to lockup module.

## Endblocker

### Withdraw tokens after unlock time mature

Once time is over, endblocker withdraw coins from matured locks and
coins are sent from lockup `ModuleAccount`.

**State modifications:**

- Fetch all unlockable `PeriodLock`s that `Owner` has not withdrawn
    yet
- Remove `PeriodLock` records from the state
- Transfer the tokens from lockup `ModuleAccount` to the
    `MsgUnlockTokens.Owner`.

### Remove synthetic locks after removal time mature

For synthetic lockups, no coin movement is made, but lockup record and
reference queues are removed.

**State modifications:**

- Fetch all synthetic lockups that is matured
- Remove `SyntheticLock` records from the state along with reference
    queues

## Transactions

### lock-tokens

Bond tokens in a LP for a set duration

```sh
osmosisd tx lockup lock-tokens [tokens] --duration --from --chain-id
```

::: details Example

To lockup `15.527546134174465309gamm/pool/3` tokens for a `one day` bonding period from `WALLET_NAME` on the osmosis mainnet:

```bash
osmosisd tx lockup lock-tokens 15527546134174465309gamm/pool/3 --duration="24h" --from WALLET_NAME --chain-id osmosis-1
```

To lockup `25.527546134174465309gamm/pool/13` tokens for a `one week` bonding period from `WALLET_NAME` on the osmosis testnet:

```bash
osmosisd tx lockup lock-tokens 25527546134174465309gamm/pool/13 --duration="168h" --from WALLET_NAME --chain-id osmo-test-4
```

To lockup `35.527546134174465309 gamm/pool/197` tokens for a `two week` bonding period from `WALLET_NAME` on the osmosis mainnet:

```bash
osmosisd tx lockup lock-tokens 35527546134174465309gamm/pool/197 --duration="336h" --from WALLET_NAME --chain-id osmosis-1
```
:::


### begin-unlock-by-id

Begin the unbonding process for tokens given their unique lock ID

```sh
osmosisd tx lockup begin-unlock-by-id [id] --from --chain-id
```

::: details Example

To begin the unbonding time for all bonded tokens under id `75` from `WALLET_NAME` on the osmosis mainnet:

```bash
osmosisd tx lockup begin-unlock-by-id 75 --from WALLET_NAME --chain-id osmosis-1
```
:::
::: warning Note
The ID corresponds to the unique ID given to your lockup transaction (explained more in lock-by-id section)
:::

### begin-unlock-tokens

Begin unbonding process for all bonded tokens in a wallet

```sh
osmosisd tx lockup begin-unlock-tokens --from --chain-id
```

::: details Example

To begin unbonding time for ALL pools and ALL bonded tokens in `WALLET_NAME` on the osmosis mainnet:


```bash
osmosisd tx lockup begin-unlock-tokens --from=WALLET_NAME --chain-id=osmosis-1 --yes
```
:::

## Queries

In this section we describe the queries required on grpc server.

``` protobuf
// Query defines the gRPC querier service.
service Query {
    // Return full balance of the module
 rpc ModuleBalance(ModuleBalanceRequest) returns (ModuleBalanceResponse);
 // Return locked balance of the module
 rpc ModuleLockedAmount(ModuleLockedAmountRequest) returns (ModuleLockedAmountResponse);

 // Returns unlockable coins which are not withdrawn yet
 rpc AccountUnlockableCoins(AccountUnlockableCoinsRequest) returns (AccountUnlockableCoinsResponse);
 // Returns unlocking coins
   rpc AccountUnlockingCoins(AccountUnlockingCoinsRequest) returns (AccountUnlockingCoinsResponse) {}
 // Return a locked coins that can't be withdrawn
 rpc AccountLockedCoins(AccountLockedCoinsRequest) returns (AccountLockedCoinsResponse);

 // Returns locked records of an account with unlock time beyond timestamp
 rpc AccountLockedPastTime(AccountLockedPastTimeRequest) returns (AccountLockedPastTimeResponse);
 // Returns locked records of an account with unlock time beyond timestamp excluding tokens started unlocking
 rpc AccountLockedPastTimeNotUnlockingOnly(AccountLockedPastTimeNotUnlockingOnlyRequest) returns (AccountLockedPastTimeNotUnlockingOnlyResponse) {}
 // Returns unlocked records with unlock time before timestamp
 rpc AccountUnlockedBeforeTime(AccountUnlockedBeforeTimeRequest) returns (AccountUnlockedBeforeTimeResponse);

 // Returns lock records by address, timestamp, denom
 rpc AccountLockedPastTimeDenom(AccountLockedPastTimeDenomRequest) returns (AccountLockedPastTimeDenomResponse);
 // Returns lock record by id
 rpc LockedByID(LockedRequest) returns (LockedResponse);

 // Returns account locked records with longer duration
 rpc AccountLockedLongerDuration(AccountLockedLongerDurationRequest) returns (AccountLockedLongerDurationResponse);
 // Returns account locked records with longer duration excluding tokens started unlocking
   rpc AccountLockedLongerDurationNotUnlockingOnly(AccountLockedLongerDurationNotUnlockingOnlyRequest) returns (AccountLockedLongerDurationNotUnlockingOnlyResponse) {}
 // Returns account's locked records for a denom with longer duration
 rpc AccountLockedLongerDurationDenom(AccountLockedLongerDurationDenomRequest) returns (AccountLockedLongerDurationDenomResponse);

 // Returns account locked records with a specific duration
 rpc AccountLockedDuration(AccountLockedDurationRequest) returns (AccountLockedDurationResponse);
}
```

### account-locked-beforetime

Query an account's unlocked records after a specified time (UNIX) has passed

In other words, if an account unlocked all their bonded tokens the moment the query was executed, only the locks that would have completed their bond time requirement by the time the `TIMESTAMP` is reached will be returned.

::: details Example

In this example, the current UNIX time is `1639776682`, 2 days from now is approx `1639971082`, and 15 days from now is approx `1641094282`.

An account's `ADDRESS` is locked in both the `1 day` and `1 week` gamm/pool/3. To query the `ADDRESS` with a timestamp 2 days from now `1639971082`:

```bash
osmosisd query lockup account-locked-beforetime ADDRESS 1639971082
```

In this example will output the `1 day` lock but not the `1 week` lock:

```bash
locks:
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

If querying the same `ADDRESS` with a timestamp 15 days from now `1641094282`:

```bash
osmosisd query lockup account-locked-beforetime ADDRESS 1641094282
```

In this example will output both the `1 day` and `1 week` lock:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```
:::


### account-locked-coins

Query an account's locked (bonded) LP tokens

```sh
osmosisd query lockup account-locked-coins [address]
```

:::: details Example

```bash
osmosisd query lockup account-locked-coins osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

An example output:

```bash
coins:
- amount: "413553955105681228583"
  denom: gamm/pool/1
- amount: "32155370994266157441309"
  denom: gamm/pool/10
- amount: "220957857520769912023"
  denom: gamm/pool/3
- amount: "31648237936933949577"
  denom: gamm/pool/42
- amount: "14162624050980051053569"
  denom: gamm/pool/5
- amount: "1023186951315714985896914"
  denom: gamm/pool/9
```
::: warning Note
All GAMM amounts listed are 10^18. Move the decimal place to the left 18 places to get the GAMM amount listed in the GUI.

You may also specify a --height flag to see bonded LP tokens at a specified height (note: if running a pruned node, this may result in an error)
:::
::::

### account-locked-longer-duration

Query an account's locked records that are greater than or equal to a specified lock duration

```sh
osmosisd query lockup account-locked-longer-duration [address] [duration]
```

::: details Example

Here is an example of querying an `ADDRESS` for all `1 day` or greater bonding periods:

```bash
osmosisd query lockup account-locked-longer-duration osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h
```

An example output:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "2021-12-18T23:32:58.900715388Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```
:::


### account-locked-longer-duration-denom

Query an account's locked records for a denom that is locked equal to or greater than the specified duration AND match a specified denom

```sh
osmosisd query lockup account-locked-longer-duration-denom [address] [duration] [denom]
```

::: details Example

Here is an example of an `ADDRESS` that is locked in both the `1 day` and `1 week` for both the gamm/pool/3 and gamm/pool/1, then queries the `ADDRESS` for all bonding periods equal to or greater than `1 day` for just the gamm/pool/3:

```bash
osmosisd query lockup account-locked-longer-duration-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h gamm/pool/3
```

An example output:

```bash
locks:
- ID: "571839"
  coins:
  - amount: "15527546134174465309"
    denom: gamm/pool/3
  duration: 24h
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

As shown, the gamm/pool/3 is returned but not the gamm/pool/1 due to the denom filter.
:::


### account-locked-longer-duration-not-unlocking

Query an account's locked records for a denom that is locked equal to or greater than the specified duration AND is not in the process of being unlocked

```sh
osmosisd query lockup account-locked-longer-duration-not-unlocking [address] [duration]
```

::: details Example

Here is an example of an `ADDRESS` that is locked in both the `1 day` and `1 week` gamm/pool/3, begins unlocking process for the `1 day` bond, and queries the `ADDRESS` for all bonding periods equal to or greater than `1 day` that are not unbonding:

```bash
osmosisd query lockup account-locked-longer-duration-not-unlocking osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 24h
```

An example output:

```bash
locks:
- ID: "571839"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

The `1 day` bond does not show since it is in the process of unbonding.
:::


### account-locked-pasttime

Query the locked records of an account with the unlock time beyond timestamp (UNIX)

```bash
osmosisd query lockup account-locked-pasttime [address] [timestamp]
```

::: details Example

Here is an example of an account that is locked in both the `1 day` and `1 week` gamm/pool/3. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`)

```bash
osmosisd query lockup account-locked-pasttime osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082
```

The example output:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` lock ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input.
:::


### account-locked-pasttime-denom

Query the locked records of an account with the unlock time beyond timestamp (unix) and filter by a specific denom

```bash
osmosisd query lockup account-locked-pasttime-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 [timestamp] [denom]
```

::: details Example

Here is an example of an account that is locked in both the `1 day` and `1 week` gamm/pool/3 and `1 day` and `1 week` gamm/pool/1. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`) and filters for gamm/pool/3

```bash
osmosisd query lockup account-locked-pasttime-denom osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082 gamm/pool/3
```

The example output:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` lock ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input. Additionally, neither of the `1 day` or `1 week` lock IDs for gamm/pool/1 showed due to the denom filter.
:::


### account-locked-pasttime-not-unlocking

Query the locked records of an account with the unlock time beyond timestamp (unix) AND is not in the process of unlocking

```sh
osmosisd query lockup account-locked-pasttime [address] [timestamp]
```

::: details Example

Here is an example of an account that is locked in both the `1 day` and `1 week` gamm/pool/3. In this example, the UNIX time is currently `1639776682` and queries an `ADDRESS` for UNIX time two days later from the current time (which in this example would be `1639971082`) AND is not unlocking:

```bash
osmosisd query lockup account-locked-pasttime osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259 1639971082
```

The example output:

```bash
locks:
- ID: "572027"
  coins:
  - amount: "16120691802759484268"
    denom: gamm/pool/3
  duration: 604800.000006193s
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Note that the `1 day` lock ID did not display because, if the unbonding time began counting down from the time the command was executed, the bonding period would be complete before the two day window given by the UNIX timestamp input. Additionally, if ID 572027 were to begin the unlocking process, the query would have returned blank.
:::


### account-unlockable-coins

Query an address's LP shares that have completed the unlocking period and are ready to be withdrawn

```bash
osmosisd query lockup account-unlockable-coins ADDRESS
```



### account-unlocking-coins

Query an address's LP shares that are currently unlocking

```sh
osmosisd query lockup account-unlocking-coins [address]
```

::: details Example

```bash
osmosisd query lockup account-unlocking-coins osmo1xqhlshlhs5g0acqgrkafdemvf5kz4pp4c2x259
```

Example output:

```bash
coins:
- amount: "15527546134174465309"
  denom: gamm/pool/3
```
:::


### lock-by-id

Query a lock record by its ID

```sh
osmosisd query lockup lock-by-id [id]
```

::: details Example

Every time a user bonds tokens to an LP, a unique lock ID is created for that transaction.

Here is an example viewing the lock record for ID 9:

```bash
osmosisd query lockup lock-by-id 9
```

And its output:

```bash
lock:
  ID: "9"
  coins:
  - amount: "2449472670508255020346507"
    denom: gamm/pool/2
  duration: 336h
  end_time: "0001-01-01T00:00:00Z"
  owner: osmo16r39ghhwqjcwxa8q3yswlz8jhzldygy66vlm82
```

In summary, this shows wallet `osmo16r39ghhwqjcwxa8q3yswlz8jhzldygy66vlm82` bonded `2449472.670 gamm/pool/2` LP shares for a `2 week` locking period.
:::


### module-balance

Query the balance of all LP shares (bonded and unbonded)

```sh
osmosisd query lockup module-balance
```

::: details Example

```bash
osmosisd query lockup module-balance
```

An example output:

```bash
coins:
- amount: "118851922644152734549498647"
  denom: gamm/pool/1
- amount: "2165392672114512349039263626"
  denom: gamm/pool/10
- amount: "9346769826591025900804"
  denom: gamm/pool/13
- amount: "229347389639275840044722315"
  denom: gamm/pool/15
- amount: "81217698776012800247869"
  denom: gamm/pool/183
- amount: "284253336860259874753775"
  denom: gamm/pool/197
- amount: "664300804648059580124426710"
  denom: gamm/pool/2
- amount: "5087102794776326441530430"
  denom: gamm/pool/22
- amount: "178900843925960029029567880"
  denom: gamm/pool/3
- amount: "64845148811263846652326124"
  denom: gamm/pool/4
- amount: "177831279847453210600513"
  denom: gamm/pool/42
- amount: "18685913727862493301261661338"
  denom: gamm/pool/5
- amount: "23579028640963777558149250419"
  denom: gamm/pool/6
- amount: "1273329284855460149381904976"
  denom: gamm/pool/7
- amount: "625252103927082207683116933"
  denom: gamm/pool/8
- amount: "1148475247281090606949382402"
  denom: gamm/pool/9
```
:::


### module-locked-amount

Query the balance of all bonded LP shares

```sh
osmosisd query lockup module-locked-amount
```

::: details Example

```bash
osmosisd query lockup module-locked-amount
```

An example output:

```bash

  "coins":
    {
      "denom": "gamm/pool/1",
      "amount": "247321084020868094262821308"
    },
    {
      "denom": "gamm/pool/10",
      "amount": "2866946821820635047398966697"
    },
    {
      "denom": "gamm/pool/13",
      "amount": "9366580741745176812984"
    },
    {
      "denom": "gamm/pool/15",
      "amount": "193294911294343602187680438"
    },
    {
      "denom": "gamm/pool/183",
      "amount": "196722012808526595790871"
    },
    {
      "denom": "gamm/pool/197",
      "amount": "1157025085661167198918241"
    },
    {
      "denom": "gamm/pool/2",
      "amount": "633051132033131378888258047"
    },
    {
      "denom": "gamm/pool/22",
      "amount": "3622601406125950733194696"
    },
...

```

NOTE: This command seems to only work on gRPC and on CLI returns an EOF error.
:::



### output-all-locks

Output all locks into a json file

```sh
osmosisd query lockup output-all-locks [max lock ID]
```

:::: details Example

This example command outputs locks 1-1000 and saves to a json file:

```bash
osmosisd query lockup output-all-locks 1000
```
::: warning Note
If a lockup has been completed, the lockup status will show as "0" (or successful) and no further information will be available. To get further information on a completed lock, run the lock-by-id query.
:::
::::


### total-locked-of-denom

Query locked amount for a specific denom in the duration provided

```sh
osmosisd query lockup total-locked-of-denom [denom] --min-duration
```

:::: details Example

This example command outputs the amount of `gamm/pool/2` LP shares that locked in the `2 week` bonding period:

```bash
osmosisd query lockup total-locked-of-denom gamm/pool/2 --min-duration "336h"
```

Which, at the time of this writing outputs `14106985399822075248947045` which is equivalent to `14106985.3998 gamm/pool/2`

NOTE: As of this writing, there is a bug that defaults the min duration to days instead of seconds. Ensure you specify the time in seconds to get the correct response.
:::

## Commands

```sh
# 1 day 100stake lock-tokens command
osmosisd tx lockup lock-tokens 200stake --duration="86400s" --from=validator --chain-id=testing --keyring-backend=test --yes

# 5s 100stake lock-tokens command
osmosisd tx lockup lock-tokens 100stake --duration="5s" --from=validator --chain-id=testing --keyring-backend=test --yes

# begin unlock tokens, NOTE: add more gas when unlocking more than two locks in a same command
osmosisd tx lockup begin-unlock-tokens --from=validator --gas=500000 --chain-id=testing --keyring-backend=test --yes

# unlock tokens, NOTE: add more gas when unlocking more than two locks in a same command
osmosisd tx lockup unlock-tokens --from=validator --gas=500000 --chain-id=testing --keyring-backend=test --yes

# unlock specific period lock
osmosisd tx lockup unlock-by-id 1 --from=validator --chain-id=testing --keyring-backend=test --yes

# account balance
osmosisd query bank balances $(osmosisd keys show -a validator --keyring-backend=test)

# query module balance
osmosisd query lockup module-balance

# query locked amount
osmosisd query lockup module-locked-amount

# query lock by id
osmosisd query lockup lock-by-id 1

# query account unlockable coins
osmosisd query lockup account-unlockable-coins $(osmosisd keys show -a validator --keyring-backend=test)

# query account locks by denom past time
osmosisd query lockup account-locked-pasttime-denom $(osmosisd keys show -a validator --keyring-backend=test) 1611879610 stake

# query account locks past time
osmosisd query lockup account-locked-pasttime $(osmosisd keys show -a validator --keyring-backend=test) 1611879610

# query account locks by denom with longer duration
osmosisd query lockup account-locked-longer-duration-denom $(osmosisd keys show -a validator --keyring-backend=test) 5.1s stake

# query account locks with longer duration
osmosisd query lockup account-locked-longer-duration $(osmosisd keys show -a validator --keyring-backend=test) 5.1s

# query account locked coins
osmosisd query lockup account-locked-coins $(osmosisd keys show -a validator --keyring-backend=test)

# query account locks before time
osmosisd query lockup account-locked-beforetime $(osmosisd keys show -a validator --keyring-backend=test) 1611879610
```
