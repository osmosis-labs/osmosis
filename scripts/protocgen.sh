#!/usr/bin/env bash

set -eo pipefail

# get protoc executions
go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos 2>/dev/null
# get cosmos sdk from github
go get github.com/cosmos/cosmos-sdk 2>/dev/null

# move the vendor folder to a temp dir so that go list works properly
temp_dir="f29ea6aa861dc4b083e8e48f67cce"
if [ -d vendor ]; then
  mv ./vendor ./$temp_dir
fi

# Get the path of the cosmos-sdk repo from go/pkg/mod
cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk)

# move the vendor folder back to ./vendor
if [ -d $temp_dir ]; then
  mv ./$temp_dir ./vendor
fi

proto_dirs=$(find . \( -path ./third_party -o -path ./vendor \) -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)

for dir in $proto_dirs; do
  # generate protobuf bind
  buf protoc \
  -I "proto" \
  -I "$cosmos_sdk_dir/third_party/proto" \
  -I "$cosmos_sdk_dir/proto" \
  --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  $(find "${dir}" -name '*.proto')

  # generate grpc gateway
  buf protoc \
  -I "proto" \
  -I "$cosmos_sdk_dir/third_party/proto" \
  -I "$cosmos_sdk_dir/proto" \
  --grpc-gateway_out=logtostderr=true:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')
done

# command to generate docs using protoc-gen-doc
buf protoc \
  -I "proto" \
  -I "$cosmos_sdk_dir/third_party/proto" \
  -I "$cosmos_sdk_dir/proto" \
  --doc_out=./docs/core \
  --doc_opt=./docs/protodoc-markdown.tmpl,proto-docs.md \
  $(find "$(pwd)/proto" -maxdepth 4 -name '*.proto')

cp -r ./github.com/osmosis-labs/osmosis/v*/x/* x/
rm -rf ./github.com
