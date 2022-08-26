#!/usr/bin/env bash

set -eo pipefail
oIFS="$IFS"; IFS=, ; set -- $1 ; IFS="$oIFS"

DISABLED_MUTATORS='branch/*'

# Only consider the following:
# * go files in types, keeper, or module root directories
# * ignore test and Protobuf files
go_file_exclusions="-name '*.go' -and -not -name '*_test.go' -and -not -name '*pb*' -and -not -name 'module.go'"
MUTATION_SOURCES=$(find ./x -type f -path '*/keeper/*' -or -path '*/types/*' ${!go_file_exclusions} )
MUTATION_SOURCES+=$(find ./x -maxdepth 2 -type f ${!go_file_exclusions} )

# Filter on a module-by-module basis as provided by input
arg_len=$#

for i in "$@"; do
  if [ $arg_len -gt 1 ]; then
    MODULE_FORMAT+="./x/$i\|"
    MODULE_NAMES+="${i} "
    let "arg_len--"
  else
    MODULE_FORMAT+="./x/$i"
    MODULE_NAMES+="${i}"
  fi
done

MUTATION_SOURCES=$(echo "$MUTATION_SOURCES" | grep "$MODULE_FORMAT")

#Collect multiple lines into a single line to be fed into go-mutesting
MUTATION_SOURCES=$(echo $MUTATION_SOURCES | tr '\n' ' ')

echo "running mutation tests for the following module(s): $MODULE_NAMES"
OUTPUT=$(go run github.com/osmosis-labs/go-mutesting/cmd/go-mutesting --disable=$DISABLED_MUTATORS $MUTATION_SOURCES)

# Fetch the final result output and the overall mutation testing score
RESULT=$(echo "$OUTPUT" | grep 'The mutation score')
SCORE=$(echo "$RESULT" | grep -Eo '[[:digit:]]\.[[:digit:]]+')

echo "writing mutation test result to mutation_test_result.txt"
echo "$OUTPUT" > mutation_test_result.txt

# Print the mutation score breakdown
echo $RESULT

# Return a non-zero exit code if the score is below 75%
if (( $(echo "$SCORE < 0.75" |bc -l) )); then
  echo "Mutation testing score below desired level ($SCORE < 0.75)"
  exit 1
fi
