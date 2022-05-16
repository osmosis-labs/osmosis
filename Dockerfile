# syntax=docker/dockerfile:1

ARG BASE_IMG_TAG=nonroot 

## Build Image
FROM golang:1.18-bullseye as build

WORKDIR /osmosis
COPY . /osmosis

# From https://github.com/CosmWasm/wasmd/blob/master/Dockerfile
# For more details see https://github.com/CosmWasm/wasmvm#builds-of-libwasmvm 
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta10/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep 2f44efa9c6c1cda138bd1f46d8d53c5ebfe1f4a53cf3457b01db86472c4917ac
RUN BUILD_TAGS=muslc make build

## Deploy image
FROM gcr.io/distroless/base-debian11:${BASE_IMG_TAG}

COPY --from=build /osmosis/build/osmosisd /bin/osmosisd

ENV HOME /osmosis
WORKDIR $HOME

EXPOSE 26656 
EXPOSE 26657
EXPOSE 1317

ENTRYPOINT ["osmosisd"]
CMD [ "start" ]
