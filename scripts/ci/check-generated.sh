#!/usr/bin/env bash

set -euo pipefail

make proto-gen
make run-querygen

#Specificially ignore all differences in go.mod / go.sum.
if ! git diff --stat --exit-code . ':(exclude)*.mod' ':(exclude)*.sum'; then
    echo ">> ERROR:"
    echo ">>"
    echo ">> Protobuf generated code requires update (either tools or .proto files may have changed)."
    echo ">> Ensure your tools are up-to-date, re-run 'make proto-all' and update this PR."
    echo ">>"
    exit 1
fi
