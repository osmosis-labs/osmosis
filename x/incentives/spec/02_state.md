<!--
order: 2
-->

# State

## Incentives management

All the incentives that are going to be provided are locked into `IncentivePool` until released to the appropriate recipients after a specific period of time.

### Gauge

Rewards to be distributed are organized by `Gauge`. The `Gauge` describes how users can get reward, stores the amount of coins in the gauge, the cadence at which rewards are to be distributed, and the number of epochs to distribute the reward over.

```protobuf
enum LockQueryType {
  option (gogoproto.goproto_enum_prefix) = false;

  ByDuration = 0; // locks which has more than specific duration
  ByTime = 1; // locks which are started before specific time
}

message QueryCondition {
  LockQueryType lock_query_type = 1; // type of lock, ByLockDuration | ByLockTime
  string denom = 2; // lock denom
  google.protobuf.Duration duration = 3; // condition for lock duration, only valid if positive
  google.protobuf.Timestamp timestamp = 4; // condition for lock start time, not valid if unset value
}

message Gauge {
  uint64 id = 1; // unique ID of a Gauge
  QueryCondition distribute_to = 2; // distribute condition of a lock which meet one of these conditions
  repeated cosmos.base.v1beta1.Coin coins = 3; // can distribute multiple coins
  google.protobuf.Timestamp start_time = 4; // condition for lock start time, not valid if unset value
  uint64 num_epochs_paid_over = 5; // number of epochs distribution will be done 
}
```

### Gauge queues

#### Upcoming queue

To start release `Gauges` at a specific time, we schedule distribution start time with time key queue.

#### Active queue

Active queue has all the `Gauges` that are distributing and after distribution period finish, it's removed from the queue.
