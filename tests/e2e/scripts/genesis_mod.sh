#!/bin/sh

# change staking denom to uosmo
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# update staking genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["unbonding_time"]="240s"' | sponge $HOME/.osmosisd/config/genesis.json
# update crisis variable to uosmo
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# udpate gov genesis
# cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="60s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# update epochs genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["epochs"]["epochs"][1]["duration"]="60s"' | sponge $HOME/.osmosisd/config/genesis.json
# update poolincentives genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][0]="120s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][1]="180s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][2]="240s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["params"]["minted_denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# update incentives genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][0]="1s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][1]="120s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][2]="180s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][3]="240s"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["params"]["distr_epoch_identifier"]="day"' | sponge $HOME/.osmosisd/config/genesis.json
# update mint genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["mint"]["params"]["epoch_identifier"]="day"' | sponge $HOME/.osmosisd/config/genesis.json
# update gamm genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="uosmo"' | sponge $HOME/.osmosisd/config/genesis.json
# remove seeds
sed -i.bak -E 's#^(seeds[[:space:]]+=[[:space:]]+).*$#\1""#' ~/.osmosisd/config/config.toml

osmosisd start
