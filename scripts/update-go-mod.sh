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

# Define modules
modules=("osmoutils" "osmomath" "x/ibc-hooks" "x/epochs")

# Find all go.mod files in the repo
go_mod_files=$(find . -name go.mod)

# Loop over each go.mod file
for file in $go_mod_files; do
  # Get the directory of the go.mod file
  dir=$(dirname $file)

  # Change to that directory
  cd $dir

  # Loop over each module
  for module in ${modules[@]}; do
    # Check if the module is a direct requirement
    if grep -q "github.com/osmosis-labs/osmosis/$module" go.mod; then
      # If it is, run go get with the provided commit
      go get "github.com/osmosis-labs/osmosis/$module@$commit_after"
    fi
  done

  # Run go mod tidy and go work sync
  go mod tidy
  go work sync

  # Return to the root directory
  cd - > /dev/null
done
