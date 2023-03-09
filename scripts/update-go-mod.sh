#!/bin/bash

# Script for checking `git diff` between two commits and updating osmoutils, osmomath or ibc-hooks if any were between two commits
# Used by Go Mod Auto Version Update workflow
# First argument: sha of a first commit
# Second argument: sha of a second commit

is_updated() {
    if [ "${1}" != "" ]
    then
        return 1
    fi
    return 0
}

commit_before=$1
commit_after=$2

changed_osmoutils=$(git diff --name-only $commit_before $commit_after | grep osmoutils)
changed_osmomath=$(git diff --name-only $commit_before $commit_after | grep osmomath)
changed_ibc_hooks=$(git diff --name-only $commit_before $commit_after | grep x/ibc-hooks)

is_updated $changed_osmoutils
update_osmoutils=$?

is_updated $changed_osmomath
update_osmomath=$?

is_updated $changed_ibc_hooks
update_ibc_hooks=$?

if [ $update_osmoutils -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/osmoutils@$commit_after
fi

if [ $update_osmomath -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/osmomath@$commit_after
fi

if [ $update_ibc_hooks -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/x/ibc-hooks@$commit_after
fi

go mod tidy