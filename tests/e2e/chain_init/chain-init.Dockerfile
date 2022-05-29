# syntax=docker/dockerfile:1

## Build Image
FROM golang:1.18.2-alpine3.15 as build

RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git

# needed by github.com/zondax/hid
RUN apk add linux-headers

WORKDIR /osmosis
COPY . /osmosis

# From https://github.com/CosmWasm/wasmd/blob/master/Dockerfile
# For more details see https://github.com/CosmWasm/wasmvm#builds-of-libwasmvm 
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep f6282df732a13dec836cda1f399dd874b1e3163504dbd9607c6af915b2740479
RUN BUILD_TAGS=muslc LINK_STATICALLY=true make build-e2e-chain-init

## Deploy image
FROM ubuntu

COPY --from=build /osmosis/build/chain_init /bin/chain_init

ENV HOME /osmosis
WORKDIR $HOME

ENTRYPOINT [ "chain_init" ]
