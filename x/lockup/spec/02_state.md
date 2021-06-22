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
  ID         uint64
  Owner      sdk.AccAddress
  Duration   time.Duration
  UnlockTime time.Time
  Coins      sdk.Coins
}
```

### Period lock queues

To provide time efficient queries, several queues are managed by denom, unlock time, and duration.

All queues objects are sorted by timestamp. The time used within any queue is first rounded to the nearest nanosecond then sorted. The sortable time format used is a slight modification of the RFC3339Nano and uses the the format string
`"2006-01-02T15:04:05.000000000"`.

In all cases, the stored timestamp represents the maturation time of the queue element.

