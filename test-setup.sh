#!/bin/bash

# Variables
CHAIN_ID="localosmosis"
MONIKER="localosmosis"
KEYRING="test"
KEY="mykey"
TOKEN1="1000000000uosmo"
TOKEN2="1000000000uion"
POOL_FILE="pool.json"

# Initialize the chain
build/osmosisd init $MONIKER --chain-id $CHAIN_ID

# Add a key to the keyring
build/osmosisd keys add $KEY --keyring-backend $KEYRING

# Add genesis accounts
build/osmosisd add-genesis-account $(build/osmosisd keys show $KEY -a --keyring-backend $KEYRING) $TOKEN1,$TOKEN2

GENESIS_FILE=~/.osmosisd/config/genesis.json
# Replace all occurrences of "stake" with "uosmo" in the genesis file
sed -i '' 's/"stake"/"uosmo"/g' $GENESIS_FILE
sed -i '' 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/g' ~/.osmosisd/config/config.toml

# Reduce pool creation fee to 1000uosmo
jq '.app_state.poolmanager.params.pool_creation_fee[0].amount = "1000"' $GENESIS_FILE > tmp.json && mv tmp.json $GENESIS_FILE
jq '.app_state.poolmanager.params.taker_fee_params.default_taker_fee = "0.1"' $GENESIS_FILE > tmp.json && mv tmp.json $GENESIS_FILE

# Generate a genesis transaction
build/osmosisd gentx $KEY 1000000uosmo --chain-id $CHAIN_ID --keyring-backend $KEYRING

# Collect genesis transactions
build/osmosisd collect-gentxs

# Validate the genesis file
# build/osmosisd validate-genesis

# Create a pool file
cat > $POOL_FILE <<EOL
{
    "weights": "1uosmo,1uion",
    "initial-deposit": "500000uosmo,500000uion",
    "swap-fee": "0.01",
    "exit-fee": "0.01",
    "future-governor": "24h"
}
EOL

# Start the chain
build/osmosisd start &

# Wait for the chain to start
sleep 10

# Create the pool
build/osmosisd tx gamm create-pool --pool-file $POOL_FILE --from $KEY --keyring-backend $KEYRING --chain-id $CHAIN_ID --fees 5000uosmo --yes --gas 500000

# Add two more keys to the keyring
# build/osmosisd keys add account1 --keyring-backend $KEYRING
# build/osmosisd keys add account2 --keyring-backend $KEYRING
# osmo13jwg8f7hk9d6ys6f5wg5kxvud8r7l526rd8dcq
build/osmosisd keys add account1 --keyring-backend $KEYRING --recover <<< "connect balance decrease flip trade indicate search hotel donor wet venue jacket insect chicken garlic vacuum use screen future pull option bid uncover demand"
# osmo1xkm3xa030cwk79w50nxdlvcu68exmguehucttf
build/osmosisd keys add account2 --keyring-backend $KEYRING --recover <<< "sponsor thing any tool globe fly barrel silent symbol uncover pool scrub hand hard address master club bring close crater boring nephew angry glow"

# Fund the new accounts
build/osmosisd tx bank send $KEY $(build/osmosisd keys show account1 -a --keyring-backend $KEYRING) 5000000uosmo --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees 5000uosmo --yes
sleep 10
build/osmosisd tx bank send $KEY $(build/osmosisd keys show account2 -a --keyring-backend $KEYRING) 5000000uosmo --chain-id $CHAIN_ID --keyring-backend $KEYRING --fees 5000uosmo --yes

# Affiliate account2 with account1
build/osmosisd tx poolmanager revenue-share $(build/osmosisd keys show account1 -a --keyring-backend $KEYRING) --from account2 --keyring-backend test --chain-id $CHAIN_ID --fees 20000uosmo --yes

# Check if the account is affiliated
curl -s http://localhost:1317/osmosis/poolmanager/v2/affiliated/$(build/osmosisd keys show account2 -a --keyring-backend $KEYRING) | jq

# Check balance of account1 before
build/osmosisd query bank balances $(build/osmosisd keys show account1 -a --keyring-backend $KEYRING)

# Perform a swap using account2
build/osmosisd tx gamm swap-exact-amount-in \
    --from account2 \
    1000000uosmo 1 \
    --chain-id $CHAIN_ID \
    --keyring-backend $KEYRING \
    --fees 20000uosmo \
    --yes \
    --swap-route-denoms uion \
    --swap-route-pool-ids 1

# Check balance of account1 after
build/osmosisd query bank balances $(build/osmosisd keys show account1 -a --keyring-backend $KEYRING)