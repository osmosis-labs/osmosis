<!--
order: 7
-->

# Queries

In this section we describe the queries required on grpc server.

```go
// QueryServer is the server API for Query service.
type QueryServer interface {
	// Return full balance of the module
	GetModuleBalance(sdk.Context, *ModuleBalanceRequest) (*ModuleBalanceResponse, error)
	// Return locked balance of the module
	GetModuleLockedAmount(sdk.Context, *ModuleLockedAmountRequest) (*ModuleLockedAmountResponse, error)

	// Returns whole unlockable coins which are not withdrawn yet
	GetAccountUnlockableCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	// Return a locked coins that can't be withdrawn
	GetAccountLockedCoins(sdk.Context, sdk.AccAddress) sdk.Coins

	// Returns the total number of tokens of an account whose unlock time is beyond timestamp
	GetAccountLockedPastTime(sdk.Context, *ModuleAccountLockedPastTimeRequest) (*ModuleAccountLockedPastTimeResponse, error)
	// Same as GetAccountLockedPastTime but denom specific
	GetAccountLockedPastTimeDenom(sdk.Context, *AccountLockedPastTimeDenomRequest) (*AccountLockedPastTimeDenomResponse, error)
	// Returns the length of the initial lock time when the lock was created
	GetAccountLockPeriod(sdk.Context, *AccountAccountLockPeriodRequest) (*AccountAccountLockPeriodResponse, error) 
}
```