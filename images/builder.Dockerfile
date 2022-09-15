# syntax=docker/dockerfile:1

ARG RUNNER_IMAGE="gcr.io/distroless/static"
# Rebuild with make docker-build-setup
ARG SETUP_VERSION

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM osmosis-setup:${SETUP_VERSION}

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
