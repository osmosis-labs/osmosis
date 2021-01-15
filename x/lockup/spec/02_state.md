<!--
order: 2
-->

# State

## Locked coins management

Locked coins are all stored in module account for `lockup` module which is called `LockPool`.
When user lock coins within `lockup` module, it's moved from user account to `LockPool` and a record (`PeriodLock` struct) is created.

Once the period is over, user can withdraw it at anytime from `LockPool`.
User can withdraw by PeriodLock ID or withdraw all `UnlockableCoins` at a time.

### Period Lock

A `PeriodLock` is a single unit of lock by period. It's a record of locked coin at a specific time.
It stores owner, duration, unlock time and the amount of coins locked.

```go
type PeriodLock struct {
  ID         uint64         // unique ID of a lock
	Owner      sdk.AccAddress
	Duration   time.Duration
	UnlockTime time.Time
  Coins      sdk.Coins
}
```

```protobuf
message PeriodLock {
  uint64 ID = 1;
  string owner = 2;
  google.protobuf.Duration duration = 3;
  google.protobuf.Timestamp unlock_time = 4;
  repeated cosmos.base.v1beta1.Coin coins = 5;
}
```