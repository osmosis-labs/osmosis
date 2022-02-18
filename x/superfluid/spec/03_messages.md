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
- Lookup `lock` by `LockID`
- Safety Checks
    - Check that `Sender` is the owner of `lock`
    - Check that `lock` corresponds to a single locked asset
    - Check that `lock` is not unlocking
    - Check that `lock` is locked for at least the unbonding period
    - Check that this `LockID` is not already superfluided
    - FIXME something to do with getting the unstaking synthetic lockup?
- Get the `IntermediaryAccount` for this `lock` `Denom` and `ValAddr` pair.
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
- Undelegate `Osmo` from `IntermediaryAccount` to `Validator`
    - `Osmo` will be burned from `IntermediaryAccount` on epoch after unbonding finishes
- Delete the connection betweene `lockID` and `IntermediaryAccount`
