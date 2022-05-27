# syntax=docker/dockerfile:1

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:1.18.2-alpine3.15 as build

ARG NAME="osmosis"
ARG APP_NAME="osmosisd"
ARG VERSION 
ARG COMMIT
ARG COSMWASM_VERSION="v1.0.0"
ARG BUILD_TAGS="netgo ledger muslc"

RUN set -eux; apk add --no-cache ca-certificates build-base;
RUN apk add git
# Needed by github.com/zondax/hid
RUN apk add linux-headers

WORKDIR /osmosis
COPY . /osmosis

# CosmWasm: see https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/$COSMWASM_VERSION/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/$COSMWASM_VERSION/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a

# CosmWasm: copy the right library according to architecture. The final location will be found by the linker flag `-lwasmvm_muslc`
RUN cp /lib/libwasmvm_muslc.$(uname -m).a /lib/libwasmvm_muslc.a

RUN go build \
    -mod=readonly \
    -tags "$BUILD_TAGS" \
    -ldflags "-X github.com/osmosis-labs/osmosis/version.Name=$NAME -X github.com/osmosis-labs/osmosis/version.AppName=$APP_NAME -X github.com/osmosis-labs/osmosis/version.Version=$VERSION -X github.com/osmosis-labs/osmosis/version.Commit=$COMMIT -X github.com/osmosis-labs/osmosis/version.BuildTags='netgo,ledger,muslc' -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
    -trimpath \
    -o /osmosis/build/ \
    ./...

ENTRYPOINT ["ash"]