# syntax=docker/dockerfile:1

## Build Image
FROM golang:1.18.2-alpine3.15 as build

ARG E2E_SCRIPT_NAME

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
RUN BUILD_TAGS=muslc LINK_STATICALLY=true E2E_SCRIPT_NAME=${E2E_SCRIPT_NAME} make build-e2e-script

## Deploy image
FROM ubuntu

# Args only last for a single build stage - renew
ARG E2E_SCRIPT_NAME

COPY --from=build /osmosis/build/${E2E_SCRIPT_NAME} /bin/${E2E_SCRIPT_NAME}

# ARGs are not expanded in ENTRYPOINT - use ENV
ENV ENTER_SCRIPT=$E2E_SCRIPT_NAME

ENV HOME /osmosis
WORKDIR $HOME

ENTRYPOINT $ENTER_SCRIPT
