# syntax=docker/dockerfile:1

ARG BASE_IMG_TAG=nonroot

# --------------------------------------------------------
# Build 
# --------------------------------------------------------

FROM golang:1.18.2-alpine3.15 as build

# linux-headers needed by github.com/zondax/hid
RUN set -eux; apk add --no-cache ca-certificates build-base; apk add git linux-headers

# CosmWasm: see https://github.com/CosmWasm/wasmvm/releases
# 1) Add the releases into /lib/
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.aarch64.a \ 
  https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.x86_64.a \ 
  /lib/
# 2) Verify their hashes
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 7d2239e9f25e96d0d4daba982ce92367aacf0cbd95d2facb8442268f2b1cc1fc; \
  sha256sum /lib/libwasmvm_muslc.x86_64.a | grep f6282df732a13dec836cda1f399dd874b1e3163504dbd9607c6af915b2740479
# Copy the right library according to architecture. The final location will be found by the linker flag `-lwasmvm_muslc`
RUN cp /lib/libwasmvm_muslc.$(uname -m).a /lib/libwasmvm_muslc.a

# Get go dependencies first
COPY go.mod go.sum /osmosis/
WORKDIR /osmosis
RUN --mount=type=cache,target=/root/.cache/go-build \
 --mount=type=cache,target=/root/go/pkg/mod \
 go mod download

# Copy all files into our docker repo
COPY . /osmosis

# build

RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/root/go/pkg/mod \
  BUILD_TAGS=muslc LINK_STATICALLY=true make build

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM gcr.io/distroless/base-debian11:${BASE_IMG_TAG}

COPY --from=build /osmosis/build/osmosisd /bin/osmosisd

ENV HOME /osmosis
WORKDIR $HOME

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317

ENTRYPOINT ["osmosisd"]
CMD [ "start" ]
