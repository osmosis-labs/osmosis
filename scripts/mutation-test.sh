#!/usr/bin/env bash

set -eo pipefail

DISABLED_MUTATORS='branch/*'

# Only consider the following:
# * go files in types or keeper packages
# * ignore test and Protobuf files
MUTATION_SOURCES=$(find ./x -type f \( -path '*/keeper/*' -or -path '*/types/*' \) \( -name '*.go' -and -not -name '*_test.go' -and -not -name '*pb*' \))

# XXX: Filter on a module-by-module basis and expand when we think other modules
# are ready. Once all modules are considered stable enough to be tested, remove
# this filter entirely.
MUTATION_SOURCES=$(echo "$MUTATION_SOURCES" | grep './x/tokenfactory')

# Collect multiple lines into a single line to be fed into go-mutesting
MUTATION_SOURCES=$(echo $MUTATION_SOURCES | tr '\n' ' ')

OUTPUT=$(go run github.com/osmosis-labs/go-mutesting/cmd/go-mutesting --disable=$DISABLED_MUTATORS $MUTATION_SOURCES)

# Fetch the final result output and the overall mutation testing score
RESULT=$(echo "$OUTPUT" | grep 'The mutation score')
SCORE=$(echo "$RESULT" | grep -Eo '[[:digit:]]\.[[:digit:]]+')

echo "writing mutation test result to mutation_test_result.txt"
echo "$OUTPUT" > mutation_test_result.txt

echo $RESULT

# Return a non-zero exit code if the score is below 75%
if (( $(echo "$SCORE < 0.75" |bc -l) )); then
  echo "Mutation testing score below desired level ($SCORE < 0.75)"
  exit 1
fi
