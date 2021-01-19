<!--
order: 5
-->

# Keepers

## Lock Keeper

Lock Keeper provide utility functions to query user's locks.

```go
// Keeper is the interface for lockup module keeper
type Keeper interface {
	// Return full balance of the module
	GetModuleBalance(sdk.Context) sdk.Coins
	// Return locked balance of the module
	GetModuleLockedAmount(sdk.Context) sdk.Coins

	// Returns whole unlockable coins which are not withdrawn yet
	GetAccountUnlockableCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	// Return a locked coins that can't be withdrawn
	GetAccountLockedCoins(sdk.Context, sdk.AccAddress) sdk.Coins

	// Returns the total locks of an account whose unlock time is beyond timestamp
	GetAccountLockedPastTime(sdk.AccAddress, timestamp time.Time) []types.PeriodLock
	// Returns the total unlocks of an account whose unlock time is before timestamp
	GetAccountUnlockedBeforeTime(sdk.AccAddress, timestamp time.Time) []types.PeriodLock

	// Same as GetAccountLockedPastTime but denom specific
	GetAccountLockedPastTimeDenom(sdk.AccAddress, denom string, timestamp time.Time) []types.PeriodLock
	// Get iterator for all locks of a denom token that unlocks after timestamp
	IteratorAccountsLockedPastTimeDenom(denom string, timestamp time.Time) db.Iterator
	// Returns all the accounts that locked coins for longer than time.Duration.  Doesn't matter how long is left until unlock.  Only based on initial locktimes
	IteratorLockPeriodsDenom(denom string, time.Duration) []types.PeriodLock
	// Returns the length of the initial lock time when the lock was created
	GetAccountLockPeriod(sdk.AccAddress, lockID uint64)

	// Unlock all unlockable coins 
	UnlockAllUnlockableCoins(sdk.Context, sdk.AccAddress) error
	// unlock by period lock ID
	UnlockPeriodLockByID(sdk.Context, sdk.AccAddress, LockID uint64) error
}
```

# Lock Admin Keeper

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
	// @sunny, how amount ratio could be checked for pool1 LP and pool2 LP tokens?
	RageQuit(sdk.Context, lockID uint64, newCoins sdk.Coins) error
}
```
