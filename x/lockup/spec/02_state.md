<!--
order: 2
-->

# State

## Locked coins management

Locked coins are all stored in module account for `lockup` module which is called `LockPool`.
When user lock coins within `lockup` module, it's moved from user account to `LockPool` and a record (`PeriodLock` struct) is created.

Once the period is over, user can withdraw it at anytime from `LockPool`.
Note:
- Do we need automate withdraw from the pool?
- When querying for full lock amount, unlocked amount but which is in lock pool yet should be queried?
- Is it needed to withdraw by record or by coins? (Or even withdraw full for unlocked?)

### Period Lock

A `PeriodLock` is a single unit of lock by period. It's a record of locked coin at a specific time.
It stores owner, duration, unlock time and the amount of coins locked.

```go
type PeriodLock struct {
	owner      sdk.AccAddress
	duration   time.Duration
	unlockTime time.Time
	coins      sdk.Coins
}

type UnlockedTokens struct {
	owner      sdk.AccAddress
	coins      sdk.Coins
}
```

```protobuf
message PeriodLock {
  string owner = 1;
  google.protobuf.Duration duration = 2;
  google.protobuf.Timestamp unlock_time = 3;
  repeated cosmos.base.v1beta1.Coin coins = 4;
}

message UnlockedTokens {
  string owner = 1;
  repeated cosmos.base.v1beta1.Coin coins = 2;
}
```