
FROM golang:1.18-bullseye

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get --no-install-recommends -y install \
    pciutils build-essential git wget \
    lsb-release dpkg-dev curl bsdmainutils fakeroot
RUN mkdir -p /usr/local/share/osmosis/

# Deploy the shell functions library.
COPY buildlib.sh /usr/local/share/osmosis/

# Create the 'builder' user.
RUN useradd -ms /bin/bash -U builder
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
RUN sha256sum /usr/local/lib/libwasmvm_muslc.a | grep 2f44efa9c6c1cda138bd1f46d8d53c5ebfe1f4a53cf3457b01db86472c4917ac

# Drop root privileges.
USER builder:builder
WORKDIR /sources

# Run the application's build.sh.
ENTRYPOINT [ "/bin/bash", "-c", "${BUILD_SCRIPT}" ]
