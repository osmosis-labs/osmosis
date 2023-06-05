#!/bin/bash

set -euo pipefail

NEXT_MAJOR_VERSION=$1
import_path_to_replace=$(go list -m)

version_to_replace=$(echo $import_path_to_replace | sed -n 's/.*v\([0-9]*\).*/\1/p') 

echo $version_to_replace
echo Current import paths are $version_to_replace, replacing with $NEXT_MAJOR_VERSION

# list all folders containing Go modules.
modules=$(go list -tags e2e ./... | sed "s/g.*v${version_to_replace}\///")

while IFS= read -r line; do
  modules_to_upgrade_manually+=("$line")
done < <(find . -name go.mod -exec grep -l "github.com/osmosis-labs/osmosis/v16" {} \; | grep -v  "^./go.mod$" | sed 's|/go.mod||' | sed 's|^./||')

replace_paths() {
    file="${1}"
    sed -i "s/github.com\/osmosis-labs\/osmosis\/v${version_to_replace}/github.com\/osmosis-labs\/osmosis\/v${NEXT_MAJOR_VERSION}/g" ${file}
}

echo "Replacing import paths in all files"

while IFS= read -r line; do
  files+=("$line")
done < <(find ./ -type f -not \( -path "./vendor*" -or -path "./.git*" -or -name "*.md" \))

echo "Updating all files"

for file in "${files[@]}"; do
    if test -f "$file"; then
        # skip files that need manual upgrading 
        for excluded_file in "${modules_to_upgrade_manually[@]}"; do
            if [[ "$file" == *"$excluded_file"* ]]; then
                continue 2
            fi
        done
        replace_paths $file
    fi
done

exit 0

echo "Updating go.mod and vendoring"
# go.mod
replace_paths "go.mod"
go mod vendor >/dev/null

# ensure that generated files are updated.
# N.B.: This must be run after go mod vendor.
echo "running make proto-gen"
make proto-gen >/dev/null

echo "Run go mod vendor after proto-gen to avoid vendoring issues"
go mod vendor >/dev/null

echo "running make run-querygen"
make run-querygen >/dev/null
