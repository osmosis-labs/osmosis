#!/usr/bin/env bash

set -eo pipefail

# get protoc executions
go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos 2>/dev/null

cd proto
buf mod update
cd ..
buf generate

cp -r ./github.com/osmosis-labs/osmosis/v*/x/* x/
rm -rf ./github.com
