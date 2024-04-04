#!/bin/bash

set -euo pipefail

# http_get function downloads a file from a URL and saves it to a specified output file.
# if the HTTP status code is not 200 OK, the function will print an error message and exit.
function http_get() {
    # Check if the output file argument is provided
    if [[ $# -lt 2 ]]; then
        echo "Usage: http_get <URL> output_file"
        return 1
    fi

	# Define the URL and output file
	url=$1
    output_file=$2

	# Download the file, redirect output to specified file and get the HTTP status code
    status=$(curl -v --silent --output "$output_file" --write-out "%{response_code}" "$url")
	
    # Remove leading zeros (if any) to prevent bash interpreting result as octal
    status=${status##+(0)}

	# Check if the HTTP status code is not 200
	if [[ $status -ne 200 ]]; then
        echo "ERROR - Request failed with ${status}"
		exit 1
	fi
}

# Download mainnet asset list
http_get \
	"https://raw.githubusercontent.com/osmosis-labs/assetlists/main/osmosis-1/osmosis-1.assetlist.json" \
	./cmd/osmosisd/cmd/osmosis-1-assetlist.json

# Download testnet asset list
http_get \
	"https://raw.githubusercontent.com/osmosis-labs/assetlists/main/osmo-test-5/osmo-test-5.assetlist.json" \
	./cmd/osmosisd/cmd/osmo-test-5-assetlist.json
