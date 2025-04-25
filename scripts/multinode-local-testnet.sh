#!/bin/bash
set -e

CHAIN_ID="testing"

# always returns true so set -e doesn't exit if it is not running.
killall symphonyd || true
rm -rf $HOME/.symphonyd/

# make 5 symphony directories
mkdir $HOME/.symphonyd
mkdir $HOME/.symphonyd/validator1
mkdir $HOME/.symphonyd/validator2
mkdir $HOME/.symphonyd/validator3
mkdir $HOME/.symphonyd/validator4

# init all 4 validators
symphonyd init --chain-id=$CHAIN_ID validator1 --home=$HOME/.symphonyd/validator1
symphonyd prepare-genesis mainnet $CHAIN_ID --home=$HOME/.symphonyd/validator1
symphonyd init --chain-id=$CHAIN_ID validator2 --home=$HOME/.symphonyd/validator2
symphonyd init --chain-id=$CHAIN_ID validator3 --home=$HOME/.symphonyd/validator3
symphonyd init --chain-id=$CHAIN_ID validator4 --home=$HOME/.symphonyd/validator4
# create keys for all 4 validators
symphonyd keys add validator1 --keyring-backend=test --home=$HOME/.symphonyd/validator1
symphonyd keys add validator2 --keyring-backend=test --home=$HOME/.symphonyd/validator2
symphonyd keys add validator3 --keyring-backend=test --home=$HOME/.symphonyd/validator3
symphonyd keys add validator4 --keyring-backend=test --home=$HOME/.symphonyd/validator4
# create key for first node
symphonyd keys add tax_receiver_addr --keyring-backend=test --home=$HOME/.symphonyd/validator1

update_genesis () {    
    cat $HOME/.symphonyd/validator1/config/genesis.json | jq "$1" > $HOME/.symphonyd/validator1/config/tmp_genesis.json && mv $HOME/.symphonyd/validator1/config/tmp_genesis.json $HOME/.symphonyd/validator1/config/genesis.json
}

# create validator node with tokens
symphonyd add-genesis-account $(symphonyd keys show validator1 -a --keyring-backend=test --home=$HOME/.symphonyd/validator1) 100000000000note,10000000usdr --home=$HOME/.symphonyd/validator1
symphonyd gentx validator1 5000000000note --moniker="validator1" --chain-id=$CHAIN_ID --keyring-backend=test --home=$HOME/.symphonyd/validator1 --chain-id=$CHAIN_ID
symphonyd collect-gentxs --home=$HOME/.symphonyd/validator1

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
update_genesis '.app_state["mint"]["params"]["epoch_identifier"]="minute"'

# update gamm genesis
#update_genesis '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="note"'

# update concentratedliquidity genesis
#update_genesis '.app_state["concentratedliquidity"]["params"]["is_permissionless_pool_creation_enabled"]=true'

# update txfees genesis
update_genesis '.app_state["txfees"]["params"]["swap_fees_epoch_identifier"]="minute"'

# update oracle genesis by adding test tokens to whitelist
#update_genesis '.app_state["oracle"]["params"]["whitelist"][0]["name"]="peppe"'
update_genesis '.app_state["oracle"]["tobin_taxes"][3]["denom"]="peppe"'
update_genesis '.app_state["oracle"]["tobin_taxes"][3]["tobin_tax"]="0.015"'

#update_genesis '.app_state["oracle"]["params"]["whitelist"][1]["name"]="usdr"'
update_genesis '.app_state["oracle"]["tobin_taxes"][4]["denom"]="usdr"'
update_genesis '.app_state["oracle"]["tobin_taxes"][4]["tobin_tax"]="0.01"'

# update oracle genesis by adding exchange rate for test tokens
update_genesis '.app_state["oracle"]["exchange_rates"][0]["denom"]="peppe"'
update_genesis '.app_state["oracle"]["exchange_rates"][0]["exchange_rate"]="1.7"'

update_genesis '.app_state["oracle"]["exchange_rates"][1]["denom"]="usdr"'
update_genesis '.app_state["oracle"]["exchange_rates"][1]["exchange_rate"]="2.0"'


update_genesis '.app_state["stablestakingincentives"]["params"]["distribution_contract_address"] = "symphony14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s748pj4"'

update_genesis '.app_state["market"]["params"]["tax_receiver"]="'$(symphonyd keys show tax_receiver_addr -a --keyring-backend=test --home=$HOME/.symphonyd/validator1)'"'

# copy validator1 genesis file to validator2-4
cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator2/config/genesis.json
cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator3/config/genesis.json
cp $HOME/.symphonyd/validator1/config/genesis.json $HOME/.symphonyd/validator4/config/genesis.json

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


# copy tendermint node id of validator1 to persistent peers of validator2-4
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656\"|g" $HOME/.symphonyd/validator2/config/config.toml
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656\"|g" $HOME/.symphonyd/validator3/config/config.toml
sed -i -E "s|persistent_peers = \"\"|persistent_peers = \"$(symphonyd tendermint show-node-id --home=$HOME/.symphonyd/validator1)@localhost:26656\"|g" $HOME/.symphonyd/validator4/config/config.toml

# start all three validators
tmux new -s validator1 -d symphonyd start --home=$HOME/.symphonyd/validator1
#tmux new -s validator2 -d symphonyd start --home=$HOME/.symphonyd/validator2
#tmux new -s validator3 -d symphonyd start --home=$HOME/.symphonyd/validator3
#tmux new -s validator4 -d symphonyd start --home=$HOME/.symphonyd/validator4

# send note from first validator to second validator
echo "Waiting 7 seconds to send funds to validators 2, 3, and 4..."
#sleep 7

# send note to other validators
# send note from 1-st validator to 2-nd validator
#symphonyd tx bank send validator1 $(symphonyd keys show validator2 -a --keyring-backend=test --home=$HOME/.symphonyd/validator2) 500000000note --keyring-backend=test --home=$HOME/.symphonyd/validator1 --chain-id=$CHAIN_ID --broadcast-mode sync --yes --fees 1000000note
#sleep 5

# send note from 1-st validator to 3-rd validator
#symphonyd tx bank send validator1 $(symphonyd keys show validator3 -a --keyring-backend=test --home=$HOME/.symphonyd/validator3) 400000000note --keyring-backend=test --home=$HOME/.symphonyd/validator1 --chain-id=$CHAIN_ID --broadcast-mode sync --yes --fees 1000000note
#sleep 5

# send note from 1-st validator to 4-rd validator
#symphonyd tx bank send validator1 $(symphonyd keys show validator4 -a --keyring-backend=test --home=$HOME/.symphonyd/validator4) 400000000note --keyring-backend=test --home=$HOME/.symphonyd/validator1 --chain-id=$CHAIN_ID --broadcast-mode sync --yes --fees 1000000note
#sleep 5

# add 2 - 4 validators
#symphonyd tx staking create-validator --amount=500000000note --from=validator2 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd/validator2) --moniker="validator2" --chain-id=$CHAIN_ID --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator1 --broadcast-mode sync  --yes --fees 1000000note
#sleep 5
#symphonyd tx staking create-validator --amount=500000000note --from=validator3 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd/validator3) --moniker="validator3" --chain-id=$CHAIN_ID --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator1 --broadcast-mode sync  --yes --fees 1000000note
#sleep 5
#symphonyd tx staking create-validator --amount=500000000note --from=validator4 --pubkey=$(symphonyd tendermint show-validator --home=$HOME/.symphonyd/validator4) --moniker="validator4" --chain-id=$CHAIN_ID --commission-rate="0.1" --commission-max-rate="0.2" --commission-max-change-rate="0.05" --min-self-delegation="500000000" --keyring-backend=test --home=$HOME/.symphonyd/validator1 --broadcast-mode sync  --yes --fees 1000000note
#sleep 5

echo "All 4 Validators are up and running!"