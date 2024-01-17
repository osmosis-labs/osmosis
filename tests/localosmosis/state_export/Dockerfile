# syntax=docker/dockerfile:1

ARG GO_VERSION="1.20"
ARG RUNNER_IMAGE="alpine:3.17"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine as builder

COPY . osmosis/

RUN apk add --no-cache \
    ca-certificates \
    build-base \
    linux-headers \
    git \
    && rm -rf /var/cache/apk/*

RUN cd osmosis/

# Download go dependencies
WORKDIR /go/osmosis
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Cosmwasm - Download correct libwasmvm version
RUN ARCH=$(uname -m) && WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | sed 's/.* //') && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$ARCH.a \
        -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc.$ARCH | cut -d ' ' -f 1)

# Build osmosisd binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go build \
        -mod=readonly \
        -trimpath \
        -tags "netgo,ledger,muslc,osusergo" \
        -ldflags \
            "-X github.com/cosmos/cosmos-sdk/version.Name="osmosis" \
            -X github.com/cosmos/cosmos-sdk/version.AppName="osmosisd" \
            -X github.com/cosmos/cosmos-sdk/version.Version=${GIT_REF} \
            -X github.com/cosmos/cosmos-sdk/version.Commit=${GIT_COMMIT} \
            -X github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger,muslc \
            -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
        -trimpath \
        -o /go/osmosis/build/osmosisd \
        /go/osmosis/cmd/osmosisd/main.go

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ${RUNNER_IMAGE}

ARG VALIDATOR_MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"
ARG VALIDATOR_ADDRESS="osmo12smx2wdlyttvyzvzg54y2vnqwq2qjateuf7thj"

COPY --from=builder /go/osmosis/build/osmosisd /bin/osmosisd

ENV HOME /osmosis
WORKDIR $HOME

# Add opionanted node configuration
# This files are copied at runtime so it is possible to override them
COPY ./tests/localosmosis/state_export/config/config.toml /etc/osmosis/config/config.toml
COPY ./tests/localosmosis/state_export/config/app.toml /etc/osmosis/config/app.toml
COPY ./tests/localosmosis/state_export/config/client.toml /etc/osmosis/config/client.toml

# Add localosmosis validator keys (not sensitive info)
COPY ./tests/localosmosis/state_export/config/node_key.json /etc/osmosis/config/node_key.json
COPY ./tests/localosmosis/state_export/config/priv_validator_key.json /etc/osmosis/config/priv_validator_key.json

# Save mnemonic for simplicity
RUN echo $VALIDATOR_MNEMONIC >> /etc/osmosis/mnemonic
RUN echo $VALIDATOR_ADDRESS > /etc/osmosis/address

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
EXPOSE 9090

COPY ./tests/localosmosis/state_export/start.sh /osmosis/start.sh
RUN chmod +x /osmosis/start.sh
ENTRYPOINT ["/osmosis/start.sh"] 
