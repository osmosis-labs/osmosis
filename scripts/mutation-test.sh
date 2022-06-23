#!/usr/bin/env bash

set -eo pipefail

# Only consider the following:
# * go files in types or keeper packages
# * ignore test and Protobuf files
MUTATION_SOURCES=$(find ./x -type f \( -path '*/keeper/*' -or -path '*/types/*' \) \( -name '*.go' -and -not -name '*_test.go' -and -not -name '*pb*' \))

# XXX: Filter on a module-by-module basis and expand when we think other modules
# are ready.
MUTATION_SOURCES=$(echo "$MUTATION_SOURCES" | grep './x/tokenfactory')

# Collect multiple lines into a single line to be fed into go-mutesting
MUTATION_SOURCES=$(echo $MUTATION_SOURCES | tr '\n' ' ')

OUTPUT=$(go run github.com/zimmski/go-mutesting/cmd/go-mutesting --blacklist=mutation.blacklist $MUTATION_SOURCES)
echo "$OUTPUT" | grep 'The mutation score'
