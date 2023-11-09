#!/bin/sh
set -e

MONIKER=localosmosis
OSMOSIS_HOME=/osmosis/.osmosisd
GENESIS_URL=https://osmosis-dev.fra1.digitaloceanspaces.com/localosmosis/genesis.json

# Initialize osmosis home
echo -e "\n🚀 Initializing osmosis home..."
osmosisd init $MONIKER --home $OSMOSIS_HOME > /dev/null 2>&1

# Customize config.toml, app.toml and client.toml
echo -e "\n🧾 Copying config.toml, app.toml, and client.toml from /etc/osmosis/..."
if [ -f /etc/osmosis/config/config.toml ]; then
    cp /etc/osmosis/config/config.toml $OSMOSIS_HOME/config/config.toml
fi
if [ -f /etc/osmosis/config/client.toml ]; then
    cp /etc/osmosis/config/client.toml $OSMOSIS_HOME/config/client.toml
fi
if [ -f /etc/osmosis/config/app.toml ]; then
    cp /etc/osmosis/config/app.toml $OSMOSIS_HOME/config/app.toml
fi

# Validator keys
echo -e "\n🔑 Restoring validator keys..."
cp /etc/osmosis/config/node_key.json $OSMOSIS_HOME/config/node_key.json
cp /etc/osmosis/config/priv_validator_key.json $OSMOSIS_HOME/config/priv_validator_key.json

# Add key to test-keyring
echo -e "\n🔑 Adding localosmosis key to test keyring-backend..."
cat /etc/osmosis/mnemonic | osmosisd keys add $MONIKER --recover --keyring-backend test > /dev/null 2>&1

echo -e "\n🔑 Your validator mnemonic is:\n$(cat /etc/osmosis/mnemonic)"
echo -e "\n📍 Your validator address is:\n$(cat /etc/osmosis/address)\n"

# Download genesis
echo "🌐 Downloading latest localosmosis genesis..."
wget -q $GENESIS_URL -O $OSMOSIS_HOME/config/genesis.json

echo -e "\n🧪 Starting localosmosis...\n"
echo -e "⏳ It will take some time to hit your first blocks...\n"

osmosisd start --home $OSMOSIS_HOME --x-crisis-skip-assert-invariants
