#!/bin/bash

KEY="test"
CHAINID="${CHAINID:-osmosis-1}"
# check if CHAINID is not defined
if [ -z "$CHAINID" ];
then
    CHAINID="osmosis-1"
fi
KEYRING="test"
MONIKER="localtestnet"
KEYALGO="secp256k1"
LOGLEVEL="info"

# retrieve all args
WILL_RECOVER=0
WILL_INSTALL=0
WILL_CONTINUE=0
INITIALIZE_ONLY=0
# $# is to check number of arguments
if [ $# -gt 0 ];
then
    # $@ is for getting list of arguments
    for arg in "$@"; do
        case $arg in
        --initialize)
            INITIALIZE_ONLY=1
            shift
            ;;
        --recover)
            WILL_RECOVER=1
            shift
            ;;
        --install)
            WILL_INSTALL=1
            shift
            ;;
        --continue)
            WILL_CONTINUE=1
            shift
            ;;
        *)
            printf >&2 "wrong argument somewhere"; exit 1;
            ;;
        esac
    done
fi

# continue running if everything is configured
if [ $WILL_CONTINUE -eq 1 ];
then
    # Start the node (remove the --pruning=nothing flag if historical queries are not needed)
    osmosisd start --pruning=nothing --log_level $LOGLEVEL --minimum-gas-prices=0.0001uosmo --p2p.laddr tcp://0.0.0.0:2280 --grpc.address 0.0.0.0:2282 --grpc-web.address 0.0.0.0:2283
    exit 1;
fi

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }
# command -v toml > /dev/null 2>&1 || { echo >&2 "toml not installed. More info: https://github.com/mrijken/toml-cli"; exit 1; }

# install babyd if not exist
if [ $WILL_INSTALL -eq 0 ];
then 
    command -v osmosisd > /dev/null 2>&1 || { echo >&1 "installing osmosisd"; make install; }
else
    echo >&1 "installing osmosisd"
    rm -rf $HOME/.osmosisd*
    make install
fi

osmosisd config keyring-backend $KEYRING
osmosisd config chain-id $CHAINID

# determine if user wants to recorver or create new
MNEMONIC=""
if [ $WILL_RECOVER -eq 0 ];
then
    MNEMONIC=$(osmosisd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --output json | jq -r '.mnemonic')
else
    MNEMONIC=$(osmosisd keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover --output json | jq -r '.mnemonic')
fi

echo "MNEMONIC=$MNEMONIC" >> client/.env
echo "MNEMONIC for $(osmosisd keys show $KEY -a --keyring-backend $KEYRING) = $MNEMONIC" >> scripts/mnemonic.txt

echo >&1 "\n"

# init chain
osmosisd init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to uosmo
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json
cat $HOME/.osmosisd/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="uosmo"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json

# Set gas limit in genesis
cat $HOME/.osmosisd/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="10000000"' > $HOME/.osmosisd/config/tmp_genesis.json && mv $HOME/.osmosisd/config/tmp_genesis.json $HOME/.osmosisd/config/genesis.json


# create more test key
MNEMONIC_1=$(osmosisd keys add test1 --keyring-backend $KEYRING --algo $KEYALGO --output json | jq -r '.mnemonic')
TO_ADDRESS=$(osmosisd keys show test1 -a --keyring-backend $KEYRING)
echo "MNEMONIC for $TO_ADDRESS = $MNEMONIC_1" >> scripts/mnemonic.txt
echo "TO_ADDRESS=$TO_ADDRESS" >> client/.env

# Allocate genesis accounts (cosmos formatted addresses)
osmosisd add-genesis-account $KEY 1000000000000uosmo --keyring-backend $KEYRING
osmosisd add-genesis-account test1 1000000000000uosmo --keyring-backend $KEYRING

# Sign genesis transaction
osmosisd gentx $KEY 1000000uosmo --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
osmosisd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
osmosisd validate-genesis

# if initialize only, exit
if [ $INITIALIZE_ONLY -eq 1 ];
then
    exit 0;
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
osmosisd start --pruning=nothing --log_level $LOGLEVEL --minimum-gas-prices=0.0001uosmo --p2p.laddr tcp://0.0.0.0:2280 --rpc.laddr tcp://0.0.0.0:2281 --grpc.address 0.0.0.0:2282 --grpc-web.address 0.0.0.0:2283