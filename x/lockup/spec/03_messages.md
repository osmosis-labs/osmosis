<!--
order: 3
-->

# Messages

## Lock Tokens

`MsgLockTokens` can be submitted by any token holder via a `MsgLockTokens` transaction.

```go
type MsgLockTokens struct {
	Proposer   sdk.AccAddress
	UnlockTime time.Time
	Coins      sdk.Coins
}
```

**State modifications:**

- Validate `Proposer` has enough tokens
- Generate new `PeriodLock` record
- Save the record inside the keeper's time basis unlock queue
- Transfer the tokens from the `Proposer` to lockup `ModuleAccount`.

## Unlock Tokens

Once time is over, users can withdraw unlocked coins from lockup `ModuleAccount`.

```go
type MsgUnlockTokens struct {
  Proposer   sdk.AccAddress
  Coins      sdk.Coins
}
```

**State modifications:**

- Validate `Proposer` has unlocked coins within lockup `ModuleAccount`
- 3 options to manage the records (Option 3 seems to be best personally for implementation simplicity)
 1) Remove `PeriodLock` record, in this case, user can withdraw only by record basis
 2) Add `Withdrawn` Flag to `true`, in this case, user can withdraw only by record basis
 3) Manage `WithdrawnTokens` records separately and user can withdraw `UnlockedTokens - WithdrawnTokens`
- Save the record inside the keeper's time basis unlock queue
- Transfer the tokens from lockup `ModuleAccount` to the `Proposer`.

