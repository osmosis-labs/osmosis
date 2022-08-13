#!/bin/bash

set -euo pipefail

NEXT_MAJOR_VERSION=$1
import_path_to_replace=$(go list -m)

version_to_replace=$(echo $import_path_to_replace | sed 's/g.*v//') 

echo Current import paths are $version_to_replace, replacing with $NEXT_MAJOR_VERSION

# list all folders containing Go modules.
modules=$(go list -tags e2e ./... | sed "s/g.*v${version_to_replace}\///")

replace_paths() {
    file="${1}"
    sed -i "s/github.com\/osmosis-labs\/osmosis\/v${version_to_replace}/github.com\/osmosis-labs\/osmosis\/v${NEXT_MAJOR_VERSION}/g" ${file}
}

echo "Replacing import paths in all files"

files=$(find ./ -type f -and -not \( -path "./vendor*" -or -path "./.git*" -or -name "*.md" \))

echo "Updating all files"
for file in $files; do
    replace_paths ${file}
done

echo "Updating go.mod and vendoring"
# go.mod
replace_paths "go.mod"
go mod vendor >/dev/null

# ensure that generated files are updated.
# N.B.: This must be run after go mod vendor.
echo "running make proto-gen"
make proto-gen >/dev/null
