#!/bin/sh
set -eo pipefail

DEFAULT_CHAIN_ID="localjuno"
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

CHAIN_HOME=$HOME/.junod
CONFIG_FOLDER=$CHAIN_HOME/config

install_prerequisites () {
    apk add dasel
}

edit_genesis () {

    GENESIS=$CONFIG_FOLDER/genesis.json

    # Update staking module
    dasel put string -f $GENESIS '.app_state.staking.params.bond_denom' 'ujuno'
    dasel put string -f $GENESIS '.app_state.staking.params.unbonding_time' '10000s'
}

add_genesis_accounts () {
    
    # Validator
    echo "⚖️ Add validator account"
    echo $VALIDATOR_MNEMONIC | junod keys add $VALIDATOR_MONIKER --recover --keyring-backend=test --home $CHAIN_HOME
    VALIDATOR_ACCOUNT=$(junod keys show -a $VALIDATOR_MONIKER --keyring-backend test --home $CHAIN_HOME)
    junod add-genesis-account $VALIDATOR_ACCOUNT 100000000000ujuno --home $CHAIN_HOME
    
    # Faucet
    echo "🚰 Add faucet account"
    echo $FAUCET_MNEMONIC | junod keys add faucet --recover --keyring-backend=test --home $CHAIN_HOME
    FAUCET_ACCOUNT=$(junod keys show -a faucet --keyring-backend test --home $CHAIN_HOME)
    junod add-genesis-account $FAUCET_ACCOUNT 100000000000ujuno --home $CHAIN_HOME

    # Relayer
    echo "🔗 Add relayer account"
    echo $RELAYER_MNEMONIC | junod keys add relayer --recover --keyring-backend=test --home $CHAIN_HOME
    RELAYER_ACCOUNT=$(junod keys show -a relayer --keyring-backend test --home $CHAIN_HOME)
    junod add-genesis-account $RELAYER_ACCOUNT 1000000000ujuno --home $CHAIN_HOME
    
    junod gentx $VALIDATOR_MONIKER 500000000ujuno --keyring-backend=test --chain-id=$CHAIN_ID --home $CHAIN_HOME
    junod collect-gentxs --home $CHAIN_HOME
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
    echo $VALIDATOR_MNEMONIC | junod init -o --chain-id=$CHAIN_ID --home $CHAIN_HOME --recover $VALIDATOR_MONIKER
    edit_genesis
    add_genesis_accounts
    edit_config
fi

echo "🏁 Starting $CHAIN_ID..."
junod start --home $CHAIN_HOME
