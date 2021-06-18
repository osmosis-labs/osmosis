<!--
order: 5
-->

# Keepers

## Lockup Keeper

Lockup keeper provides utility functions to store lock queues and query locks.

```go
// Keeper is the interface for lockup module keeper
type Keeper interface {
	GetModuleBalance(sdk.Context) sdk.Coins
	GetModuleLockedCoins(sdk.Context) sdk.Coins
	GetAccountUnlockableCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetAccountUnlockingCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetAccountLockedCoins(sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetAccountLockedPastTime(sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock
	GetAccountUnlockedBeforeTime(sdk.Context, addr sdk.AccAddress, timestamp time.Time) []types.PeriodLock
	GetAccountLockedPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock
	GetAccountLockedLongerDuration(sdk.Context, addr sdk.AccAddress, duration time.Duration) []types.PeriodLock
	GetAccountLockedLongerDurationDenom(sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock
	GetLocksPastTimeDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock
	GetLocksLongerThanDurationDenom(ctx sdk.Context, addr sdk.AccAddress, denom string, duration time.Duration) []types.PeriodLock
	GetLockByID(sdk.Context, lockID uint64) (*types.PeriodLock, error)
	GetPeriodLocks(sdk.Context) ([]types.PeriodLock, error)
	UnlockAllUnlockableCoins(sdk.Context, account sdk.AccAddress) (sdk.Coins, error)
	UnlockPeriodLockByID(sdk.Context, LockID uint64) (*types.PeriodLock, error)
	LockTokens(sdk.Context, owner sdk.AccAddress, coins sdk.Coins, duration time.Duration) (types.PeriodLock, error)
	Lock(sdk.Context, lock types.PeriodLock) error
	Unlock(sdk.Context, lock types.PeriodLock) error
}
```

# Lock Admin Keeper

Lockup admin keeper provides god privilege functions to remove tokens from locks and create new locks.

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
