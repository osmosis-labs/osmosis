# syntax=docker/dockerfile:1

ARG BASE_IMG_TAG=nonroot

# --------------------------------------------------------
# Build
# --------------------------------------------------------

FROM golang:1.18.2-alpine3.15 as build

RUN set -eux; apk add --no-cache ca-certificates build-base;
RUN apk add git
# Needed by github.com/zondax/hid
RUN apk add linux-headers

WORKDIR /osmosis
COPY . /osmosis

# CosmWasm: see https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 7d2239e9f25e96d0d4daba982ce92367aacf0cbd95d2facb8442268f2b1cc1fc
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep f6282df732a13dec836cda1f399dd874b1e3163504dbd9607c6af915b2740479

# CosmWasm: copy the right library according to architecture. The final location will be found by the linker flag `-lwasmvm_muslc`
RUN cp /lib/libwasmvm_muslc.$(uname -m).a /lib/libwasmvm_muslc.a

RUN BUILD_TAGS=muslc LINK_STATICALLY=true make build

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM ubuntu

COPY --from=build /osmosis/build/osmosisd /bin/osmosisd

ENV HOME /osmosis
WORKDIR $HOME
RUN apt-get update && apt-get install -y jq
# change staking denom to uosmo
RUN osmosisd init --chain-id=localosmosis val
RUN echo "satisfy adjust timber high purchase tuition stool faith fine install that you unaware feed domain license impose boss human eager hat rent enjoy dawn" | osmosisd keys add val --recover --keyring-backend=test
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN osmosisd add-genesis-account osmo1phaxpevm5wecex2jyaqty2a4v02qj7qmlmzk5a 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo1cyyzpxplxdzkeea7kwsydadg87357qnahakaks 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo18s5lynnmx37hq4wlrw9gdn68sg2uxp5rgk26vv 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo1qwexv7c6sm95lwhzn9027vyu2ccneaqad4w8ka 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo14hcxlnwlqtq75ttaxf674vk6mafspg8xwgnn53 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo12rr534cer5c0vj53eq4y32lcwguyy7nndt0u2t 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo1nt33cjd5auzh36syym6azgc8tve0jlvklnq7jq 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo10qfrpash5g2vk3hppvu45x0g860czur8ff5yx0 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo1f4tvsdukfwh6s9swrc24gkuz23tp8pd3e9r5fa 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo1myv43sqgnj5sm4zl98ftl45af9cfzk7nhjxjqh 100000000000uosmo,100000000000uion
RUN osmosisd add-genesis-account osmo14gs9zqh8m49yy9kscjqu9h72exyf295afg6kgk 100000000000uosmo,100000000000uion
RUN osmosisd gentx val 500000000uosmo --keyring-backend=test --chain-id=localosmosis
RUN osmosisd collect-gentxs
# update staking genesis
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["unbonding_time"]="240s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
# update crisis variable to uosmo
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
# udpate gov genesis
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="60s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
# update epochs genesis
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["epochs"]["epochs"][1]["duration"]="60s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
# update poolincentives genesis
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][0]="120s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][1]="180s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][2]="240s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["params"]["minted_denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
# update incentives genesis
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][0]="1s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][1]="120s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][2]="180s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][3]="240s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["params"]["distr_epoch_identifier"]="day"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
# update mint genesis
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["mint"]["params"]["epoch_identifier"]="day"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
# update gamm genesis
RUN cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
# remove seeds
RUN sed -i.bak -E 's#^(seeds[[:space:]]+=[[:space:]]+).*$#\1""#' ~/.osmosisd/config/config.toml

EXPOSE 26656
EXPOSE 26657
EXPOSE 1317

ENTRYPOINT ["osmosisd"]
CMD ["start"]
