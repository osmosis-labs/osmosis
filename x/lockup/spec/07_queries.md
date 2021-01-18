<!--
order: 7
-->

# Queries

In this section we describe the queries required on grpc server.

```protobuf
// Query defines the gRPC querier service.
service Query {
    // Return full balance of the module
	rpc GetModuleBalance(ModuleBalanceRequest) returns (ModuleBalanceResponse);
	// Return locked balance of the module
	rpc GetModuleLockedAmount(ModuleLockedAmountRequest) returns (ModuleLockedAmountResponse);

	// Returns whole unlockable coins which are not withdrawn yet
	rpc GetAccountUnlockableCoins(AccountUnlockableCoinsRequest) returns (AccountUnlockableCoinsResponse);
	// Return a locked coins that can't be withdrawn
	rpc GetAccountLockedCoins(AccountLockedCoinsRequest) returns (AccountLockedCoinsResponse);

	// Returns the total number of tokens of an account whose unlock time is beyond timestamp
	rpc GetAccountLockedPastTime(ModuleAccountLockedPastTimeRequest) returns (ModuleAccountLockedPastTimeResponse);
	// Same as GetAccountLockedPastTime but denom specific
	rpc GetAccountLockedPastTimeDenom(AccountLockedPastTimeDenomRequest) returns (AccountLockedPastTimeDenomResponse);
	// Returns the length of the initial lock time when the lock was created
	rpc GetAccountLockPeriod(AccountLockPeriodRequest) returns (AccountLockPeriodResponse);
}
```