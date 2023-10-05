#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen

# move the vendor folder to a temp dir so that go list works properly
temp_dir=$(mktemp -d)
if [ -d vendor ]; then
  mv ./vendor "${temp_dir}"
fi

# Get the path of the cosmos-sdk repo from go/pkg/mod
cosmos_sdk_dir=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk) || { echo "Error: Failed to find github.com/cosmos/cosmos-sdk"; exit 1; }

# move the vendor folder back to ./vendor
if [ -d "${temp_dir}" ]; then
  mv "${temp_dir}" ./vendor
fi

if [ -d "${cosmos_sdk_dir}/proto" ]; then
  proto_dirs=$(find ./proto "${cosmos_sdk_dir}/proto" -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
  for dir in $proto_dirs; do
    # generate swagger files (filter query files)
    query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
    if [[ ! -z "$query_file" ]]; then
      protoc  \
      -I "proto" \
      -I "${cosmos_sdk_dir}/third_party/proto" \
      -I "${cosmos_sdk_dir}/proto" \
        "${query_file}" \
      --swagger_out ./tmp-swagger-gen \
      --swagger_opt logtostderr=true \
      --swagger_opt fqn_for_swagger_name=true \
      --swagger_opt simple_operation_ids=true
    fi
  done
fi

if [ -d "./client/docs" ]; then
  cd ./client/docs
  yarn install
  yarn combine
  yarn convert
  yarn build

  # check if yq is installed, if not, install it.
  if ! command -v yq &> /dev/null; then
    echo "yq not found! Installing yq..."
    wget https://github.com/mikefarah/yq/releases/download/v4.6.1/yq_linux_amd64 -O /usr/bin/yq && chmod +x /usr/bin/yq
  fi

  #Add public servers to spec file for Osmosis testnet and mainnet
  yq -i '."servers"+=[{"url":"https://lcd.osmosis.zone","description":"Osmosis mainnet node"},{"url":"https://lcd-test.osmosis.zone","description":"Osmosis testnet node"}]' static/openapi/openapi.yaml

  cd ../../
fi

# clean swagger files
rm -rf ./tmp-swagger-gen

# remove temporary directory
rm -rf "${temp_dir}"
