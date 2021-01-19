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

	// Returns the total locks of an account whose unlock time is beyond timestamp
	rpc GetAccountLockedPastTime(AccountLockedPastTimeRequest) returns (AccountLockedPastTimeResponse);
	// Returns the total unlocks of an account whose unlock time is before timestamp
	rpc GetAccountUnlockedBeforeTime(AccountUnlockedBeforeTimeRequest) returns (AccountUnlockedBeforeTimeResponse);

	// Same as GetAccountLockedPastTime but denom specific
	rpc GetAccountLockedPastTimeDenom(AccountLockedPastTimeDenomRequest) returns (AccountLockedPastTimeDenomResponse);
	// Returns the length of the initial lock time when the lock was created
	rpc GetLock(LockRequest) returns (LockResponse);

	// Returns account locked with duration longer than specified
	rpc GetAccountLockedLongerThanDuration(AccountLockedLongerDurationRequest) returns (AccountLockedLongerDurationResponse);
	// Returns account locked with duration longer than specified with specific denom
	rpc GetAccountLockedLongerThanDurationDenom(AccountLockedLongerDurationDenomRequest) returns (AccountLockedLongerDurationDenomResponse;
}
```