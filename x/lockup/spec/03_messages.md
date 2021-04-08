<!--
order: 3
-->

# Messages

## Lock Tokens

`MsgLockTokens` can be submitted by any token holder via a `MsgLockTokens` transaction.

```go
type MsgLockTokens struct {
	Owner    sdk.AccAddress
	Duration time.Duration
	Coins    sdk.Coins
}
```

**State modifications:**

- Validate `Owner` has enough tokens
- Generate new `PeriodLock` record
- Save the record inside the keeper's time basis unlock queue
- Transfer the tokens from the `Owner` to lockup `ModuleAccount`.

## Unlock Tokens

Once time is over, users can withdraw unlocked coins from lockup `ModuleAccount`.

```go
type MsgUnlockTokens struct {
  Owner sdk.AccAddress
}
```

**State modifications:**

- Fetch all unlockable `PeriodLock`s that `Owner` has not withdrawn yet
- Remove `PeriodLock` records from the state
- Transfer the tokens from lockup `ModuleAccount` to the `MsgUnlockTokens.Owner`.

## Unlock PeriodLock

Once time is over, users can withdraw unlocked coins from lockup `ModuleAccount`.

```go
type MsgUnlockPeriodLock struct {
  Owner  sdk.AccAddress
  LockID uint64
}
```

**State modifications:**

- Check `PeriodLock` with `LockID` specified by `MsgUnlockPeriodLock` is available and not withdrawn already
- Check `PeriodLock` owner is same as `MsgUnlockPeriodLock.Owner`
- Remove `PeriodLock` record from the state
- Transfer the tokens from lockup `ModuleAccount` to the `Owner`.

Note: If another module needs past `PeriodLock` item, it can log the details themselves using the hooks.
