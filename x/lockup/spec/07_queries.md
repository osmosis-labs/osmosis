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

	// Returns whole balance deposited by the user which is not withdrawn yet
	GetAccountPoolBalance(sdk.Context, *ModuleAccountPoolBalanceRequest) (*ModuleAccountPoolBalanceResponse, error)
	// Return a locked amount that can't be withdrawn
	GetAccountLockedAmount(sdk.Context, *ModuleAccountLockedAmountRequest) (*ModuleAccountLockedAmountResponse, error)

	// Returns the total number of tokens of an account whose unlock time is beyond timestamp
	GetAccountLockedPastTime(sdk.Context, *ModuleAccountLockedPastTimeRequest) (*ModuleAccountLockedPastTimeResponse, error)
	// Same as GetAccountLockedPastTime but denom specific
	GetAccountLockedPastTimeDenom(sdk.Context, *AccountLockedPastTimeDenomRequest) (*AccountLockedPastTimeDenomResponse, error)
	// @sunny, single account can have multiple period locks, and it seems to be not valid query
	// Returns the length of the initial lock time when the lock was created
	// GetAccountLockPeriod(sdk.Context, *AccountAccountLockPeriodRequest) (*AccountAccountLockPeriodResponse, error) 
}
```