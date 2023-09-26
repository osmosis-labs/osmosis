#!/bin/bash

# Check if exactly 2 arguments are passed
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <upgrade_version> <upgrade_height>"
    exit 1
fi

# Extract arguments
UPGRADE_VERSION="$1"
UPGRADE_HEIGHT="$2"

# Run the Python script and redirect output to chain.schema.json
python update_chain_registry.py --upgrade_version "$UPGRADE_VERSION" --upgrade_height "$UPGRADE_HEIGHT" > ../../../chain.schema.json
