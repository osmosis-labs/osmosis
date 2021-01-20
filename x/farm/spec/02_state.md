<!--
order: 2
-->

# State

### Period
Only incrementally increasing value such as a block height or timestamp can be used.

```proto
message Farm {
    uint64 farmId = 1 [(gogoproto.moretags) = "yaml:\"farm_id\""];
    string totalShare = 2 [
        (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
        (gogoproto.moretags) = "yaml:\"total_share\"",
        (gogoproto.nullable) = false
    ];
    int64 current_period = 3 [(gogoproto.moretags) = "yaml:\"current_period\""];
    repeated cosmos.base.v1beta1.DecCoin current_rewards = 4 [
        (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.DecCoins",
        (gogoproto.nullable) = false
    ];
    int64 last_period = 5 [(gogoproto.moretags) = "yaml:\"last_period\""];
}
```
`x/farm` stores the shares deposited, currently ongoing period, and the rewards of the current period. It also stores the last processed priod.

```proto
message HistoricalRecord {
    repeated cosmos.base.v1beta1.DecCoin cumulative_reward_ratio = 1 [
        (gogoproto.moretags)     = "yaml:\"cumulative_reward_ratio\"",
        (gogoproto.castrepeated) = "github.com/cosmos/cosmos-sdk/types.DecCoins",
        (gogoproto.nullable)     = false
    ];
}
```
`HistoricalRecord` is stored in  "farm{farmId}/records/{period}" key and allows getting the historical record of the specified period and farmId.

```proto
message Farmer {
    uint64 farmId = 1 [(gogoproto.moretags) = "yaml:\"farm_id\""];
    string address = 2;
    string share = 3 [
        (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
        (gogoproto.moretags) = "yaml:\"share\"",
        (gogoproto.nullable) = false
    ];
    int64 last_withdrawn_period = 4 [(gogoproto.moretags) = "yaml:\"last_withdrawn_period\""];
}
```

