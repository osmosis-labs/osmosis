# syntax=docker/dockerfile:1

## Build
FROM golang:1.17-bullseye as build

WORKDIR /osmosis
COPY . /osmosis
RUN make build

## Deploy
FROM gcr.io/distroless/base-debian11

WORKDIR /

# wasm dependencies
COPY --from=build /go/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.0.0-beta5/api/libwasmvm.so /lib/
COPY --from=build /lib/x86_64-linux-gnu/libgcc_s.so.1 /lib/

# osmosisd binary
COPY --from=build /osmosis/build/osmosisd /osmosisd

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
EXPOSE 9090

ENTRYPOINT ["/osmosisd"]