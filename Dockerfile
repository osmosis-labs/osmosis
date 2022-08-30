# syntax=docker/dockerfile:1

ARG GO_VERSION="1.18"
ARG RUNNER_IMAGE="gcr.io/distroless/static"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine as builder

RUN set -eux; apk add --no-cache ca-certificates build-base; apk add git linux-headers

# Download go dependencies
WORKDIR /osmosis
COPY go.* .
COPY x/epochs/go.mod x/epochs/go.sum /osmosis/x/epochs/

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

# Copy the remaining files
COPY . .

# Build osmosisd binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    VERSION=$(echo $(git describe --tags) | sed 's/^v//') && \
    COMMIT=$(git log -1 --format='%H') && \
    go build \
      -mod=readonly \
      -tags "netgo,ledger,muslc" \
      -ldflags "-X github.com/cosmos/cosmos-sdk/version.Name="osmosis" \
              -X github.com/cosmos/cosmos-sdk/version.AppName="osmosisd" \
              -X github.com/cosmos/cosmos-sdk/version.Version=$VERSION \
              -X github.com/cosmos/cosmos-sdk/version.Commit=$COMMIT \
              -X github.com/cosmos/cosmos-sdk/version.BuildTags='netgo,ledger,muslc' \
              -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
      -trimpath \
      -o /osmosis/build/ \
      ./...

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ${RUNNER_IMAGE}

COPY --from=builder /osmosis/build/osmosisd /bin/osmosisd

ENV HOME /osmosis
WORKDIR $HOME

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317

ENTRYPOINT ["osmosisd"]
