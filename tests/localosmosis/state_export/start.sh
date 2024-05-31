#!/bin/sh
set -e

MONIKER=localsymphony
OSMOSIS_HOME=/symphony/.symphonyd
GENESIS_URL=https://symphony-dev.fra1.digitaloceanspaces.com/localsymphony/genesis.json

# Initialize symphony home
echo -e "\n🚀 Initializing symphony home..."
symphonyd init $MONIKER --home $OSMOSIS_HOME > /dev/null 2>&1

# Customize config.toml, app.toml and client.toml
echo -e "\n🧾 Copying config.toml, app.toml, and client.toml from /etc/symphony/..."
if [ -f /etc/symphony/config/config.toml ]; then
    cp /etc/symphony/config/config.toml $OSMOSIS_HOME/config/config.toml
fi
if [ -f /etc/symphony/config/client.toml ]; then
    cp /etc/symphony/config/client.toml $OSMOSIS_HOME/config/client.toml
fi
if [ -f /etc/symphony/config/app.toml ]; then
    cp /etc/symphony/config/app.toml $OSMOSIS_HOME/config/app.toml
fi

# Validator keys
echo -e "\n🔑 Restoring validator keys..."
cp /etc/symphony/config/node_key.json $OSMOSIS_HOME/config/node_key.json
cp /etc/symphony/config/priv_validator_key.json $OSMOSIS_HOME/config/priv_validator_key.json

# Add key to test-keyring
echo -e "\n🔑 Adding localsymphony key to test keyring-backend..."
cat /etc/symphony/mnemonic | symphonyd keys add $MONIKER --recover --keyring-backend test > /dev/null 2>&1

echo -e "\n🔑 Your validator mnemonic is:\n$(cat /etc/symphony/mnemonic)"
echo -e "\n📍 Your validator address is:\n$(cat /etc/symphony/address)\n"

# Download genesis
echo "🌐 Downloading latest localsymphony genesis..."
wget -q $GENESIS_URL -O $OSMOSIS_HOME/config/genesis.json

echo -e "\n🧪 Starting localsymphony...\n"
echo -e "⏳ It will take some time to hit your first blocks...\n"

symphonyd start --home $OSMOSIS_HOME --x-crisis-skip-assert-invariants
