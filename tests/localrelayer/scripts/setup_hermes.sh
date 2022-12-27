#!/bin/sh
set -e

# Default values for the two chains
DEFAULT_CHAIN_A_ID="localosmosis-a"
DEFAULT_CHAIN_A_MNEMONIC="black frequent sponsor nice claim rally hunt suit parent size stumble expire forest avocado mistake agree trend witness lounge shiver image smoke stool chicken"
DEFAULT_CHAIN_A_HOST="localosmosis-a"
DEFAULT_CHAIN_A_ADDRESS_PREFIX=osmo
DEFAULT_CHAIN_A_TOKEN=uosmo

DEFAULT_CHAIN_B_ID="localosmosis-b"
DEFAULT_CHAIN_B_MNEMONIC="black frequent sponsor nice claim rally hunt suit parent size stumble expire forest avocado mistake agree trend witness lounge shiver image smoke stool chicken"
DEFAULT_CHAIN_B_HOST="localosmosis-b"
DEFAULT_CHAIN_B_ADDRESS_PREFIX=osmo
DEFAULT_CHAIN_B_TOKEN=uosmo

# Override defaults via environment variables
CHAIN_A_MNEMONIC=${CHAIN_A_MNEMONIC:-$DEFAULT_CHAIN_A_MNEMONIC}
CHAIN_A_ID=${CHAIN_A_ID:-$DEFAULT_CHAIN_A_ID}
CHAIN_A_HOST=${CHAIN_A_HOST:-$DEFAULT_CHAIN_A_HOST}
CHAIN_A_ADDRESS_PREFIX=${CHAIN_A_ADDRESS_PREFIX:-DEFAULT_CHAIN_A_ADDRESS_PREFIX}
CHAIN_A_ADDRESS_TOKEN=${CHAIN_A_ADDRESS_TOKEN:-DEFAULT_CHAIN_A_ADDRESS_TOKEN}

CHAIN_B_MNEMONIC=${CHAIN_B_MNEMONIC:-$DEFAULT_CHAIN_B_MNEMONIC}
CHAIN_B_ID=${CHAIN_B_ID:-$DEFAULT_CHAIN_B_ID}
CHAIN_B_HOST=${CHAIN_B_HOST:-$DEFAULT_CHAIN_B_HOST}
CHAIN_B_ADDRESS_PREFIX=${CHAIN_B_ADDRESS_PREFIX:-DEFAULT_CHAIN_B_ADDRESS_PREFIX}
CHAIN_B_ADDRESS_TOKEN=${CHAIN_B_ADDRESS_TOKEN:-DEFAULT_CHAIN_B_ADDRESS_TOKEN}

install_prerequisites(){
    echo "🧰 Install prerequisites"
    apt update
    apt -y install curl gettext-base
}

create_config(){
    mkdir -p /root/.hermes/
    # Replace environment variables in config file
    envsubst < /etc/hermes/config.toml | tee /root/.hermes/config.toml > /dev/null
}

add_keys(){

    echo "🔑 Adding key for $CHAIN_A_ID"
    mkdir -p /home/hermes/mnemonics/
    echo $CHAIN_A_MNEMONIC > /home/hermes/mnemonics/$CHAIN_A_ID

    hermes keys add \
    --chain $CHAIN_A_ID \
    --mnemonic-file /home/hermes/mnemonics/$CHAIN_A_ID \
    --key-name $CHAIN_A_ID \
    --overwrite

    echo "🔑 Adding key for $CHAIN_B_ID"
    echo $CHAIN_B_MNEMONIC > /home/hermes/mnemonics/$CHAIN_B_ID

    hermes keys add \
    --chain $CHAIN_B_ID \
    --mnemonic-file /home/hermes/mnemonics/$CHAIN_B_ID \
    --key-name $CHAIN_B_ID \
    --overwrite
}

create_channel(){
    echo "🥱 Waiting for $CHAIN_A_ID to start"
    COUNTER=0
    until $(curl --output /dev/null --silent --head --fail http://$CHAIN_A_ID:26657/status); do
        printf '.'
        sleep 2
    done

    echo "🥱 Waiting for $CHAIN_B_ID to start"
    COUNTER=0
    until $(curl --output /dev/null --silent --head --fail http://$CHAIN_B_ID:26657/status); do
        printf '.'
        sleep 5
    done

    echo "📺 Creating channel $CHAIN_A_ID <> $CHAIN_B_ID"
    hermes create channel \
    --a-chain $CHAIN_A_ID \
    --b-chain $CHAIN_B_ID \
    --a-port transfer \
    --b-port transfer \
    --new-client-connection --yes
}

install_prerequisites
create_config
add_keys
create_channel

echo "✉️ Start Hermes"
hermes start
