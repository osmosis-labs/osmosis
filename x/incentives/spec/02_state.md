<!--
order: 2
-->

# State

## Incentives management

All the incentives that is going to be provided are locked into `IncentivePool` until it's released to the users after a specific period of time.

### Pot

Rewards to be distributed are organized by `Pot`. The `Pot` describes how users can get reward, stores the amount of coins in the pot, the cadence at which rewards are to be distributed, and the number of epochs to distribute the reward over.

```go
type LockType int
const (
  ByLockDuration LockType = iota // locks which has more than specific duration
  ByLockTime // locks which are started before specific time
)

type DistrCondition struct {
  LockType LockType // type of lock condition
  Denom  string // lock denom
  Duration time.Duration // condition for lock duration, only valid if positive
  Timestamp time.Time // condition for lock start time, not valid if unset value
}

type Pot struct {
  ID           uint64 // unique ID of a Pot
  DistributeTo []DistrCondition // distribute condition of a lock which meet one of these conditions
  TotalRewards sdk.Coins // can distribute multiple coins
  StartTime    time.Time // start time to start distribution
  NumEpochs    uint64 // number of epochs distribution will be done 
}
```

```protobuf
enum LockType {
    option (gogoproto.goproto_enum_prefix) = false;

    by_duration = 0; // locks which has more than specific duration
    by_time = 1; // locks which are started before specific time
}

message DistrCondition {
  LockType lock_type = 1; // type of lock, ByLockDuration | ByLockTime
  string denom = 2; // lock denom
  google.protobuf.Duration duration = 3; // condition for lock duration, only valid if positive
  google.protobuf.Timestamp timestamp = 4; // condition for lock start time, not valid if unset value
}

message Pot {
  uint64 id = 1; // unique ID of a Pot
  repeated DistrCondition distribute_to = 2; // distribute condition of a lock which meet one of these conditions
  repeated cosmos.base.v1beta1.Coin coins = 3; // can distribute multiple coins
  google.protobuf.Timestamp start_time = 4; // condition for lock start time, not valid if unset value
  uint64 num_epochs = 5; // number of epochs distribution will be done 
}
```

### Pot queues

#### Upcoming queue

To release from `Pots` every epoch, we schedule distribution start time with time key queue.

#### Active queue

Active queue has all the `Pots` that are distributing and after distribution period finish, it's removed from the queue.
