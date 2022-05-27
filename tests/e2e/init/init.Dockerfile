# syntax=docker/dockerfile:1

## Build Image
FROM golang:1.18-bullseye as build

WORKDIR /osmosis
COPY . /osmosis

# From https://github.com/CosmWasm/wasmd/blob/master/Dockerfile
# For more details see https://github.com/CosmWasm/wasmvm#builds-of-libwasmvm 
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta7/libwasmvm_muslc.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep d0152067a5609bfdfb3f0d5d6c0f2760f79d5f2cd7fd8513cafa9932d22eb350
RUN BUILD_TAGS=muslc make build-e2e-chain-init

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
