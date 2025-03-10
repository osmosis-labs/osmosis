<!--
order: 2
-->

# State

## State Objects

The `x/cron` module keeps the following object in the state: CronJob.

This object is used to store the state of a

- `CronJob` - to store the details of the cron jobs

```go
// this defines the details of the cronjob
message CronJob {
    // id is the unique identifier for the cron job
    uint64 id = 1;
    // name is the name of the cron job
    string name = 2;
    // description is the description of the cron job
    string description = 3;
    // Msgs that will be executed every period amount of time
    repeated MsgContractCron msg_contract_cron = 4 [(gogoproto.nullable) = false];
    // set cron enabled or not
    bool enable_cron = 5;
  }
```

```go
message MsgContractCron {
    // Contract is the address of the smart contract
    string contract_address = 1;
    // Msg is json encoded message to be passed to the contract
    string json_msg = 2;
  }
```

## Genesis & Params

The `x/cron` module's `GenesisState` defines the state necessary for initializing the chain from a previously exported height. It contains the module Parameters and Cron jobs. The params are used to control the Security Address which is responsible to register cron operations. This value can be modified with a governance proposal.

```go
// GenesisState defines the cron module's genesis state.
message GenesisState {
  Params params = 1 [
    (gogoproto.moretags) = "yaml:\"params\"",
    (gogoproto.nullable) = false
  ];
  repeated CronJob cron_jobs  = 2  [
    (gogoproto.moretags) = "yaml:\"cron_jobs\"",
    (gogoproto.nullable) = false
  ];
}
```

```go
// Params defines the parameters for the module.
message Params {
  // Security address that can whitelist/delist contract
  repeated string security_address = 1 [
    (gogoproto.jsontag) = "security_address,omitempty",
    (gogoproto.moretags) = "yaml:\"security_address\""
  ];
}
```

## State Transitions

The following state transitions are possible:

- Register the cron job
- Update the cron job
- Delete the cron job
