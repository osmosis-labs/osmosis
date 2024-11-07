# syntax=docker/dockerfile:1

ARG GO_VERSION="1.21"
ARG RUNNER_IMAGE="alpine:3.19"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine as builder

COPY . symphony/

RUN apk add --no-cache \
    ca-certificates \
    build-base \
    linux-headers \
    git \
    && rm -rf /var/cache/apk/*

RUN cd symphony/

# Download go dependencies
WORKDIR /go/symphony
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Cosmwasm - Download correct libwasmvm version
RUN ARCH=$(uname -m) && WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm/v2 | sed 's/.* //') && \
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$ARCH.a \
        -O /lib/libwasmvm_muslc.a && \
    # verify checksum
    wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/checksums.txt -O /tmp/checksums.txt && \
    sha256sum /lib/libwasmvm_muslc.a | grep $(cat /tmp/checksums.txt | grep libwasmvm_muslc.$ARCH | cut -d ' ' -f 1)

# Build symphonyd binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go build \
        -mod=readonly \
        -trimpath \
        -tags "netgo,ledger,muslc,osusergo" \
        -ldflags \
            "-X github.com/cosmos/cosmos-sdk/version.Name="symphony" \
            -X github.com/cosmos/cosmos-sdk/version.AppName="symphonyd" \
            -X github.com/cosmos/cosmos-sdk/version.Version=${GIT_REF} \
            -X github.com/cosmos/cosmos-sdk/version.Commit=${GIT_COMMIT} \
            -X github.com/cosmos/cosmos-sdk/version.BuildTags=netgo,ledger,muslc \
            -w -s -linkmode=external -extldflags '-Wl,-z,muldefs -static'" \
        -trimpath \
        -o /go/symphony/build/symphonyd \
        /go/symphony/cmd/symphonyd/main.go

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ${RUNNER_IMAGE}

ARG VALIDATOR_MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"
ARG VALIDATOR_ADDRESS="symphony1p7mp7r9f9f6sf2c95ht42ncm6ga96ha8xghdeg"

COPY --from=builder /go/symphony/build/symphonyd /bin/symphonyd

ENV HOME /symphony
WORKDIR $HOME

# Add opionanted node configuration
# This files are copied at runtime so it is possible to override them
COPY ./tests/localsymphony/state_export/config/config.toml /etc/symphony/config/config.toml
COPY ./tests/localsymphony/state_export/config/app.toml /etc/symphony/config/app.toml
COPY ./tests/localsymphony/state_export/config/client.toml /etc/symphony/config/client.toml

# Add localsymphony validator keys (not sensitive info)
COPY ./tests/localsymphony/state_export/config/node_key.json /etc/symphony/config/node_key.json
COPY ./tests/localsymphony/state_export/config/priv_validator_key.json /etc/symphony/config/priv_validator_key.json

# Save mnemonic for simplicity
RUN echo $VALIDATOR_MNEMONIC >> /etc/symphony/mnemonic
RUN echo $VALIDATOR_ADDRESS > /etc/symphony/address

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
EXPOSE 9090

COPY ./tests/localsymphony/state_export/start.sh /symphony/start.sh
RUN chmod +x /symphony/start.sh
ENTRYPOINT ["/symphony/start.sh"] 
