#!/bin/bash

# Script for checking `git diff` between two commits and updating osmoutils, osmomath, epochs or ibc-hooks if any were changed between two commits
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
changed_epochs=$(git diff --name-only $commit_before $commit_after | grep x/epochs)

is_updated $changed_osmoutils
update_osmoutils=$?

is_updated $changed_osmomath
update_osmomath=$?

is_updated $changed_ibc_hooks
update_ibc_hooks=$?

is_updated $changed_epochs
update_epochs=$?

if [ $update_osmoutils -eq 1 ]
then 
	go get github.com/osmosis-labs/osmosis/osmoutils@$commit_after

    # x/epochs depends on osmoutils
    cd x/epochs
    go get github.com/osmosis-labs/osmosis/osmoutils@$commit_after
    go mod tidy

    # x/ibc-hooks depends on osmoutils
    cd ../ibc-hooks
    go get github.com/osmosis-labs/osmosis/osmoutils@$commit_after
    go mod tidy
    
    # return to root
    cd ../..
fi

if [ $update_osmomath -eq 1 ]
then 
	go get github.com/osmosis-labs/osmosis/osmomath@$commit_after
fi

if [ $update_ibc_hooks -eq 1 ]
then 
	go get github.com/osmosis-labs/osmosis/x/ibc-hooks@$commit_after
fi

if [ $update_epochs -eq 1 ]
then 
	go get github.com/osmosis-labs/osmosis/x/epochs@$commit_after
fi

go mod tidy
go work sync