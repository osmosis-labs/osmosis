```html
<!--
order: 5
-->
```

Keepers

Lockup Keeper

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

# Lock Admin Keeper

Lockup admin keeper provides god privilege functions to remove tokens
from locks and create new locks.

``` go
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
