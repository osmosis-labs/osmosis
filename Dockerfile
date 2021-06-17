# Simple usage with a mounted data directory:
# > docker build -t osmosis .
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.osmosisd:/osmosis/.osmosisd -v ~/.osmosiscli:/osmosis/.osmosiscli osmosis osmosisd init
# > docker run -it -p 46657:46657 -p 46656:46656 -v ~/.osmosisd:/osmosis/.osmosisd -v ~/.osmosiscli:/osmosis/.osmosiscli osmosis osmosisd start
FROM faddat/archlinux

ENV GOPATH=/go
ENV PATH=$PATH:/go/bin

# Set up dependencies
RUN pacman -Syyu --noconfirm curl make git go gcc linux-headers python base-devel protobufs wget && \
    wget -O /genesis.json https://github.com/osmosis-labs/networks/raw/main/osmosis-1/genesis.json


# Add source files
COPY . /osmosis

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN cd /osmosis && \
    go mod download && \
    make install

# Final image
FROM faddat/archlinux

# Install ca-certificates
RUN pacman -Syyu --noconfirm

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/osmosisd /usr/bin/osmosisd
COPY --from=build-env /genesis.json /genesis.json

# Run osmosisd by default, omit entrypoint to ease using container with osmosiscli
EXPOSE 26656
EXPOSE 26657

CMD ["osmosisd init ease$RANDOM && cp /genesis.json ~/.osmosis/config/genesis.json && osmosisd start --persistent-peers 8d9967d5f865c68f6fe2630c0f725b0363554e77@134.255.252.173:26656,778fdedf6effe996f039f22901a3360bc838b52e@161.97.187.189:36657,64d36f3a186a113c02db0cf7c588c7c85d946b5b@209.97.132.170:26656,4d9ac3510d9f5cfc975a28eb2a7b8da866f7bc47@37.187.38.191:26656,2f9c16151400d8516b0f58c030b3595be20b804c@37.120.245.167:26656,bada684070727cb3dda430bcc79b329e93399665@173.212.240.91:26656"]
