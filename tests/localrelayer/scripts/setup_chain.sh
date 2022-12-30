#!/bin/sh
set -eo pipefail

DEFAULT_CHAIN_ID="localosmosis"
DEFAULT_VALIDATOR_MONIKER="validator"
DEFAULT_VALIDATOR_MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"
DEFAULT_FAUCET_MNEMONIC="increase bread alpha rigid glide amused approve oblige print asset idea enact lawn proof unfold jeans rabbit audit return chuckle valve rather cactus great"
DEFAULT_RELAYER_MNEMONIC="black frequent sponsor nice claim rally hunt suit parent size stumble expire forest avocado mistake agree trend witness lounge shiver image smoke stool chicken"

# Override default values with environment variables
CHAIN_ID=${CHAIN_ID:-$DEFAULT_CHAIN_ID}
VALIDATOR_MONIKER=${VALIDATOR_MONIKER:-$DEFAULT_VALIDATOR_MONIKER}
VALIDATOR_MNEMONIC=${VALIDATOR_MNEMONIC:-$DEFAULT_VALIDATOR_MNEMONIC}
FAUCET_MNEMONIC=${FAUCET_MNEMONIC:-$DEFAULT_FAUCET_MNEMONIC}
RELAYER_MNEMONIC=${RELAYER_MNEMONIC:-$DEFAULT_RELAYER_MNEMONIC}

OSMOSIS_HOME=$HOME/.osmosisd
CONFIG_FOLDER=$OSMOSIS_HOME/config

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

    # Update gamm module
    dasel put string -f $GENESIS '.app_state.gamm.params.pool_creation_fee.[0].denom' "uosmo"

    # Update txfee basedenom
    dasel put string -f $GENESIS '.app_state.txfees.basedenom' "uosmo"

    # Update wasm permission (Nobody or Everybody)
    dasel put string -f $GENESIS '.app_state.wasm.params.code_upload_access.permission' "Everybody"
}

add_genesis_accounts () {
    
    # Validator
    echo "⚖️ Add validator account"
    echo $VALIDATOR_MNEMONIC | osmosisd keys add $VALIDATOR_MONIKER --recover --keyring-backend=test --home $OSMOSIS_HOME
    VALIDATOR_ACCOUNT=$(osmosisd keys show -a $VALIDATOR_MONIKER --keyring-backend test --home $OSMOSIS_HOME)
    osmosisd add-genesis-account $VALIDATOR_ACCOUNT 100000000000uosmo,100000000000uion,100000000000stake --home $OSMOSIS_HOME
    
    # Faucet
    echo "🚰 Add faucet account"
    echo $FAUCET_MNEMONIC | osmosisd keys add faucet --recover --keyring-backend=test --home $OSMOSIS_HOME
    FAUCET_ACCOUNT=$(osmosisd keys show -a faucet --keyring-backend test --home $OSMOSIS_HOME)
    osmosisd add-genesis-account $FAUCET_ACCOUNT 100000000000uosmo,100000000000uion,100000000000stake --home $OSMOSIS_HOME

    # Relayer
    echo "🔗 Add relayer account"
    echo $RELAYER_MNEMONIC | osmosisd keys add relayer --recover --keyring-backend=test --home $OSMOSIS_HOME
    RELAYER_ACCOUNT=$(osmosisd keys show -a relayer --keyring-backend test --home $OSMOSIS_HOME)
    osmosisd add-genesis-account $RELAYER_ACCOUNT 1000000000uosmo,1000000000uion,1000000000stake --home $OSMOSIS_HOME
    
    osmosisd gentx $VALIDATOR_MONIKER 500000000uosmo --keyring-backend=test --chain-id=$CHAIN_ID --home $OSMOSIS_HOME
    osmosisd collect-gentxs --home $OSMOSIS_HOME
}

edit_config () {
    # Remove seeds
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' ''

    # Expose the rpc
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' "tcp://0.0.0.0:26657"
}

if [[ ! -d $CONFIG_FOLDER ]]
then
    install_prerequisites
    echo "🧪 Creating Osmosis home for $VALIDATOR_MONIKER"
    echo $VALIDATOR_MNEMONIC | osmosisd init -o --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --recover $VALIDATOR_MONIKER
    edit_genesis
    add_genesis_accounts
    edit_config
fi

echo "🏁 Starting $CHAIN_ID..."
osmosisd start --home $OSMOSIS_HOME
