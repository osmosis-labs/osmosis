<!--
order: 3
-->

# Messages

## Create Pot

`MsgCreatePot` can be submitted by any account to create a `Pot`.

```go
type MsgCreatePot struct {
	Owner    sdk.AccAddress
  DistributeTo []DistrCondition
  Rewards sdk.Coins
  StartTime    time.Time // start time to start distribution
  NumEpochs    uint64 // number of epochs distribution will be done
}
```

**State modifications:**

- Validate `Owner` has enough tokens for rewards
- Generate new `Pot` record
- Save the record inside the keeper's time basis unlock queue
- Transfer the tokens from the `Owner` to incentives `ModuleAccount`.

## Adding balance to Pot

`MsgAddToPot` can be submitted by any account to add more incentives to a `Pot`.

```go
type MsgAddToPot struct {
	PotID uint64
  Rewards sdk.Coins
}
```

**State modifications:**

- Validate `Owner` has enough tokens for rewards
- Check if `Pot` with specified `msg.PotID` is available
- Modify the `Pot` record by adding `msg.Rewards`
- Transfer the tokens from the `Owner` to incentives `ModuleAccount`.
