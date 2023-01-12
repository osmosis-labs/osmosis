#!/bin/bash
set -e

# always returns true so set -e doesn't exit if it is not running.
killall osmosisd || true
rm -rf $HOME/.osmosisd/

# make four osmosis directories
mkdir $HOME/.osmosisd
mkdir $HOME/.osmosisd/validator1

# init all three validators
osmosisd init --chain-id=testing validator1 --home=$HOME/.osmosisd/validator1

# create keys for all three validators
osmosisd keys add validator1 --keyring-backend=test --home=$HOME/.osmosisd/validator1

update_genesis () {    
    cat $HOME/.osmosisd/validator1/config/genesis.json | jq "$1" > $HOME/.osmosisd/validator1/config/tmp_genesis.json && mv $HOME/.osmosisd/validator1/config/tmp_genesis.json $HOME/.osmosisd/validator1/config/genesis.json
}

# change staking denom to uosmo
update_genesis '.app_state["staking"]["params"]["bond_denom"]="uosmo"'

# create validator node with tokens to transfer to the three other nodes
osmosisd add-genesis-account $(osmosisd keys show validator1 -a --keyring-backend=test --home=$HOME/.osmosisd/validator1) 100000000000uosmo,100000000000stake,100000000000uatom,2000000uakt --home=$HOME/.osmosisd/validator1
osmosisd gentx validator1 500000000uosmo --keyring-backend=test --home=$HOME/.osmosisd/validator1 --chain-id=testing
osmosisd collect-gentxs --home=$HOME/.osmosisd/validator1


# update staking genesis
update_genesis '.app_state["staking"]["params"]["unbonding_time"]="240s"'

# update crisis variable to uosmo
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="uosmo"'

# udpate gov genesis
update_genesis '.app_state["gov"]["voting_params"]["voting_period"]="60s"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="uosmo"'

# update epochs genesis
update_genesis '.app_state["epochs"]["epochs"][1]["duration"]="60s"'

# update poolincentives genesis
update_genesis '.app_state["poolincentives"]["lockable_durations"][0]="120s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][1]="180s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][2]="240s"'
update_genesis '.app_state["poolincentives"]["params"]["minted_denom"]="uosmo"'

# update incentives genesis
update_genesis '.app_state["incentives"]["lockable_durations"][0]="1s"'
update_genesis '.app_state["incentives"]["lockable_durations"][1]="120s"'
update_genesis '.app_state["incentives"]["lockable_durations"][2]="180s"'
update_genesis '.app_state["incentives"]["lockable_durations"][3]="240s"'
update_genesis '.app_state["incentives"]["params"]["distr_epoch_identifier"]="day"'

# update mint genesis
update_genesis '.app_state["mint"]["params"]["mint_denom"]="uosmo"'
update_genesis '.app_state["mint"]["params"]["epoch_identifier"]="day"'

# update gamm genesis
update_genesis '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="uosmo"'


# port key (validator1 uses default ports)
# validator1 1317, 9090, 9091, 26658, 26657, 26656, 6060
# validator2 1316, 9088, 9089, 26655, 26654, 26653, 6061
# validator3 1315, 9086, 9087, 26652, 26651, 26650, 6062

# change config.toml values
VALIDATOR1_CONFIG=$HOME/.osmosisd/validator1/config/config.toml

# validator1
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG

# start all three validators
osmosisd start --home=$HOME/.osmosisd/validator1

echo "1 Validators are up and running!"