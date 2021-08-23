FROM faddat/archlinux AS build

ENV GOPATH=/go
ENV PATH=$PATH:/go/bin

# Set up dependencies
RUN pacman -Syyu --noconfirm curl make git go gcc linux-headers python base-devel protobuf wget && \
    wget -O /genesis.json https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json


# Add source files
COPY . /osmosis

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN cd /osmosis && \
    make install

# Final image
FROM faddat/archlinux

RUN pacman -Syyu --noconfirm 

# Copy over binaries from the build-env
COPY --from=build /go/bin/osmosisd /usr/bin/osmosisd
COPY --from=build /genesis.json /genesis.json

# Run osmosisd by default, omit entrypoint to ease using container with osmosiscli
EXPOSE 26656
EXPOSE 26657
EXPOSE 1317
EXPOSE 9090

