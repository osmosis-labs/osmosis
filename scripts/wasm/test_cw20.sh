#!/bin/bash

# this is cw20 code from install_contracts.sh, true if this is the first proposal
CODE=1
CHAIN_ID=wasm-2

osmosisd keys add demo --keyring-backend test
VAL=$(osmosisd keys show -a validator --keyring-backend test)
DEMO=$(osmosisd keys show -a demo --keyring-backend test)

# string interpolation in JSON blobs in bash just sucks... I usually use cosmjs for this, but that takes more setup to run.
INIT=$(cat <<EOF
{
  "name": "My first token",
  "symbol": "FRST",
  "decimals": 6,
  "initial_balances": [{
    "address": "$VAL",
    "amount": "123456789000"
  }]
}
EOF
)

osmosisd tx wasm instantiate $CODE "$INIT" --label "First Coin" --no-admin --from validator \
    --keyring-backend test --chain-id $CHAIN_ID -y -b block --gas 500000 --gas-prices 0.025stake

# Ideally we could parse the results of above, this will always be the address of the first contract with code id 1
CONTRACT=osmo14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sq2r9g9

QUERY=$(cat <<EOF
{ "balance": { "address": "$VAL" }}
EOF
)
QUERY_DEMO=$(cat <<EOF
{ "balance": { "address": "$DEMO" }}
EOF
)

# check initial balance
echo "Validator Balance:"
osmosisd query wasm contract-state smart $CONTRACT "$QUERY"
echo "Demo Balance:"
osmosisd query wasm contract-state smart $CONTRACT "$QUERY_DEMO"

# send some tokens
TRANSFER=$(cat <<EOF
{
  "transfer": {
    "recipient": "$DEMO",
    "amount": "987654321"
  }
}
EOF
)
osmosisd tx wasm execute $CONTRACT "$TRANSFER" --from validator --keyring-backend test \
    --chain-id $CHAIN_ID -y -b block --gas 500000 --gas-prices 0.025stake

# check final balance
echo "Validator Balance:"
osmosisd query wasm contract-state smart $CONTRACT "$QUERY"
echo "Demo Balance:"
osmosisd query wasm contract-state smart $CONTRACT "$QUERY_DEMO"
