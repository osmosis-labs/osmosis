#!/usr/bin/env bash

set -eox pipefail

# get protoc executions
go get github.com/cosmos/gogoproto/protoc-gen-gocosmos 2>/dev/null

cd proto
buf mod update
cd ..

echo "generating"
buf generate

cp -r ./github.com/osmosis-labs/osmosis/v*/x/* x/
rm -rf ./github.com
