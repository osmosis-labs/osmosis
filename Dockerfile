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
    go mod download && \
    make install

# Final image
FROM faddat/archlinux

# Install ca-certificates
RUN pacman -Syyu --noconfirm

# Copy over binaries from the build-env
COPY --from=build /go/bin/osmosisd /usr/bin/osmosisd
COPY --from=build /genesis.json /genesis.json

# Run osmosisd by default, omit entrypoint to ease using container with osmosiscli
EXPOSE 26656
EXPOSE 26657

CMD osmosisd init ease$RANDOM && cp /genesis.json ~/.osmosisd/config/genesis.json && osmosisd start --p2p.seeds e437756a853061cc6f1639c2ac997d9f7e84be67@144.76.183.180:26656,83adaa38d1c15450056050fd4c9763fcc7e02e2c@ec2-44-234-84-104.us-west-2.compute.amazonaws.com:26656,23142ab5d94ad7fa3433a889dcd3c6bb6d5f247d@95.217.193.163:26656 --p2p.persistent_peers 8d9967d5f865c68f6fe2630c0f725b0363554e77@134.255.252.173:26656,778fdedf6effe996f039f22901a3360bc838b52e@161.97.187.189:36657,64d36f3a186a113c02db0cf7c588c7c85d946b5b@209.97.132.170:26656,4d9ac3510d9f5cfc975a28eb2a7b8da866f7bc47@37.187.38.191:26656,2f9c16151400d8516b0f58c030b3595be20b804c@37.120.245.167:26656,bada684070727cb3dda430bcc79b329e93399665@173.212.240.91:26656,2115945f074ddb038de5d835e287fa03e32f0628@95.217.43.85:26656,778fdedf6effe996f039f22901a3360bc838b52e@161.97.187.189:36656,785bc83577e3980545bac051de8f57a9fd82695f@194.233.164.146:26656,e7916387e05acd53d1b8c0f842c13def365c7bb6@176.9.64.212:26656  --p2p.seed_mode true
