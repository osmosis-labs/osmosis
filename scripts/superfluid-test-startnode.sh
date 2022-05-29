#!/bin/bash

rm -rf $HOME/.osmosisd/

cd $HOME

osmosisd init --chain-id=testing testing --home=$HOME/.osmosisd
osmosisd keys add validator --keyring-backend=test --home=$HOME/.osmosisd
osmosisd add-genesis-account $(osmosisd keys show validator -a --keyring-backend=test --home=$HOME/.osmosisd) 100000000000stake,100000000000valtoken --home=$HOME/.osmosisd
osmosisd gentx validator 500000000stake --keyring-backend=test --home=$HOME/.osmosisd --chain-id=testing
osmosisd collect-gentxs --home=$HOME/.osmosisd

# update staking genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["unbonding_time"]="120s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

# update governance genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="10s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

# update epochs genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["epochs"]["epochs"][0]["identifier"]="min"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["epochs"]["epochs"][0]["duration"]="60s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

# update poolincentives genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][0]="120s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][1]="180s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][2]="240s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

# update incentives genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["params"]["distr_epoch_identifier"]="min"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][0]="1s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][1]="120s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][2]="180s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][3]="240s"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

# update mint genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["mint"]["params"]["epoch_identifier"]="min"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

# update gamm genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="stake"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

# update superfluid genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["superfluid"]["params"]["refresh_epoch_identifier"]="min"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

osmosisd start --home=$HOME/.osmosisd
