# Simple usage with a mounted data directory:
# > docker build -t osmosis .
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.osmosisd:/osmosis/.osmosisd -v ~/.osmosiscli:/osmosis/.osmosiscli osmosis osmosisd init
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.osmosisd:/osmosis/.osmosisd -v ~/.osmosiscli:/osmosis/.osmosiscli osmosis osmosisd start
FROM golang:alpine AS build-env

# Set up dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3

# Set working directory for the build
WORKDIR /go/src/github.com/osmosis-labs/osmosis

# Add source files
COPY . .

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make install

# Final image
FROM alpine:edge

ENV osmosis /osmosis

# Install ca-certificates
RUN apk add --update ca-certificates

RUN addgroup osmosis && \
    adduser -S -G osmosis osmosis -h "$OSMOSIS"

USER osmosis

WORKDIR $OSMOSIS

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/osmosisd /usr/bin/osmosisd

# Run osmosisd by default, omit entrypoint to ease using container with osmosiscli
CMD ["osmosisd"]