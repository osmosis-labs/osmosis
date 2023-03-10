#!/bin/sh

DEFAULT_CHAIN_ID="localosmosis"
DEFAULT_VALIDATOR_MONIKER="localosmosis-validator"
DEFAULT_VALIDATOR_MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"

# Override default values with environment variables

CHAIN_ID=${CHAIN_ID:-$DEFAULT_CHAIN_ID}
VALIDATOR_MNEMONIC=${MNEMONIC:-$DEFAULT_VALIDATOR_MNEMONIC}
VALIDATOR_MONIKER=${MONIKER:-$DEFAULT_VALIDATOR_MONIKER}

OSMOSIS_HOME=$HOME/.osmosisd
CONFIG_FOLDER=$OSMOSIS_HOME/config
POOL_DIR="/osmosis/pools"

POOLS_MNEMONIC="traffic cool olive pottery elegant innocent aisle dial genuine install shy uncle ride federal soon shift flight program cave famous provide cute pole struggle"
POOLS_KEY=localosmosis-pools

FAUCET_MNEMONIC="only item always south dry begin barely seed wire praise chapter bomb remind abandon erase safe point vehicle tuition release half denial receive water"
FAUCET_KEY=localosmosis-faucet

TX_FEES=10000uosmo

install_prerequisites () {
    apk add dasel
}

edit_genesis () {

    GENESIS=$CONFIG_FOLDER/genesis.json

    # Update staking module
    dasel put string -f $GENESIS '.app_state.staking.params.bond_denom' 'uosmo'
    dasel put string -f $GENESIS '.app_state.staking.params.unbonding_time' '240s'

    # Update crisis module
    dasel put string -f $GENESIS '.app_state.crisis.constant_fee.denom' 'uosmo'

    # Udpate gov module
    dasel put string -f $GENESIS '.app_state.gov.voting_params.voting_period' '60s'
    dasel put string -f $GENESIS '.app_state.gov.deposit_params.min_deposit.[0].denom' 'uosmo'

    # Update epochs module
    dasel put string -f $GENESIS '.app_state.epochs.epochs.[1].duration' "60s"

    # Update poolincentives module
    dasel put string -f $GENESIS '.app_state.poolincentives.lockable_durations.[0]' "120s"
    dasel put string -f $GENESIS '.app_state.poolincentives.lockable_durations.[1]' "180s"
    dasel put string -f $GENESIS '.app_state.poolincentives.lockable_durations.[2]' "240s"
    dasel put string -f $GENESIS '.app_state.poolincentives.params.minted_denom' "uosmo"

    # Update incentives module
    dasel put string -f $GENESIS '.app_state.incentives.lockable_durations.[0]' "1s"
    dasel put string -f $GENESIS '.app_state.incentives.lockable_durations.[1]' "120s"
    dasel put string -f $GENESIS '.app_state.incentives.lockable_durations.[2]' "180s"
    dasel put string -f $GENESIS '.app_state.incentives.lockable_durations.[3]' "240s"
    dasel put string -f $GENESIS '.app_state.incentives.params.distr_epoch_identifier' "day"

    # Update mint module
    dasel put string -f $GENESIS '.app_state.mint.params.mint_denom' "uosmo"
    dasel put string -f $GENESIS '.app_state.mint.params.epoch_identifier' "day"

    # Update poolmanager module
    dasel put string -f $GENESIS '.app_state.poolmanager.params.pool_creation_fee.[0].denom' "uosmo"

    # Update txfee basedenom
    dasel put string -f $GENESIS '.app_state.txfees.basedenom' "uosmo"

    # Update wasm permission (Nobody or Everybody)
    dasel put string -f $GENESIS '.app_state.wasm.params.code_upload_access.permission' "Everybody"
}

add_genesis_accounts () {

    # Add keys to keyring
    echo $VALIDATOR_MNEMONIC | osmosisd keys add $VALIDATOR_MONIKER --recover --keyring-backend=test --home $OSMOSIS_HOME
    echo $POOLS_MNEMONIC     | osmosisd keys add $POOLS_KEY --recover --keyring-backend=test --home $OSMOSIS_HOME
    echo $FAUCET_MNEMONIC    | osmosisd keys add $FAUCET_KEY --recover --keyring-backend=test --home $OSMOSIS_HOME

    # Compute account addresses
    VALIDATOR_ACCOUNT_ADDRESS=$(osmosisd keys show -a --keyring-backend test --bech acc $VALIDATOR_MONIKER --home $OSMOSIS_HOME)
    POOLS_ACCOUNT_ADDRESS=$(osmosisd keys show -a --keyring-backend test --bech acc $POOLS_KEY --home $OSMOSIS_HOME)
    FAUCET_ACCOUNT_ADDRESS=$(osmosisd keys show -a --keyring-backend test --bech acc $FAUCET_KEY --home $OSMOSIS_HOME)

    # Add validator account
    osmosisd add-genesis-account $VALIDATOR_ACCOUNT_ADDRESS 10000000000000000uosmo,10000000000000000uion,10000000000000000stake --home $OSMOSIS_HOME

    # Add pools account
    osmosisd add-genesis-account $POOLS_ACCOUNT_ADDRESS 10000000000000000uosmo,10000000000000000uion,10000000000000000stake --home $OSMOSIS_HOME

    # Add faucet account
    osmosisd add-genesis-account $FAUCET_ACCOUNT_ADDRESS 10000000000000000uosmo,10000000000000000uion,10000000000000000stake --home $OSMOSIS_HOME
    
    # Add gen-tx to create the genesis validator
    osmosisd gentx $VALIDATOR_MONIKER 5000000000000000uosmo --keyring-backend=test --chain-id=$CHAIN_ID --home $OSMOSIS_HOME
    osmosisd collect-gentxs --home $OSMOSIS_HOME
}

edit_config () {
    # Remove seeds
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' ''

    # Expose the rpc
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' "tcp://0.0.0.0:26657"
}

create_pool () {

    local pool_file="$1"

    substring='code: 0'
    COUNTER=0
    while [ $COUNTER -lt 15 ]; do
        output=$(osmosisd tx gamm create-pool \
            --pool-file=$pool_file \
            --from $POOLS_KEY \
            --chain-id=$CHAIN_ID \
            --home $OSMOSIS_HOME \
            --fees $TX_FEES \
            --keyring-backend=test \
            -b block --yes  2>&1)
        if [ "$output" != "${output%"$substring"*}" ]; then
            echo "✅ Created pool from $POOL_JSON"
            break
        else
            let COUNTER=COUNTER+1
            sleep 0.5
        fi
    done
}

if [[ ! -d $CONFIG_FOLDER ]]
then
    echo $VALIDATOR_MNEMONIC | osmosisd init -o --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --recover $VALIDATOR_MONIKER
    install_prerequisites
    edit_genesis
    add_genesis_accounts
    edit_config

    osmosisd start --home $OSMOSIS_HOME &

    # Create a pool for each file in the pools directory
    for pool_file in "${POOL_DIR}"/*.json; do
        echo "Creating pool from file: ${pool_file}"
        create_pool "${pool_file}"
    done

    wait
else
    osmosisd start --home $OSMOSIS_HOME
fi


