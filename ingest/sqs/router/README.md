# Router

## Algorithm

In this section, we describe the general router algorithm.

1. Retrieve pools from storage.
2. Filter out low liquidity pools.
3. Rank pools by several heuristics such as:
 - liquidity
 - pool type (priority: transmuter, concentrated, stableswap, balancer)
 - presence of error in TVL computation.
4. Compute candidate routes
   * For the given token in and token out denom, find all possible routes
   between them using the pool ranking discussed above as well as by limiting
   the algorithm per configuration.
   * The configurations are:
      * Max Hops: The maximum number of hops allowed in a route.
      * Max Routes: The maximum number of routes to consider.
   * The algorithm that is currently used is breadth first search.
5. Compute the best quote when swapping amount in in-full directly over each route.
6. Sort routes by best quote.
7. Keep "Max Splittable Routes" and attempt to determine an optimal quote split across them
   * If the split quote is more optimal, return that. Otherwise, return the best single direct quote.

## Caching

We perform caching of routes to avoid having to recompute them on every request.
The routes are cached in a Redis instance.

There is a configuration parameter that enables the route cache to be updated every X blocks.
However, that is an experimental feature. See the configuration section for details.

The router also caches the routes when it computes it for the first time for a given token in and token out denom.
As of now, the cache is cleared at the end of very block. We should investigate only clearing pool data but persisting
the routes for longer while allowing for manual updates and invalidation.

## Configuration

The router has several configuration parameters that are set via `app.toml`.

See the recommended enabled configuration below:
```toml
###############################################################################
###              Osmosis Sidecar Query Server Configuration                 ###
###############################################################################

[osmosis-sqs]

# SQS service is disabled by default.
is-enabled = "true"

# The hostname and address of the sidecar query server storage.
db-host = "localhost"
db-port = "6379"

# Defines the web server configuration.
server-address = ":9092"
timeout-duration-secs = "2"

# Defines the logger configuration.
logger-filename = "sqs.log"
logger-is-production = "true"
logger-level = "info"

# Defines the gRPC gateway endpoint of the chain.
grpc-gateway-endpoint = "http://localhost:26657"

# The list of preferred poold IDs in the router.
# These pools will be prioritized in the candidate route selection, ignoring all other
# heuristics such as TVL.
preferred-pool-ids = []

# The maximum number of pools to be included in a single route.
max-pools-per-route = "4"

# The maximum number of routes to be returned in candidate route search.
max-routes = "20"

# The maximum number of routes to be split across. Must be smaller than or
# equal to max-routes.
max-split-routes = "3"

# The maximum number of iterations to split a route across.
max-split-iterations = "10"

# The minimum liquidity of a pool to be included in a route.
min-osmo-liquidity = "10000"

# The height interval at which the candidate routes are recomputed and updated in
# Redis
route-update-height-interval = "0"

# Whether to enable candidate route caching in Redis.
route-cache-enabled = "true"
```
