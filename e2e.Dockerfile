# syntax=docker/dockerfile:1
ARG IMG_TAG=latest

## Build Image
FROM golang:1.17-bullseye as build

WORKDIR /osmosis
COPY . /osmosis

# From https://github.com/CosmWasm/wasmd/blob/master/Dockerfile
# For more details see https://github.com/CosmWasm/wasmvm#builds-of-libwasmvm 
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta7/libwasmvm_muslc.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep d0152067a5609bfdfb3f0d5d6c0f2760f79d5f2cd7fd8513cafa9932d22eb350
RUN BUILD_TAGS=muslc make build

## Deploy image
FROM gcr.io/distroless/cc:$IMG_TAG
ARG IMG_TAG
COPY --from=build /osmosis/build/osmosisd /bin/osmosisd

ENV HOME /osmosis
WORKDIR $HOME

EXPOSE 26656 26657 1317 9090

ENTRYPOINT ["osmosisd", "start"]
