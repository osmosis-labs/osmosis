#!/bin/bash

rm -rf $HOME/.symphonyd/

cd $HOME

symphonyd init --chain-id=testing testing --home=$HOME/.symphonyd
symphonyd keys add validator --keyring-backend=test --home=$HOME/.symphonyd
symphonyd add-genesis-account $(symphonyd keys show validator -a --keyring-backend=test --home=$HOME/.symphonyd) 100000000000stake,100000000000valtoken --home=$HOME/.symphonyd
symphonyd gentx validator 500000000stake --keyring-backend=test --home=$HOME/.symphonyd --chain-id=testing
symphonyd collect-gentxs --home=$HOME/.symphonyd

update_genesis () {    
    cat $HOME/.symphonyd/validator1/config/genesis.json | jq "$1" > $HOME/.symphonyd/validator1/config/tmp_genesis.json && mv $HOME/.symphonyd/validator1/config/tmp_genesis.json $HOME/.symphonyd/validator1/config/genesis.json
}

# update staking genesis
update_genesis '.app_state["staking"]["params"]["unbonding_time"]="120s"'

# update governance genesis
update_genesis '.app_state["gov"]["voting_params"]["voting_period"]="10s"'

# update epochs genesis
update_genesis '.app_state["epochs"]["epochs"][0]["identifier"]="min"'
update_genesis '.app_state["epochs"]["epochs"][0]["duration"]="60s"'

# update poolincentives genesis
update_genesis '.app_state["poolincentives"]["lockable_durations"][0]="120s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][1]="180s"'
update_genesis '.app_state["poolincentives"]["lockable_durations"][2]="240s"'

# update incentives genesis
update_genesis '.app_state["incentives"]["params"]["distr_epoch_identifier"]="min"'
update_genesis '.app_state["incentives"]["lockable_durations"][0]="1s"'
update_genesis '.app_state["incentives"]["lockable_durations"][1]="120s"'
update_genesis '.app_state["incentives"]["lockable_durations"][2]="180s"'
update_genesis '.app_state["incentives"]["lockable_durations"][3]="240s"'

# update mint genesis
update_genesis '.app_state["mint"]["params"]["epoch_identifier"]="min"'

# update gamm genesis
update_genesis '.app_state["gamm"]["params"]["pool_creation_fee"][0]["denom"]="stake"'

# update superfluid genesis
update_genesis '.app_state["superfluid"]["params"]["minimum_risk_factor"]="0.500000000000000000"'

symphonyd start --home=$HOME/.symphonyd
