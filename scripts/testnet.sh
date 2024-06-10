#!/bin/bash
set -e

CHAIN_ID="symphony-testnet-1"

if [ "$#" -ne 4 ]; then
    echo "Usage: $0 <NODE1_API> <NODE2_API> <NODE3_API> <NODE4_API>"
    exit 1
fi

NODE1_API=$1
NODE2_API=$2
NODE3_API=$3
NODE4_API=$4

killall symphonyd || true
rm -rf $HOME/.symphonyd/

# make 5 symphony directories
mkdir $HOME/.symphonyd
mkdir $HOME/.symphonyd/validator1
mkdir $HOME/.symphonyd/validator2
mkdir $HOME/.symphonyd/validator3
mkdir $HOME/.symphonyd/validator4

mkdir -p ~/.symphonyd/cosmovisor
mkdir -p ~/.symphonyd/cosmovisor/genesis
mkdir -p ~/.symphonyd/cosmovisor/genesis/bin
mkdir -p ~/.symphonyd/cosmovisor/upgrades

cp ../build/symphonyd ~/.symphonyd/cosmovisor/genesis/bin

# init all 4 validators
symphonyd init --chain-id=$CHAIN_ID validator1 --home=$HOME/.symphonyd/validator1
symphonyd init --chain-id=$CHAIN_ID validator2 --home=$HOME/.symphonyd/validator2
symphonyd init --chain-id=$CHAIN_ID validator3 --home=$HOME/.symphonyd/validator3
symphonyd init --chain-id=$CHAIN_ID validator4 --home=$HOME/.symphonyd/validator4
# create keys for all 4 validators
symphonyd keys add validator1 --keyring-backend=test --home=$HOME/.symphonyd/validator1
symphonyd keys add validator2 --keyring-backend=test --home=$HOME/.symphonyd/validator2
symphonyd keys add validator3 --keyring-backend=test --home=$HOME/.symphonyd/validator3
symphonyd keys add validator4 --keyring-backend=test --home=$HOME/.symphonyd/validator4

update_genesis () {    
    cat $HOME/.symphonyd/validator1/config/genesis.json | jq "$1" > $HOME/.symphonyd/validator1/config/tmp_genesis.json && mv $HOME/.symphonyd/validator1/config/tmp_genesis.json $HOME/.symphonyd/validator1/config/genesis.json
}

# change staking denom to note
update_genesis '.app_state["staking"]["params"]["bond_denom"]="note"'

# update staking genesis
update_genesis '.app_state["staking"]["params"]["unbonding_time"]="240s"'

# update crisis variable to note
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="note"'

# update gov genesis
update_genesis '.app_state["gov"]["voting_params"]["voting_period"]="60s"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="note"'
update_genesis '.app_state["gov"]["params"]["min_deposit"][0]["denom"]="note"'
update_genesis '.app_state["gov"]["params"]["expedited_min_deposit"][0]["denom"]="note"'

# update epochs genesis
update_genesis '.app_state["epochs"]["epochs"][1]["duration"]="60s"'

# update poolincentives genesis
update_genesis '.app_state["poolincentives"]["lockable_durations"][0]="120s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][1]="180s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][2]="240s"'
update_genesis '.app_state["poolincentives"]["params"]["minted_denom"]="note"'

# update incentives genesis
update_genesis '.app_state["incentives"]["lockable_durations"][0]="1s"'
update_genesis '.app_state["incentives"]["lockable_durations"][1]="120s"'
update_genesis '.app_state["incentives"]["lockable_durations"][2]="180s"'
update_genesis '.app_state["incentives"]["lockable_durations"][3]="240s"'
update_genesis '.app_state["incentives"]["params"]["distr_epoch_identifier"]="day"'

# update mint genesis
update_genesis '.app_state["mint"]["params"]["mint_denom"]="note"'
update_genesis '.app_state["mint"]["params"]["epoch_identifier"]="day"'

# update gamm genesis
update_genesis '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="note"'

# update concentratedliquidity genesis
update_genesis '.app_state["concentratedliquidity"]["params"]["is_permissionless_pool_creation_enabled"]=true'

# update txfees genesis
update_genesis '.app_state["txfees"]["basedenom"]="note"'


# create validator node with tokens
symphonyd add-genesis-account $(symphonyd keys show validator1 -a --keyring-backend=test --home=$HOME/.symphonyd/validator1) 100000000000note --home=$HOME/.symphonyd/validator1
symphonyd gentx validator1 5000000000note --moniker="validator1" --chain-id= --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator1 --chain-id=$CHAIN_ID
symphonyd collect-gentxs --home=$HOME/.symphonyd/validator1

# copy validator1 genesis file to validator2-4
 cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator2/config/genesis.json
 cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator3/config/genesis.json
 cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator4/config/genesis.json

# port key (validators uses default ports)
# validator 1317, 9050, 9091, 26658, 26657, 26656, 6060, 26660

# change app.toml values
VALIDATOR1_APP_TOML=$HOME/.symphonyd/validator1/config/app.toml
VALIDATOR2_APP_TOML=$HOME/.symphonyd/validator2/config/app.toml
VALIDATOR3_APP_TOML=$HOME/.symphonyd/validator3/config/app.toml
VALIDATOR4_APP_TOML=$HOME/.symphonyd/validator4/config/app.toml

# validator2
sed -i -E 's|adaptive-fee-enabled = "false"|adaptive-fee-enabled = "true"|g' $VALIDATOR3_APP_TOML

# validator3
sed -i -E 's|adaptive-fee-enabled = "false"|adaptive-fee-enabled = "true"|g' $VALIDATOR3_APP_TOML

# validator4
sed -i -E 's|adaptive-fee-enabled = "false"|adaptive-fee-enabled = "true"|g' $VALIDATOR4_APP_TOML


# change config.toml values
VALIDATOR1_CONFIG=$HOME/.symphonyd/validator1/config/config.toml
VALIDATOR2_CONFIG=$HOME/.symphonyd/validator2/config/config.toml
VALIDATOR3_CONFIG=$HOME/.symphonyd/validator3/config/config.toml
VALIDATOR4_CONFIG=$HOME/.symphonyd/validator4/config/config.toml

# validator1
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR1_CONFIG

# validator2
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR2_CONFIG

# validator3
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR3_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR3_CONFIG

# validator4
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR4_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR4_CONFIG

# persistent_peers

# copy tendermint node id of validator 2,3,4 to 1
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@$NODE2_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@$NODE3_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@$NODE4_API:26656\"|g" $HOME/.symphonyd/validator1/config/config.toml
# copy tendermint node id of 1,3,4 to 2
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@$NODE1_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@$NODE3_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@$NODE4_API:26656\"|g" $HOME/.symphonyd/validator2/config/config.toml
# copy tendermint node id of 1,2,4 to 3
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@$NODE1_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@$NODE2_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@$NODE4_API:26656\"|g" $HOME/.symphonyd/validator3/config/config.toml
# copy tendermint node id of 1,2,3 to 4
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@$NODE1_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@$NODE2_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@$NODE3_API:26656\"|g" $HOME/.symphonyd/validator4/config/config.toml

# rpc_servers

# copy tendermint node id  2,3,4 to 1
sed -i -E "s|rpc_servers = \"\"|rpc_servers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@$NODE2_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@$NODE3_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@$NODE4_API:26656\"|g" $HOME/.symphonyd/validator1/config/config.toml
# copy tendermint node id of 1,3,4 to 2
sed -i -E "s|rpc_servers = \"\"|rpc_servers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@$NODE1_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@$NODE3_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@$NODE4_API:26656\"|g" $HOME/.symphonyd/validator2/config/config.toml
# copy tendermint node id of 1,2,4 to 3
sed -i -E "s|rpc_servers = \"\"|rpc_servers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@$NODE1_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@$NODE2_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@$NODE4_API:26656\"|g" $HOME/.symphonyd/validator3/config/config.toml
# copy tendermint node id of 1,2,3 to 4
sed -i -E "s|rpc_servers = \"\"|rpc_servers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@$NODE1_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@$NODE2_API:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@$NODE3_API:26656\"|g" $HOME/.symphonyd/validator4/config/config.toml

echo "All 4 Validators configuration are created!"
