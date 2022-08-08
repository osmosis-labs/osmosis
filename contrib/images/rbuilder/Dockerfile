
FROM golang:1.19-alpine3.15

ENV DEBIAN_FRONTEND=noninteractive
RUN apk add pciutils build-base git wget linux-headers \
    dpkg-dev curl gcc abuild binutils bash fakeroot

RUN mkdir -p /usr/local/share/osmosis/

# Deploy the shell functions library.
COPY buildlib.sh /usr/local/share/osmosis/

# Create the 'builder' user.
RUN addgroup -S builder && adduser -S builder -G builder
ARG APP
ARG DEBUG
ARG TARGET_PLATFORMS
ENV APP ${APP:-app}
ENV DEBUG ${DEBUG}
ENV VERSION unknown
ENV COMMIT unknown
ENV LEDGER_ENABLE true
ENV TARGET_PLATFORMS ${TARGET_PLATFORMS:-linux/amd64}
ENV BUILD_SCRIPT ${BUILD_SCRIPT:-/sources/.build.sh}

# From https://github.com/CosmWasm/wasmd/blob/master/Dockerfile
# For more details see https://github.com/CosmWasm/wasmvm#builds-of-libwasmvm 
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.x86_64.a /usr/local/lib/libwasmvm_muslc.a
RUN chown builder /usr/local/lib/libwasmvm_muslc.a 
RUN sha256sum /usr/local/lib/libwasmvm_muslc.a | grep f6282df732a13dec836cda1f399dd874b1e3163504dbd9607c6af915b2740479

# Drop root privileges.
USER builder:builder
WORKDIR /sources

RUN git config --global --add safe.directory /sources

# Run the application's build.sh.
ENTRYPOINT [ "/bin/bash", "-c", "${BUILD_SCRIPT}" ]
