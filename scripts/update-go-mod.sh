#!/bin/bash

check_update() {
    if [ "${1}" != "" ]
    then
        return 1
    fi
    return 0
}

main_commit=$(git rev-parse origin/$GITHUB_BASE_REF)
head_commit=$(git rev-parse origin/$GITHUB_HEAD_REF)
echo main commit: $main_commit
echo head commit: $head_commit


changed_osmoutils=$(git diff --name-only $main_commit $head_commit | grep osmoutils)
changed_osmomath=$(git diff --name-only $main_commit $head_commit | grep osmomath)
changed_ibc_hooks=$(git diff --name-only $main_commit $head_commit | grep x/ibc-hooks)

check_update $changed_osmoutils
update_osmoutils=$?

check_update $changed_osmomath
update_osmomath=$?

check_update $changed_ibc_hooks
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

echo any updated $any_updated
if [ $any_updated -eq 1 ]
then
    go mod tidy
    echo "MAKE_PULL_REQUEST=1" >> $GITHUB_ENV
fi
exit