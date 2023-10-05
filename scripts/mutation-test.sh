#!/usr/bin/env bash

set -eo pipefail
IFS=' '
MODULE=$1
SPECIFIC_FILES=$2

DISABLED_MUTATORS='branch/*'

# Check if specific files are provided
if [ -n "$SPECIFIC_FILES" ]; then
    IFS=','
    read -ra FILE_ARRAY <<< "$SPECIFIC_FILES"
    for file in "${FILE_ARRAY[@]}"; do
        MUTATION_SOURCES+=$(find $file)
    done
    echo "Running mutation tests on the following file(s): $SPECIFIC_FILES"
else
    # Only consider the following:
    # * go files in types, keeper, or module root directories
    # * ignore test and Protobuf files
    go_file_exclusions="-type f ! -path */client/* -name *.go -and -not -name *_test.go -and -not -name *pb* -and -not -name module.go -and -not -name sim_msgs.go -and -not -name codec.go -and -not -name errors.go"
    MUTATION_SOURCES=$(find ../x/$MODULE $go_file_exclusions)
    MUTATION_SOURCES+=$(printf '\n'; find ../x/$MODULE -maxdepth 2 $go_file_exclusions)
    echo "No specific files provided, running mutation tests on all Go files in the module: $MODULE"
fi

# Collect multiple lines into a single line to be fed into go-mutesting
MUTATION_SOURCES=$(echo $MUTATION_SOURCES | tr '\n' ' ' | sed 's/^ *//;s/ *$//')

OUTPUT=$(go run github.com/osmosis-labs/go-mutesting/cmd/go-mutesting --disable=$DISABLED_MUTATORS $MUTATION_SOURCES)

# Fetch the final result output and the overall mutation testing score
RESULT=$(echo "$OUTPUT" | grep 'The mutation score')
SCORE=$(echo "$RESULT" | grep -Eo '[[:digit:]]\.[[:digit:]]+')

echo "Writing mutation test result to mutation_test_result.txt"
echo "$OUTPUT" > mutation_test_result.txt

# Print the mutation score breakdown
echo $RESULT

# Return a non-zero exit code if the score is below 75%
if (( $(echo "$SCORE < 0.75" |bc -l) )); then
  echo "Mutation testing score below desired level ($SCORE < 0.75)"
  exit 1
fi
