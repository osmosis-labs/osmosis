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

	// returns Pot by id
	rpc PotByID(PotByIDRequest) returns PotByIDResponse;
	// returns pots both upcoming and active
	rpc Pots(PotsRequest) returns PotsResponse;
	// returns active pots
	rpc ActivePots(ActivePotsRequest) returns ActivePotsResponse;
	// returns scheduled pots
	rpc UpcomingPots(UpcomingPotsRequest) returns UpcomingPotsResponse;
	// returns rewards estimation at a future specific time
	rpc RewardsEst(RewardsEstRequest) returns RewardsEstResponse;
}
```