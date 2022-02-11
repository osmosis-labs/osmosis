#!/bin/bash
rm -rf $HOME/.osmosisd/


#make four osmosis directories
mkdir $HOME/.osmosisd
mkdir $HOME/.osmosisd/validator1
mkdir $HOME/.osmosisd/validator2
mkdir $HOME/.osmosisd/validator3

#init first osmosis directory to create genesis file
osmosisd init --chain-id=testing validator1 --home=$HOME/.osmosisd/validator1
osmosisd keys add validator1 --keyring-backend=test --home=$HOME/.osmosisd/validator1

# change staking denom to uosmo
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="uosmo"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json

#create validator node with tokens to transfer to the three other nodes
osmosisd add-genesis-account $(osmosisd keys show validator1 -a --keyring-backend=test --home=$HOME/.osmosisd/validator1) 100000000000uosmo --home=$HOME/.osmosisd/validator1
osmosisd gentx validator1 500000000uosmo --keyring-backend=test --home=$HOME/.osmosisd/validator1 --chain-id=testing
osmosisd collect-gentxs --home=$HOME/.osmosisd/validator1


# update staking genesis
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["staking"]["params"]["unbonding_time"]="120s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json

# update governance genesis
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["gov"]["voting_params"]["voting_period"]="10s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json

# update epochs genesis
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["epochs"]["epochs"][0]["identifier"]="min"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["epochs"]["epochs"][0]["duration"]="60s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json

# update poolincentives genesis
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][0]="120s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][1]="180s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["poolincentives"]["lockable_durations"][2]="240s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json

# update incentives genesis
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["incentives"]["params"]["distr_epoch_identifier"]="min"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][0]="1s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][1]="120s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][2]="180s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["incentives"]["lockable_durations"][3]="240s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json

# update mint genesis
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["mint"]["params"]["epoch_identifier"]="min"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json

# update gamm genesis
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="stake"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json

# update superfluid genesis
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["superfluid"]["params"]["refresh_epoch_identifier"]="min"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
cat $HOME/.osmosisd/validator1/config/genesis.json | jq '.app_state["superfluid"]["params"]["unbonding_duration"]="120s"' > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json




# init validator2-3
osmosisd init --chain-id=testing validator2 --home=$HOME/.osmosisd/validator2
osmosisd init --chain-id=testing validator3 --home=$HOME/.osmosisd/validator3
