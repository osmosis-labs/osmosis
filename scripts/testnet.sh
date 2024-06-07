#!/bin/bash
set -e

CHAIN_ID="symphony-testnet-1"

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

# # copy validator1 genesis file to validator2-4
# cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator2/config/genesis.json
# cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator3/config/genesis.json
# cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator4/config/genesis.json

# symphonyd add-genesis-account $(symphonyd keys show validator3 -a --keyring-backend=test --home=$HOME/.symphonyd/validator3) 100000000000note --home=$HOME/.symphonyd/validator3
# symphonyd gentx validator3 500000000note --moniker="validator3" --chain-id=$CHAIN_ID --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator3 --chain-id=$CHAIN_ID
# symphonyd collect-gentxs --home=$HOME/.symphonyd/validator3

# symphonyd add-genesis-account $(symphonyd keys show validator4 -a --keyring-backend=test --home=$HOME/.symphonyd/validator4) 100000000000note --home=$HOME/.symphonyd/validator4
# symphonyd gentx validator4 500000000note --moniker="validator4" --chain-id=$CHAIN_ID --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator4 --chain-id=$CHAIN_ID
# symphonyd collect-gentxs --home=$HOME/.symphonyd/validator4

# port key (validator1 uses default ports)
# validator1 1317, 9050, 9091, 26658, 26657, 26656, 6060, 26660
# validator2 1316, 9088, 9089, 26655, 26654, 26653, 6061, 26630
# validator3 1315, 9086, 9087, 26652, 26651, 26650, 6062, 26620
# validator4 1314, 9084, 9085, 26649, 26648, 26647, 6063, 26610

# change app.toml values
VALIDATOR1_APP_TOML=$HOME/.symphonyd/validator1/config/app.toml
VALIDATOR2_APP_TOML=$HOME/.symphonyd/validator2/config/app.toml
VALIDATOR3_APP_TOML=$HOME/.symphonyd/validator3/config/app.toml
VALIDATOR4_APP_TOML=$HOME/.symphonyd/validator4/config/app.toml

# validator1
sed -i -E 's|0.0.0.0:9090|0.0.0.0:9050|g' $VALIDATOR1_APP_TOML

# validator2
sed -i -E 's|localhost:1317|localhost:1316|g' $VALIDATOR2_APP_TOML
sed -i -E 's|localhost:9090|localhost:9088|g' $VALIDATOR2_APP_TOML
sed -i -E 's|localhost:9091|localhost:9089|g' $VALIDATOR2_APP_TOML

# validator3
sed -i -E 's|localhost:1317|localhost:1315|g' $VALIDATOR3_APP_TOML
sed -i -E 's|localhost:9090|localhost:9086|g' $VALIDATOR3_APP_TOML
sed -i -E 's|localhost:9091|localhost:9087|g' $VALIDATOR3_APP_TOML
sed -i -E 's|adaptive-fee-enabled = "false"|adaptive-fee-enabled = "true"|g' $VALIDATOR3_APP_TOML

# validator4
sed -i -E 's|localhost:1317|localhost:1314|g' $VALIDATOR4_APP_TOML
sed -i -E 's|localhost:9090|localhost:9084|g' $VALIDATOR4_APP_TOML
sed -i -E 's|localhost:9091|localhost:9085|g' $VALIDATOR4_APP_TOML
sed -i -E 's|adaptive-fee-enabled = "false"|adaptive-fee-enabled = "true"|g' $VALIDATOR4_APP_TOML


# change config.toml values
VALIDATOR1_CONFIG=$HOME/.symphonyd/validator1/config/config.toml
VALIDATOR2_CONFIG=$HOME/.symphonyd/validator2/config/config.toml
VALIDATOR3_CONFIG=$HOME/.symphonyd/validator3/config/config.toml
VALIDATOR4_CONFIG=$HOME/.symphonyd/validator4/config/config.toml

# validator1
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR1_CONFIG
# sed -i -E 's|version = "v0"|version = "v1"|g' $VALIDATOR1_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR1_CONFIG

# validator2
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26655|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26654|g' $VALIDATOR2_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26653|g' $VALIDATOR2_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR2_CONFIG
sed -i -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26630"|g' $VALIDATOR2_CONFIG

# validator3
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26652|g' $VALIDATOR3_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26651|g' $VALIDATOR3_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26650|g' $VALIDATOR3_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR3_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR3_CONFIG
sed -i -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26620"|g' $VALIDATOR3_CONFIG

# validator4
sed -i -E 's|tcp://127.0.0.1:26658|tcp://127.0.0.1:26649|g' $VALIDATOR4_CONFIG
sed -i -E 's|tcp://127.0.0.1:26657|tcp://127.0.0.1:26648|g' $VALIDATOR4_CONFIG
sed -i -E 's|tcp://0.0.0.0:26656|tcp://0.0.0.0:26647|g' $VALIDATOR4_CONFIG
sed -i -E 's|allow_duplicate_ip = false|allow_duplicate_ip = true|g' $VALIDATOR4_CONFIG
sed -i -E 's|prometheus = false|prometheus = true|g' $VALIDATOR4_CONFIG
sed -i -E 's|prometheus_listen_addr = ":26660"|prometheus_listen_addr = ":26610"|g' $VALIDATOR4_CONFIG


# persistent_peers

# copy tendermint node id of validator 2,3,4 to 1
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@localhost:26654,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@localhost:26651,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@localhost:26648\"|g" $HOME/.symphonyd/validator1/config/config.toml
# copy tendermint node id of 1,3,4 to 2
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@localhost:26651,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@localhost:26648\"|g" $HOME/.symphonyd/validator2/config/config.toml
# copy tendermint node id of 1,2,4 to 3
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@localhost:26654,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@localhost:26648\"|g" $HOME/.symphonyd/validator3/config/config.toml
# copy tendermint node id of 1,2,3 to 4
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@localhost:26654,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@localhost:26651\"|g" $HOME/.symphonyd/validator4/config/config.toml

# rpc_servers

# copy tendermint node id  2,3,4 to 1
sed -i -E "s|rpc_servers = \"\"|rpc_servers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@localhost:26654,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@localhost:26651,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@localhost:26648\"|g" $HOME/.symphonyd/validator1/config/config.toml
# copy tendermint node id of 1,3,4 to 2
sed -i -E "s|rpc_servers = \"\"|rpc_servers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@localhost:26651,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@localhost:26648\"|g" $HOME/.symphonyd/validator2/config/config.toml
# copy tendermint node id of 1,2,4 to 3
sed -i -E "s|rpc_servers = \"\"|rpc_servers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@localhost:26654,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator4)@localhost:26648\"|g" $HOME/.symphonyd/validator3/config/config.toml
# copy tendermint node id of 1,2,3 to 4
sed -i -E "s|rpc_servers = \"\"|rpc_servers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator2)@localhost:26654,$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator3)@localhost:26651\"|g" $HOME/.symphonyd/validator4/config/config.toml

echo "All 4 Validators are up!"

# copy validator1 genesis file to validator2-4

# start separately all 4 validators on different nodes
# symphonyd start --home=$HOME/.symphonyd

# send note to other validators
# send note from 1-st validator to 2-nd validator
# symphonyd tx bank send validator1 $(symphonyd keys show validator2 -a --keyring-backend=test --home=$HOME/.symphonyd/validator2) 500000000note --keyring-backend=test --home=$HOME/.symphonyd/validator1 --chain-id=$CHAIN_ID --broadcast-mode sync --yes --fees 1000000note

# send note from 1-st validator to 3-rd validator
# symphonyd tx bank send validator1 $(symphonyd keys show validator3 -a --keyring-backend=test --home=$HOME/.symphonyd/validator3) 400000000note --keyring-backend=test --home=$HOME/.symphonyd/validator1 --chain-id=$CHAIN_ID --broadcast-mode sync --yes --fees 1000000note

# send note from 1-st validator to 4-rd validator
# symphonyd tx bank send validator1 $(symphonyd keys show validator4 -a --keyring-backend=test --home=$HOME/.symphonyd/validator4) 400000000note --keyring-backend=test --home=$HOME/.symphonyd/validator1 --chain-id=$CHAIN_ID --broadcast-mode sync --yes --fees 1000000note

# add 2- 4 validators
# symphonyd tx staking create-validator --amount=500000000note --from=validator1 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd/validator2) --moniker="validator2" --chain-id=$CHAIN_ID --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator1 --broadcast-mode sync  --yes --fees 1000000note
# symphonyd tx staking create-validator --amount=500000000note --from=validator1 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd/validator3) --moniker="validator3" --chain-id=$CHAIN_ID --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator1 --broadcast-mode sync  --yes --fees 1000000note
# symphonyd tx staking create-validator --amount=500000000note --from=validator1 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd/validator4) --moniker="validator4" --chain-id=$CHAIN_ID --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator1 --broadcast-mode sync  --yes --fees 1000000note