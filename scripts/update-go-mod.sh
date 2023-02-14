#!/bin/bash

is_updated() {
    if [ "${1}" != "" ]
    then
        return 1
    fi
    return 0
}

main_commit=$(git rev-parse origin/$GITHUB_BASE_REF)
head_commit=$(git rev-parse origin/$GITHUB_HEAD_REF)

changed_osmoutils=$(git diff --name-only $main_commit $head_commit | grep osmoutils)
changed_osmomath=$(git diff --name-only $main_commit $head_commit | grep osmomath)
changed_ibc_hooks=$(git diff --name-only $main_commit $head_commit | grep x/ibc-hooks)

is_updated $changed_osmoutils
update_osmoutils=$?

is_updated $changed_osmomath
update_osmomath=$?

is_updated $changed_ibc_hooks
update_ibc_hooks=$?

any_updated=0 # we do not want to run `go mod tidy`` in case none of these files have changed

if [ $update_osmoutils -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/osmoutils@$head_commit
    any_updated=1
fi

if [ $update_osmomath -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/osmomath@$head_commit
    any_updated=1
fi

if [ $update_ibc_hooks -eq 1 ]
then 
    go get github.com/osmosis-labs/osmosis/x/ibc-hooks@$head_commit
    any_updated=1
fi

if [ $any_updated -eq 1 ]
then
    go mod tidy
fi