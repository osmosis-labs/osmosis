<!--
order: 3
-->

# Messages

## Superfluid Delegate

Owners of locks on approved pools can submit `MsgSuperfluidDelegate` transactions to delegate the Osmo in their locks to a selected validator.

```go
type MsgSuperfluidDelegate struct {
 Sender  string 
 LockId  uint64 
 ValAddr string 
}
```

**State Modifications:**

- Safety Checks that are being done before running superfluid logic:
  - Check that `Sender` is the owner of `lock`
  - Check that `lock` corresponds to a single locked asset
  - Check that `lock` is not unlocking
  - Check that `lock` is locked for at least the unbonding period
  - Check that this `LockID` is not already superfluided
  - Check that the same lock isn't being unbonded
- Get the `IntermediaryAccount` for this lock's `Denom` and `ValAddr` pair.
  - Create it + a new gauge for the synthetic denom, if it does not yet exist.
- Create a SyntheticLockup.
- Calculate `Osmo` to delegate on behalf of this `lock`, as `Osmo Equivalent Multiplier` * `# LP Shares` * `Risk Adjustment Factor`
  - If this amount is less than 0.000001 `Osmo` (`1 uosmo`) reject the transaction, as it would be delegating `0 uosmo`
- Mint `Osmo` to match this amount and send to `IntermediaryAccount`
- Create a delegation from `IntermediaryAccount` to `Validator`
- Create a new perpetual `Gauge` for distributing staking payouts to locks of a synethic asset based on this `Validator` / `Denom` pair.
- Create a connection between this `lockID` and this `IntermediaryAccount`

## Superfluid Undelegate

```go
type MsgSuperfluidUndelegate struct {
 Sender string
 LockId uint64
}
```

**State Modifications:**

- Lookup `lock` by `LockID`
- Check that `Sender` is the owner of `lock`
- Get the `IntermediaryAccount` for this `lockID`
- Delete the `SyntheticLockup` associated to this `lockID` + `ValAddr` pair
- Create a new `SyntheticLockup` which is unbonding
- Calculate the amount of `Osmo` delegated on behalf of this `lock` as `Osmo Equivalent Multipler` * `# LP Shares` * `Risk Adjustment Factor`
  - If this amount is less than 0.000001 `Osmo`, there is no delegated `Osmo` to undelegate and burn
- Use `InstantUndelegate` to instantly remove delegation from `IntermediaryAccount` to `Validator`
- Immediately burn undelegated `Osmo`
- Delete the connection between `lockID` and `IntermediaryAccount`

## Lock and Superfluid Delegate

```go
type MsgLockAndSuperfluidDelegate struct {
 Sender string
 Coins sdk.Coins
 ValAddr string
}
```

This is effectively a multimsg tx of lockup's `MsgLockTokens` and superfluid's `MsgSuperfluidDelegate`,
but it is implemented as a single msg, because currently we don't have a way of passing the lockid
outputted by `MsgLockTokens` as an input into the `MsgSuperfluidDelegate` prior to execution.

**State Modifications:**

- Ensures that Coins has a length of only 1 (we use sdk.Coins instead of sdk.Coin in order to allow more flexibility in the future)
- Creates a lockup with Coins of a lock duration equivalent to the unstaking period from the staking module
  - Uses the lockup module's MsgServer
- Gets the lock id of the created lock, and uses it generate and execute a MsgSuperfluidDelegate message
  - Uses the SuperfluidDelegate function on this msg server

## Superfluid Unbond Lock

```go
type MsgSuperfluidUnbondLock struct {
	Sender string
	LockId uint64
}
```

This message does all the functionality of `MsgSuperfluidUndelegate` but also starts unbonding the underlying
lock as well, allowing both the unstaking and unlocking to complete at the same time.  Without using this function, a user will not be able to start unbonding their underlying lock until after the the unstaking has finished.

**State Modifications:**

- This runs the functionality of `MsgSuperfluidUndelegate`
- It then triggers a force unbond of the underlying lock id