#!/bin/bash

set -euo pipefail

NEXT_MAJOR_VERSION=$1
import_path_to_replace=$(go list -m)

version_to_replace=$(echo $import_path_to_replace | sed 's/g.*v//') 

echo Current import paths are $version_to_replace, replacing with $NEXT_MAJOR_VERSION

# list all folders containing Go modules.
modules=$(go list ./... | sed "s/g.*v${version_to_replace}\///")

replace_paths() {
    file="${1}"
    sed -i "s/github.com\/osmosis-labs\/osmosis\/v${version_to_replace}/github.com\/osmosis-labs\/osmosis\/v${NEXT_MAJOR_VERSION}/g" ${file}
}

echo "Replacing import paths in Go files"
for mod in $modules;
do
    for file in $mod/*; do
        if [ -f "${file}" ]; then
            replace_paths $file
        fi
    done
done

echo "Replacing import paths in proto files"
for file in $(find proto/osmosis -type f -name "*.proto"); do
    replace_paths $file
done

# protocgen.sh
replace_paths "scripts/protocgen.sh"

echo "Updating go.mod and vendoring"
# go.mod
replace_paths "go.mod"
go mod vendor >/dev/null

echo "running make proto-gen"
# ensure protos match generated Go files
# N.B.: This must be run after go mod vendor.
make proto-gen >/dev/null
