# syntax=docker/dockerfile:1

## Build
FROM golang:1.17-bullseye as build

WORKDIR /osmosis
COPY . /osmosis
RUN make build

## Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /osmosis/build/osmosisd /osmosisd

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
EXPOSE 9090

ENTRYPOINT ["/osmosisd"]