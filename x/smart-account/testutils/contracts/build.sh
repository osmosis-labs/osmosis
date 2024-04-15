#!/usr/bin/env bash

set -euxo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# for all dirs in SCRIPT_DIR
for dir in $SCRIPT_DIR/*; do
    # if dir is a directory
    if [ -d "$dir" ]; then
        # if dir has a Cargo.toml
        if [ -f "$dir/Cargo.toml" ]; then
            # build the contract
            pushd $dir
            cargo wasm
            cp target/wasm32-unknown-unknown/release/*.wasm "$SCRIPT_DIR/../bytecode/"
            popd
        fi
    fi
done



