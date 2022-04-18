
FROM golang:1.18-bullseye

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get --no-install-recommends -y install \
    pciutils build-essential git wget \
    lsb-release dpkg-dev curl bsdmainutils fakeroot
RUN mkdir -p /usr/local/share/osmosis/

# Deploy the shell functions library.
COPY buildlib.sh /usr/local/share/osmosis/

# From https://github.com/CosmWasm/wasmd/blob/master/Dockerfile
# For more details see https://github.com/CosmWasm/wasmvm#builds-of-libwasmvm 
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0-beta7/libwasmvm_muslc.a /lib/libwasmvm_muslc.a
RUN sha256sum /lib/libwasmvm_muslc.a | grep d0152067a5609bfdfb3f0d5d6c0f2760f79d5f2cd7fd8513cafa9932d22eb350

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

WORKDIR /sources

# Mount the application's sources.
VOLUME [ "/sources" ]

# Run the application's build.sh.
ENTRYPOINT [ "/bin/bash", "-c", "${BUILD_SCRIPT}" ]
