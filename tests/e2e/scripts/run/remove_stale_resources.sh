#!/bin/bash

# Filtering by containers belonging to the "osmosis-testnet" network
LIST_CONTAINERS_CMD=$(docker ps -a --filter network=osmosis-testnet --format {{.ID}})
LIST_NETWORKS_CMD=$(docker network ls --filter name=osmosis-testnet --format {{.ID}})

if [[ "$LIST_CONTAINERS_CMD" != "" ]]; then
    echo "Removing stale e2e containers"
    docker container rm -f $LIST_CONTAINERS_CMD
else
    echo "No stale e2e containers found"
fi

if [[ "$LIST_NETWORKS_CMD" != "" ]]; then
    echo "Removing stale e2e networks"
    docker network rm $LIST_NETWORKS_CMD
else
    echo "No stale e2e networks found"
fi
