# syntax=docker/dockerfile:1

ARG GO_VERSION="1.18"

# --------------------------------------------------------
# Pre-build Setup
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine

RUN set -eux; apk add --no-cache ca-certificates build-base; apk add git linux-headers

# Download go dependencies
WORKDIR /osmosis
COPY go.* .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Cosmwasm - download correct libwasmvm version
RUN WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
      -O /lib/libwasmvm_muslc.a

# Cosmwasm - verify checksum
RUN wget https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep $(uname -m) | cut -d ' ' -f 1)
