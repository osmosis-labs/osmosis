#!/bin/bash
# script parameters, in same order as: osmoutils, osmomath, ibc-hooks

check_update() {
    if [ "${1}" != "" ]
    then
        return 1
    fi
    return 0
}

check_update $1
update_osmoutils=$?

check_update $2
update_osmomath=$?

check_update $3
update_ibc_hooks=$?

latest_commit=$(git rev-parse HEAD)
any_updated=0 # we do not want to run `go mod tidy`` in case none of these files have changed

if [ $update_osmoutils -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/osmoutils@$latest_commit
    any_updated=1
fi

if [ $update_osmomath -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/osmomath@$latest_commit
    any_updated=1
fi

if [ $update_ibc_hooks -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/x/ibc-hooks@$latest_commit
    any_updated=1
fi

if [ $any_updated -eq 1 ]
then
    go mod tidy
    exit 1
fi