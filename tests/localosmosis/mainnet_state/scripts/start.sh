#!/bin/sh
set -e 
set -o pipefail

CHAIN_ID=localosmosis
OSMOSIS_HOME=$HOME/.osmosisd
CONFIG_FOLDER=$OSMOSIS_HOME/config
MONIKER=val

MNEMONIC="bottom loan skill merry east cradle onion journey palm apology verb edit desert impose absurd oil bubble sweet glove shallow size build burst effort"

install_prerequisites () {
    apk add --no-cache dasel
    apk add --no-cache python3 py3-pip
}

edit_config () {
    # Remove seeds
    dasel put string -f $CONFIG_FOLDER/config.toml '.p2p.seeds' ''

    # Disable fast_sync
    dasel put bool -f $CONFIG_FOLDER/config.toml '.fast_sync' 'false'

    # Expose the rpc
    dasel put string -f $CONFIG_FOLDER/config.toml '.rpc.laddr' "tcp://0.0.0.0:26657"
}

install_prerequisites

echo $MNEMONIC | osmosisd init -o --chain-id=$CHAIN_ID --home $OSMOSIS_HOME --recover $MONIKER 2>& /dev/null

VALIDATOR_PUBKEY_JSON=$(osmosisd tendermint show-validator --home $OSMOSIS_HOME)
VALIDATOR_PUBKEY=$(echo $VALIDATOR_PUBKEY_JSON | dasel -r json '.key' --plain)
VALIDATOR_HEX_ADDRESS=$(osmosisd debug pubkey $VALIDATOR_PUBKEY_JSON 2>&1 --home $OSMOSIS_HOME | grep Address | cut -d " " -f 2)
VALIDATOR_ACCOUNT_ADDRESS=$(osmosisd debug addr $VALIDATOR_HEX_ADDRESS 2>&1  --home $OSMOSIS_HOME | grep Acc | cut -d " " -f 3)
VALIDATOR_OPERATOR_ADDRESS=$(osmosisd debug addr $VALIDATOR_HEX_ADDRESS 2>&1  --home $OSMOSIS_HOME | grep Val | cut -d " " -f 3)
VALIDATOR_CONSENSUS_ADDRESS=$(osmosisd debug bech32-convert $VALIDATOR_OPERATOR_ADDRESS -p osmovalcons  --home $OSMOSIS_HOME 2>&1)

python3 -u testnetify.py \
   -i /osmosis/genesis.json \
   -o $CONFIG_FOLDER/genesis.json \
   -c $CHAIN_ID \
   --validator-hex-address $VALIDATOR_HEX_ADDRESS \
   --validator-operator-address $VALIDATOR_OPERATOR_ADDRESS \
   --validator-consensus-address $VALIDATOR_CONSENSUS_ADDRESS \
   --validator-pubkey $VALIDATOR_PUBKEY \
   -v --pretty-output \
#    --account-pubkey $ACCOUNT_PUBKEY \
#    --account-address $ACCOUNT_ADDRESS \

edit_config
osmosisd start --home $OSMOSIS_HOME --x-crisis-skip-assert-invariants
