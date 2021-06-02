<!--
order: 7
-->

# Queries

In this section we describe the queries required on grpc server.

```protobuf
// Query defines the gRPC querier service.
service Query {
	// returns coins that is going to be distributed
	rpc ModuleToDistributeCoins(ModuleToDistributeCoinsRequest) returns ModuleToDistributeCoinsResponse;
	// returns coins that are distributed by module so far
	rpc ModuleDistributedCoins(ModuleDistributedCoinsRequest) returns ModuleToDistributeCoinsResponse;

	// returns Gauge by id
	rpc GaugeByID(GaugeByIDRequest) returns GaugeByIDResponse;
	// returns gauges both upcoming and active
	rpc Gauges(GaugesRequest) returns GaugesResponse;
	// returns active gauges
	rpc ActiveGauges(ActiveGaugesRequest) returns ActiveGaugesResponse;
	// returns scheduled gauges
	rpc UpcomingGauges(UpcomingGaugesRequest) returns UpcomingGaugesResponse;
	// returns rewards estimation at a future specific time
	rpc RewardsEst(RewardsEstRequest) returns RewardsEstResponse;
}
```