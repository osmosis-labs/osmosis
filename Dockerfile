# syntax=docker/dockerfile:1

# Rebuild with make docker-build-builder
ARG BUILDER_VERSION="v1"
ARG RUNNER_IMAGE="gcr.io/distroless/static"

FROM osmosis-builder:${BUILDER_VERSION} as builder

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
