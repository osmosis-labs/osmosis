#!/usr/bin/env bash

set -eo pipefail

cd proto
proto_dirs=$(find ./osmosis -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep go_package $file &>/dev/null; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

# move proto files to the right places
#
# Note: Proto files are suffixed with the current binary version.
cp -r github.com/osmosis-labs/osmosis/v13/* ./
rm -rf github.com

go mod tidy -compat=1.18
