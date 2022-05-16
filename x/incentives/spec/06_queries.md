```{=html}
<!--
order: 7
-->
```

Queries
=======

In this section we describe the queries required on grpc server.

``` {.protobuf}
// Query defines the gRPC querier service.
service Query {
  // returns coins that is going to be distributed
  rpc ModuleToDistributeCoins(ModuleToDistributeCoinsRequest) returns (ModuleToDistributeCoinsResponse) {}
  // returns coins that are distributed by module so far
  rpc ModuleDistributedCoins(ModuleDistributedCoinsRequest) returns (ModuleDistributedCoinsResponse) {}
  // returns Gauge by id
  rpc GaugeByID(GaugeByIDRequest) returns (GaugeByIDResponse) {}
  // returns gauges both upcoming and active
  rpc Gauges(GaugesRequest) returns (GaugesResponse) {}
  // returns active gauges
  rpc ActiveGauges(ActiveGaugesRequest) returns (ActiveGaugesResponse) {}
  // returns scheduled gauges
  rpc UpcomingGauges(UpcomingGaugesRequest) returns (UpcomingGaugesResponse) {}
  // RewardsEst returns an estimate of the rewards at a future specific time.
  // The querier either provides an address or a set of locks
  // for which they want to find the associated rewards.
  rpc RewardsEst(RewardsEstRequest) returns (RewardsEstResponse) {}
  // returns lockable durations that are valid to give incentives
  rpc LockableDurations(QueryLockableDurationsRequest) returns (QueryLockableDurationsResponse) {}
}
```
