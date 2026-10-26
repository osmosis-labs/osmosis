#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

# Check if exactly 2 arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <upgrade_version> <upgrade_height>"
    exit 1
fi

# Extract arguments
UPGRADE_VERSION="$1"
UPGRADE_HEIGHT="$2"

# Replace the registry only after the command succeeds, including environment setup.
OUTPUT="$(mktemp "$SCRIPT_DIR/../../../chain.schema.json.XXXXXX")"
trap 'rm -f "$OUTPUT"' EXIT
uv run --project "$SCRIPT_DIR/.." --locked python "$SCRIPT_DIR/update_chain_registry.py" \
    --upgrade_version "$UPGRADE_VERSION" --upgrade_height "$UPGRADE_HEIGHT" > "$OUTPUT"
cat "$OUTPUT" > "$SCRIPT_DIR/../../../chain.schema.json"
