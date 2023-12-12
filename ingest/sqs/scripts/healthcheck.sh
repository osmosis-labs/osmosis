#!/bin/bash

validate_http_code() {
    local http_status=$1
    local check_url=$2
    # Check if the HTTP status code is 200
    if [ "$http_status" -eq 200 ]; then
        echo "Status check successful (HTTP 200 OK) " $check_url
    else
        echo "Health check failed with HTTP $http_status"
        error_message=$(curl -s -o /dev/null -w "%{stderr}" $check_url)
        echo "Error message: $error_message"
        exit 1
    fi
}

# This script performs a health check on the node by making a GET request to the health check URL
# and checking if the HTTP status code is 200. If the HTTP status code is not 200, the script
# will exit with a non-zero exit code and print the error message.
perform_health_check() {
    local url=$1

    health_check_url=$url/healthcheck

    # Make a GET request to the health check URL and capture the HTTP status code and error message
    code_str=$(curl --write-out '%{http_code}' --silent --output /dev/null $health_check_url)

    validate_http_code $code_str $health_check_url
}

# This script performs a status check on the node by making a POST request to the status check URL via grpc
# and checking if the HTTP status code is 200. If the HTTP status code is not 200, the script
# will exit with a non-zero exit code and print the error message.
# It also prints height and whether a node is syncing or not.
perform_status_check() {
    local status_check_url=$1

    full_status_url=$status_check_url/status

    code_str=$(curl --write-out '%{http_code}' --silent --output /dev/null $full_status_url)

    validate_http_code $code_str $full_status_url

    # Make a POST request to the health check URL and capture the full response
    full_response=$(curl -X POST -H "Content-Type: application/json" -d '{
    "jsonrpc": "2.0",
    "method": "status",
    "id": 1
    }' $status_check_url)

    # Extract the status code using jq
    block_height=$(echo "$full_response" | jq .result.sync_info.latest_block_height)
    echo "Height" $block_height

    is_syncing=$(echo "$full_response" | jq .result.sync_info.catching_up)
    echo "Is Synching" $is_syncing
}
