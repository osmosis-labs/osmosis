# Messages

## Lock Tokens

`MsgLockTokens` can be submitted by any token holder via a
`MsgLockTokens` transaction.

``` {.go}
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

## Begin Unlock of all locks

Once time is over, users can withdraw unlocked coins from lockup
`ModuleAccount`.

``` {.go}
type MsgBeginUnlockingAll struct {
 Owner string
}
```

**State modifications:**

- Fetch all unlockable `PeriodLock`s that has not started unlocking
    yet
- Set `PeriodLock`'s unlock time
- Remove lock references from `NotUnlocking` queue
- Add lock references to `Unlocking` queue

## Begin unlock for a lock

Once time is over, users can withdraw unlocked coins from lockup
`ModuleAccount`.

``` {.go}
type MsgBeginUnlocking struct {
 Owner string
 ID    uint64
}
```

**State modifications:**

- Check `PeriodLock` with `ID` specified by `MsgBeginUnlocking` is not
    started unlocking yet
- Set `PeriodLock`'s unlock time
- Remove lock references from `NotUnlocking` queue
- Add lock references to `Unlocking` queue

Note: If another module needs past `PeriodLock` item, it can log the
details themselves using the hooks.
