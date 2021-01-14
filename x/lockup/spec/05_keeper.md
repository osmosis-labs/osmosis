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

	// Returns the total number of tokens of an account whose unlock time is beyond timestamp
	GetAccountLockedPastTime(sdk.AccAddress, timestamp time.Time) (sdk.Coins | []types.PeriodLock)
	// Same as GetAccountLockedPastTime but denom specific
	GetAccountLockedPastTimeDenom(sdk.AccAddress, denom string, timestamp time.Time) (sdk.Coins | []types.PeriodLock)
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
