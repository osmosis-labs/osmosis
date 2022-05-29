#!/bin/bash

VERSION="v0.12.1"
CONTRACTS="cw20_base cw1_whitelist"

for CONTRACT in $CONTRACTS; do
  curl -s -L -O https://github.com/CosmWasm/cw-plus/releases/download/$VERSION/$CONTRACT.wasm
done