<!--
order: 3
-->

# Messages

## Create Gauge

`MsgCreateGauge` can be submitted by any account to create a `Gauge`.

```go
type MsgCreateGauge struct {
	Owner             sdk.AccAddress
  DistributeTo      QueryCondition
  Rewards           sdk.Coins
  StartTime         time.Time // start time to start distribution
  NumEpochsPaidOver uint64 // number of epochs distribution will be done
}
```

**State modifications:**

- Validate `Owner` has enough tokens for rewards
- Generate new `Gauge` record
- Save the record inside the keeper's time basis unlock queue
- Transfer the tokens from the `Owner` to incentives `ModuleAccount`.

## Adding balance to Gauge

`MsgAddToGauge` can be submitted by any account to add more incentives to a `Gauge`.

```go
type MsgAddToGauge struct {
	GaugeID uint64
  Rewards sdk.Coins
}
```

**State modifications:**

- Validate `Owner` has enough tokens for rewards
- Check if `Gauge` with specified `msg.GaugeID` is available
- Modify the `Gauge` record by adding `msg.Rewards`
- Transfer the tokens from the `Owner` to incentives `ModuleAccount`.

## Claim Lock Reward

`MsgClaimLockReward` can be submitted by any account to claim accumulated rewards for a lock.

```go
type MsgClaimLockReward struct {
  Owner string
  ID    uint64
}
```

**State modifications:**

- Validate `Owner` has any rewards to claim
- Check if `PeriodLock` with specified `msg.ID` is available
- Transfer rewards to `Owner`

## Claim Lock Reward All

`MsgClaimLockRewardAll` can be submitted by any account to claim accumulated rewards for all locks.

```go
type MsgClaimLockRewardAll struct {
  Owner string
}
```

**State modifications:**

- Validate `Owner` has any rewards to claim
- Transfer rewards to `Owner`