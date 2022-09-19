# [WIP] Interchain Name Service

The `Interchain Name Service` module allows for the mapping of [interchain accounts](https://github.com/cosmos/interchain-accounts-demo) to a human-readable name.

TODO:

- [x] Initial proof-of-concept contract
- [x] Setup beaker and deploy system
- [x] Registrar and name lookup
- [x] Annual rent mechanism (subject to change)
- [ ] Reverse registrar (given an address, look up a name if it exists)
- [ ] Name integrates with interchain accounts, finalize name format
- [ ] Potential harbebger / anti-squatting [mechanism](https://vitalik.eth.limo/general/2022/09/09/ens.html)
- [ ] Potentially integrate with [interchain NFTs](https://github.com/cosmos/ibc/blob/main/spec/app/ics-721-nft-transfer/README.md)?

# Local setup

1. Follow the instructions to install [localosmosis](https://docs.osmosis.zone/developing/dapps/get_started/cosmwasm-localosmosis.html#setup-localosmosis) and set up a [local key](https://docs.osmosis.zone/developing/dapps/get_started/cosmwasm-localosmosis.html#created-a-local-key).

2. Install [beaker](https://docs.osmosis.zone/developing/tools/beaker/#installation) and then setup the initial rust project.

```
cd x/interchain-name-service
cargo build
```

3. Compile, deploy, and instantiate the `name-service` contract.

```
beaker wasm deploy name-service --signer-account test1 --no-wasm-opt  --raw '{"purchase_price":{"amount":"100","denom":"uosmo"},"transfer_price":{"amount":"999","denom":"uosmo"},"annual_rent_amount":"20"}'
```

4. Execute an example transaction on localosmosis!

```
beaker wasm execute name-service --raw '{"register":{"name":"johndoe","years":"1"}}' --signer-account test1 --funds 120uosmo
```
