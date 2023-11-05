# Sidecar Query Server

This is a sidecar query server that is used for performing query tasks outside of the main chain.
The high-level architecture is that the chain reads data at the end of the block, parses it
and then writes into a Redis instance.

The sidecar query server then reads the parsed data from Redis and serves it to the client
via HTTP endpoints.

The use case for this is performing certain data and computationally intensive tasks outside of
the chain node or the clients. For example, routing falls under this category because it requires
all pool data for performing the complex routing algorithm.

## Data

### Pools

For every chain pool, its pool model is written to Redis.

Additionally, we instrument each pool model with bank balances and OSMO-denominated TVL.

Some pool models do not contain balances by default. As a result, clients have to requery balance
for each pool model directly from chain. Having the balances in Redis allows us to avoid this and serve
pools with balances directly.

The routing algorithm requires the knowledge of TVL for prioritizing pools. As a result, each pool model
is instrumented with OSMO-denominated TVL.

### Router

For routing, we must know about the taker fee for every denom pair. As a result, in the router
repository, we stote the taker fee keyed by the denom pair.

These taker fees are then read from Redis to initialize the router.

## Open Questions

- How to handle atomicity between ticks and pools? E.g. let's say a block is written between the time initial pools are read
and the time the ticks are read. Now, we have data that is partially up-to-date.