#!/bin/sh

# initialize osmosis
osmosisd init --chain-id=localosmosis val
# remove seeds
sed -i.bak -E 's#^(seeds[[:space:]]+=[[:space:]]+).*$#\1""#' ~/.osmosisd/config/config.toml
sed -i.bak -E 's#^(fast_sync[[:space:]]+=[[:space:]]+).*$#\1false#' ~/.chain-maind/config/config.toml
python3 testnetify.py
cp testnet_genesis.json .osmosisd/config/genesis.json