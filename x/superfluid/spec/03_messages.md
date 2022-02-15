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
- Lookup `lock` by `LockId`
- Safety Checks
    - Check that `Sender` is the owner of `lock`
    - Check that `lock` corresponds to a single locked asset
    - Check that `lock` is not unlocking
    - Check that `lock` is locked for at least the unbonding period
    - Check that this `LockId` is not already superfluided
    - FIXME something to do with getting the unstaking synthetic lockup?
- Get the `IntermediaryAccount` for this `lock` `Denom` and `ValAddr` pair.
- Mint `Osmo` to match amount in `lock` and send to `IntermediaryAccount`
- Create a delegation from `IntermediaryAccount` to `Validator`
- Create a new perpetual `Gauge` for distributing staking payouts to locks of a synethic asset based on this `Validator` / `Denom` pair.
- Create a connection between this `lockID` and this `IntermediaryAccount`

### Questions

- Where are we checking that the lock is of an LP share?
- Where are we checking that the pool contains Osmo?
- Where are we checking that this pool is approved for superluid?
- How does twap work here? It looks like it expects twap to already be stored in state?
    - Why do we need to mint osmo now, instead of in the epoch? (presumably during epoch the amount will be changed to match live twap/spot anyway?)


## Superfluid Undelegate
```go
type MsgSuperfluidUndelegate struct {
	Sender string
	LockId uint64
}
```