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
- Get the `IntermediaryAccount` for this `lock` `Denom` and `ValAddr` pair.
  - Create it + a new gauge for the synthetic denom, if it does not yet exist.
- Create a SyntheticLockup.
- Mint `Osmo` to match amount in `lock` (based on LP to Osmo ratio at last epoch) and send to `IntermediaryAccount`
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
- Calculate the amount of `Osmo` delegated on behalf of this `lock`
- Use `InstantUndelegate` to instantly remove delegation from `IntermediaryAccount` to `Validator`
- Immediately burn undelegated `Osmo`
- Delete the connection betweene `lockID` and `IntermediaryAccount`
