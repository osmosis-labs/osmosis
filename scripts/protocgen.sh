#!/usr/bin/env bash

set -eo pipefail

go get github.com/cosmos/gogoproto 2>/dev/null
go get github.com/cosmos/cosmos-sdk 2>/dev/null

echo "Generating gogo proto code"
ls -a
cd proto
ls -a

proto_dirs=$(find ./osmosis -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep go_package $file &>/dev/null; then
      echo "Generating gogo proto code for $file"
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done
ls -a
cd ..
ls -a

# move proto files to the right places
#
# Note: Proto files are suffixed with the current binary version.
cp -r github.com/osmosis-labs/osmosis/v20/* ./
cp -r github.com/osmosis-labs/osmosis/osmoutils ./
cp -r github.com/osmosis-labs/osmosis/x/epochs ./x/
#rm -rf github.com

#go mod tidy

# TODO: Uncomment once ORM/Pulsar support is needed.
#
# Ref: https://github.com/osmosis-labs/osmosis/pull/1589