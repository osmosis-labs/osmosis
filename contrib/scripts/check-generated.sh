#!/usr/bin/env bash

set -euo pipefail
# Install buf and gogo tools, so that differences that arise from
# toolchain differences are also caught.
readonly tools="$(mktemp -d)"
go install github.com/bufbuild/buf/cmd/buf
go install github.com/gogo/protobuf/protoc-gen-gogofaster@latest

make proto-gen

#Specificially ignore all differences in go.mod / go.sum.
if ! git diff --stat --exit-code . ':(exclude)*.mod' ':(exclude)*.sum'; then
    echo ">> ERROR:"
    echo ">>"
    echo ">> Protobuf generated code requires update (either tools or .proto files may have changed)."
    echo ">> Ensure your tools are up-to-date, re-run 'make proto-all' and update this PR."
    echo ">>"
    exit 1
fi
