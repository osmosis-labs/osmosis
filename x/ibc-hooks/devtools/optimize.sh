#!/bin/bash
set -o errexit -o nounset -o pipefail
command -v shellcheck >/dev/null && shellcheck "$0"

# This script lives in a ./devtools sub directory.
# Navigate to the repo root first.
SCRIPT_DIR="$(realpath "$(dirname "$0")")"
cd "${SCRIPT_DIR}/.."

# compile all contracts
for C in ./contracts/*/ ; do
  echo "Compiling $(basename "$C")..."
  (cd "$C" && RUSTFLAGS='-C link-arg=-s' cargo build --release --target wasm32-unknown-unknown --lib --locked)
done

mkdir -p ./wasm_outputs

for SRC in ./target/wasm32-unknown-unknown/release/*.wasm; do
  FILENAME=$(basename "$SRC")
  if command -v wasm-opt >/dev/null ; then
    wasm-opt -Os "$SRC" -o "./wasm_outputs/$FILENAME"
    chmod -x "./wasm_outputs/$FILENAME"
  else
    cp "$SRC" "./wasm_outputs/$FILENAME"
  fi
done

ls -l ./wasm_outputs
