# Router

## Query


## Quote

```bash
curl "localhost:9092/quote?tokenIn=5000000uosmo&tokenOutDenom=uusdc" | jq .
```

### Pools

```bash
curl "localhost:9092/all-pools" | jq .
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
osmosisd tx gov submit-proposal upload-code-id-and-whitelist x/cosmwasmpool/bytecode/transmuter_migrate.wasm --title "T" --description "T" --from lo-test1 --keyring-backend test --chain-id localosmosis --gas=50000000 --fees 625000uosmo -b=block

osmosisd tx gov deposit 1 10000000uosmo --from val --keyring-backend test --chain-id localosmosis -b=block --fees=125000uosmo --gas=50000000

osmosisd tx gov vote 1 yes --from val --keyring-backend test --chain-id localosmosis -b=block --fees=6250uosmo --gas=2500000

osmosisd tx cosmwasmpool create-pool 1 "{\"pool_asset_denoms\":[\"uion\",\"uosmo\"]}" --from lo-test1 --keyring-backend test --chain-id localosmosis --fees 8750uosmo -b=block --gas=3500000
```
