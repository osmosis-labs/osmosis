#!/bin/sh
set -eo pipefail

DEFAULT_CHAIN_ID="localsymphony"
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

OSMOSIS_HOME=$HOME/.symphonyd
CONFIG_FOLDER=$OSMOSIS_HOME/config

install_prerequisites () {
    apk add dasel
}

edit_genesis () {

    echo "Editing genesis.json"

    GENESIS=$CONFIG_FOLDER/genesis.json

    echo "Update staking module"

    # Update staking module
    dasel put -t string -f $GENESIS '.app_state.staking.params.bond_denom' -v 'note'
    dasel put -t string -f $GENESIS '.app_state.staking.params.unbonding_time' -v '240s'

    echo "Update crisis module"

    # Update crisis module
    dasel put -t string -f $GENESIS '.app_state.crisis.constant_fee.denom' -v 'note'

    # Update gov module
    dasel put -t string -f $GENESIS '.app_state.gov.voting_params.voting_period' -v '60s'
    dasel put -t string -f $GENESIS '.app_state.gov.params.min_deposit.[0].denom' -v 'note'

    # Update epochs module
    dasel put -t string -f $GENESIS '.app_state.epochs.epochs.[1].duration' -v "60s"

    echo "Update poolincentives module"

    # Update poolincentives module
    dasel put -t string -f $GENESIS '.app_state.poolincentives.lockable_durations.[0]' -v "120s"
    dasel put -t string -f $GENESIS '.app_state.poolincentives.lockable_durations.[1]' -v "180s"
    dasel put -t string -f $GENESIS '.app_state.poolincentives.lockable_durations.[2]' -v "240s"
    dasel put -t string -f $GENESIS '.app_state.poolincentives.params.minted_denom' -v "note"

    # Update incentives module
    dasel put -t string -f $GENESIS '.app_state.incentives.lockable_durations.[0]' -v "1s"
    dasel put -t string -f $GENESIS '.app_state.incentives.lockable_durations.[1]' -v "120s"
    dasel put -t string -f $GENESIS '.app_state.incentives.lockable_durations.[2]' -v "180s"
    dasel put -t string -f $GENESIS '.app_state.incentives.lockable_durations.[3]' -v "240s"
    dasel put -t string -f $GENESIS '.app_state.incentives.params.distr_epoch_identifier' -v "day"

    echo "Update mint module"

    # Update mint module
    dasel put -t string -f $GENESIS '.app_state.mint.params.mint_denom' -v "note"
    dasel put -t string -f $GENESIS '.app_state.mint.params.epoch_identifier' -v "day"

    # Update gamm module
    dasel put -t string -f $GENESIS '.app_state.gamm.params.pool_creation_fee.[0].denom' -v "note"

    # Update txfee basedenom
    dasel put -t string -f $GENESIS '.app_state.txfees.basedenom' -v "note"

    # Update wasm permission (Nobody or Everybody)
    dasel put -t string -f $GENESIS '.app_state.wasm.params.code_upload_access.permission' -v "Everybody"
}

add_genesis_accounts () {
    
    # Validator
    echo "⚖️ Add validator account"
    echo $VALIDATOR_MNEMONIC | symphonyd keys add $VALIDATOR_MONIKER --recover --keyring-backend=test --home $OSMOSIS_HOME
    VALIDATOR_ACCOUNT=$(symphonyd keys show -a $VALIDATOR_MONIKER --keyring-backend test --home $OSMOSIS_HOME)
    symphonyd add-genesis-account $VALIDATOR_ACCOUNT 100000000000note,100000000000uion --home $OSMOSIS_HOME
    
    # Faucet
    echo "🚰 Add faucet account"
    echo $FAUCET_MNEMONIC | symphonyd keys add faucet --recover --keyring-backend=test --home $OSMOSIS_HOME
    FAUCET_ACCOUNT=$(symphonyd keys show -a faucet --keyring-backend test --home $OSMOSIS_HOME)
    symphonyd add-genesis-account $FAUCET_ACCOUNT 100000000000note,100000000000uion --home $OSMOSIS_HOME

    # Relayer
    echo "🔗 Add relayer account"
    echo $RELAYER_MNEMONIC | symphonyd keys add relayer --recover --keyring-backend=test --home $OSMOSIS_HOME
    RELAYER_ACCOUNT=$(symphonyd keys show -a relayer --keyring-backend test --home $OSMOSIS_HOME)
    symphonyd add-genesis-account $RELAYER_ACCOUNT 1000000000note,1000000000uion --home $OSMOSIS_HOME
    
    symphonyd gentx $VALIDATOR_MONIKER 500000000note --keyring-backend=test --chain-id=$CHAIN_ID --home $OSMOSIS_HOME
    symphonyd collect-gentxs --home $OSMOSIS_HOME
}

edit_config () {
    # Remove seeds
    dasel put -t string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' -v ''

    # Expose the rpc
    dasel put -t string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' -v "tcp://0.0.0.0:26657"

    # Expose the grpc
    dasel put -t string -f $CONFIG_FOLDER/app.toml -v "0.0.0.0:9090" '.grpc.address'
}

if [[ ! -d $CONFIG_FOLDER ]]
then
    install_prerequisites
    echo "🧪 Creating Symphony home for $VALIDATOR_MONIKER"
    echo $VALIDATOR_MNEMONIC | symphonyd init $VALIDATOR_MONIKER -o --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --recover
    edit_genesis
    add_genesis_accounts
    edit_config
fi

echo "🏁 Starting $CHAIN_ID..."
symphonyd start --home $OSMOSIS_HOME
