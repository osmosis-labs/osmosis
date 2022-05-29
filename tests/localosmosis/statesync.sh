#!/bin/bash

# initialize osmosis
osmosisd init --chain-id=localosmosis val
# remove seeds
sed -i.bak -E 's#^(seeds[[:space:]]+=[[:space:]]+).*$#\1""#' ~/.osmosisd/config/config.toml
sed -i.bak -E 's#^(fast_sync[[:space:]]+=[[:space:]]+).*$#\1false#' ~/.osmosisd/config/config.toml
python3 /osmosis/testnetify.py
cp testnet_genesis.json .osmosisd/config/genesis.json