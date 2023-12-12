# Router

## Query


### Quote

```bash
curl "localhost:9092/router/quote?tokenIn=5000000uosmo&tokenOutDenom=uion" | jq .
```

### Pools

```bash
curl "localhost:9092/pools/all" | jq .
```

## Trade-offs To Re-evaluate

- Router skips found route if token OUT is found in the intermediary
path by calling `validateAndFilterRoutes` function
- Router skips found route if token IN is found in the intermediary
path by calling `validateAndFilterRoutes` function
- In the above 2 cases, we could exit early instead of continuing to search for such routes

## Manually Create Transmuter Commands

- uion / uosmo pool

```bash
# Gov prop for code upload
osmosisd tx gov submit-proposal upload-code-id-and-whitelist x/cosmwasmpool/bytecode/transmuter_migrate.wasm --title "T" --description "T" --from lo-test1 --keyring-backend test --chain-id localosmosis --gas=50000000 --fees 625000uosmo -b=block

# Deposit gov prop
osmosisd tx gov deposit 1 10000000uosmo --from val --keyring-backend test --chain-id localosmosis -b=block --fees=125000uosmo --gas=50000000

# Vote yes on gov prop
osmosisd tx gov vote 1 yes --from val --keyring-backend test --chain-id localosmosis -b=block --fees=6250uosmo --gas=2500000

# Create transmuter
osmosisd tx cosmwasmpool create-pool 1 "{\"pool_asset_denoms\":[\"uion\",\"uosmo\"]}" --from lo-test1 --keyring-backend test --chain-id localosmosis --fees 8750uosmo -b=block --gas=3500000

# Lp into transmuter
osmosisd tx wasm execute osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9 "{\"join_pool\":{}  }" --amount 1000000uosmo,2000000uion --from lo-test1 --keyring-backend test --chain-id localosmosis --fees 8750uosmo -b=block --gas=3500000
```

### Plan

1. Let's make sure that tick model can be retrieved from storage for CL pools independently. (complete)
2. Let's not retrieve tick model with pools. (complete)
3. Let's cache CL pools IDs together with routes. That way, we know what pool IDs need to query ticks (complete)
4. Only get all pools and taker fees if fail to get routes from Redis
5. Only get ticks for pools in the route, do not pre-load all ticks.

