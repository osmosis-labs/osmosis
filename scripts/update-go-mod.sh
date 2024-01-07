#!/bin/bash

# Script for updating osmoutils, osmomath, epochs and ibc-hooks
# Used by Go Mod Auto Version Update workflow
# Argument: sha of commit on target branch

commit_after=$1

# UPDATE OSMOUTILS
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

# UPDATE OSMOMATH
go get github.com/osmosis-labs/osmosis/osmomath@$commit_after

# UPDATE IBC HOOKS
go get github.com/osmosis-labs/osmosis/x/ibc-hooks@$commit_after

# UPDATE EPOCHS
go get github.com/osmosis-labs/osmosis/x/epochs@$commit_after

go mod tidy
go work sync